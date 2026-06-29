package scraper

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AntiBanConfig controls anti-bot detection bypass behavior.
type AntiBanConfig struct {
	MinDelay    time.Duration // minimum delay between requests
	MaxDelay    time.Duration // maximum delay between requests
	MaxRetries  int           // max retry attempts
	RetryDelay  time.Duration // base delay for exponential backoff
	UserAgents  []string      // pool of user agents to rotate
	Proxies     []string      // proxy URLs (http://user:pass@host:port)
	EnableCookies bool        // enable cookie jar for session persistence
}

// DefaultAntiBanConfig returns a conservative config for slow lane scraping.
func DefaultAntiBanConfig() AntiBanConfig {
	return AntiBanConfig{
		MinDelay:      1500 * time.Millisecond,
		MaxDelay:      3000 * time.Millisecond,
		MaxRetries:    3,
		RetryDelay:    2 * time.Second,
		EnableCookies: true,
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		},
	}
}

// AntiBanTransport wraps http.RoundTripper with anti-detection features.
type AntiBanTransport struct {
	cfg    AntiBanConfig
	jar    *cookiejar.Jar
	mu     sync.Mutex
	proxy  func(*http.Request) (*url.URL, error)
}

// NewAntiBanTransport creates a transport with anti-bot features.
func NewAntiBanTransport(cfg AntiBanConfig) *AntiBanTransport {
	var jar *cookiejar.Jar
	if cfg.EnableCookies {
		jar = newCookieJar()
	}
	return &AntiBanTransport{
		cfg:   cfg,
		jar:   jar,
		proxy: pickRandomProxy(cfg.Proxies),
	}
}

func newCookieJar() *cookiejar.Jar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}
	return jar
}

func pickRandomProxy(proxies []string) func(*http.Request) (*url.URL, error) {
	if len(proxies) == 0 {
		return nil
	}
	idx := rand.Intn(len(proxies))
	proxyURL := proxies[idx]
	parsed, _ := url.Parse(proxyURL)
	return func(req *http.Request) (*url.URL, error) {
		return parsed, nil
	}
}

// RoundTrip executes a single HTTP request with anti-detection.
func (t *AntiBanTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rotate User-Agent
	t.mu.Lock()
	uaIdx := rand.Intn(len(t.cfg.UserAgents))
	req.Header.Set("User-Agent", t.cfg.UserAgents[uaIdx])
	t.mu.Unlock()

	// Set Referer if not already set
	if req.Header.Get("Referer") == "" {
		req.Header.Set("Referer", "https://www.eastmoney.com/")
	}

	// Set Accept headers
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json, text/plain, */*")
	}
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	}
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}
	if req.Header.Get("Connection") == "" {
		req.Header.Set("Connection", "keep-alive")
	}

	// Add random delay
	delay := time.Duration(rand.Int63n(int64(t.cfg.MaxDelay-t.cfg.MinDelay) + int64(t.cfg.MinDelay)))
	time.Sleep(delay)

	// Execute with retry
	return t.executeWithRetry(req)
}

func (t *AntiBanTransport) executeWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= t.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := t.cfg.RetryDelay * time.Duration(1<<uint(attempt-1))
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			time.Sleep(backoff + jitter)
		}

		// Clone request to avoid modifying original
		reqCopy := cloneRequest(req)

		var resp *http.Response
		var err error

		roundTripper := &cookieAwareTransport{
			base: &http.Transport{},
			jar:  t.jar,
		}
		if t.proxy != nil {
			roundTripper.base.Proxy = t.proxy
		}

		resp, err = roundTripper.RoundTrip(reqCopy)

		if err != nil {
			lastErr = err
			continue
		}

		// Check for anti-bot response indicators
		if resp.StatusCode == 403 || resp.StatusCode == 429 || resp.StatusCode == 503 {
			resp.Body.Close()
			lastErr = fmt.Errorf("anti-bot detected: status %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all retries failed: %w", lastErr)
}

// cookieAwareTransport wraps http.Transport to handle cookies.
type cookieAwareTransport struct {
	base *http.Transport
	jar  *cookiejar.Jar
}

func (ct *cookieAwareTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add cookies if jar exists
	if ct.jar != nil {
		if cookies := ct.jar.Cookies(req.URL); len(cookies) > 0 {
			req.Header.Set("Cookie", strings.Join(cookieStrings(cookies), "; "))
		}
	}

	resp, err := ct.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Save cookies from response
	if ct.jar != nil && resp.Header.Get("Set-Cookie") != "" {
		ct.jar.SetCookies(req.URL, resp.Cookies())
	}

	return resp, nil
}

func cookieStrings(cookies []*http.Cookie) []string {
	strs := make([]string, len(cookies))
	for i, c := range cookies {
		strs[i] = c.Name + "=" + c.Value
	}
	return strs
}

// cloneRequest creates a shallow copy of the request.
func cloneRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

// AntiBanClient wraps http.Client with anti-detection transport.
type AntiBanClient struct {
	client *http.Client
}

// NewAntiBanClient creates a client with anti-bot features.
func NewAntiBanClient(cfg AntiBanConfig) *AntiBanClient {
	transport := NewAntiBanTransport(cfg)
	return &AntiBanClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// Get performs a GET request with anti-detection.
func (c *AntiBanClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// Do performs a request with anti-detection.
func (c *AntiBanClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Client returns the underlying http.Client.
func (c *AntiBanClient) Client() *http.Client {
	return c.client
}

// RateLimiter provides token bucket rate limiting.
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a rate limiter with given tokens per second.
func NewRateLimiter(tokensPerSecond float64, maxTokens float64) *RateLimiter {
	if maxTokens <= 0 {
		maxTokens = tokensPerSecond
	}
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: tokensPerSecond,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available.
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	for rl.tokens < 1 {
		rl.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		rl.mu.Lock()
		now = time.Now()
		elapsed = now.Sub(rl.lastRefill).Seconds()
		rl.tokens += elapsed * rl.refillRate
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
	rl.tokens -= 1
}

// TryWait returns true if a token is available immediately.
func (rl *RateLimiter) TryWait() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens -= 1
		return true
	}
	return false
}

// Decryptor provides utilities for decrypting encrypted API responses.
type Decryptor struct{}

// DecryptAES decodes AES-encrypted data (common in EastMoney API).
func (d *Decryptor) DecryptAES(encrypted, key string) (string, error) {
	// Placeholder: implement actual AES decryption logic
	// EastMoney often uses AES-CBC with a dynamic key derived from JS
	return "", fmt.Errorf("AES decryption not implemented - requires analyzing frontend JS")
}

// DecryptRSA decrypts RSA-encrypted data.
func (d *Decryptor) DecryptRSA(encrypted, privateKey string) (string, error) {
	return "", fmt.Errorf("RSA decryption not implemented")
}

// AnalyzeJS extracts encryption logic from a JS file.
func (d *Decryptor) AnalyzeJS(jsSource string) map[string]interface{} {
	// Parse JS to find encryption functions
	// This would use a JS parser or regex to extract patterns
	return map[string]interface{}{
		"status": "not_implemented",
		"note":   "Use a JS engine (e.g., goja) to execute and trace encryption",
	}
}
