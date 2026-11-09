package subnet

const (
	awsMinIPv4Prefix = 28
	awsMaxIPv4Prefix = 16
	awsMinIPv6Prefix = 64
	awsMaxIPv6Prefix = 44
)

// AWS returns the addresses Amazon reserves in a VPC subnet, plus the
// remaining usable range. AWS IPv4 subnets must be /16–/28; IPv6 /44–/64.
func AWS(info Info) CloudReport {
	prefixOK := false
	minPrefix := awsMinIPv4Prefix
	note := ""
	if info.Version == 4 {
		prefixOK = info.Bits >= awsMaxIPv4Prefix && info.Bits <= awsMinIPv4Prefix
		if !prefixOK {
			note = "VPC IPv4 subnets must be between /16 and /28"
		}
	} else {
		minPrefix = awsMinIPv6Prefix
		prefixOK = info.Bits >= awsMaxIPv6Prefix && info.Bits <= awsMinIPv6Prefix && info.Bits%4 == 0
		if !prefixOK {
			note = "VPC IPv6 subnets must be between /44 and /64 (step /4)"
		}
	}

	specs := []reservedSpec{
		{0, false, "Network", "Network address"},
		{1, false, "VPC router", "Reserved by AWS for the VPC router"},
	}
	if info.Version == 4 {
		specs = append(specs,
			reservedSpec{2, false, "Amazon DNS", "Reserved by AWS for the Amazon DNS server"},
			reservedSpec{3, false, "Future use", "Reserved by AWS for future use"},
			reservedSpec{0, true, "Broadcast", "Network broadcast address (unused in a VPC, still reserved)"},
		)
	} else {
		specs = append(specs,
			reservedSpec{2, false, "Reserved", "Reserved by AWS"},
			reservedSpec{3, false, "Reserved", "Reserved by AWS"},
			reservedSpec{0, true, "Last address", "Reserved by AWS; IPv6 has no broadcast"},
		)
	}

	return buildCloud(info, "aws", minPrefix, 0, prefixOK, note, specs, 4, -1)
}
