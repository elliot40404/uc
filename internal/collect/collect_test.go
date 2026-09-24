package collect

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/usage"
)

type countingReporter struct {
	active atomic.Int32
	peak   atomic.Int32
}

func (r *countingReporter) Report(_ context.Context, name, dir string, _ time.Time) usage.Report {
	n := r.active.Add(1)
	for {
		peak := r.peak.Load()
		if n <= peak || r.peak.CompareAndSwap(peak, n) {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	r.active.Add(-1)
	return usage.Report{Account: name, Dir: dir, Status: usage.StatusOK}
}

func TestAllBoundsConcurrentRequestsAndPreservesOrder(t *testing.T) {
	accounts := make([]discover.Account, 50)
	for i := range accounts {
		accounts[i] = discover.Account{Provider: "claude", Name: fmt.Sprintf("account-%d", i)}
	}
	reporter := new(countingReporter)
	reports := All(t.Context(), accounts, map[string]Reporter{"claude": reporter}, time.Now())
	if reporter.peak.Load() < 2 || reporter.peak.Load() > maxConcurrent {
		t.Fatalf("peak concurrent requests = %d", reporter.peak.Load())
	}
	for i, r := range reports {
		if r.Account != accounts[i].Name {
			t.Fatalf("report %d = %q", i, r.Account)
		}
	}
}
