package tdx

import "context"

const (
	TQLEXBaseURL = "http://tdxhub.icfqs.com:7615/TQLEX"
	RAGBaseURL   = "https://ai.icfqs.com:8965/v1/rag-entity-retrieve"
)

type TQLEXResponse struct {
	Data  interface{} `json:"Data"`
	Error string      `json:"Error"`
}

type RAGRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

type RAGResponse struct {
	Results []RAGResult `json:"results"`
	Error   string      `json:"error,omitempty"`
}

type RAGResult struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Score   float64 `json:"score"`
}

type Client interface {
	TQLEXQuery(ctx context.Context, entry string, body interface{}) (*TQLEXResponse, error)
	RAGQuery(ctx context.Context, query string, topK int) (*RAGResponse, error)
}

// TDXHead is the common header for TDX API requests.
type TDXHead struct {
	Target  string `json:"Target"`
	CharSet string `json:"CharSet"`
}

// QuoteRequest is the request body for tdx_quotes.
type QuoteRequest struct {
	Head        TDXHead `json:"Head"`
	Code        string  `json:"Code"`
	Setcode     string  `json:"Setcode"`
	HasHQInfo   string  `json:"HasHQInfo"`
	HasExtInfo  string  `json:"HasExtInfo"`
	BspNum      string  `json:"BspNum"`
	HasProInfo  string  `json:"HasProInfo"`
	HasCalcInfo string  `json:"HasCalcInfo"`
	HasCwInfo   string  `json:"HasCwInfo"`
	HasStatInfo string  `json:"HasStatInfo"`
	StatParam   string  `json:"StatParam,omitempty"`
}

// KlineRequest is the request body for tdx_kline.
type KlineRequest struct {
	Head          TDXHead `json:"Head"`
	Code          string  `json:"Code"`
	Setcode       int     `json:"Setcode"`
	Period        int     `json:"Period"`
	Startxh       int     `json:"Startxh"`
	WantNum       int     `json:"WantNum"`
	TQFlag        int     `json:"TQFlag"`
	MPData        int     `json:"MPData"`
	HasAttachInfo int     `json:"HasAttachInfo"`
	HasLtgb       int     `json:"HasLtgb"`
	ForRefresh    int     `json:"ForRefresh"`
	HasIpoPrice   int     `json:"HasIpoPrice"`
}

// ScreenerItem is one element in the screener request array.
type ScreenerItem struct {
	Message  string `json:"message"`
	Rang     string `json:"rang"`
	PageNo   string `json:"pageNo"`
	PageSize string `json:"pageSize"`
}

// ScreenerRequest is the request body for tdx_screener (array).
type ScreenerRequest []ScreenerItem

// IndicatorSelectRequest is the request body for tdx_indicator_select.
type IndicatorSelectRequest struct {
	Message string `json:"message"`
	Rang    string `json:"rang"`
}

// ApiDataRequest is the request body for tdx_api_data.
type ApiDataRequest struct {
	Params []interface{} `json:"Params"`
}

// PeriodToCode converts a human-readable period string to the TDX numeric code.
func PeriodToCode(period string) int {
	switch period {
	case "day", "日线":
		return 4
	case "week", "周线":
		return 5
	case "month", "月线":
		return 6
	case "60min", "60分钟":
		return 3
	case "1min", "1分钟":
		return 9
	case "5min", "5分钟":
		return 10
	case "15min", "15分钟":
		return 11
	case "30min", "30分钟":
		return 12
	default:
		return 4
	}
}
