package cli

import (
	"encoding/json"
	"io"
	"net/netip"

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

type hostsJSON struct {
	CIDR      string        `json:"cidr"`
	Addresses string        `json:"addresses"`
	Limit     int           `json:"limit"`
	Truncated bool          `json:"truncated"`
	Hosts     []hostRowJSON `json:"hosts"`
}

type hostRowJSON struct {
	Index   int    `json:"index"`
	Address string `json:"address"`
	Role    string `json:"role"`
	You     bool   `json:"you,omitempty"`
}

type sheetJSON struct {
	Version int            `json:"version"`
	Rows    []sheetRowJSON `json:"rows"`
}

type sheetRowJSON struct {
	Prefix    int    `json:"prefix"`
	Mask      string `json:"mask"`
	Wildcard  string `json:"wildcard"`
	Addresses string `json:"addresses"`
	Usable    string `json:"usable"`
	Note      string `json:"note,omitempty"`
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
	return encodeJSON(w, out)
}

func writeHostsJSON(w io.Writer, info subnet.Info, aws bool, limit int) error {
	n, truncated := hostLimit(info, limit)
	var report subnet.AWSReport
	if aws {
		report = subnet.AWS(info)
	}
	out := hostsJSON{
		CIDR:      info.NetworkCIDR(),
		Addresses: info.Total.String(),
		Limit:     limit,
		Truncated: truncated,
		Hosts:     make([]hostRowJSON, 0, n),
	}
	for i := range n {
		addr, ok := info.AddressAt(uint64(i))
		if !ok {
			break
		}
		role, _ := hostRole(info, addr, aws, report)
		out.Hosts = append(out.Hosts, hostRowJSON{
			Index:   i,
			Address: addr.String(),
			Role:    role,
			You:     addr == info.Addr,
		})
	}
	return encodeJSON(w, out)
}

func writeSheetJSON(w io.Writer, ipv6 bool) error {
	rows := subnet.IPv4Sheet()
	version := 4
	if ipv6 {
		rows = subnet.IPv6Sheet()
		version = 6
	}
	out := sheetJSON{
		Version: version,
		Rows:    make([]sheetRowJSON, 0, len(rows)),
	}
	for _, r := range rows {
		out.Rows = append(out.Rows, sheetRowJSON{
			Prefix:    r.Prefix,
			Mask:      r.Mask,
			Wildcard:  r.Wildcard,
			Addresses: r.Addresses,
			Usable:    r.Usable,
			Note:      r.Note,
		})
	}
	return encodeJSON(w, out)
}

func encodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func hostLimit(info subnet.Info, limit int) (n int, truncated bool) {
	n = limit
	truncated = true
	if info.Total.IsInt64() {
		size := int(info.Total.Int64())
		if size <= limit {
			return size, false
		}
	}
	return n, truncated
}

func hostRole(info subnet.Info, addr netip.Addr, aws bool, report subnet.AWSReport) (role, style string) {
	role = "usable"
	style = "ok"
	if aws {
		if r := subnet.AWSRole(report, addr); r != "" {
			return r, "bad"
		}
		return role, style
	}
	if info.HasBroadcast && (addr == info.Network || addr == info.Last) {
		if addr == info.Network {
			return "network", "warn"
		}
		return "broadcast", "warn"
	}
	if addr == info.Network && info.Version == 6 && info.HostBits > 1 {
		return "subnet-router anycast", "warn"
	}
	return role, style
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
