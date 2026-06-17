// Package dnsmap converts DNS messages into matcher maps.
//
// It lives in its own package (rather than in mapsutil) so that the heavy
// github.com/miekg/dns dependency is only linked into binaries that actually
// convert DNS messages, instead of every consumer of the generic map helpers.
package dnsmap

import (
	"fmt"

	"github.com/miekg/dns"
)

const defaultFormat = "%s"

// DNSToMap converts a DNS message to a matcher map.
func DNSToMap(msg *dns.Msg, format string) (m map[string]interface{}) {
	m = make(map[string]interface{})

	if format == "" {
		format = defaultFormat
	}

	m[fmt.Sprintf(format, "rcode")] = msg.Rcode

	var qs string
	for _, question := range msg.Question {
		qs += fmt.Sprintln(question.String())
	}
	m[fmt.Sprintf(format, "question")] = qs

	var exs string
	for _, extra := range msg.Extra {
		exs += fmt.Sprintln(extra.String())
	}
	m[fmt.Sprintf(format, "extra")] = exs

	var ans string
	for _, answer := range msg.Answer {
		ans += fmt.Sprintln(answer.String())
	}
	m[fmt.Sprintf(format, "answer")] = ans

	var nss string
	for _, ns := range msg.Ns {
		nss += fmt.Sprintln(ns.String())
	}
	m[fmt.Sprintf(format, "ns")] = nss

	m[fmt.Sprintf(format, "raw")] = msg.String()

	return m
}
