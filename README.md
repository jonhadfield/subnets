# subnets

IPv4/IPv6 subnet calculator: CIDR details, prefix cheatsheet, AWS VPC reserved addresses.

## Install

```bash
go install github.com/jonhadfield/subnets/cmd/subnets@latest
```

Requires Go 1.27 or later.

## Usage

```bash
# CIDR / address details
subnets 192.168.0.3/28

# Same, with AWS VPC reservations
subnets 192.168.0.0/28 --aws

# Prefix / mask cheatsheet
subnets sheet
subnets sheet --ipv6

# List addresses (AWS roles optional)
subnets hosts 192.168.0.0/28 --aws
```
