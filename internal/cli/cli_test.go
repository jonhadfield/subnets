package cli

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRoot()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return ansi.ReplaceAllString(out.String()+errOut.String(), ""), err
}

func TestShowCIDR(t *testing.T) {
	out, err := run(t, "192.168.0.3/28")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"192.168.0.3/28",
		"192.168.0.0/28",
		"255.255.255.240",
		"0xFFFFFFF0",
		"192.168.0.1",
		"192.168.0.14",
		"192.168.0.15",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func TestShowAWS(t *testing.T) {
	out, err := run(t, "192.168.0.3/28", "--aws")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"AWS VPC reserved",
		"Reserved by AWS for the VPC router",
		"Amazon DNS",
		"192.168.0.4 – 192.168.0.14",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func TestShowJSON(t *testing.T) {
	out, err := run(t, "192.168.0.3/28", "--json", "--aws")
	if err != nil {
		t.Fatal(err)
	}
	var got showJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.CIDR != "192.168.0.0/28" {
		t.Fatalf("cidr: got %q", got.CIDR)
	}
	if got.Mask != "255.255.255.240" {
		t.Fatalf("mask: got %q", got.Mask)
	}
	if got.Broadcast != "192.168.0.15" {
		t.Fatalf("broadcast: got %q", got.Broadcast)
	}
	if got.HostIndex != "3" {
		t.Fatalf("host_index: got %q", got.HostIndex)
	}
	if got.AWS == nil || got.AWS.UsableRange == nil {
		t.Fatal("expected aws usable_range")
	}
	if got.AWS.UsableRange.First != "192.168.0.4" || got.AWS.UsableRange.Last != "192.168.0.14" {
		t.Fatalf("aws usable_range: %+v", got.AWS.UsableRange)
	}
}

func TestSheet(t *testing.T) {
	out, err := run(t, "sheet")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/24") || !strings.Contains(out, "255.255.255.0") {
		t.Fatalf("cheatsheet missing /24\n%s", out)
	}
}

func TestSheetIPv6(t *testing.T) {
	out, err := run(t, "sheet", "--ipv6")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"/64", "ffff:ffff:ffff:ffff::", "LAN / SLAAC"} {
		if !strings.Contains(out, want) {
			t.Fatalf("IPv6 cheatsheet missing %q\n%s", want, out)
		}
	}
}

func TestShowIPv6(t *testing.T) {
	out, err := run(t, "2001:db8::1/64")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"2001:db8::1/64",
		"2001:db8::/64",
		"IPv6",
		"Documentation",
		"2001:db8::ffff:ffff:ffff:ffff",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func TestHostsAWS(t *testing.T) {
	out, err := run(t, "hosts", "192.168.0.0/28", "--aws")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "VPC router") {
		t.Fatalf("hosts missing AWS roles\n%s", out)
	}
}

func TestHostsLimit(t *testing.T) {
	out, err := run(t, "hosts", "10.0.0.0/24", "--limit", "4")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "showing first 4 of 256 addresses") {
		t.Fatalf("missing truncation note\n%s", out)
	}
	if !strings.Contains(out, "raise --limit") {
		t.Fatalf("missing --limit hint\n%s", out)
	}
	if strings.Contains(out, "10.0.0.4") {
		t.Fatalf("printed past --limit\n%s", out)
	}
}

func TestHostsJSON(t *testing.T) {
	out, err := run(t, "hosts", "192.168.0.0/28", "--json", "--aws", "--limit", "5")
	if err != nil {
		t.Fatal(err)
	}
	var got hostsJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.CIDR != "192.168.0.0/28" {
		t.Fatalf("cidr: got %q", got.CIDR)
	}
	if !got.Truncated || len(got.Hosts) != 5 {
		t.Fatalf("truncated=%v hosts=%d", got.Truncated, len(got.Hosts))
	}
	if got.Hosts[1].Role != "VPC router" {
		t.Fatalf("index 1 role: got %q", got.Hosts[1].Role)
	}
	if got.Hosts[4].Role != "usable" {
		t.Fatalf("index 4 role: got %q", got.Hosts[4].Role)
	}
}

func TestSheetJSON(t *testing.T) {
	out, err := run(t, "sheet", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got sheetJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.Version != 4 || len(got.Rows) != 33 {
		t.Fatalf("version=%d rows=%d", got.Version, len(got.Rows))
	}
	if got.Rows[24].Mask != "255.255.255.0" || got.Rows[24].Note != "Class C" {
		t.Fatalf("/24 row: %+v", got.Rows[24])
	}
}

func TestShowJSONAWSIPv6BadStep(t *testing.T) {
	out, err := run(t, "2001:db8::/62", "--json", "--aws")
	if err != nil {
		t.Fatal(err)
	}
	var got showJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.AWS == nil || got.AWS.PrefixOK {
		t.Fatalf("expected prefix_ok false: %+v", got.AWS)
	}
	if !strings.Contains(got.AWS.Note, "step /4") {
		t.Fatalf("note: got %q", got.AWS.Note)
	}
}

func TestInvalid(t *testing.T) {
	_, err := run(t, "not-an-ip")
	if err == nil {
		t.Fatal("expected error")
	}
}
