package subnet

import (
	"encoding/binary"
	"math/big"
	"net/netip"
)

func addrToInt(addr netip.Addr) *big.Int {
	return new(big.Int).SetBytes(addr.AsSlice())
}

func intToAddr(n *big.Int, is4 bool) netip.Addr {
	if is4 {
		var b [4]byte
		n.FillBytes(b[:])
		return netip.AddrFrom4(b)
	}
	var b [16]byte
	n.FillBytes(b[:])
	return netip.AddrFrom16(b)
}

func offset(addr netip.Addr, delta int64) netip.Addr {
	n := addrToInt(addr)
	n.Add(n, big.NewInt(delta))
	return intToAddr(n, addr.Is4())
}

func lastAddr(prefix netip.Prefix) netip.Addr {
	network := prefix.Masked().Addr()
	hostBits := prefix.Addr().BitLen() - prefix.Bits()
	if network.Is4() {
		n := binary.BigEndian.Uint32(network.AsSlice())
		var hostMask uint32
		if hostBits == 32 {
			hostMask = ^uint32(0)
		} else {
			hostMask = uint32(1<<hostBits) - 1
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], n|hostMask)
		return netip.AddrFrom4(b)
	}
	raw := network.As16()
	remaining := hostBits
	for i := 15; i >= 0 && remaining > 0; i-- {
		n := min(remaining, 8)
		raw[i] |= byte(1<<n) - 1
		remaining -= n
	}
	return netip.AddrFrom16(raw)
}

func maskAddr(is4 bool, bits int) netip.Addr {
	if is4 {
		var mask uint32
		if bits > 0 {
			mask = ^uint32(0) << (32 - bits)
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], mask)
		return netip.AddrFrom4(b)
	}
	var raw [16]byte
	full := bits / 8
	for i := range full {
		raw[i] = 0xff
	}
	if rem := bits % 8; rem > 0 && full < 16 {
		raw[full] = byte(^uint8(0) << (8 - rem))
	}
	return netip.AddrFrom16(raw)
}

func wildcardAddr(mask netip.Addr) netip.Addr {
	raw := mask.AsSlice()
	for i := range raw {
		raw[i] = ^raw[i]
	}
	if mask.Is4() {
		return netip.AddrFrom4([4]byte(raw))
	}
	return netip.AddrFrom16([16]byte(raw))
}
