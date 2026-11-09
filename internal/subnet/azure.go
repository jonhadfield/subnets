package subnet

const (
	azureMinIPv4Prefix = 29
	azureMaxIPv4Prefix = 2
)

// Azure returns addresses Azure reserves in a VNet subnet (first four and last).
// IPv4 subnets are typically /2–/29.
func Azure(info Info) CloudReport {
	prefixOK := true
	minPrefix := azureMinIPv4Prefix
	note := ""
	if info.Version == 4 {
		prefixOK = info.Bits >= azureMaxIPv4Prefix && info.Bits <= azureMinIPv4Prefix
		if !prefixOK {
			note = "VNet IPv4 subnets must be between /2 and /29"
		}
	} else {
		// Azure IPv6 subnets are /64; reservation model differs — still show classic ends.
		minPrefix = 64
		prefixOK = info.Bits == 64
		if !prefixOK {
			note = "VNet IPv6 subnets are /64"
		}
	}

	specs := []reservedSpec{
		{0, false, "Network", "Network address"},
		{1, false, "Default gateway", "Reserved by Azure for the default gateway"},
		{2, false, "Azure DNS", "Reserved by Azure to map Azure DNS"},
		{3, false, "Azure DNS", "Reserved by Azure to map Azure DNS"},
	}
	if info.Version == 4 {
		specs = append(specs, reservedSpec{0, true, "Broadcast", "Network broadcast address"})
	} else {
		specs = append(specs, reservedSpec{0, true, "Last address", "Reserved by Azure"})
	}

	return buildCloud(info, "azure", minPrefix, 0, prefixOK, note, specs, 4, -1)
}
