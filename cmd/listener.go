package cmd

import (
	"context"
	"fmt"
	"foremandns/util"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

var domainsToAddresses *util.TTLMap

func init() {
	domainsToAddresses = util.New(ttl)
}

type handler struct{}

func (dnsHandler *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domainOriginal := msg.Question[0].Name

		domain := strings.TrimSuffix(domainOriginal, zone)
		domain = strings.TrimSuffix(domain, ".")

		log.Debug("Domain is ", domain)

		address := domainsToAddresses.Get(domain)
		if address != "" {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domainOriginal, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		} else {
			host, _, err := client.Hosts.Get(context.Background(), domain)
			if err != nil {
				log.Error("Hosts.Get returned error:", err)
			} else if host.IP != nil && *host.IP != "" {
				log.Debug(fmt.Sprintf("Host value is %v \n", host))
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: domainOriginal, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP(*host.IP),
				})
				domainsToAddresses.Put(domain, *host.IP)
			}
		}
	}
	w.WriteMsg(&msg)
}
