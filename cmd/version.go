package cmd

import "github.com/spf13/cobra"

func newVersionCmd() *cobra.Command { return &cobra.Command{Use: "version"} }
