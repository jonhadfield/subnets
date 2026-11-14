package subnet

const (
	gcpMinIPv4Prefix = 29
	gcpMaxIPv4Prefix = 8
)

// GCP returns addresses Google Cloud reserves in a subnet: the first two and
// the last two of the primary IPv4 range. Subnets are typically /8–/29.
func GCP(info Info) CloudReport {
	prefixOK := true
	minPrefix := gcpMinIPv4Prefix
	note := ""
	if info.Version == 4 {
		prefixOK = info.Bits >= gcpMaxIPv4Prefix && info.Bits <= gcpMinIPv4Prefix
		if !prefixOK {
			note = "VPC IPv4 subnets must be between /8 and /29"
		}
	} else {
		// GCP assigns a /64 IPv6 range per subnet (ULA or GUA).
		minPrefix = 64
		prefixOK = info.Bits == 64
		if !prefixOK {
			note = "VPC IPv6 subnets are /64"
		}
	}

	specs := []reservedSpec{
		{0, false, "Network", "Network address"},
		{1, false, "Gateway", "Reserved by Google Cloud for the default gateway"},
		{-1, true, "Google reserved", "Reserved by Google Cloud for future use"},
	}
	if info.Version == 4 {
		specs = append(specs, reservedSpec{0, true, "Broadcast", "Network broadcast address"})
	} else {
		specs = append(specs, reservedSpec{0, true, "Last address", "Reserved by Google Cloud"})
	}

	// Usable: .2 through second-to-last-1 → firstUsable +2, lastUsable -2
	return buildCloud(info, "gcp", minPrefix, 0, prefixOK, note, specs, 2, -2)
}
