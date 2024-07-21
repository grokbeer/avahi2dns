package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/holoplot/go-avahi"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type resolveHostNameResult struct {
	HostName avahi.HostName
	Error    error
}

func createDNSReply(logger *logrus.Entry, cfg *config, aserver *avahi.Server, r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range r.Question {
			switch q.Qtype {
			case dns.TypeA:
				if cfg.IPv6Only { // Skip TypeA if only IPv6 is enabled
					logger.Info("IPv4 disabled, skipping A query...")
					continue
				}
				rr, err := avahiToRecord(logger, cfg, aserver, q.Name, avahi.ProtoInet, "A")
				if err != nil {
					logger.WithError(err).Error("avahi A lookup failed, skipping query...")
					continue
				}
				m.Answer = append(m.Answer, rr)

			case dns.TypeAAAA:
				if cfg.IPv4Only { // Skip TypeAAAA if only IPv4 is enabled
					logger.Info("IPv6 disabled, skipping AAAA query...")
					continue
				}
				rr, err := avahiToRecord(logger, cfg, aserver, q.Name, avahi.ProtoInet6, "AAAA")
				if err != nil {
					logger.WithError(err).Error("avahi AAAA lookup failed, skipping query...")
					continue
				}
				m.Answer = append(m.Answer, rr)

			default:
				logger.WithField("type", q.Qtype).Warning("unsupported question")
			}
		}

	default:
		logger.WithField("opcode", r.Opcode).Warning("unsupported opcode")
	}

	return m
}

// TimedResolveHostName wraps the avahi.ResolveHostName method with a timeout.
func TimedResolveHostName(timeoutSecs uint8, aserver *avahi.Server, iface int32, protocol int32, name string, aprotocol int32, flags uint32) (avahi.HostName, error) {
	resultChan := make(chan resolveHostNameResult, 1)
	go func() {
		hn, err := aserver.ResolveHostName(iface, protocol, name, aprotocol, flags)
		resultChan <- resolveHostNameResult{HostName: hn, Error: err}
	}()
	select {
	case <-time.After(time.Duration(timeoutSecs) * time.Second):
		return avahi.HostName{}, errors.New("timed out")
	case result := <-resultChan:
		return result.HostName, result.Error
	}
}

func avahiToRecord(logger *logrus.Entry, cfg *config, aserver *avahi.Server, name string, proto int32, recordType string) (dns.RR, error) {
	logger.WithFields(logrus.Fields{
		"name":     name,
		"type":     recordType,
		"protocol": proto,
	}).Info("forwarding query to avahi")
	var hn avahi.HostName
	var err error
	if cfg.TimeoutSecs > 0 {
		hn, err = TimedResolveHostName(cfg.TimeoutSecs, aserver, avahi.InterfaceUnspec, proto, name, proto, 0)
	} else {
		hn, err = aserver.ResolveHostName(avahi.InterfaceUnspec, proto, name, proto, 0)
	}
	if err != nil {
		return nil, fmt.Errorf("avahi resolve failure: %w", err)
	}
	rr, err := dns.NewRR(fmt.Sprintf("%s %s %s", name, recordType, hn.Address))
	if err != nil {
		return nil, fmt.Errorf("failured to create record: %w", err)
	}
	return rr, err
}
