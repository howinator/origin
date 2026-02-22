package adblock

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/howinator/house/internal/pihole"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Re-enable ad blocking on Pi-hole servers",
	RunE:  runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	passwords := make(map[string]string)
	var missing []string
	for _, s := range servers {
		p := os.Getenv(s.passEnv)
		if p == "" {
			missing = append(missing, s.passEnv)
		}
		passwords[s.passEnv] = p
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	ctx := cmd.Context()
	var wg sync.WaitGroup
	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(servers))

	for _, s := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			client := pihole.NewClient(s.host)
			err := client.EnableBlocking(ctx, passwords[s.passEnv])
			results <- result{name: s.name, err: err}
		}(s)
	}

	wg.Wait()
	close(results)

	var failed bool
	for r := range results {
		if r.err != nil {
			if errors.Is(r.err, pihole.ErrAlreadyEnabled) {
				fmt.Printf("OK   %s: ad blocking already enabled\n", r.name)
			} else {
				fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", r.name, r.err)
				failed = true
			}
		} else {
			fmt.Printf("OK   %s: ad blocking enabled\n", r.name)
		}
	}

	if failed {
		return fmt.Errorf("one or more servers failed")
	}
	return nil
}
