package adblock

import "github.com/spf13/cobra"

// Cmd is the parent command for ad blocking operations.
var Cmd = &cobra.Command{
	Use:   "adblock",
	Short: "Manage ad blocking on Pi-hole servers",
}

func init() {
	Cmd.AddCommand(disableCmd)
}
