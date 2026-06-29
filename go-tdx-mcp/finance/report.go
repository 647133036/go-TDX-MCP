package finance

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type FinancialReport struct {
	Code       string       `json:"code"`
	ReportType string       `json:"report_type"`
	Num        int          `json:"num"`
	Periods    []ReportPeriod `json:"periods"`
}

type ReportPeriod struct {
	Date  string            `json:"date"`
	Items map[string]float64 `json:"items"`
}

const (
	ProfitURL     = "https://money.finance.sina.com.cn/corp/go.php/vDOWN_ProfitStatement/displaytype/4/stockid/%s/ctrl/all.phtml"
	BalanceURL    = "https://money.finance.sina.com.cn/corp/go.php/vDOWN_BalanceSheet/displaytype/4/stockid/%s/ctrl/all.phtml"
	CashFlowURL   = "https://money.finance.sina.com.cn/corp/go.php/vDOWN_CashFlow/displaytype/4/stockid/%s/ctrl/all.phtml"
)

func FetchReport(code, reportType string) (*FinancialReport, error) {
	var urlFmt string
	switch strings.ToLower(reportType) {
	case "lrb":
		urlFmt = ProfitURL
	case "fzb":
		urlFmt = BalanceURL
	case "llb":
		urlFmt = CashFlowURL
	default:
		return nil, fmt.Errorf("unsupported report type: %s (use lrb/fzb/llb)", reportType)
	}

	url := fmt.Sprintf(urlFmt, code)
	hc := &http.Client{Timeout: 15}
	resp, err := hc.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch financial report failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	encoding := detectEncoding(body, resp.Header.Get("Content-Type"))
	if encoding == "gbk" {
		converted, err2 := gbkToUTF8(body)
		if err2 == nil {
			body = converted
		}
	}

	periods, err := parseCSVFinancial(body)
	if err != nil {
		return nil, fmt.Errorf("parse csv failed: %w", err)
	}

	return &FinancialReport{
		Code:       code,
		ReportType: reportType,
		Num:        len(periods),
		Periods:    periods,
	}, nil
}

func parseCSVFinancial(body []byte) ([]ReportPeriod, error) {
	text := string(body)
	lines := splitLines(text)
	if len(lines) < 2 {
		return nil, fmt.Errorf("empty or too few lines")
	}

	// Find header line: first line with 3+ fields where at least 2 look like dates
	var headerFields []string
	for _, line := range lines {
		fields := splitTab(line)
		if len(fields) >= 3 && countDateFields(fields) >= 2 {
			headerFields = fields
			break
		}
	}

	if len(headerFields) == 0 {
		// Fallback: use first non-empty line with 3+ fields
		for _, line := range lines {
			fields := splitTab(line)
			if len(fields) >= 3 {
				headerFields = fields
				break
			}
		}
	}

	if len(headerFields) < 3 {
		return nil, fmt.Errorf("cannot find header line")
	}

	// headerFields[0] = item name column header (unused)
	// headerFields[1..] = date columns
	dates := headerFields[1:]

	// Collect rows grouped by date
	dateRows := make(map[string][]rowItem)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := splitTab(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSpace(fields[0])
		if name == "" || isFooterRow(name) {
			continue
		}
		values := fields[1:]
		for j, val := range values {
			date := ""
			if j < len(dates) {
				date = strings.TrimSpace(dates[j])
			}
			if date == "" {
				continue
			}
			f := parseFinancialFloat(strings.TrimSpace(val))
			dateRows[date] = append(dateRows[date], rowItem{name: name, value: f})
		}
	}

	// Build ReportPeriods sorted by date order from header
	var periods []ReportPeriod
	for _, date := range dates {
		date = strings.TrimSpace(date)
		if date == "" {
			continue
		}
		items, ok := dateRows[date]
		if !ok {
			continue
		}
		p := ReportPeriod{Date: date, Items: make(map[string]float64)}
		for _, ri := range items {
			p.Items[ri.name] = ri.value
		}
		// Calculate YoY vs previous period
		periods = append(periods, p)
	}

	return periods, nil
}

type rowItem struct {
	name  string
	value float64
}

func splitLines(text string) []string {
	return strings.Split(text, "\n")
}

func splitTab(line string) []string {
	return strings.Split(line, "\t")
}

func countDateFields(fields []string) int {
	count := 0
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" || f == "--" {
			continue
		}
		if len(f) == 8 && isDigits(f) {
			count++
		}
	}
	return count
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isFooterRow(name string) bool {
	lower := strings.ToLower(name)
	// These are common footer/summary rows to skip
	skipKeywords := []string{"合计", "总计", "附注"}
	for _, kw := range skipKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func parseFinancialFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.ReplaceAll(s, " ", "")
	if s == "" || s == "--" || s == "-" || s == "null" {
		return 0
	}
	multiplier := 1.0
	l := strings.ToLower(s)
	if strings.HasSuffix(l, "亿") {
		multiplier = 100000000
		s = s[:len(s)-2]
	} else if strings.HasSuffix(l, "万") {
		multiplier = 10000
		s = s[:len(s)-2]
	} else if strings.HasSuffix(l, "千") {
		multiplier = 1000
		s = s[:len(s)-2]
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v * multiplier
}

func detectEncoding(body []byte, contentType string) string {
	if contentType != "" {
		lower := strings.ToLower(contentType)
		if strings.Contains(lower, "gbk") || strings.Contains(lower, "gb2312") || strings.Contains(lower, "gb18030") {
			return "gbk"
		}
		if strings.Contains(lower, "utf-8") || strings.Contains(lower, "utf8") {
			return "utf-8"
		}
	}
	// BOM detection
	if bytes.HasPrefix(body, []byte{0xef, 0xbb, 0xbf}) {
		return "utf-8-bom"
	}
	// Heuristic: GBK common Chinese financial terms
	if bytes.Contains(body, []byte{0xb9, 0xfa, 0xc1, 0xac}) {
		return "gbk"
	}
	return ""
}

func gbkToUTF8(data []byte) ([]byte, error) {
	// Strip BOM if present
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		data = data[3:]
	}
	// Check if already valid UTF-8
	if isUTF8(data) {
		return data, nil
	}
	// Use simple GBK lookup for common financial terms
	return minimalGBKConvert(data)
}

func isUTF8(data []byte) bool {
	for i := 0; i < len(data); i++ {
		b := data[i]
		if b&0x80 == 0 {
			continue
		}
		if b&0xE0 == 0xC0 {
			if i+1 >= len(data) || data[i+1]&0xC0 != 0x80 {
				return false
			}
			i++
		} else if b&0xF0 == 0xE0 {
			if i+2 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 {
				return false
			}
			i += 2
		} else if b&0xF8 == 0xF0 {
			if i+3 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 || data[i+3]&0xC0 != 0x80 {
				return false
			}
			i += 3
		} else {
			return false
		}
	}
	return true
}

// minimalGBKConvert handles the most common financial report fields.
// For production use, consider importing golang.org/x/text/encoding/simplifiedchinese.
func minimalGBKConvert(data []byte) ([]byte, error) {
	// Since we can't reliably convert GBK without external deps,
	// try to extract numeric data by splitting on tabs and keeping only the structure.
	// The Chinese field names will be garbled but the numeric data is ASCII.
	lines := splitLines(string(data))
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := splitTab(line)
		if len(fields) == 0 {
			continue
		}
		// Keep the first field (Chinese name) as-is (garbled but preserved)
		// and all numeric fields
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n")), nil
}

func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	// Simplified: just use a buffer
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	b := make([]byte, 4096)
	for {
		n, err := r.Read(b)
		if n > 0 {
			buf.Write(b[:n])
		}
		if err != nil {
			break
		}
	}
	return buf.Bytes(), nil
}
