// Package ipcalc provides network utilities for performing IP address arithmetic.
//
// Note that next/prev functions are subject to wrapping,
// e.g., NextIP(255.255.255.255) -> 0.0.0.0.
package ipcalc

import (
	"net"
	"strconv"
	"strings"
)

// CopyIP returns a copy of a net.IP address.
func CopyIP(ip net.IP) net.IP {
	return append(net.IP(nil), ip...)
}

// IP returns an IP address of the correct byte length.
func IP(ip net.IP) net.IP {
	if x := ip.To4(); x != nil {
		return CopyIP(x)
	}
	return CopyIP(ip)
}

// IPVersion returns the IP address version for the given net.IP.
func IPVersion(ip net.IP) int {
	if len(IP(ip)) == net.IPv4len {
		return 4
	}
	return 6
}

// IPSize returns the address size in bytes for the given net.IP.
func IPSize(ip net.IP) int {
	if IPVersion(ip) == 4 {
		return net.IPv4len
	}
	return net.IPv6len
}

// ParseMask returns a net.IPMask from a string representation.
func ParseMask(mask string) net.IPMask {
	return net.IPMask(IP(net.ParseIP(mask)))
}

// ParseIPMask returns a net.IP and net.IPMask from an ip[/mask] string representation.
// The IPMask need not be in CIDR format, if it starts with ~ then it will be inverted.
func ParseIPMask(addr string) (net.IP, net.IPMask, error) {
	v := strings.Split(addr, "/")
	ip := net.ParseIP(v[0])
	if ip == nil {
		return nil, nil, &net.ParseError{Type: "IP address", Text: v[0]}
	}
	var mask net.IPMask
	if len(v) > 2 {
		return nil, nil, &net.ParseError{Type: "IP/Mask", Text: addr}
	} else if len(v) == 2 {
		size := IPSize(ip) * 8
		wildcard := false
		if strings.HasPrefix(v[1], "~") {
			wildcard = true
			v[1] = v[1][1:]
		}
		if bits, err := strconv.Atoi(v[1]); err == nil {
			mask = net.CIDRMask(bits, size)
		} else {
			mask = ParseMask(v[1])
		}
		if wildcard {
			mask = Complement(mask)
		}
		if mask == nil {
			return nil, nil, &net.ParseError{Type: "Mask", Text: v[1]}
		}
	}
	return ip, mask, nil
}

// Complement returns the complement of a given net.IPMask, commonly used as a Wildcard Mask.
// e.g., Complement(255.255.254.0) -> 0.0.1.255.
func Complement(mask net.IPMask) net.IPMask {
	if mask == nil {
		return nil
	}
	w := make(net.IPMask, len(mask))
	for i := 0; i < len(mask); i++ {
		w[i] = ^mask[i]
	}
	return w
}

// NextIP returns the next IP address.
// e.g., NextIP(192.168.0.0) -> 192.168.0.1.
func NextIP(ip net.IP) net.IP {
	ip = IP(ip)
	for i := IPSize(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0x00 {
			break
		}
	}
	return ip
}

// PrevIP returns the previous IP address.
// e.g., PrevIP(192.168.0.1) -> 192.168.0.0.
func PrevIP(ip net.IP) net.IP {
	ip = IP(ip)
	for i := IPSize(ip) - 1; i >= 0; i-- {
		ip[i]--
		if ip[i] != 0xff {
			break
		}
	}
	return ip
}

// Add returns the sum of two net.IP addresses with the given mask.
// e.g., Add(192.168.0.1, 192.168.0.2, 0.0.0.255) -> 192.168.0.3.
func Add(a, b net.IP, mask net.IPMask) net.IP {
	a = IP(a)
	b = IP(b)
	for i := len(mask) - 1; i >= 0; i-- {
		prev := a[i]
		a[i] += b[i] & mask[i]
		if a[i] < prev && i > 0 {
			for j := i - 1; j >= 0; j-- {
				a[j]++
				if a[j] != 0x00 {
					break
				}
			}
		}
	}
	return a
}

// Substract returns the difference of two net.IP addresses with the given mask.
// e.g., Substract(192.168.0.3, 192.168.0.1, 0.0.0.255) -> 192.168.0.2.
func Substract(a, b net.IP, mask net.IPMask) net.IP {
	a = IP(a)
	b = IP(b)
	for i := len(mask) - 1; i >= 0; i-- {
		prev := a[i]
		a[i] -= b[i] & mask[i]
		if a[i] > prev && i > 0 {
			for j := i - 1; j >= 0; j-- {
				a[j]--
				if a[j] != 0xff {
					break
				}
			}
		}
	}
	return a
}

// And returns the bitwise AND of two net.IP addresses.
// e.g., And(192.168.0.255, 192.168.255.128) -> 192.168.0.128.
func And(a, b net.IP) net.IP {
	a = IP(a)
	b = IP(b)
	for i := 0; i < IPSize(b); i++ {
		a[i] &= b[i]
	}
	return a
}

// Or returns the bitwise OR of two net.IP addresses.
// e.g., Or(192.168.0.15, 192.168.10.128) -> 192.168.0.128.
func Or(a, b net.IP) net.IP {
	a = IP(a)
	b = IP(b)
	for i := 0; i < IPSize(b); i++ {
		a[i] |= b[i]
	}
	return a
}

// Xor returns the bitwise XOR of two net.IP addresses.
// e.g., Xor(192.0.2.1, 172.31.128.17) -> 108.31.130.16.
func Xor(a, b net.IP) net.IP {
	a = IP(a)
	b = IP(b)
	for i := 0; i < IPSize(b); i++ {
		a[i] ^= b[i]
	}
	return a
}

// Merge combines two net.IP addresses with the given mask.
// For bit i, if mask[i] is set then b[i] is returned, otherwise a[i] is returned.
// e.g., Merge(192.168.0.1, 172.16.32.100, 0.0.0.255) -> 192.168.0.100.
func Merge(a, b net.IP, mask net.IPMask) net.IP {
	a = IP(a)
	b = IP(b)
	ip := make(net.IP, len(mask))
	for i := 0; i < len(mask); i++ {
		ip[i] = a[i]&^mask[i] | b[i]&mask[i]
	}
	return ip
}

// Broadcast returns the broadcast IP address for the given IPNet.
func Broadcast(n net.IPNet) net.IP {
	ip := make(net.IP, IPSize(n.IP))
	for i := 0; i < IPSize(n.IP); i++ {
		ip[i] = n.IP[i] | ^n.Mask[i]
	}
	return ip
}

// NextSubnet returns the next subnet.
// e.g., NextSubnet(192.168.0.0/24) -> 192.168.1.0/24.
func NextSubnet(n net.IPNet) net.IPNet {
	return net.IPNet{
		IP:   NextIP(Broadcast(n)),
		Mask: n.Mask,
	}
}

// PrevSubnet returns the previous subnet.
// e.g., PrevSubnet(192.168.1.0/24) -> 192.168.0.0/24.
func PrevSubnet(n net.IPNet) net.IPNet {
	return net.IPNet{
		IP:   PrevIP(n.IP).Mask(n.Mask),
		Mask: n.Mask,
	}
}

// Contains returns whether the first net.IPNet wholly contains the second one.
// If either net has a non-standard mask then the result is undefined.
func Contains(a, b net.IPNet) bool {
	return a.Contains(b.IP) && a.Contains(Broadcast(b))
}
