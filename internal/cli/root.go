package cli

import (
	"fmt"
	"os"

	"github.com/jonhadfield/subnets/internal/subnet"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

type options struct {
	aws   bool
	ipv6  bool
	limit int
}

func NewRoot() *cobra.Command {
	opts := &options{limit: 256}

	root := &cobra.Command{
		Use:           "subnets [cidr]",
		Short:         "IP subnet calculator and cheatsheet",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		Long: `subnets shows subnet details for an address or CIDR.

Examples:
  subnets 192.168.0.3/28
  subnets 192.168.0.0/28 --aws
  subnets sheet
  subnets hosts 192.168.0.0/28 --aws`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			info, err := subnet.Parse(args[0])
			if err != nil {
				return err
			}
			writeShow(cmd.OutOrStdout(), info, opts.aws)
			return nil
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.Flags().BoolVar(&opts.aws, "aws", false, "show addresses reserved by AWS VPC")

	root.AddCommand(newSheetCmd(opts), newHostsCmd(opts))
	return root
}

func newSheetCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheet",
		Short: "Prefix length and subnet mask cheatsheet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			writeSheet(cmd.OutOrStdout(), opts.ipv6)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.ipv6, "ipv6", false, "show the IPv6 prefix cheatsheet")
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
			writeHosts(cmd.OutOrStdout(), info, opts.aws, opts.limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.aws, "aws", false, "mark addresses reserved by AWS VPC")
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
