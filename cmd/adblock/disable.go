package adblock

import (
	"fmt"
	"os"
	"sync"

	"github.com/howinator/house/internal/pihole"
	"github.com/spf13/cobra"
)

const timerSeconds = 900 // 15 minutes

type server struct {
	name    string
	host    string
	passEnv string
}

var servers = []server{
	{name: "pihole1", host: "pihole.ui.sparky.best", passEnv: "HOUSE_PIHOLE1_PASSWORD"},
	{name: "pihole2", host: "pihole2.ui.sparky.best", passEnv: "HOUSE_PIHOLE2_PASSWORD"},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Temporarily disable ad blocking for 30 minutes",
	RunE:  runDisable,
}

func runDisable(cmd *cobra.Command, args []string) error {
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
			err := client.DisableBlocking(ctx, passwords[s.passEnv], timerSeconds)
			results <- result{name: s.name, err: err}
		}(s)
	}

	wg.Wait()
	close(results)

	var failed bool
	for r := range results {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", r.name, r.err)
			failed = true
		} else {
			fmt.Printf("OK   %s: ad blocking disabled for 30 minutes\n", r.name)
		}
	}

	if failed {
		return fmt.Errorf("one or more servers failed")
	}
	return nil
}
