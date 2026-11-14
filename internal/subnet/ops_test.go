package subnet

import (
	"net/netip"
	"testing"
)

func TestParseDottedMask(t *testing.T) {
	info, err := Parse("192.168.0.3/255.255.255.240")
	if err != nil {
		t.Fatal(err)
	}
	if info.Bits != 28 || info.NetworkCIDR() != "192.168.0.0/28" {
		t.Fatalf("got bits=%d cidr=%s", info.Bits, info.NetworkCIDR())
	}
}

func TestParseAddrAndMask(t *testing.T) {
	info, err := ParseAddrMask("10.0.0.5", "255.255.255.0")
	if err != nil {
		t.Fatal(err)
	}
	if info.Bits != 24 || info.Network != netip.MustParseAddr("10.0.0.0") {
		t.Fatalf("got bits=%d net=%s", info.Bits, info.Network)
	}
}

func TestParseMaskOnly(t *testing.T) {
	info, err := ParseMaskPrefix("255.255.255.240")
	if err != nil {
		t.Fatal(err)
	}
	if info.Bits != 28 || info.NetworkCIDR() != "0.0.0.0/28" {
		t.Fatalf("got bits=%d cidr=%s", info.Bits, info.NetworkCIDR())
	}
}

func TestParseBareMaskLikeAddressIsHost(t *testing.T) {
	info, err := Parse("128.0.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if info.Bits != 32 || info.NetworkCIDR() != "128.0.0.0/32" {
		t.Fatalf("got bits=%d cidr=%s (bare address must not become a mask)", info.Bits, info.NetworkCIDR())
	}
}

func TestParseNonContiguousMask(t *testing.T) {
	_, err := ParseMaskPrefix("255.255.0.255")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContainsSubnet(t *testing.T) {
	got, err := Contains("10.0.0.0/8", "10.1.2.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Contains || got.Relation != "subnet" {
		t.Fatalf("%+v", got)
	}
}

func TestContainsAddress(t *testing.T) {
	got, err := Contains("10.0.0.0/8", "10.1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Contains || got.Relation != "address" {
		t.Fatalf("%+v", got)
	}
}

func TestContainsDisjoint(t *testing.T) {
	got, err := Contains("10.0.0.0/8", "192.168.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	if got.Contains || got.Relation != "disjoint" {
		t.Fatalf("%+v", got)
	}
}

func TestContainsEqualUnmasked(t *testing.T) {
	got, err := Contains("10.0.0.0/24", "10.0.0.128/24")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Contains || got.Relation != "equal" {
		t.Fatalf("%+v", got)
	}
	if got.Outer != "10.0.0.0/24" || got.Inner != "10.0.0.0/24" {
		t.Fatalf("cidrs: %+v", got)
	}
}

func TestContainsSupernet(t *testing.T) {
	got, err := Contains("10.0.0.0/25", "10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if got.Contains || got.Relation != "supernet" {
		t.Fatalf("%+v", got)
	}
}

func TestContainsNestedNotOverlap(t *testing.T) {
	// Aligned CIDRs that share addresses are containment, not "overlap".
	got, err := Contains("10.0.0.0/24", "10.0.0.128/25")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Contains || got.Relation != "subnet" {
		t.Fatalf("%+v", got)
	}
}

func TestSplit(t *testing.T) {
	parts, err := Split("10.0.0.0/24", 26)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 4 {
		t.Fatalf("len=%d", len(parts))
	}
	want := []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"}
	for i, w := range want {
		if parts[i].NetworkCIDR() != w {
			t.Fatalf("[%d] got %s want %s", i, parts[i].NetworkCIDR(), w)
		}
	}
}

func TestSplitCount(t *testing.T) {
	parts, err := SplitCount("10.0.0.0/24", 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 4 || parts[0].Bits != 26 {
		t.Fatalf("len=%d bits=%d", len(parts), parts[0].Bits)
	}
}

func TestAggregate(t *testing.T) {
	got, err := Aggregate([]string{"10.0.0.0/25", "10.0.0.128/25"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].NetworkCIDR() != "10.0.0.0/24" {
		t.Fatalf("%v", got)
	}
}

func TestAggregatePartial(t *testing.T) {
	got, err := Aggregate([]string{"10.0.0.0/25", "10.0.1.0/25"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 leftovers, got %d", len(got))
	}
}

func TestAzureReserved(t *testing.T) {
	info, err := Parse("10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	r := Azure(info)
	if len(r.Reserved) != 5 {
		t.Fatalf("reserved=%d", len(r.Reserved))
	}
	if r.Reserved[1].Role != "Default gateway" {
		t.Fatalf("role=%s", r.Reserved[1].Role)
	}
	if !r.HasUsable || r.FirstUsable.String() != "10.0.0.4" || r.LastUsable.String() != "10.0.0.254" {
		t.Fatalf("usable %s-%s", r.FirstUsable, r.LastUsable)
	}
}

func TestGCPReserved(t *testing.T) {
	info, err := Parse("10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	r := GCP(info)
	if len(r.Reserved) != 4 {
		t.Fatalf("reserved=%d %+v", len(r.Reserved), r.Reserved)
	}
	if r.Reserved[2].Addr.String() != "10.0.0.254" {
		t.Fatalf("google reserved=%s", r.Reserved[2].Addr)
	}
	if !r.HasUsable || r.FirstUsable.String() != "10.0.0.2" || r.LastUsable.String() != "10.0.0.253" {
		t.Fatalf("usable %s-%s", r.FirstUsable, r.LastUsable)
	}
}

func TestAzureIPv6Prefix(t *testing.T) {
	ok, err := Parse("2001:db8::/64")
	if err != nil {
		t.Fatal(err)
	}
	if !Azure(ok).PrefixOK {
		t.Fatal("expected /64 to be a valid Azure IPv6 prefix")
	}
	bad, err := Parse("2001:db8::/56")
	if err != nil {
		t.Fatal(err)
	}
	report := Azure(bad)
	if report.PrefixOK {
		t.Fatal("/56 should fail Azure IPv6 /64 check")
	}
	if report.Note == "" {
		t.Fatal("expected Azure IPv6 note")
	}
}

func TestGCPIPv6Prefix(t *testing.T) {
	ok, err := Parse("2001:db8::/64")
	if err != nil {
		t.Fatal(err)
	}
	if !GCP(ok).PrefixOK {
		t.Fatal("expected /64 to be a valid GCP IPv6 prefix")
	}
	for _, cidr := range []string{"2001:db8::/56", "2001:db8::/128"} {
		info, err := Parse(cidr)
		if err != nil {
			t.Fatalf("%s: %v", cidr, err)
		}
		report := GCP(info)
		if report.PrefixOK {
			t.Fatalf("%s should fail GCP IPv6 /64 check", cidr)
		}
		if report.Note == "" {
			t.Fatalf("%s: expected GCP IPv6 note", cidr)
		}
	}
}
