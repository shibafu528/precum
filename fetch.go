package precum

import (
	"context"
	"fmt"
	"net/http"
)

func fetch(ctx context.Context, method string, url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch(http.NewRequest): %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch(http.Client.Do): %w", err)
	}
	return res, nil
}
