package tdx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient struct {
	httpClient *http.Client
	token      string
}

func NewHTTPClient(token string, timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPClient{
		httpClient: &http.Client{Timeout: timeout},
		token:      token,
	}
}

// TQLEXQuery sends a request to the TQLEX API server.
// entry is passed as ?Entry=<entry> query parameter.
// body is marshalled as JSON and sent as the HTTP body (may be nil for GET-like requests).
func (c *HTTPClient) TQLEXQuery(ctx context.Context, entry string, body interface{}) (*TQLEXResponse, error) {
	u, err := url.Parse(TQLEXBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	q := u.Query()
	q.Set("Entry", entry)
	u.RawQuery = q.Encode()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("token", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var tdxResp TQLEXResponse
	if err := json.Unmarshal(respBytes, &tdxResp); err != nil {
		return &TQLEXResponse{Data: string(respBytes)}, nil
	}

	if tdxResp.Error != "" {
		return &tdxResp, fmt.Errorf("TDX error: %s", tdxResp.Error)
	}

	if tdxResp.Data == nil {
		var raw interface{}
		if err := json.Unmarshal(respBytes, &raw); err != nil {
			tdxResp.Data = string(respBytes)
		} else {
			tdxResp.Data = raw
		}
	}

	return &tdxResp, nil
}

// RAGQuery sends a request to the RAG entity retrieval API.
func (c *HTTPClient) RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error) {
	if topK <= 0 {
		topK = 10
	}

	reqBody := RAGRequest{
		Query: query,
		TopK:  topK,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RAGBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("token", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var ragResp RAGResponse
	if err := json.Unmarshal(respBytes, &ragResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if ragResp.Error != "" {
		return &ragResp, fmt.Errorf("RAG error: %s", ragResp.Error)
	}

	return &ragResp, nil
}
