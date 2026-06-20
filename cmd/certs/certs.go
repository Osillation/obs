package certs

import "github.com/spf13/cobra"

func NewCertsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "Manage mTLS employee certificates",
	}
	cmd.AddCommand(newInitCACmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRevokeCmd())
	return cmd
}
