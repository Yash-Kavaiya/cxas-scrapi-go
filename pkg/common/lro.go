package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
)

type lroOperation struct {
	Name     string          `json:"name"`
	Done     bool            `json:"done"`
	Error    *lroError       `json:"error,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

type lroError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WaitForOperation polls a long-running operation until done or ctx is cancelled.
// On success it unmarshals op.response into out (if non-nil).
// baseURL should be the scheme+host only, e.g. "https://ces.googleapis.com".
func WaitForOperation(ctx context.Context, client *http.Client, baseURL, opName string, out interface{}) error {
	url := fmt.Sprintf("%s/%s/%s", baseURL, httpclient.APIVersion, opName)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var op lroOperation
			if err := httpclient.DoJSON(ctx, client, "GET", url, nil, &op); err != nil {
				return fmt.Errorf("poll operation %s: %w", opName, err)
			}
			if !op.Done {
				continue
			}
			if op.Error != nil {
				return fmt.Errorf("operation failed [%d]: %s", op.Error.Code, op.Error.Message)
			}
			if out != nil && len(op.Response) > 0 {
				return json.Unmarshal(op.Response, out)
			}
			return nil
		}
	}
}
