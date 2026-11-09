package subnet

import (
	"math/big"
	"net/netip"
)

// Cloud is a cloud provider with VPC/VNet reserved addresses.
type Cloud string

const (
	CloudAWS   Cloud = "aws"
	CloudAzure Cloud = "azure"
	CloudGCP   Cloud = "gcp"
)

// Reserved is an address a cloud provider keeps inside a subnet.
type Reserved struct {
	Addr   netip.Addr
	Role   string
	Detail string
}

// CloudReport is a cloud provider's view of a subnet's reserved addresses.
type CloudReport struct {
	Provider    string
	Reserved    []Reserved
	FirstUsable netip.Addr
	LastUsable  netip.Addr
	Usable      *big.Int
	HasUsable   bool
	TooSmall    bool
	InputRole   string
	MinPrefix   int
	PrefixOK    bool
	Note        string
}

// AWSReport is an alias kept for existing call sites.
type AWSReport = CloudReport

func (c Cloud) Title() string {
	switch c {
	case CloudAzure:
		return "Azure"
	case CloudGCP:
		return "GCP"
	default:
		return "AWS"
	}
}

// CloudFor returns reserved addresses for the given cloud provider.
func CloudFor(info Info, cloud Cloud) CloudReport {
	switch cloud {
	case CloudAzure:
		return Azure(info)
	case CloudGCP:
		return GCP(info)
	default:
		return AWS(info)
	}
}

// CloudRole returns the reservation role for addr, or "" if usable.
func CloudRole(report CloudReport, addr netip.Addr) string {
	for _, r := range report.Reserved {
		if r.Addr == addr {
			return r.Role
		}
	}
	return ""
}

// AWSRole is kept for existing call sites.
func AWSRole(report CloudReport, addr netip.Addr) string {
	return CloudRole(report, addr)
}

type reservedSpec struct {
	delta  int64
	fromEnd bool
	role   string
	detail string
}

func buildCloud(
	info Info,
	provider string,
	minPrefix, maxPrefix int,
	prefixOK bool,
	note string,
	specs []reservedSpec,
	firstUsableDelta, lastUsableDelta int64,
) CloudReport {
	report := CloudReport{
		Provider:  provider,
		Usable:    big.NewInt(0),
		MinPrefix: minPrefix,
		PrefixOK:  prefixOK,
		Note:      note,
	}
	seen := make(map[netip.Addr]struct{})
	add := func(addr netip.Addr, role, detail string) {
		if !info.Prefix.Contains(addr) {
			return
		}
		if _, ok := seen[addr]; ok {
			return
		}
		seen[addr] = struct{}{}
		report.Reserved = append(report.Reserved, Reserved{Addr: addr, Role: role, Detail: detail})
	}
	for _, s := range specs {
		var addr netip.Addr
		if s.fromEnd {
			addr = offset(info.Last, s.delta)
		} else {
			addr = offset(info.Network, s.delta)
		}
		add(addr, s.role, s.detail)
	}

	first := offset(info.Network, firstUsableDelta)
	last := offset(info.Last, lastUsableDelta)
	if info.Prefix.Contains(first) && info.Prefix.Contains(last) && addrToInt(first).Cmp(addrToInt(last)) <= 0 {
		report.FirstUsable = first
		report.LastUsable = last
		report.HasUsable = true
		report.Usable = new(big.Int).Add(new(big.Int).Sub(addrToInt(last), addrToInt(first)), big.NewInt(1))
	} else {
		report.TooSmall = true
	}

	for _, r := range report.Reserved {
		if r.Addr == info.Addr {
			report.InputRole = r.Role
			break
		}
	}
	return report
}
