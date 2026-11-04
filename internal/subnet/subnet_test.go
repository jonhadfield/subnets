package subnet

import (
	"math/big"
	"net/netip"
	"testing"
)

func TestParseIPv4Subnet(t *testing.T) {
	info, err := Parse("192.168.0.3/28")
	if err != nil {
		t.Fatal(err)
	}
	if info.Network != netip.MustParseAddr("192.168.0.0") {
		t.Fatalf("network: got %s", info.Network)
	}
	if info.Last != netip.MustParseAddr("192.168.0.15") {
		t.Fatalf("last: got %s", info.Last)
	}
	if info.FirstHost != netip.MustParseAddr("192.168.0.1") {
		t.Fatalf("first host: got %s", info.FirstHost)
	}
	if info.LastHost != netip.MustParseAddr("192.168.0.14") {
		t.Fatalf("last host: got %s", info.LastHost)
	}
	if info.Mask != netip.MustParseAddr("255.255.255.240") {
		t.Fatalf("mask: got %s", info.Mask)
	}
	if Hex(info.Mask) != "0xFFFFFFF0" {
		t.Fatalf("mask hex: got %s", Hex(info.Mask))
	}
	if info.Wildcard != netip.MustParseAddr("0.0.0.15") {
		t.Fatalf("wildcard: got %s", info.Wildcard)
	}
	if info.Total.Cmp(big.NewInt(16)) != 0 {
		t.Fatalf("total: got %s", info.Total)
	}
	if info.Usable.Cmp(big.NewInt(14)) != 0 {
		t.Fatalf("usable: got %s", info.Usable)
	}
	if info.HostIndex().Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("host index: got %s", info.HostIndex())
	}
	if info.Kind.Name != "Private" {
		t.Fatalf("kind: got %s", info.Kind.Name)
	}
	if info.LegacyClass() != "C" {
		t.Fatalf("class: got %s", info.LegacyClass())
	}
	if info.NetworkCIDR() != "192.168.0.0/28" {
		t.Fatalf("network cidr: got %s", info.NetworkCIDR())
	}
}

func TestParseBareAddress(t *testing.T) {
	info, err := Parse("10.1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if info.Bits != 32 {
		t.Fatalf("bits: got %d", info.Bits)
	}
	if info.Network != info.Addr {
		t.Fatalf("network should equal host for /32")
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("not-an-ip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSlash31And32(t *testing.T) {
	p31, err := Parse("10.0.0.0/31")
	if err != nil {
		t.Fatal(err)
	}
	if p31.Usable.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("/31 usable: got %s", p31.Usable)
	}
	if p31.HasBroadcast {
		t.Fatal("/31 should not have a broadcast")
	}

	p32, err := Parse("10.0.0.1/32")
	if err != nil {
		t.Fatal(err)
	}
	if p32.Usable.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("/32 usable: got %s", p32.Usable)
	}
}

func TestAWSReservedIPv4(t *testing.T) {
	info, err := Parse("192.168.0.3/28")
	if err != nil {
		t.Fatal(err)
	}
	report := AWS(info)
	if !report.PrefixOK {
		t.Fatal("expected /28 to be a valid AWS prefix")
	}
	if report.InputRole != "Future use" {
		t.Fatalf("input role: got %q", report.InputRole)
	}
	if !report.HasUsable {
		t.Fatal("expected usable AWS range")
	}
	if report.FirstUsable != netip.MustParseAddr("192.168.0.4") {
		t.Fatalf("first usable: got %s", report.FirstUsable)
	}
	if report.LastUsable != netip.MustParseAddr("192.168.0.14") {
		t.Fatalf("last usable: got %s", report.LastUsable)
	}
	if report.Usable.Cmp(big.NewInt(11)) != 0 {
		t.Fatalf("aws usable: got %s", report.Usable)
	}
	want := []string{
		"192.168.0.0",
		"192.168.0.1",
		"192.168.0.2",
		"192.168.0.3",
		"192.168.0.15",
	}
	if len(report.Reserved) != len(want) {
		t.Fatalf("reserved count: got %d", len(report.Reserved))
	}
	for i, addr := range want {
		if report.Reserved[i].Addr.String() != addr {
			t.Fatalf("reserved[%d]: got %s want %s", i, report.Reserved[i].Addr, addr)
		}
	}
}

func TestAWSTooSmall(t *testing.T) {
	info, err := Parse("10.0.0.0/30")
	if err != nil {
		t.Fatal(err)
	}
	report := AWS(info)
	if report.PrefixOK {
		t.Fatal("/30 is smaller than AWS allows")
	}
	if !report.TooSmall {
		t.Fatal("expected too-small AWS usable range")
	}
}

func TestIPv6(t *testing.T) {
	info, err := Parse("2001:db8::1/64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != 6 {
		t.Fatalf("version: got %d", info.Version)
	}
	if info.Network != netip.MustParseAddr("2001:db8::") {
		t.Fatalf("network: got %s", info.Network)
	}
	if info.Last != netip.MustParseAddr("2001:db8::ffff:ffff:ffff:ffff") {
		t.Fatalf("last: got %s", info.Last)
	}
	if info.HasBroadcast {
		t.Fatal("IPv6 should not have broadcast")
	}
	if info.Kind.Name != "Documentation" {
		t.Fatalf("kind: got %s", info.Kind.Name)
	}
	report := AWS(info)
	if !report.PrefixOK {
		t.Fatal("expected /64 to be a valid AWS IPv6 prefix")
	}
	if report.Reserved[0].Addr != info.Network {
		t.Fatalf("first reserved: got %s", report.Reserved[0].Addr)
	}
	if report.Reserved[len(report.Reserved)-1].Addr != info.Last {
		t.Fatalf("last reserved: got %s", report.Reserved[len(report.Reserved)-1].Addr)
	}
}

func TestIPv4Sheet(t *testing.T) {
	rows := IPv4Sheet()
	if len(rows) != 33 {
		t.Fatalf("rows: got %d", len(rows))
	}
	row := rows[28]
	if row.Mask != "255.255.255.240" {
		t.Fatalf("/28 mask: got %s", row.Mask)
	}
	if row.Usable != "14" {
		t.Fatalf("/28 usable: got %s", row.Usable)
	}
	if rows[24].Note != "Class C" {
		t.Fatalf("/24 note: got %s", rows[24].Note)
	}
}

func TestFormatNumber(t *testing.T) {
	if got := FormatNumber(big.NewInt(16)); got != "16" {
		t.Fatalf("got %s", got)
	}
	if got := FormatNumber(big.NewInt(4294967296)); got != "4,294,967,296" {
		t.Fatalf("got %s", got)
	}
}

func TestAddressAt(t *testing.T) {
	info, err := Parse("192.168.0.0/28")
	if err != nil {
		t.Fatal(err)
	}
	addr, ok := info.AddressAt(3)
	if !ok || addr != netip.MustParseAddr("192.168.0.3") {
		t.Fatalf("got %s ok=%v", addr, ok)
	}
	if _, ok := info.AddressAt(16); ok {
		t.Fatal("index 16 should be out of range")
	}
}
