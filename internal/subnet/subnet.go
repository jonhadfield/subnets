package subnet

import (
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"strings"
)

// Info holds subnet details for an address or CIDR.
type Info struct {
	Input        string
	Addr         netip.Addr
	Prefix       netip.Prefix
	Bits         int
	HostBits     int
	Network      netip.Addr
	Last         netip.Addr
	Mask         netip.Addr
	Wildcard     netip.Addr
	FirstHost    netip.Addr
	LastHost     netip.Addr
	Total        *big.Int
	Usable       *big.Int
	Version      int
	HasBroadcast bool
	Kind         Kind
}

// Kind is a short address classification label.
type Kind struct {
	Name string
}

var ErrInvalid = errors.New("invalid address or CIDR")

// Parse accepts an address ("192.168.0.3") or CIDR ("192.168.0.3/28").
// A bare IPv4 address is treated as /32; a bare IPv6 address as /128.
func Parse(s string) (Info, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Info{}, fmt.Errorf("%w: empty input", ErrInvalid)
	}
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return fromPrefix(prefix, s), nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return Info{}, fmt.Errorf("%w: %s", ErrInvalid, s)
	}
	return fromPrefix(netip.PrefixFrom(addr, addr.BitLen()), s), nil
}

func fromPrefix(prefix netip.Prefix, input string) Info {
	addr := prefix.Addr().Unmap()
	prefix = netip.PrefixFrom(addr, prefix.Bits())
	network := prefix.Masked().Addr()
	last := lastAddr(prefix)
	bits := prefix.Bits()
	bitLen := addr.BitLen()
	hostBits := bitLen - bits
	is4 := addr.Is4()
	mask := maskAddr(is4, bits)
	total := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))

	info := Info{
		Input:        input,
		Addr:         addr,
		Prefix:       prefix,
		Bits:         bits,
		HostBits:     hostBits,
		Network:      network,
		Last:         last,
		Mask:         mask,
		Wildcard:     wildcardAddr(mask),
		Total:        total,
		Version:      4,
		HasBroadcast: is4 && bits <= 30,
		Kind:         classify(addr),
	}
	if !is4 {
		info.Version = 6
	}
	info.FirstHost, info.LastHost, info.Usable = classicUsable(info)
	return info
}

func classicUsable(info Info) (first, last netip.Addr, usable *big.Int) {
	switch {
	case info.HostBits == 0:
		return info.Network, info.Network, big.NewInt(1)
	case info.HostBits == 1:
		return info.Network, info.Last, big.NewInt(2)
	case info.HasBroadcast:
		return offset(info.Network, 1), offset(info.Last, -1), new(big.Int).Sub(info.Total, big.NewInt(2))
	default:
		// IPv6 has no broadcast; the subnet-router anycast is the network address.
		return offset(info.Network, 1), info.Last, new(big.Int).Sub(info.Total, big.NewInt(1))
	}
}

// Broadcast is the IPv4 directed broadcast address. The zero address if none.
func (i Info) Broadcast() netip.Addr {
	if !i.HasBroadcast {
		return netip.Addr{}
	}
	return i.Last
}

// HostIndex is the zero-based offset of the input address within the subnet.
func (i Info) HostIndex() *big.Int {
	return new(big.Int).Sub(addrToInt(i.Addr), addrToInt(i.Network))
}

// AddressAt returns the address at a zero-based offset from the network.
func (i Info) AddressAt(index uint64) (netip.Addr, bool) {
	if i.Total.Cmp(new(big.Int).SetUint64(index)) <= 0 {
		return netip.Addr{}, false
	}
	return offset(i.Network, int64(index)), true
}

// NetworkCIDR is the canonical network prefix, e.g. 192.168.0.0/28.
func (i Info) NetworkCIDR() string {
	return netip.PrefixFrom(i.Network, i.Bits).String()
}

func classify(addr netip.Addr) Kind {
	switch {
	case addr.IsUnspecified():
		return Kind{Name: "Unspecified"}
	case addr.IsLoopback():
		return Kind{Name: "Loopback"}
	case addr.IsMulticast():
		return Kind{Name: "Multicast"}
	case addr.IsLinkLocalUnicast():
		return Kind{Name: "Link-local"}
	case isSharedAddressSpace(addr):
		return Kind{Name: "Shared (CGNAT)"}
	case isDocumentation(addr):
		return Kind{Name: "Documentation"}
	case addr.IsPrivate():
		if addr.Is4() {
			return Kind{Name: "Private"}
		}
		return Kind{Name: "Unique local"}
	case addr.IsGlobalUnicast():
		return Kind{Name: "Public"}
	default:
		return Kind{Name: "Other"}
	}
}

func isSharedAddressSpace(addr netip.Addr) bool {
	if !addr.Is4() {
		return false
	}
	return netip.MustParsePrefix("100.64.0.0/10").Contains(addr)
}

func isDocumentation(addr netip.Addr) bool {
	docs := []string{
		"192.0.2.0/24",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"2001:db8::/32",
	}
	for _, cidr := range docs {
		if netip.MustParsePrefix(cidr).Contains(addr) {
			return true
		}
	}
	return false
}

// LegacyClass is the historical IPv4 class (A–E). Empty for IPv6.
func (i Info) LegacyClass() string {
	if i.Version != 4 {
		return ""
	}
	first := i.Addr.As4()[0]
	switch {
	case first < 128:
		return "A"
	case first < 192:
		return "B"
	case first < 224:
		return "C"
	case first < 240:
		return "D (multicast)"
	default:
		return "E (reserved)"
	}
}

func FormatNumber(n *big.Int) string {
	s := n.String()
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	rem := len(s) % 3
	if rem > 0 {
		b.WriteString(s[:rem])
		if len(s) > rem {
			b.WriteByte(',')
		}
	}
	for i := rem; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}

func Hex(addr netip.Addr) string {
	raw := addr.AsSlice()
	var b strings.Builder
	b.WriteString("0x")
	for _, c := range raw {
		fmt.Fprintf(&b, "%02X", c)
	}
	return b.String()
}

func Binary(addr netip.Addr) string {
	raw := addr.AsSlice()
	parts := make([]string, 0, len(raw))
	for _, b := range raw {
		parts = append(parts, fmt.Sprintf("%08b", b))
	}
	sep := "."
	if addr.Is6() {
		sep = ":"
	}
	return strings.Join(parts, sep)
}
