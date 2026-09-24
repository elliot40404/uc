package collect

import (
	"context"
	"sync"
	"time"

	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/usage"
)

type Reporter interface {
	Report(ctx context.Context, name, dir string, now time.Time) usage.Report
}

const maxConcurrent = 8

func All(ctx context.Context, accounts []discover.Account, reporters map[string]Reporter, now time.Time) []usage.Report {
	out := make([]usage.Report, len(accounts))
	if len(accounts) == 0 {
		return out
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(maxConcurrent, len(accounts)) {
		wg.Go(func() {
			for i := range jobs {
				a := accounts[i]
				if rep, ok := reporters[a.Provider]; ok {
					out[i] = rep.Report(ctx, a.Name, a.Dir, now)
				}
			}
		})
	}
	for i := range accounts {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}
