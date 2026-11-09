package subnet

import (
	"fmt"
	"net/netip"
	"strings"
)

// MaskBits returns the prefix length of a contiguous IPv4/IPv6 subnet mask,
// or an error if the address is not a valid contiguous mask.
func MaskBits(mask netip.Addr) (int, error) {
	raw := mask.AsSlice()
	bits := 0
	seenZero := false
	for _, b := range raw {
		for i := 7; i >= 0; i-- {
			on := b&(1<<uint(i)) != 0
			if on {
				if seenZero {
					return 0, fmt.Errorf("%w: non-contiguous mask %s", ErrInvalid, mask)
				}
				bits++
			} else {
				seenZero = true
			}
		}
	}
	return bits, nil
}

// ParseMaskPrefix accepts a dotted/colonic mask alone (e.g. "255.255.255.240")
// and returns a 0.0.0.0/n or ::/n Info for that prefix length.
func ParseMaskPrefix(s string) (Info, error) {
	s = strings.TrimSpace(s)
	mask, err := netip.ParseAddr(s)
	if err != nil {
		return Info{}, fmt.Errorf("%w: %s", ErrInvalid, s)
	}
	mask = mask.Unmap()
	bits, err := MaskBits(mask)
	if err != nil {
		return Info{}, err
	}
	var network netip.Addr
	if mask.Is4() {
		network = netip.IPv4Unspecified()
	} else {
		network = netip.IPv6Unspecified()
	}
	input := s
	return fromPrefix(netip.PrefixFrom(network, bits), input), nil
}

// ParseAddrMask builds Info from an address and a separate subnet mask.
func ParseAddrMask(addrStr, maskStr string) (Info, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(addrStr))
	if err != nil {
		return Info{}, fmt.Errorf("%w: %s", ErrInvalid, addrStr)
	}
	mask, err := netip.ParseAddr(strings.TrimSpace(maskStr))
	if err != nil {
		return Info{}, fmt.Errorf("%w: %s", ErrInvalid, maskStr)
	}
	addr = addr.Unmap()
	mask = mask.Unmap()
	if addr.Is4() != mask.Is4() {
		return Info{}, fmt.Errorf("%w: address and mask IP versions differ", ErrInvalid)
	}
	bits, err := MaskBits(mask)
	if err != nil {
		return Info{}, err
	}
	input := addr.String() + " " + mask.String()
	return fromPrefix(netip.PrefixFrom(addr, bits), input), nil
}

// parseAddrSlashMask handles "192.168.0.3/255.255.255.240".
func parseAddrSlashMask(s string) (Info, bool, error) {
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return Info{}, false, nil
	}
	addrPart, maskPart := s[:i], s[i+1:]
	if _, err := netip.ParseAddr(maskPart); err != nil {
		return Info{}, false, nil
	}
	if _, err := netip.ParseAddr(addrPart); err != nil {
		return Info{}, false, nil
	}
	info, err := ParseAddrMask(addrPart, maskPart)
	if err != nil {
		return Info{}, true, err
	}
	info.Input = s
	return info, true, nil
}
