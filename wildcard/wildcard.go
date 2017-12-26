// Package wildcard provides utilities for working with Wildcard Masks.
//
// Wildcard masks are often thought of as "inverse subnet masks", there are some differences:
// * Matching logic is reversed from subnet masks, meaning
// * 0 means the equivalent bit must match
// * 1 means the equivalent bit does not matter
// * 0s and 1s need not be contiguous
//
// Examples below, format is IP/Wildcard:
// * 192.0.2.0/0.0.0.255 matches 192.0.2.*
// * 192.0.2.10/0.0.255.0 matches 192.0.*.10
// * 192.0.2.1/0.0.255.254 matches 192.0.*.{1,3,5,7,...,255}
//
// Use ipcalc.Complement to convert a subnet mask to its wildcard counterpart.
package wildcard

import (
	"net"

	"github.com/hazaelsan/ipcalc"
)

// Wildcard represents a wildcard mask.
type Wildcard struct {
	ip   net.IP
	bits net.IP
	mask net.IP
}

// New returns a Wildcard from a given IP address and wildcard mask.
func New(ip net.IP, wildcard net.IPMask) Wildcard {
	ip = ipcalc.IP(ip)
	mask := ipcalc.IP(net.IP(ipcalc.Complement(wildcard)))
	return Wildcard{
		ip:   ip,
		bits: ipcalc.And(ip, mask),
		mask: mask,
	}
}

// IP returns the current IP address for a Wildcard.
func (w Wildcard) IP() net.IP {
	return w.ip
}

// Wildcard returns the wildcard mask for a Wildcard.
func (w Wildcard) Wildcard() net.IPMask {
	return ipcalc.Complement(net.IPMask(w.mask))
}

// Matches returns whether an IP address matches the Wildcard.
func (w Wildcard) Matches(ip net.IP) bool {
	return ipcalc.And(ip, w.mask).Equal(w.bits)
}

// First returns a Wildcard with the lowest IP address matching the Wildcard.
// e.g., New(192.0.2.128, 0.0.0.254).First() -> Wildcard(192.0.2.0, 0.0.0.254).
func (w Wildcard) First() Wildcard {
	return Wildcard{
		ip:   ipcalc.IP(w.bits),
		bits: ipcalc.IP(w.bits),
		mask: ipcalc.IP(w.mask),
	}
}

// Last returns a Wildcard with the highest IP address matching the Wildcard.
// e.g., New(192.0.2.128, 0.0.0.254).Last() -> Wildcard(192.0.2.254, 0.0.0.254).
func (w Wildcard) Last() Wildcard {
	return Wildcard{
		ip:   ipcalc.Or(w.bits, net.IP(ipcalc.Complement(net.IPMask(w.mask)))),
		bits: ipcalc.IP(w.bits),
		mask: ipcalc.IP(w.mask),
	}
}

// Next returns the next IP address matching the Wildcard.
// e.g., New(192.0.2.128, 0.0.0.254).Next() -> 192.0.2.130.
func (w *Wildcard) Next() net.IP {
	for i := ipcalc.IPSize(w.ip) - 1; i >= 0; i-- {
		for j := uint8(0); j < 8; j++ {
			if bit(w.mask[i], j) {
				continue
			}
			if !bit(w.ip[i], j) {
				w.ip[i] |= 1 << j
				return w.ip
			}
			w.ip[i] &= ^(1 << j)
		}
	}
	return w.ip
}

// Prev returns the previous IP address matching the Wildcard.
// e.g., New(192.0.2.128, 0.0.0.254).Prev() -> 192.0.2.126.
func (w *Wildcard) Prev() net.IP {
	for i := ipcalc.IPSize(w.ip) - 1; i >= 0; i-- {
		for j := uint8(0); j < 8; j++ {
			if bit(w.mask[i], j) {
				continue
			}
			if bit(w.ip[i], j) {
				w.ip[i] &= ^(1 << j)
				return w.ip
			}
			w.ip[i] |= 1 << j
		}
	}
	return w.ip
}

func bit(b byte, i uint8) bool {
	return b>>i&1 == 1
}

// FindWildcard returns the most specific Wildcard matching several IP addresses.
func FindWildcard(a, b net.IP, extra ...net.IP) Wildcard {
	and, xor := findWildcard(a, b)
	for _, ip := range extra {
		and, xor = findWildcard(and, ip)
	}
	return New(and, xor)
}

func findWildcard(a, b net.IP) (net.IP, net.IPMask) {
	return ipcalc.And(a, b), net.IPMask(ipcalc.Xor(a, b))
}
