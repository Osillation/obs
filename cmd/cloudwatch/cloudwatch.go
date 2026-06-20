package cloudwatch

import "github.com/spf13/cobra"

func NewCloudwatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloudwatch",
		Short: "Connect AWS CloudWatch as a data source",
	}
	cmd.AddCommand(newConnectCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDisconnectCmd())
	return cmd
}
