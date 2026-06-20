package cmd

import "github.com/spf13/cobra"

func newDeployCmd() *cobra.Command { return &cobra.Command{Use: "deploy"} }
