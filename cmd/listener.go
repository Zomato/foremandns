package cmd

import (
	"context"
	"fmt"
	"github.com/karlseguin/ccache"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

var localCache = ccache.New(ccache.Configure().MaxSize(1000).ItemsToPrune(100))

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

		var address string
		if cacheType == "redis" {
			addressVal, err := redisClient.Get(domain).Result()
			if err != nil {
				log.Error(fmt.Printf("Redis Error %v \n", err))
			}
			address = addressVal
		} else {
			addressVal := localCache.Get(domain)
			if addressVal != nil {
				address = addressVal.Value().(string)
			}
		}
		if address != "" {
			log.Info("Cached domain ", domain, " and the IP is ", address)
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
				if cacheType == "redis" {
					redisClient.Set(domain, *host.IP, time.Duration(ttl)*time.Second)
				} else {
					localCache.Set(domain, *host.IP, time.Duration(ttl)*time.Second)
				}
			}
		}
	}
	w.WriteMsg(&msg)
}
