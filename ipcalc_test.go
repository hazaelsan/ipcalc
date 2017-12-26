package ipcalc

import (
	"bytes"
	"net"
	"testing"
)

func TestIP(t *testing.T) {
	tests := map[string]int{
		"192.0.2.0":        net.IPv4len,
		"::ffff:192.0.2.0": net.IPv4len,
		"2001:db8::1":      net.IPv6len,
		"::1":              net.IPv6len,
		"invalid":          0,
	}
	for ip, want := range tests {
		if got := len(IP(net.ParseIP(ip))); got != want {
			t.Errorf("IP(%v) = %v, want %v", ip, got, want)
		}
	}
}

func TestIPVersion(t *testing.T) {
	tests := map[string]int{
		"192.0.2.0":   4,
		"2001:db8::1": 6,
	}
	for ip, want := range tests {
		if got := IPVersion(net.ParseIP(ip)); got != want {
			t.Errorf("IPVersion(%v) = %v, want %v", ip, got, want)
		}
	}
}

func TestIPSize(t *testing.T) {
	tests := map[string]int{
		"192.0.2.0":   4,
		"2001:db8::1": 16,
	}
	for ip, want := range tests {
		if got := IPSize(net.ParseIP(ip)); got != want {
			t.Errorf("IPSize(%v) = %v, want %v", ip, got, want)
		}
	}
}

func TestComplement(t *testing.T) {
	tests := map[string]string{
		"255.255.255.0":         "0.0.0.255",
		"255.0.224.1":           "0.255.31.254",
		"ffff:ffff:ffff:fffc::": "::3:ffff:ffff:ffff:ffff",
		"ffff:fffe:aabb::ffff":  "0:1:5544:ffff:ffff:ffff:ffff:0",
		"0.0.0.0":               "255.255.255.255",
		"255.255.255.255":       "0.0.0.0",
		"invalid":               "",
	}
	for mask, comp := range tests {
		got := Complement(ParseMask(mask))
		want := ParseMask(comp)
		if !bytes.Equal(got, want) {
			t.Errorf("Complement(%v) = %v, want %v", mask, got, comp)
		}
	}
}

func TestParseIPMask(t *testing.T) {
	tests := []struct {
		addr string
		ip   string
		mask string
		ok   bool
	}{
		{"192.0.2.10/24", "192.0.2.10", "255.255.255.0", true},
		{"192.0.2.10/255.255.255.0", "192.0.2.10", "255.255.255.0", true},
		{"192.0.2.10", "192.0.2.10", "", true},
		{"192.0.2.10/~", "", "", false},
		{"192.0.2.10/~12", "192.0.2.10", "0.15.255.255", true},
		{"192.0.2.10/~255.240.0.0", "192.0.2.10", "0.15.255.255", true},
		{"0.0.0.0/0", "0.0.0.0", "0.0.0.0", true},
		{"invalid/24", "", "", false},
		{"192.0.2.0/invalid", "", "", false},
		{"192.0.2.0/24/1", "", "", false},
		{"2001:db8::/64", "2001:db8::", "ffff:ffff:ffff:ffff::", true},
		{"2001:db8::/ffff::", "2001:db8::", "ffff::", true},
		{"2001:db8::", "2001:db8::", "", true},
		{"foo", "0.0.0.0", "0.0.0.0", false},
	}
	for _, tt := range tests {
		ip, mask, err := ParseIPMask(tt.addr)
		if err != nil {
			if tt.ok {
				t.Errorf("ParseIPMask(%v) error = %v", tt.addr, err)
			}
			continue
		}
		if !tt.ok {
			t.Errorf("ParseIPMask(%v) error = nil", tt.addr)
		} else if !bytes.Equal(IP(ip), IP(net.ParseIP(tt.ip))) {
			t.Errorf("ParseIPMask(%v) ip = %v, want %v", tt.addr, ip, tt.ip)
		} else if !bytes.Equal(IP(net.IP(mask)), IP(net.ParseIP(tt.mask))) {
			t.Errorf("ParseIPMask(%v) mask = %v, want %v", tt.addr, mask, tt.mask)
		}
	}
}

func TestNextIP(t *testing.T) {
	tests := map[string]string{
		"0.0.0.0":         "0.0.0.1",
		"192.0.2.0":       "192.0.2.1",
		"192.0.2.255":     "192.0.3.0",
		"255.255.255.255": "0.0.0.0",
		"::":              "::1",
		"::1":             "::2",
		"2001:db8::ffff:ffff":                     "2001:db8::1:0:0",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff": "::",
	}
	for ip, want := range tests {
		if got := NextIP(net.ParseIP(ip)); !bytes.Equal(got.To16(), net.ParseIP(want)) {
			t.Errorf("NextIP(%v) = %v, want %v", ip, got, want)
		}
	}
}

func TestPrevIP(t *testing.T) {
	tests := map[string]string{
		"0.0.0.1":         "0.0.0.0",
		"192.0.2.1":       "192.0.2.0",
		"192.0.2.0":       "192.0.1.255",
		"0.0.0.0":         "255.255.255.255",
		"::2":             "::1",
		"2001:db8::1:0:0": "2001:db8::ffff:ffff",
		"::":              "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
	}
	for ip, want := range tests {
		if got := PrevIP(net.ParseIP(ip)); !bytes.Equal(got.To16(), net.ParseIP(want)) {
			t.Errorf("PrevIP(%v) = %v, want %v", ip, got, want)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		mask string
		want string
	}{
		{"192.0.2.1", "0.0.0.1", "0.0.0.255", "192.0.2.2"},
		{"192.0.2.1", "192.0.2.1", "0.0.0.255", "192.0.2.2"},
		{"192.0.2.1", "1.1.1.1", "255.255.255.255", "193.1.3.2"},
		{"192.0.2.255", "0.0.1.2", "0.0.1.255", "192.0.4.1"},
		{"255.255.255.255", "1.1.1.1", "255.255.255.0", "1.1.0.255"},
		{"2001:db8::ff", "::ff01", "::ffff", "2001:db8::1:0"},
	}
	for _, tt := range tests {
		got := Add(net.ParseIP(tt.a), net.ParseIP(tt.b), ParseMask(tt.mask))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("Add(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.mask, got, want)
		}
	}
}

func TestSubstract(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		mask string
		want string
	}{
		{"192.0.2.2", "0.0.0.1", "0.0.0.255", "192.0.2.1"},
		{"192.0.2.2", "192.0.2.1", "0.0.0.255", "192.0.2.1"},
		{"192.0.2.2", "1.1.1.1", "255.255.255.255", "190.255.1.1"},
		{"192.0.2.1", "0.0.1.2", "0.0.1.255", "192.0.0.255"},
		{"1.1.0.255", "1.1.1.1", "255.255.255.0", "255.255.255.255"},
		{"2001:db8::1:0", "::ff01", "::ffff", "2001:db8::ff"},
	}
	for _, tt := range tests {
		got := Substract(net.ParseIP(tt.a), net.ParseIP(tt.b), ParseMask(tt.mask))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("Substract(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.mask, got, want)
		}
	}
}

func TestAnd(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want string
	}{
		{"192.0.2.1", "0.0.0.0", "0.0.0.0"},
		{"192.0.2.1", "255.255.255.255", "192.0.2.1"},
		{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
		{"192.0.2.1", "0.0.0.1", "0.0.0.1"},
		{"192.0.2.1", "192.0.2.100", "192.0.2.0"},
		{"192.0.15.155", "190.130.11.127", "128.0.11.27"},
		{"255.255.255.255", "1.1.1.1", "1.1.1.1"},
		{"2001:db8:9::ae", "2001:db8:5::ff01", "2001:db8:1::"},
	}
	for _, tt := range tests {
		got := And(net.ParseIP(tt.a), net.ParseIP(tt.b))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("And(%v, %v) = %v, want %v", tt.a, tt.b, got, want)
		}
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want string
	}{
		{"192.0.2.31", "0.0.0.0", "192.0.2.31"},
		{"192.0.2.31", "255.255.255.255", "255.255.255.255"},
		{"192.0.2.31", "192.0.2.31", "192.0.2.31"},
		{"192.0.2.31", "192.0.2.144", "192.0.2.159"},
		{"192.0.2.110", "0.0.1.125", "192.0.3.127"},
		{"192.0.2.1", "192.0.2.100", "192.0.2.101"},
		{"192.0.15.155", "190.130.11.127", "254.130.15.255"},
		{"255.255.255.255", "1.1.1.1", "255.255.255.255"},
		{"2001:db8:9::ae", "2001:db8:5::ff01", "2001:db8:d::ffaf"},
	}
	for _, tt := range tests {
		got := Or(net.ParseIP(tt.a), net.ParseIP(tt.b))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("Or(%v, %v) = %v, want %v", tt.a, tt.b, got, want)
		}
	}
}

func TestXor(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want string
	}{
		{"192.0.2.1", "0.0.0.0", "192.0.2.1"},
		{"192.0.2.31", "192.0.2.31", "0.0.0.0"},
		{"192.0.2.110", "63.255.253.145", "255.255.255.255"},
		{"192.0.2.1", "255.255.255.255", "63.255.253.254"},
		{"192.0.2.1", "172.31.128.17", "108.31.130.16"},
		{"2001:db8:9::ae", "2001:db8:5::ff01", "0:0:c::ffaf"},
	}
	for _, tt := range tests {
		got := Xor(net.ParseIP(tt.a), net.ParseIP(tt.b))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("Xor(%v, %v) = %v, want %v", tt.a, tt.b, got, want)
		}
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		mask string
		want string
	}{
		{"192.0.2.133", "172.16.32.5", "0.0.255.255", "192.0.32.5"},
		{"192.0.2.133", "172.16.32.5", "255.255.255.0", "172.16.32.133"},
		{"127.127.127.127", "31.42.237.38", "255.20.15.209", "31.107.125.46"},
		{"2001:db8::9:ae", "2001:db8::5:ff01", "::dead:beef", "2001:db8::5:be01"},
	}
	for _, tt := range tests {
		got := Merge(net.ParseIP(tt.a), net.ParseIP(tt.b), ParseMask(tt.mask))
		want := IP(net.ParseIP(tt.want))
		if !bytes.Equal(got, want) {
			t.Errorf("Merge(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.mask, got, want)
		}
	}
}

func TestBroadcast(t *testing.T) {
	tests := map[string]string{
		"192.0.2.0/24":  "192.0.2.255",
		"192.0.2.0/31":  "192.0.2.1",
		"192.0.2.0/32":  "192.0.2.0",
		"2001:db8::/64": "2001:db8:0:0:ffff:ffff:ffff:ffff",
	}
	for addr, bcast := range tests {
		_, n, err := net.ParseCIDR(addr)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", addr, err)
			continue
		}
		want := IP(net.ParseIP(bcast))
		if got := Broadcast(*n); !bytes.Equal(got, want) {
			t.Errorf("Broadcast(%v) = %v, want %v", addr, got, want)
		}
	}
}

func TestNextSubnet(t *testing.T) {
	tests := map[string]string{
		"192.0.2.0/23":         "192.0.4.0/23",
		"192.0.2.0/24":         "192.0.3.0/24",
		"192.0.2.0/25":         "192.0.2.128/25",
		"192.0.2.0/31":         "192.0.2.2/31",
		"192.0.2.0/32":         "192.0.2.1/32",
		"255.255.255.0/24":     "0.0.0.0/24",
		"2001:db8::/64":        "2001:db8:0:1::/64",
		"2001:db8:9:fffe::/63": "2001:db8:a::/63",
	}
	for addr, subnet := range tests {
		_, n, err := net.ParseCIDR(addr)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", addr, err)
			continue
		}
		_, want, err := net.ParseCIDR(subnet)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", subnet, err)
		}
		if got := NextSubnet(*n); !bytes.Equal(got.IP, want.IP) || !bytes.Equal(got.Mask, want.Mask) {
			t.Errorf("NextSubnet(%v) = %v, want %v", n, got.String(), want)
		}
	}
}

func TestPrevSubnet(t *testing.T) {
	tests := map[string]string{
		"192.0.2.0/23":      "192.0.0.0/23",
		"192.0.2.0/24":      "192.0.1.0/24",
		"192.0.2.128/25":    "192.0.2.0/25",
		"192.0.2.2/31":      "192.0.2.0/31",
		"192.0.2.1/32":      "192.0.2.0/32",
		"0.0.0.0/24":        "255.255.255.0/24",
		"2001:db8:0:1::/64": "2001:db8::/64",
		"2001:db8:a::/63":   "2001:db8:9:fffe::/63",
	}
	for addr, subnet := range tests {
		_, n, err := net.ParseCIDR(addr)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", addr, err)
			continue
		}
		_, want, err := net.ParseCIDR(subnet)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", subnet, err)
		}
		if got := PrevSubnet(*n); !bytes.Equal(got.IP, want.IP) || !bytes.Equal(got.Mask, want.Mask) {
			t.Errorf("PrevSubnet(%v) = %v, want %v", n, got.String(), want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want bool
	}{
		{"192.0.2.0/24", "192.0.2.0/24", true},
		{"192.0.2.0/24", "192.0.2.0/23", false},
		{"192.0.2.0/24", "192.0.2.0/25", true},
		{"192.0.2.0/24", "192.0.1.0/24", false},
		{"2001:db8:a::/48", "2001:db8:a:1::/64", true},
		{"2001:db8:a::/48", "2001:db8:b:1::/64", false},
	}
	for _, tt := range tests {
		_, a, err := net.ParseCIDR(tt.a)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", tt.a, err)
			continue
		}
		_, b, err := net.ParseCIDR(tt.b)
		if err != nil {
			t.Errorf("ParseCIDR(%v) error = %v", tt.b, err)
			continue
		}
		if got := Contains(*a, *b); got != tt.want {
			t.Errorf("Contains(%v, %v) = %v, want %v", a, b, got, tt.want)
		}
	}
}
