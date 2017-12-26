package wildcard

import (
	"bytes"
	"net"
	"testing"

	"github.com/hazaelsan/ipcalc"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		addr    string
		matches map[string]bool
	}{
		{
			addr: "192.0.2.0/0.0.0.254",
			matches: map[string]bool{
				"192.0.2.0": true,
				"192.0.2.1": false,
				"192.0.2.2": true,
				"192.0.0.0": false,
				"192.0.4.0": false,
			},
		},
		{
			addr: "192.0.2.1/0.0.0.254",
			matches: map[string]bool{
				"192.0.2.1": true,
				"192.0.2.2": false,
				"192.0.2.3": true,
				"192.0.0.1": false,
				"192.0.4.1": false,
			},
		},
		{
			addr: "2001:db8::/::fffe",
			matches: map[string]bool{
				"2001:db8::2":   true,
				"2001:db8::4":   true,
				"2001:db8:a::2": false,
				"2001:db8::3":   false,
			},
		},
	}
	for _, tt := range tests {
		ip, mask, err := ipcalc.ParseIPMask(tt.addr)
		if err != nil {
			t.Errorf("ParseIPMask(%v) error = %v", tt.addr, err)
			continue
		}
		w := New(ip, mask)
		for addr, want := range tt.matches {
			if got := w.Matches(net.ParseIP(addr)); got != want {
				t.Errorf("Matches(%v) = %v, want %v", addr, got, want)
			}
		}
	}
}

func TestFirst(t *testing.T) {
	tests := map[string]string{
		"192.0.2.129/0.0.0.254":       "192.0.2.1",
		"192.0.2.255/0.0.0.0":         "192.0.2.255",
		"192.0.2.255/255.255.255.255": "0.0.0.0",
		"2001:db8::ff/::fffe":         "2001:db8::1",
		"2001:db8::ffff/::":           "2001:db8::ffff",
		"2001:db8::ffff/128":          "::",
	}
	for addr, want := range tests {
		ip, mask, err := ipcalc.ParseIPMask(addr)
		if err != nil {
			t.Errorf("ParseIPMask(%v) error = %v", addr, err)
			continue
		}
		if got := New(ip, mask).First().IP(); !got.Equal(net.ParseIP(want)) {
			t.Errorf("First(%v) = %v, want %v", addr, got, want)
		}
	}
}

func TestLast(t *testing.T) {
	tests := map[string]string{
		"192.0.2.128/0.0.0.254":       "192.0.2.254",
		"192.0.2.255/0.0.0.0":         "192.0.2.255",
		"192.0.2.255/255.255.255.255": "255.255.255.255",
		"2001:db8::f0/::fffe":         "2001:db8::fffe",
		"2001:db8::ffff/::":           "2001:db8::ffff",
		"2001:db8::ffff/128":          "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	}
	for addr, want := range tests {
		ip, mask, err := ipcalc.ParseIPMask(addr)
		if err != nil {
			t.Errorf("ParseIPMask(%v) error = %v", addr, err)
			continue
		}
		if got := New(ip, mask).Last().IP(); !got.Equal(net.ParseIP(want)) {
			t.Errorf("Last(%v) = %v, want %v", addr, got, want)
		}
	}
}

func TestNext(t *testing.T) {
	tests := map[string][]string{
		"192.0.2.0/0.0.0.254":         []string{"192.0.2.2", "192.0.2.4", "192.0.2.6"},
		"192.0.2.3/0.0.0.254":         []string{"192.0.2.5", "192.0.2.7"},
		"192.0.2.128/0.0.0.254":       []string{"192.0.2.130", "192.0.2.132"},
		"192.0.2.252/0.0.0.254":       []string{"192.0.2.254", "192.0.2.0"},
		"192.0.2.255/0.0.0.0":         []string{"192.0.2.255"},
		"192.0.2.255/255.255.255.255": []string{"192.0.3.0", "192.0.3.1"},
		"2001:db8::/::fffe":           []string{"2001:db8::2", "2001:db8::4", "2001:db8::6"},
		"2001:db8::9/::fffe":          []string{"2001:db8::b", "2001:db8::d", "2001:db8::f", "2001:db8::11"},
		"2001:db8::ffff/::fffe":       []string{"2001:db8::1", "2001:db8::3"},
		"2001:db8::ffff/::ffff:fffe":  []string{"2001:db8::1:1", "2001:db8::1:3"},
		"2001:db8::ffff/::":           []string{"2001:db8::ffff"},
		"2001:db8::ffff/128":          []string{"2001:db8::1:0", "2001:db8::1:1"},
	}
	for addr, tt := range tests {
		ip, mask, err := ipcalc.ParseIPMask(addr)
		if err != nil {
			t.Errorf("ParseIPMask(%v) error = %v", addr, err)
			continue
		}
		w := New(ip, mask)
		for i, want := range tt {
			if got := w.Next(); !got.Equal(net.ParseIP(want)) {
				t.Errorf("Next(%v, %v) = %v, want %v", addr, i, got, want)
			}
		}
	}
}

func TestPrev(t *testing.T) {
	tests := map[string][]string{
		"192.0.2.10/0.0.0.254":      []string{"192.0.2.8", "192.0.2.6", "192.0.2.4"},
		"192.0.2.3/0.0.0.254":       []string{"192.0.2.1", "192.0.2.255"},
		"192.0.2.128/0.0.0.254":     []string{"192.0.2.126", "192.0.2.124"},
		"192.0.2.255/0.0.0.0":       []string{"192.0.2.255"},
		"192.0.2.1/255.255.255.255": []string{"192.0.2.0", "192.0.1.255"},
		"2001:db8::10/::fffe":       []string{"2001:db8::e", "2001:db8::c"},
		"2001:db8::9/::fffe":        []string{"2001:db8::7", "2001:db8::5"},
		"2001:db8::a:3/::fffe":      []string{"2001:db8::a:1", "2001:db8::a:ffff"},
		"2001:db8::a:3/::ffff:fffe": []string{"2001:db8::a:1", "2001:db8::9:ffff"},
		"2001:db8::ffff/::":         []string{"2001:db8::ffff"},
		"2001:db8::a:0/128":         []string{"2001:db8::9:ffff", "2001:db8::9:fffe"},
	}
	for addr, tt := range tests {
		ip, mask, err := ipcalc.ParseIPMask(addr)
		if err != nil {
			t.Errorf("ParseIPMask(%v) error = %v", addr, err)
			continue
		}
		w := New(ip, mask)
		for i, want := range tt {
			if got := w.Prev(); !got.Equal(net.ParseIP(want)) {
				t.Errorf("Prev(%v, %v) = %v, want %v", addr, i, got, want)
			}
		}
	}
}

func TestFindWildcard(t *testing.T) {
	tests := []struct {
		ips   []string
		wip   string
		wcard string
	}{
		{[]string{"192.0.2.1", "192.0.2.1"}, "192.0.2.1", "0.0.0.0"},
		{[]string{"192.0.2.1", "0.0.0.0"}, "0.0.0.0", "192.0.2.1"},
		{[]string{"192.0.2.1", "255.255.255.255"}, "192.0.2.1", "63.255.253.254"},
		{[]string{"192.0.2.1", "192.0.2.255"}, "192.0.2.1", "0.0.0.254"},
		{[]string{"192.0.2.134", "192.0.6.6", "192.0.254.70"}, "192.0.2.6", "0.0.252.64"},
		{[]string{"2001:db0::1", "2002::dead:beef"}, "2000::1", "3:db0::dead:beee"},
	}
	for _, tt := range tests {
		var ips []net.IP
		for _, ip := range tt.ips {
			ips = append(ips, net.ParseIP(ip))
		}
		wip := net.ParseIP(tt.wip)
		wcard := ipcalc.ParseMask(tt.wcard)
		got := FindWildcard(ips[0], ips[1], ips[2:]...)
		if !got.IP().Equal(wip) || !bytes.Equal(got.Wildcard(), wcard) {
			t.Errorf("FindWildcard(%v) = %v/%v, want %v/%v", ips, got.IP(), net.IP(got.Wildcard()), wip, wcard)
		}
	}
}
