package subnet

import (
	"fmt"
	"net/netip"
)

// ContainsResult is whether outer contains inner (address or prefix).
type ContainsResult struct {
	Outer    string
	Inner    string
	Contains bool
	Relation string // "equal", "subnet", "address", "disjoint", "supernet", "overlap"
}

// Contains reports whether outerCIDR fully contains inner (CIDR or address).
func Contains(outerStr, innerStr string) (ContainsResult, error) {
	outer, err := Parse(outerStr)
	if err != nil {
		return ContainsResult{}, err
	}
	inner, err := Parse(innerStr)
	if err != nil {
		return ContainsResult{}, err
	}
	if outer.Version != inner.Version {
		return ContainsResult{}, fmt.Errorf("%w: IP versions differ", ErrInvalid)
	}

	out := ContainsResult{
		Outer: outer.NetworkCIDR(),
		Inner: inner.NetworkCIDR(),
	}
	if inner.Bits == inner.Addr.BitLen() && !stringsHasSlash(innerStr) {
		out.Inner = inner.Addr.String()
	}

	o, i := outer.Prefix, inner.Prefix
	switch {
	case o == i:
		out.Contains = true
		out.Relation = "equal"
	case o.Contains(i.Addr()) && o.Bits() < i.Bits() && prefixContainsPrefix(o, i):
		out.Contains = true
		if i.Bits() == i.Addr().BitLen() {
			out.Relation = "address"
		} else {
			out.Relation = "subnet"
		}
	case i.Contains(o.Addr()) && i.Bits() < o.Bits() && prefixContainsPrefix(i, o):
		out.Contains = false
		out.Relation = "supernet"
	case prefixesOverlap(o, i):
		out.Contains = false
		out.Relation = "overlap"
	default:
		out.Contains = false
		out.Relation = "disjoint"
	}
	return out, nil
}

func stringsHasSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}

func prefixContainsPrefix(outer, inner netip.Prefix) bool {
	if outer.Bits() > inner.Bits() {
		return false
	}
	return outer.Contains(inner.Addr()) && outer.Contains(lastAddr(inner))
}

func prefixesOverlap(a, b netip.Prefix) bool {
	return a.Overlaps(b)
}
