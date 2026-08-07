package fetcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	BaseURL = "https://cn.investing.com"
)

// Fetcher wraps the Python curl_cffi subprocess to bypass Cloudflare.
type Fetcher struct {
	pythonPath string
	scriptPath string
}

// New creates a new Fetcher that uses Python + curl_cffi to bypass Cloudflare.
func New() (*Fetcher, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get cwd: %w", err)
	}
	projectRoot := findProjectRoot(wd)
	scriptPath := filepath.Join(projectRoot, "tools", "fetch.py")

	return &Fetcher{
		pythonPath: "python3",
		scriptPath: scriptPath,
	}, nil
}

// findProjectRoot walks up from the given directory to find the project root
// (where tools/fetch.py exists).
func findProjectRoot(start string) string {
	dir := start
	for {
		scriptPath := filepath.Join(dir, "tools", "fetch.py")
		if _, err := os.Stat(scriptPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}
	// Fallback: try common paths
	for _, p := range []string{
		filepath.Join("/workspace", "investing-scrapers"),
		filepath.Join(os.Getenv("HOME"), "investing-scrapers"),
	} {
		if _, err := os.Stat(filepath.Join(p, "tools", "fetch.py")); err == nil {
			return p
		}
	}
	// Last resort: use the original working directory
	origWd, _ := os.Getwd()
	return origWd
}

// FetchPage fetches a full HTML page from cn.investing.com.
func (f *Fetcher) FetchPage(url string) (string, error) {
	cmd := exec.Command(f.pythonPath, f.scriptPath, "page", url)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("fetch page %s: %w\noutput: %s", url, err, string(output))
	}

	body := strings.TrimSpace(string(output))
	if body == "403" || len(body) < 100 {
		return "", fmt.Errorf("fetch page %s: got suspicious response (len=%d)", url, len(body))
	}

	return body, nil
}

// FetchSearch calls the Search API endpoint.
func (f *Fetcher) FetchSearch(query string) (string, error) {
	cmd := exec.Command(f.pythonPath, f.scriptPath, "search", query)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("search %s: %w\noutput: %s", query, err, string(output))
	}

	body := strings.TrimSpace(string(output))
	if body == "403" || len(body) < 10 {
		return "", fmt.Errorf("search %s: got suspicious response", query)
	}

	return body, nil
}
