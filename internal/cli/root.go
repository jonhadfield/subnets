package cli

import (
	"fmt"
	"os"

	"github.com/jonhadfield/subnets/internal/subnet"
	"github.com/spf13/cobra"
)

// Version is the build version. Override at link time:
//
//	-ldflags "-X github.com/jonhadfield/subnets/internal/cli.Version=1.2.3"
var Version = "dev"

type options struct {
	aws   bool
	ipv6  bool
	json  bool
	limit int
}

func NewRoot() *cobra.Command {
	opts := &options{limit: 256}

	root := &cobra.Command{
		Use:           "subnets [cidr]",
		Short:         "IP subnet calculator and cheatsheet",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		Long: `subnets shows subnet details for an address or CIDR.

Examples:
  subnets 192.168.0.3/28
  subnets 192.168.0.0/28 --aws
  subnets 192.168.0.3/28 --json
  subnets sheet
  subnets hosts 192.168.0.0/28 --aws
  subnets hosts 192.168.0.0/28 --json --aws`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			info, err := subnet.Parse(args[0])
			if err != nil {
				return err
			}
			if opts.json {
				return writeShowJSON(cmd.OutOrStdout(), info, opts.aws)
			}
			writeShow(cmd.OutOrStdout(), info, opts.aws)
			return nil
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.Flags().BoolVar(&opts.aws, "aws", false, "show addresses reserved by AWS VPC")
	root.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")

	root.AddCommand(newSheetCmd(opts), newHostsCmd(opts))
	return root
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := subnet.Parse(args[0])
			if err != nil {
				return err
			}
			if opts.limit < 1 {
				return fmt.Errorf("limit must be at least 1")
			}
			if opts.json {
				return writeHostsJSON(cmd.OutOrStdout(), info, opts.aws, opts.limit)
			}
			writeHosts(cmd.OutOrStdout(), info, opts.aws, opts.limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.aws, "aws", false, "mark addresses reserved by AWS VPC")
	cmd.Flags().BoolVar(&opts.json, "json", false, "emit machine-readable JSON")
	cmd.Flags().IntVar(&opts.limit, "limit", 256, "maximum addresses to print")
	return cmd
}

func Execute() error {
	root := NewRoot()
	err := root.Execute()
	if err != nil {
		writeError(os.Stderr, err)
		return err
	}
	return nil
}
