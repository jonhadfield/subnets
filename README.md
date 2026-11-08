# subnets

IPv4/IPv6 subnet calculator: CIDR details, prefix cheatsheet, AWS VPC reserved addresses.

Released under the [MIT License](LICENSE).

## Install

### curl (Linux and macOS)

```sh
curl -sL https://raw.githubusercontent.com/jonhadfield/subnets/main/install | sh
```

Installs the latest release into `/usr/local/bin` after checking it against
`checksums.txt`. On macOS it also clears the Gatekeeper quarantine attribute.
Override the destination with `SUBNETS_INSTALL_DIR` on the `sh` end of the
pipe (not on `curl`):

```sh
curl -sL https://raw.githubusercontent.com/jonhadfield/subnets/main/install | SUBNETS_INSTALL_DIR=~/.local/bin sh
```

On macOS, Homebrew below is still the better route for upgrades. Or install with
the signed, stapled package:

```sh
curl -fLO https://github.com/jonhadfield/subnets/releases/latest/download/subnets_macos.pkg
sudo installer -pkg subnets_macos.pkg -target /
```

### Homebrew

```sh
brew install --cask jonhadfield/tap/subnets
```

### Go

```sh
go install github.com/jonhadfield/subnets/cmd/subnets@latest
```

Requires Go 1.27 or later.

## Usage

```bash
# CIDR / address details
subnets 192.168.0.3/28

# Same, with AWS VPC reservations
subnets 192.168.0.0/28 --aws

# Machine-readable JSON
subnets 192.168.0.3/28 --json
subnets 192.168.0.3/28 --json --aws
subnets sheet --json
subnets hosts 192.168.0.0/28 --json --aws --limit 32

# Prefix / mask cheatsheet
subnets sheet
subnets sheet --ipv6

# List addresses (AWS roles optional; --limit caps output, default 256)
subnets hosts 192.168.0.0/28 --aws
subnets hosts 10.0.0.0/24 --limit 32
```
