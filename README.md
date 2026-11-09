# subnets

IPv4 and IPv6 subnet calculator for the terminal — CIDR details, a prefix
cheatsheet, host listings, and the addresses Amazon reserves inside a VPC.

[![ci](https://github.com/jonhadfield/subnets/actions/workflows/ci.yml/badge.svg)](https://github.com/jonhadfield/subnets/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/jonhadfield/subnets)](https://github.com/jonhadfield/subnets/releases/latest)
[![license](https://img.shields.io/github/license/jonhadfield/subnets)](LICENSE)

```
subnets  ·  192.168.0.3/28

╭───────────────────────────╮
│ Address                   │
│ Address       192.168.0.3 │
│ Version       IPv4        │
│ Type          Private     │
│ Legacy class  C           │
╰───────────────────────────╯

╭───────────────────────────────────────────────────╮
│ Subnet                                            │
│ CIDR          192.168.0.0/28                      │
│ Prefix        /28  (28 network bits, 4 host bits) │
│ Subnet mask   255.255.255.240  (0xFFFFFFF0)       │
│ Wildcard      0.0.0.15                            │
│ Network       192.168.0.0                         │
│ Broadcast     192.168.0.15                        │
│ First host    192.168.0.1                         │
│ Last host     192.168.0.14                        │
│ Addresses     16                                  │
│ Usable hosts  14                                  │
│ This host     index 3 of 16                       │
╰───────────────────────────────────────────────────╯
```

**Contents:** [install](#install) · [usage](#usage) · [show](#show) · [sheet](#sheet) ·
[hosts](#hosts) · [aws](#aws) · [json](#json) · [development](#development) ·
[license](#license)

## install

### Homebrew

```sh
brew install --cask jonhadfield/tap/subnets
```

Preferred on macOS: upgrades come through Homebrew, and the cask clears the
Gatekeeper quarantine attribute after install.

### curl (Linux and macOS)

```sh
curl -sL https://raw.githubusercontent.com/jonhadfield/subnets/main/install | sh
```

Resolves the latest release, downloads the archive for the machine it is run on,
checks it against `checksums.txt`, and installs to `/usr/local/bin`. The
directory is created if needed, and `sudo` is used only when the destination is
not writable. A download that does not match its checksum is refused. On macOS
it also clears the Gatekeeper quarantine attribute.

Optional variables (set on the `sh` end of the pipe, not on `curl`):

| variable | meaning |
| --- | --- |
| `SUBNETS_VERSION` | a tag to install, e.g. `v0.1.2`. Default: the latest release |
| `SUBNETS_INSTALL_DIR` | where to put the binary. Default: `/usr/local/bin` |
| `GITHUB_URL` | for a mirror or an enterprise host |

```sh
curl -sL https://raw.githubusercontent.com/jonhadfield/subnets/main/install | SUBNETS_INSTALL_DIR=~/.local/bin sh
```

amd64 and arm64 are built for Linux and macOS; anything else is refused with a
message rather than a failed download. On macOS, Homebrew above is still the
better route for upgrades.

### macOS package

Prefer Homebrew above. Some releases also attach `subnets_macos.pkg` — a signed,
notarized, and stapled installer. A notarization ticket cannot be stapled into a
bare Mach-O, so Gatekeeper can refuse a tarball download even after Apple
notarized it; the `.pkg` carries the ticket, and files it installs are never
quarantined.

When that asset is listed on the
[releases page](https://github.com/jonhadfield/subnets/releases), download it
and install with:

```sh
sudo installer -pkg subnets_macos.pkg -target /
```

### Go

```sh
go install github.com/jonhadfield/subnets/cmd/subnets@latest
```

Requires Go 1.27 or later.

### download

Binaries for macOS, Linux, and Windows are on the
[releases page](https://github.com/jonhadfield/subnets/releases):

| archive | platform |
| --- | --- |
| `subnets_darwin_arm64.tar.gz` | macOS Apple silicon |
| `subnets_darwin_amd64.tar.gz` | macOS Intel |
| `subnets_linux_amd64.tar.gz` | Linux x86_64 |
| `subnets_linux_arm64.tar.gz` | Linux arm64 |
| `subnets_windows_amd64.zip` | Windows x86_64 |

Each release also publishes `checksums.txt`. On Windows, unzip and put
`subnets.exe` somewhere on your `PATH`.

## usage

```text
subnets [cidr] [flags]
subnets sheet [flags]
subnets hosts <cidr> [flags]
```

Pass an address or a CIDR. A bare IPv4 address is treated as `/32`; a bare IPv6
address as `/128`.

| command | what it does |
| --- | --- |
| `subnets <addr\|cidr>` | subnet details for one address or prefix |
| `subnets sheet` | prefix length / subnet mask cheatsheet |
| `subnets hosts <cidr>` | list every address in a subnet |

| flag | where | meaning |
| --- | --- | --- |
| `--aws` | show, hosts | mark addresses reserved by AWS VPC |
| `--json` | show, sheet, hosts | emit machine-readable JSON |
| `--ipv6` | sheet | show the IPv6 prefix cheatsheet |
| `--limit N` | hosts | maximum addresses to print (default `256`) |
| `-v`, `--version` | root | print the build version |
| `-h`, `--help` | any | help for that command |

```sh
subnets 192.168.0.3/28
subnets 192.168.0.0/28 --aws
subnets 2001:db8::/64
subnets sheet
subnets sheet --ipv6
subnets hosts 192.168.0.0/28 --aws
subnets hosts 10.0.0.0/24 --limit 32
subnets 192.168.0.3/28 --json
```

## show

```sh
subnets 192.168.0.3/28
subnets 10.0.0.5
subnets 2001:db8::/64
```

Prints three panels by default:

- **Address** — the input, IP version, classification (Private, Public,
  Loopback, Link-local, Multicast, Shared/CGNAT, Documentation, Unique local,
  …), and the legacy IPv4 class when relevant
- **Subnet** — CIDR, prefix length, mask (with hex), wildcard, network,
  broadcast (IPv4) or last address (IPv6), first/last host, total and usable
  counts, and the zero-based index of the input address inside the subnet
- **Binary** — dotted/colonic binary of the address and mask (IPv4)

Add `--aws` for a fourth panel of VPC reservations; see [aws](#aws).

## sheet

```sh
subnets sheet
subnets sheet --ipv6
```

A quick reference of prefix lengths to subnet masks, wildcards, address counts,
and usable hosts. IPv4 covers `/0`–`/32` with notes for classful sizes,
point-to-point (`/30`, `/31` RFC 3021), and single host (`/32`). IPv6 covers the
common allocation sizes (`/32`, `/48`, `/56`, `/64`, …) through `/128`.

## hosts

```sh
subnets hosts 192.168.0.0/28
subnets hosts 192.168.0.0/28 --aws
subnets hosts 10.0.0.0/24 --limit 32
```

Lists every address in the subnet as a table: index, address, role, and a
`← you` mark when the input address is in range.

Without `--aws`, roles are the classic network/broadcast (IPv4) or
subnet-router anycast (IPv6), otherwise `usable`. With `--aws`, roles follow the
VPC reservation table below.

`--limit` caps how many rows are printed (default `256`). Larger subnets are
truncated rather than dumping millions of lines; the JSON form reports
`"truncated": true` when that happens.

## aws

```sh
subnets 192.168.0.0/28 --aws
subnets hosts 192.168.0.0/28 --aws
```

Amazon keeps five addresses in every VPC subnet. `--aws` surfaces them on show
output and tags them on `hosts`:

| offset | role | detail |
| --- | --- | --- |
| `.0` / first | Network | network address |
| `.1` / +1 | VPC router | reserved for the VPC router |
| `.2` / +2 | Amazon DNS | reserved for Amazon DNS (IPv4); reserved on IPv6 |
| `.3` / +3 | Future use | reserved for future use (IPv4); reserved on IPv6 |
| last | Broadcast / Last address | broadcast on IPv4 (still reserved in a VPC); last address on IPv6 |

The usable range starts at offset `+4` and runs through the address before the
last; the show panel reports it along with whether the input itself is reserved.

Prefix constraints are checked too: VPC IPv4 subnets must be `/16`–`/28`; VPC
IPv6 subnets `/44`–`/64` in steps of `/4`. Out-of-range prefixes are flagged
rather than silently inventing a usable range.

## json

```sh
subnets 192.168.0.3/28 --json
subnets 192.168.0.0/28 --json --aws
subnets sheet --json
subnets sheet --ipv6 --json
subnets hosts 192.168.0.0/28 --json --aws --limit 32
```

`--json` replaces the coloured tables with indented JSON on stdout — suitable
for scripts, `jq`, and CI. Numbers that can exceed a 64-bit integer (IPv6 host
counts) are emitted as strings.

Show output includes `input`, `address`, `version`, `type`, `cidr`, `prefix`,
`mask`, `wildcard`, `network`, `broadcast` / `last_address`, host range, counts,
`host_index`, and (for IPv4) `mask_hex` plus a `binary` object. With `--aws`, an
`aws` object adds `reserved`, `usable_range`, `prefix_ok`, and `input_role`.

Sheet output is `{ "version": 4|6, "rows": [ … ] }`. Hosts output is
`{ "cidr", "addresses", "limit", "truncated", "hosts": [ … ] }`.

## development

```sh
make build    # writes ./subnets with version from git describe
make test
make install  # go install with the same ldflags
```

## license

Released under the [MIT License](LICENSE).
