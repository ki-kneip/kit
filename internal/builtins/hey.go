package builtins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"mvdan.cc/sh/v3/interp"
)

// hey is a small HTTP load tester built into kit, so task scripts can
// smoke-test endpoints on any machine that has only the kit binary.
func hey(ctx context.Context, hc interp.HandlerContext, args []string) error {
	total, workers, method, url := 200, 50, "GET", ""

	for i := 0; i < len(args); i++ {
		next := func() (string, error) {
			i++
			if i >= len(args) {
				return "", fmt.Errorf("%s needs a value", args[i-1])
			}
			return args[i], nil
		}
		var err error
		switch args[i] {
		case "-n":
			var v string
			if v, err = next(); err == nil {
				total, err = strconv.Atoi(v)
			}
		case "-c":
			var v string
			if v, err = next(); err == nil {
				workers, err = strconv.Atoi(v)
			}
		case "-m":
			method, err = next()
		default:
			if url != "" {
				err = fmt.Errorf("unexpected argument %q", args[i])
			}
			url = args[i]
		}
		if err != nil {
			return err
		}
	}
	if url == "" || total < 1 || workers < 1 {
		return errors.New("usage: hey [-n requests] [-c concurrency] [-m method] URL")
	}
	if workers > total {
		workers = total
	}

	var (
		client   = &http.Client{Timeout: 30 * time.Second}
		mu       sync.Mutex
		latency  = make([]time.Duration, 0, total)
		statuses = map[int]int{}
		failed   int
	)

	jobs := make(chan struct{})
	var wg sync.WaitGroup
	start := time.Now()
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				req, err := http.NewRequestWithContext(ctx, method, url, nil)
				if err == nil {
					t0 := time.Now()
					var resp *http.Response
					if resp, err = client.Do(req); err == nil {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
						mu.Lock()
						latency = append(latency, time.Since(t0))
						statuses[resp.StatusCode]++
						mu.Unlock()
						continue
					}
				}
				mu.Lock()
				failed++
				mu.Unlock()
			}
		}()
	}

feed:
	for range total {
		select {
		case <-ctx.Done():
			break feed
		case jobs <- struct{}{}:
		}
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	sort.Slice(latency, func(i, j int) bool { return latency[i] < latency[j] })
	pct := func(p float64) time.Duration {
		if len(latency) == 0 {
			return 0
		}
		i := int(p * float64(len(latency)-1))
		return latency[i]
	}
	var sum time.Duration
	for _, d := range latency {
		sum += d
	}
	avg := time.Duration(0)
	if len(latency) > 0 {
		avg = sum / time.Duration(len(latency))
	}

	out := hc.Stdout
	fmt.Fprintf(out, "hey  %s %s\n", method, url)
	fmt.Fprintf(out, "  requests     %d done, %d failed\n", len(latency), failed)
	fmt.Fprintf(out, "  concurrency  %d\n", workers)
	fmt.Fprintf(out, "  total        %v  (%.1f req/s)\n", elapsed.Round(time.Millisecond), float64(len(latency))/elapsed.Seconds())
	fmt.Fprintf(out, "  latency      avg %v  p50 %v  p90 %v  p99 %v\n",
		avg.Round(time.Millisecond), pct(0.50).Round(time.Millisecond),
		pct(0.90).Round(time.Millisecond), pct(0.99).Round(time.Millisecond))
	for code, n := range statuses {
		fmt.Fprintf(out, "  status %d   ×%d\n", code, n)
	}

	if failed > 0 && len(latency) == 0 {
		return errors.New("every request failed")
	}
	return nil
}
