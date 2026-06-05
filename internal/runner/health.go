package runner

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitHealthy polls url until it returns HTTP 2xx or the context is done. It
// returns nil once healthy, or an error describing the last failure on timeout.
func WaitHealthy(ctx context.Context, url string, interval time.Duration) error {
	client := &http.Client{Timeout: interval}
	var lastErr error

	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("health check %s never passed: %w", url, lastErr)
			}
			return fmt.Errorf("health check %s never passed: %w", url, ctx.Err())
		default:
		}

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("health check %s never passed: %w", url, lastErr)
		case <-time.After(interval):
		}
	}
}
