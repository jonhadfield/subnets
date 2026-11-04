package subnet

import (
	"math/big"
	"net/netip"
)

// Reserved is an address AWS keeps inside a VPC subnet.
type Reserved struct {
	Addr   netip.Addr
	Role   string
	Detail string
}

// AWSReport is the AWS VPC view of a subnet.
type AWSReport struct {
	Reserved    []Reserved
	FirstUsable netip.Addr
	LastUsable  netip.Addr
	Usable      *big.Int
	HasUsable   bool
	TooSmall    bool
	InputRole   string
	MinPrefix   int
	PrefixOK    bool
}

const (
	awsMinIPv4Prefix = 28
	awsMaxIPv4Prefix = 16
	awsMinIPv6Prefix = 64
	awsMaxIPv6Prefix = 44
)

// AWS returns the addresses Amazon reserves in a VPC subnet, plus the
// remaining usable range. AWS IPv4 subnets must be /16–/28; IPv6 /44–/64.
func AWS(info Info) AWSReport {
	report := AWSReport{
		Usable: big.NewInt(0),
	}
	if info.Version == 4 {
		report.MinPrefix = awsMinIPv4Prefix
		report.PrefixOK = info.Bits >= awsMaxIPv4Prefix && info.Bits <= awsMinIPv4Prefix
	} else {
		report.MinPrefix = awsMinIPv6Prefix
		report.PrefixOK = info.Bits >= awsMaxIPv6Prefix && info.Bits <= awsMinIPv6Prefix && info.Bits%4 == 0
	}

	type role struct {
		delta  int64
		role   string
		detail string
	}
	roles := []role{
		{0, "Network", "Network address"},
		{1, "VPC router", "Reserved by AWS for the VPC router"},
		{2, "Amazon DNS", "Reserved by AWS for the Amazon DNS server"},
		{3, "Future use", "Reserved by AWS for future use"},
	}
	if info.Version == 6 {
		roles[2] = role{2, "Reserved", "Reserved by AWS"}
		roles[3] = role{3, "Reserved", "Reserved by AWS"}
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
	for _, r := range roles {
		add(offset(info.Network, r.delta), r.role, r.detail)
	}
	lastRole := "Broadcast"
	lastDetail := "Network broadcast address (unused in a VPC, still reserved)"
	if info.Version == 6 {
		lastRole = "Last address"
		lastDetail = "Reserved by AWS; IPv6 has no broadcast"
	}
	add(info.Last, lastRole, lastDetail)

	first := offset(info.Network, 4)
	last := offset(info.Last, -1)
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

// AWSRole returns the AWS reservation role for addr, or "" if it is usable.
func AWSRole(report AWSReport, addr netip.Addr) string {
	for _, r := range report.Reserved {
		if r.Addr == addr {
			return r.Role
		}
	}
	return ""
}
