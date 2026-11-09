package cli

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/jonhadfield/subnets/internal/subnet"
)

func writeShow(w io.Writer, info subnet.Info, cloud subnet.Cloud, useCloud bool) {
	var b strings.Builder
	b.WriteString(banner(info.Input))
	b.WriteString(addressBox(info))
	b.WriteString(subnetBox(info))
	b.WriteString(binaryBox(info))
	if useCloud {
		b.WriteString(cloudBox(info, cloud))
	}
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func writeSheet(w io.Writer, ipv6 bool) {
	var b strings.Builder
	if ipv6 {
		b.WriteString(banner("IPv6 prefix cheatsheet"))
		b.WriteString(sheetTable(subnet.IPv6Sheet(), true))
	} else {
		b.WriteString(banner("IPv4 prefix cheatsheet"))
		b.WriteString(sheetTable(subnet.IPv4Sheet(), false))
	}
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func writeHosts(w io.Writer, info subnet.Info, cloud subnet.Cloud, useCloud bool, limit int) {
	var b strings.Builder
	b.WriteString(banner(info.NetworkCIDR()))

	n, truncated := hostLimit(info, limit)

	var report subnet.CloudReport
	if useCloud {
		report = subnet.CloudFor(info, cloud)
	}

	rows := make([][]string, 0, n)
	for i := range n {
		addr, ok := info.AddressAt(uint64(i))
		if !ok {
			break
		}
		role, style := hostRole(info, addr, useCloud, report)
		mark := ""
		if addr == info.Addr {
			mark = "← you"
		}
		rows = append(rows, []string{fmt.Sprintf("%d", i), addr.String(), role, mark, style})
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(colBorder)).
		Headers("INDEX", "ADDRESS", "ROLE", " ").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			base := cellStyle
			if row%2 == 1 {
				base = oddCellStyle
			}
			if row >= 0 && row < len(rows) {
				switch rows[row][4] {
				case "ok":
					if col == 2 {
						return base.Foreground(colOk)
					}
				case "bad":
					if col == 2 {
						return base.Foreground(colBad)
					}
				case "warn":
					if col == 2 {
						return base.Foreground(colWarn)
					}
				}
				if rows[row][3] != "" && col == 1 {
					return base.Foreground(colInput).Bold(true)
				}
			}
			return base
		})
	for _, r := range rows {
		t.Row(r[0], r[1], r[2], r[3])
	}
	b.WriteString(t.String())
	if truncated {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("showing first %d of %s addresses (raise --limit for more)", n, subnet.FormatNumber(info.Total))))
	}
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func banner(input string) string {
	line := brandStyle.Render("subnets") + mutedStyle.Render("  ·  ") + inputStyle.Render(input)
	return "\n" + line + "\n"
}

func addressBox(info subnet.Info) string {
	rows := [][2]string{
		{"Address", info.Addr.String()},
		{"Version", fmt.Sprintf("IPv%d", info.Version)},
		{"Type", info.Kind.Name},
	}
	if class := info.LegacyClass(); class != "" {
		rows = append(rows, [2]string{"Legacy class", class})
	}
	return box("Address", rows)
}

func subnetBox(info subnet.Info) string {
	rows := [][2]string{
		{"CIDR", info.NetworkCIDR()},
		{"Prefix", fmt.Sprintf("/%d  (%d network bits, %d host bits)", info.Bits, info.Bits, info.HostBits)},
		{"Subnet mask", maskLine(info)},
		{"Wildcard", info.Wildcard.String()},
		{"Network", info.Network.String()},
	}
	if info.HasBroadcast {
		rows = append(rows, [2]string{"Broadcast", info.Broadcast().String()})
	} else {
		rows = append(rows, [2]string{"Last address", info.Last.String()})
	}
	rows = append(rows,
		[2]string{"First host", info.FirstHost.String()},
		[2]string{"Last host", info.LastHost.String()},
		[2]string{"Addresses", subnet.FormatNumber(info.Total)},
		[2]string{"Usable hosts", subnet.FormatNumber(info.Usable)},
	)
	if info.Prefix.Bits() < info.Addr.BitLen() {
		rows = append(rows, [2]string{
			"This host",
			fmt.Sprintf("index %s of %s", subnet.FormatNumber(info.HostIndex()), subnet.FormatNumber(info.Total)),
		})
	}
	return box("Subnet", rows)
}

func binaryBox(info subnet.Info) string {
	if info.Version != 4 {
		return ""
	}
	addrBin := subnet.Binary(info.Addr)
	maskBin := subnet.Binary(info.Mask)
	rows := [][2]string{
		{"Address", addrBin},
		{"Mask", maskBin},
	}
	return box("Binary", rows)
}

func cloudBox(info subnet.Info, cloud subnet.Cloud) string {
	report := subnet.CloudFor(info, cloud)
	title := cloud.Title() + " reserved"
	rows := make([][2]string, 0, len(report.Reserved)+4)
	for _, r := range report.Reserved {
		label := r.Addr.String()
		if r.Addr == info.Addr {
			label += "  ← you"
		}
		rows = append(rows, [2]string{label, r.Detail})
	}
	if report.HasUsable {
		rows = append(rows, [2]string{
			"Usable range",
			fmt.Sprintf("%s – %s  (%s hosts)", report.FirstUsable, report.LastUsable, subnet.FormatNumber(report.Usable)),
		})
	} else {
		rows = append(rows, [2]string{"Usable range", "none (subnet too small for " + cloud.Title() + ")"})
	}
	if !report.PrefixOK && report.Note != "" {
		rows = append(rows, [2]string{cloud.Title() + " note", report.Note})
	}
	if report.InputRole != "" {
		rows = append(rows, [2]string{"This address", "reserved by " + cloud.Title() + " (" + report.InputRole + ")"})
	}
	return box(title, rows)
}

func writeContains(w io.Writer, result subnet.ContainsResult) {
	var b strings.Builder
	b.WriteString(banner(result.Outer + " ∋ " + result.Inner))
	status := "no"
	if result.Contains {
		status = "yes"
	}
	rows := [][2]string{
		{"Outer", result.Outer},
		{"Inner", result.Inner},
		{"Contains", status},
		{"Relation", result.Relation},
	}
	b.WriteString(box("Contains", rows))
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func writeSplit(w io.Writer, input string, parts []subnet.Info) {
	var b strings.Builder
	b.WriteString(banner(input))
	rows := make([][]string, 0, len(parts))
	for i, p := range parts {
		rows = append(rows, []string{fmt.Sprintf("%d", i), p.NetworkCIDR(), subnet.FormatNumber(p.Total)})
	}
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(colBorder)).
		Headers("INDEX", "CIDR", "ADDRESSES").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row%2 == 0 {
				return cellStyle
			}
			return oddCellStyle
		})
	for _, r := range rows {
		t.Row(r[0], r[1], r[2])
	}
	b.WriteString(t.String())
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func writeSupernet(w io.Writer, inputs []string, parts []subnet.Info) {
	var b strings.Builder
	b.WriteString(banner(strings.Join(inputs, " + ")))
	rows := make([][2]string, 0, len(parts))
	for i, p := range parts {
		rows = append(rows, [2]string{fmt.Sprintf("%d", i), p.NetworkCIDR()})
	}
	b.WriteString(box("Supernet", rows))
	b.WriteByte('\n')
	lipgloss.Fprint(w, b.String())
}

func box(title string, rows [][2]string) string {
	var b strings.Builder
	keyWidth := 0
	for _, row := range rows {
		if n := len(row[0]); n > keyWidth {
			keyWidth = n
		}
	}
	for i, row := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(keyStyle.Render(padRight(row[0], keyWidth+2)))
		b.WriteString(valStyle.Render(row[1]))
	}
	inner := boxTitleStyle.Render(title) + "\n" + b.String()
	return boxStyle.Render(inner) + "\n"
}

func sheetTable(rows []subnet.SheetRow, ipv6 bool) string {
	maskHeader := "SUBNET MASK"
	if ipv6 {
		maskHeader = "MASK"
	}
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(colBorder)).
		Headers("PREFIX", maskHeader, "WILDCARD", "ADDRESSES", "USABLE", "NOTE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row%2 == 0 {
				return cellStyle
			}
			return oddCellStyle
		})
	for _, r := range rows {
		t.Row(
			fmt.Sprintf("/%d", r.Prefix),
			r.Mask,
			r.Wildcard,
			r.Addresses,
			r.Usable,
			r.Note,
		)
	}
	return t.String()
}

func maskLine(info subnet.Info) string {
	if info.Version == 4 {
		return info.Mask.String() + "  (" + subnet.Hex(info.Mask) + ")"
	}
	return info.Mask.String()
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func writeError(w io.Writer, err error) {
	lipgloss.Fprintln(w, errStyle.Render("error")+"  "+err.Error())
}
