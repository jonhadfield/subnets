package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/jonhadfield/subnets/internal/subnet"
	"github.com/spf13/cobra"
)

// ErrNotContained is returned by contains when outer does not contain inner.
// Execute exits 1 without printing an error banner (the result was already shown).
var ErrNotContained = errors.New("not contained")

// Version is the build version. Override at link time:
//
//	-ldflags "-X github.com/jonhadfield/subnets/internal/cli.Version=1.2.3"
var Version = "dev"

type options struct {
	aws    bool
	azure  bool
	gcp    bool
	ipv6   bool
	json   bool
	limit  int
	prefix int
	count  int
}

func (o *options) cloud() (subnet.Cloud, bool, error) {
	n := 0
	var c subnet.Cloud
	if o.aws {
		n++
		c = subnet.CloudAWS
	}
	if o.azure {
		n++
		c = subnet.CloudAzure
	}
	if o.gcp {
		n++
		c = subnet.CloudGCP
	}
	if n > 1 {
		return "", false, fmt.Errorf("--aws, --azure, and --gcp are mutually exclusive")
	}
	return c, n == 1, nil
}

func NewRoot() *cobra.Command {
	opts := &options{limit: 256}

	root := &cobra.Command{
		Use:           "subnets [addr|cidr] [mask]",
		Short:         "IP subnet calculator and cheatsheet",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(2),
		Long: `subnets shows subnet details for an address or CIDR.

Examples:
  subnets 192.168.0.3/28
  subnets 192.168.0.3 255.255.255.240
  subnets 192.168.0.0/28 --aws
  subnets 10.0.0.0/24 --azure
  subnets 10.0.0.0/24 --gcp
  subnets sheet
  subnets hosts 192.168.0.0/28 --aws
  subnets hosts 10.0.0.0/24 --azure
  subnets hosts 10.0.0.0/24 --gcp
  subnets contains 10.0.0.0/8 10.1.2.3
  subnets split 10.0.0.0/24 --prefix 26
  subnets supernet 10.0.0.0/25 10.0.0.128/25`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			info, err := parseInput(args)
			if err != nil {
				return err
			}
			cloud, useCloud, err := opts.cloud()
			if err != nil {
				return err
			}
			if opts.json {
				return writeShowJSON(cmd.OutOrStdout(), info, cloud, useCloud)
			}
			writeShow(cmd.OutOrStdout(), info, cloud, useCloud)
			return nil
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	addCloudFlags(root, opts)
	root.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")

	root.AddCommand(
		newSheetCmd(opts),
		newHostsCmd(opts),
		newContainsCmd(opts),
		newSplitCmd(opts),
		newSupernetCmd(opts),
	)
	return root
}

func addCloudFlags(cmd *cobra.Command, opts *options) {
	cmd.Flags().BoolVar(&opts.aws, "aws", false, "show addresses reserved by AWS VPC")
	cmd.Flags().BoolVar(&opts.azure, "azure", false, "show addresses reserved by Azure VNet")
	cmd.Flags().BoolVar(&opts.gcp, "gcp", false, "show addresses reserved by Google Cloud VPC")
}

func parseInput(args []string) (subnet.Info, error) {
	if len(args) == 2 {
		return subnet.ParseAddrMask(args[0], args[1])
	}
	return subnet.Parse(args[0])
}

func newSheetCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheet",
		Short: "Prefix length and subnet mask cheatsheet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.json {
				return writeSheetJSON(cmd.OutOrStdout(), opts.ipv6)
			}
			writeSheet(cmd.OutOrStdout(), opts.ipv6)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.ipv6, "ipv6", false, "show the IPv6 prefix cheatsheet")
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	return cmd
}

func newHostsCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hosts <cidr>",
		Short: "List every address in a subnet",
		Example: `  subnets hosts 192.168.0.0/28
  subnets hosts 192.168.0.0/28 --aws
  subnets hosts 10.0.0.0/24 --azure
  subnets hosts 10.0.0.0/24 --gcp
  subnets hosts 10.0.0.0/24 --limit 32`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := parseInput(args)
			if err != nil {
				return err
			}
			if opts.limit < 1 {
				return fmt.Errorf("limit must be at least 1")
			}
			cloud, useCloud, err := opts.cloud()
			if err != nil {
				return err
			}
			if opts.json {
				return writeHostsJSON(cmd.OutOrStdout(), info, cloud, useCloud, opts.limit)
			}
			writeHosts(cmd.OutOrStdout(), info, cloud, useCloud, opts.limit)
			return nil
		},
	}
	addCloudFlags(cmd, opts)
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&opts.limit, "limit", 256, "maximum addresses to print")
	return cmd
}

func newContainsCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contains <outer> <inner>",
		Short: "Check whether a prefix contains an address or prefix",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := subnet.Contains(args[0], args[1])
			if err != nil {
				return err
			}
			if opts.json {
				if err := writeContainsJSON(cmd.OutOrStdout(), result); err != nil {
					return err
				}
			} else {
				writeContains(cmd.OutOrStdout(), result)
			}
			if !result.Contains {
				return ErrNotContained
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	return cmd
}

func newSplitCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split <cidr>",
		Short: "Split a network into equal subnets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				parts []subnet.Info
				err   error
			)
			switch {
			case opts.prefix > 0 && opts.count > 0:
				return fmt.Errorf("use either --prefix or --count, not both")
			case opts.prefix > 0:
				parts, err = subnet.Split(args[0], opts.prefix)
			case opts.count > 0:
				parts, err = subnet.SplitCount(args[0], opts.count)
			default:
				return fmt.Errorf("required flag: --prefix or --count")
			}
			if err != nil {
				return err
			}
			if opts.json {
				return writeSplitJSON(cmd.OutOrStdout(), args[0], parts)
			}
			writeSplit(cmd.OutOrStdout(), args[0], parts)
			return nil
		},
	}
	cmd.Flags().IntVar(&opts.prefix, "prefix", 0, "new prefix length for each subnet")
	cmd.Flags().IntVar(&opts.count, "count", 0, "number of equal subnets (power of two)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	return cmd
}

func newSupernetCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supernet <cidr> [<cidr>...]",
		Short: "Aggregate prefixes into supernets where possible",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts, err := subnet.Aggregate(args)
			if err != nil {
				return err
			}
			if opts.json {
				return writeSupernetJSON(cmd.OutOrStdout(), args, parts)
			}
			writeSupernet(cmd.OutOrStdout(), args, parts)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	return cmd
}

func Execute() error {
	root := NewRoot()
	err := root.Execute()
	if err != nil {
		if errors.Is(err, ErrNotContained) {
			return err
		}
		writeError(os.Stderr, err)
		return err
	}
	return nil
}
