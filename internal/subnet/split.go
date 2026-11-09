package subnet

import (
	"fmt"
	"math/big"
	"net/netip"
	"sort"
)

// Split divides a network into equal subnets of newPrefix length.
func Split(cidr string, newPrefix int) ([]Info, error) {
	info, err := Parse(cidr)
	if err != nil {
		return nil, err
	}
	if newPrefix < info.Bits || newPrefix > info.Addr.BitLen() {
		return nil, fmt.Errorf("%w: new prefix /%d must be between /%d and /%d", ErrInvalid, newPrefix, info.Bits, info.Addr.BitLen())
	}
	if newPrefix == info.Bits {
		return []Info{info}, nil
	}
	countBits := newPrefix - info.Bits
	count := new(big.Int).Lsh(big.NewInt(1), uint(countBits))
	if !count.IsInt64() || count.Int64() > 1<<20 {
		return nil, fmt.Errorf("%w: split would produce more than 1,048,576 subnets", ErrInvalid)
	}
	n := int(count.Int64())
	step := new(big.Int).Lsh(big.NewInt(1), uint(info.Addr.BitLen()-newPrefix))
	out := make([]Info, 0, n)
	cur := addrToInt(info.Network)
	for range n {
		addr := intToAddr(cur, info.Version == 4)
		p := netip.PrefixFrom(addr, newPrefix)
		out = append(out, fromPrefix(p, p.Masked().String()))
		cur = new(big.Int).Add(cur, step)
	}
	return out, nil
}

// SplitCount divides a network into count equal subnets (count must be a power of two).
func SplitCount(cidr string, count int) ([]Info, error) {
	if count < 1 || count&(count-1) != 0 {
		return nil, fmt.Errorf("%w: count must be a power of two ≥ 1", ErrInvalid)
	}
	info, err := Parse(cidr)
	if err != nil {
		return nil, err
	}
	extra := 0
	for c := count; c > 1; c >>= 1 {
		extra++
	}
	return Split(cidr, info.Bits+extra)
}

// Aggregate merges CIDRs into the smallest set of covering supernets
// (only exact pairs that form a supernet are combined; leftovers remain).
func Aggregate(cidrs []string) ([]Info, error) {
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("%w: no prefixes", ErrInvalid)
	}
	set := map[netip.Prefix]struct{}{}
	version := 0
	for _, c := range cidrs {
		info, err := Parse(c)
		if err != nil {
			return nil, err
		}
		if version == 0 {
			version = info.Version
		} else if info.Version != version {
			return nil, fmt.Errorf("%w: mixed IP versions", ErrInvalid)
		}
		set[info.Prefix.Masked()] = struct{}{}
	}

	changed := true
	for changed {
		changed = false
		prefixes := make([]netip.Prefix, 0, len(set))
		for p := range set {
			prefixes = append(prefixes, p)
		}
		sort.Slice(prefixes, func(i, j int) bool {
			if prefixes[i].Bits() != prefixes[j].Bits() {
				return prefixes[i].Bits() > prefixes[j].Bits()
			}
			return addrToInt(prefixes[i].Addr()).Cmp(addrToInt(prefixes[j].Addr())) < 0
		})
		for _, p := range prefixes {
			if _, ok := set[p]; !ok {
				continue
			}
			if p.Bits() == 0 {
				continue
			}
			parentBits := p.Bits() - 1
			parent := netip.PrefixFrom(p.Addr(), parentBits).Masked()
			// sibling: the other half of parent
			half := new(big.Int).Lsh(big.NewInt(1), uint(p.Addr().BitLen()-p.Bits()))
			base := addrToInt(parent.Addr())
			a := intToAddr(base, p.Addr().Is4())
			b := intToAddr(new(big.Int).Add(base, half), p.Addr().Is4())
			left := netip.PrefixFrom(a, p.Bits()).Masked()
			right := netip.PrefixFrom(b, p.Bits()).Masked()
			if _, okL := set[left]; !okL {
				continue
			}
			if _, okR := set[right]; !okR {
				continue
			}
			delete(set, left)
			delete(set, right)
			set[parent] = struct{}{}
			changed = true
			break
		}
	}

	out := make([]Info, 0, len(set))
	for p := range set {
		out = append(out, fromPrefix(p, p.String()))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Bits != out[j].Bits {
			return out[i].Bits < out[j].Bits
		}
		return addrToInt(out[i].Network).Cmp(addrToInt(out[j].Network)) < 0
	})
	return out, nil
}
