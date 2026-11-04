package cli

import (
	"bytes"
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

func TestSheet(t *testing.T) {
	out, err := run(t, "sheet")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/24") || !strings.Contains(out, "255.255.255.0") {
		t.Fatalf("cheatsheet missing /24\n%s", out)
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

func TestInvalid(t *testing.T) {
	_, err := run(t, "not-an-ip")
	if err == nil {
		t.Fatal("expected error")
	}
}
