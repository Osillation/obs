package dashboard

import "github.com/spf13/cobra"

func NewDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Manage the local observability dashboard",
	}
	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newOpenCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newResetCmd())
	return cmd
}
