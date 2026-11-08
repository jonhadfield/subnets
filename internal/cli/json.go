package cli

import (
	"encoding/json"
	"io"

	"github.com/jonhadfield/subnets/internal/subnet"
)

type showJSON struct {
	Input       string      `json:"input"`
	Address     string      `json:"address"`
	Version     int         `json:"version"`
	Type        string      `json:"type"`
	LegacyClass string      `json:"legacy_class,omitempty"`
	CIDR        string      `json:"cidr"`
	Prefix      int         `json:"prefix"`
	HostBits    int         `json:"host_bits"`
	Mask        string      `json:"mask"`
	MaskHex     string      `json:"mask_hex,omitempty"`
	Wildcard    string      `json:"wildcard"`
	Network     string      `json:"network"`
	Broadcast   string      `json:"broadcast,omitempty"`
	LastAddress string      `json:"last_address"`
	FirstHost   string      `json:"first_host"`
	LastHost    string      `json:"last_host"`
	Addresses   string      `json:"addresses"`
	UsableHosts string      `json:"usable_hosts"`
	HostIndex   string      `json:"host_index,omitempty"`
	Binary      *binaryJSON `json:"binary,omitempty"`
	AWS         *awsJSON    `json:"aws,omitempty"`
}

type binaryJSON struct {
	Address string `json:"address"`
	Mask    string `json:"mask"`
}

type awsJSON struct {
	Reserved    []awsReservedJSON `json:"reserved"`
	UsableRange *awsRangeJSON     `json:"usable_range,omitempty"`
	TooSmall    bool              `json:"too_small,omitempty"`
	PrefixOK    bool              `json:"prefix_ok"`
	MinPrefix   int               `json:"min_prefix"`
	InputRole   string            `json:"input_role,omitempty"`
	Note        string            `json:"note,omitempty"`
}

type awsReservedJSON struct {
	Address string `json:"address"`
	Role    string `json:"role"`
	Detail  string `json:"detail"`
}

type awsRangeJSON struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Hosts string `json:"hosts"`
}

func writeShowJSON(w io.Writer, info subnet.Info, aws bool) error {
	out := showJSON{
		Input:       info.Input,
		Address:     info.Addr.String(),
		Version:     info.Version,
		Type:        info.Kind.Name,
		LegacyClass: info.LegacyClass(),
		CIDR:        info.NetworkCIDR(),
		Prefix:      info.Bits,
		HostBits:    info.HostBits,
		Mask:        info.Mask.String(),
		Wildcard:    info.Wildcard.String(),
		Network:     info.Network.String(),
		LastAddress: info.Last.String(),
		FirstHost:   info.FirstHost.String(),
		LastHost:    info.LastHost.String(),
		Addresses:   info.Total.String(),
		UsableHosts: info.Usable.String(),
	}
	if info.Version == 4 {
		out.MaskHex = subnet.Hex(info.Mask)
		out.Binary = &binaryJSON{
			Address: subnet.Binary(info.Addr),
			Mask:    subnet.Binary(info.Mask),
		}
	}
	if info.HasBroadcast {
		out.Broadcast = info.Broadcast().String()
	}
	if info.Prefix.Bits() < info.Addr.BitLen() {
		out.HostIndex = info.HostIndex().String()
	}
	if aws {
		out.AWS = awsJSONFrom(info)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func awsJSONFrom(info subnet.Info) *awsJSON {
	report := subnet.AWS(info)
	out := &awsJSON{
		Reserved:  make([]awsReservedJSON, 0, len(report.Reserved)),
		TooSmall:  report.TooSmall,
		PrefixOK:  report.PrefixOK,
		MinPrefix: report.MinPrefix,
		InputRole: report.InputRole,
	}
	for _, r := range report.Reserved {
		out.Reserved = append(out.Reserved, awsReservedJSON{
			Address: r.Addr.String(),
			Role:    r.Role,
			Detail:  r.Detail,
		})
	}
	if report.HasUsable {
		out.UsableRange = &awsRangeJSON{
			First: report.FirstUsable.String(),
			Last:  report.LastUsable.String(),
			Hosts: report.Usable.String(),
		}
	}
	if !report.PrefixOK {
		if info.Version == 4 {
			out.Note = "VPC IPv4 subnets must be between /16 and /28"
		} else {
			out.Note = "VPC IPv6 subnets must be between /44 and /64 (step /4)"
		}
	}
	return out
}
