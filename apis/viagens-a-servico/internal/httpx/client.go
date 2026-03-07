// Radar do Povo ETL - https://radardopovo.com
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"time"
)

const defaultUserAgent = "viagens-etl/1.0 (+https://radardopovo.com)"

type Client struct {
	httpClient *http.Client
	maxRetries int
	logger     *slog.Logger
	userAgent  string
	rng        *rand.Rand
}

func NewClient(timeout time.Duration, maxRetries int, logger *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		maxRetries: maxRetries,
		logger:     logger,
		userAgent:  defaultUserAgent,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	var lastErr error
	baseBackoff := 2 * time.Second
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err == nil {
			if !isRetryStatus(resp.StatusCode) {
				return resp, nil
			}
			bodyPreview := readSmallBody(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, bodyPreview)
		} else {
			lastErr = err
			if !isRetryableErr(err) {
				return nil, err
			}
		}

		if attempt == c.maxRetries {
			break
		}
		sleep := withJitter(baseBackoff*time.Duration(1<<attempt), c.rng)
		c.logger.Debug("http retry",
			slog.Int("attempt", attempt+1),
			slog.Duration("sleep", sleep),
			slog.String("url", url),
			slog.String("error", lastErr.Error()),
		)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
	}
	return nil, fmt.Errorf("request failed after retries: %w", lastErr)
}

func isRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests ||
		code == http.StatusInternalServerError ||
		code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.ErrUnexpectedEOF)
}

func withJitter(base time.Duration, rng *rand.Rand) time.Duration {
	factor := 1.0 + (rng.Float64()*0.4 - 0.2)
	return time.Duration(float64(base) * factor)
}

func readSmallBody(r io.Reader) string {
	if r == nil {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(r, 256))
	if err != nil {
		return ""
	}
	return string(b)
}
