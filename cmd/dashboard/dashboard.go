package dashboard

import "github.com/spf13/cobra"

func NewDashboardCmd() *cobra.Command { return &cobra.Command{Use: "dashboard"} }
