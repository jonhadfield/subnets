package subnet

import (
	"math/big"
)

// SheetRow is one line of the prefix-length cheatsheet.
type SheetRow struct {
	Prefix    int
	Mask      string
	Wildcard  string
	Addresses string
	Usable    string
	Note      string
}

// IPv4Sheet returns prefix, mask, and host counts for /0 through /32.
func IPv4Sheet() []SheetRow {
	rows := make([]SheetRow, 0, 33)
	for bits := range 33 {
		mask := maskAddr(true, bits)
		total := new(big.Int).Lsh(big.NewInt(1), uint(32-bits))
		usable := new(big.Int).Set(total)
		switch bits {
		case 32:
			usable = big.NewInt(1)
		case 31:
			usable = big.NewInt(2)
		default:
			usable.Sub(usable, big.NewInt(2))
		}
		rows = append(rows, SheetRow{
			Prefix:    bits,
			Mask:      mask.String(),
			Wildcard:  wildcardAddr(mask).String(),
			Addresses: FormatNumber(total),
			Usable:    FormatNumber(usable),
			Note:      ipv4Note(bits),
		})
	}
	return rows
}

// IPv6Sheet returns commonly used IPv6 prefix lengths.
func IPv6Sheet() []SheetRow {
	prefixes := []int{32, 48, 52, 56, 64, 96, 112, 120, 124, 126, 127, 128}
	rows := make([]SheetRow, 0, len(prefixes))
	for _, bits := range prefixes {
		mask := maskAddr(false, bits)
		total := new(big.Int).Lsh(big.NewInt(1), uint(128-bits))
		rows = append(rows, SheetRow{
			Prefix:    bits,
			Mask:      mask.String(),
			Wildcard:  wildcardAddr(mask).String(),
			Addresses: FormatNumber(total),
			Usable:    FormatNumber(total),
			Note:      ipv6Note(bits),
		})
	}
	return rows
}

func ipv4Note(bits int) string {
	switch bits {
	case 0:
		return "Default route"
	case 8:
		return "Class A"
	case 16:
		return "Class B"
	case 24:
		return "Class C"
	case 30:
		return "Point-to-point (2 hosts)"
	case 31:
		return "RFC 3021 p2p"
	case 32:
		return "Single host"
	default:
		return ""
	}
}

func ipv6Note(bits int) string {
	switch bits {
	case 32:
		return "Typical allocation"
	case 48:
		return "Site prefix"
	case 56:
		return "Small site"
	case 64:
		return "LAN / SLAAC"
	case 127:
		return "Point-to-point"
	case 128:
		return "Single host"
	default:
		return ""
	}
}
