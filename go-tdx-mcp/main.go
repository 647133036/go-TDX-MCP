package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"github.com/tdx/go-tdx-mcp/tdx"
	"github.com/tdx/go-tdx-mcp/web"
)

type Config struct {
	Token     string `json:"token"`
	Timeout   int    `json:"timeout"`
	WebPort   int    `json:"web_port"`
	TDxHost   string `json:"tdx_host"`
	TDxPort   int    `json:"tdx_port"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	if cfg.WebPort <= 0 {
		cfg.WebPort = 8000
	}
	return &cfg, nil
}

type runMode int

const (
	modeStdio          runMode = iota
	modeSSE
	modeStreamableHTTP
	modeCombined
)

func detectMode() runMode {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--sse":
			return modeSSE
		case "--streamable-http":
			return modeStreamableHTTP
		case "--web", "--serve":
			return modeCombined
		}
	}
	return modeStdio
}

func webPortFromArgs() int {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--port=") {
			var port int
			fmt.Sscanf(arg, "--port=%d", &port)
			if port > 0 && port < 65536 {
				return port
			}
		}
	}
	return 0
}

func main() {
	cfg := &Config{Timeout: 30, WebPort: 8000}

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "--") {
		var err error
		cfg, err = loadConfig(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法加载配置文件 %s: %v\n", os.Args[1], err)
			cfg = &Config{Timeout: 30, WebPort: 8000}
		}
	} else {
		if _, err := os.Stat("config.json"); err == nil {
			loaded, err := loadConfig("config.json")
			if err == nil {
				cfg = loaded
			}
		}
	}

	if p := webPortFromArgs(); p > 0 {
		cfg.WebPort = p
	}

	if tokenEnv := os.Getenv("TDX_TOKEN"); tokenEnv != "" {
		cfg.Token = tokenEnv
	}

	client := tdx.NewUnifiedClient(cfg.Token, cfg.Timeout, cfg.TDxHost, cfg.TDxPort)

	switch detectMode() {
	case modeSSE:
		runSSE(client, cfg)
	case modeStreamableHTTP:
		runStreamableHTTP(client, cfg)
	case modeCombined:
		runCombined(client, cfg)
	default:
		runMCP(client)
	}
}

func buildMCPServer(client *tdx.UnifiedClient) *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"TDX Finance MCP",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	for _, tool := range tdx.AllTools() {
		h := tdx.GetHandler(tool.Name)
		if h == nil {
			fmt.Fprintf(os.Stderr, "警告: 核心工具 '%s' 无对应处理器\n", tool.Name)
			continue
		}
		mcpServer.AddTool(tool, tdx.CreateToolHandler(client, h))
	}

	for _, tool := range tdx.GetAllExpandedTools() {
		h := tdx.GetExpandedHandler(tool.Name)
		if h == nil {
			fmt.Fprintf(os.Stderr, "警告: 扩展工具 '%s' 无对应处理器\n", tool.Name)
			continue
		}
		mcpServer.AddTool(tool, tdx.CreateToolHandler(client, h))
	}

	for _, tool := range tdx.GetAllV3Tools() {
		h := tdx.GetV3Handler(tool.Name)
		if h == nil {
			fmt.Fprintf(os.Stderr, "警告: V3工具 '%s' 无对应处理器\n", tool.Name)
			continue
		}
		mcpServer.AddTool(tool, tdx.CreateToolHandler(client, h))
	}

	for _, tool := range tdx.GetAllNewTools() {
		h := tdx.GetNewHandler(tool.Name)
		if h == nil {
			fmt.Fprintf(os.Stderr, "警告: 新增工具 '%s' 无对应处理器\n", tool.Name)
			continue
		}
		mcpServer.AddTool(tool, tdx.CreateToolHandler(client, h))
	}

	mcpServer.AddPrompts(tdx.AllServerPrompts()...)
	return mcpServer
}

func toolCount() int {
	return len(tdx.AllTools()) + len(tdx.GetAllExpandedTools()) + len(tdx.GetAllV3Tools()) + len(tdx.GetAllNewTools())
}

func runMCP(client *tdx.UnifiedClient) {
	mcpServer := buildMCPServer(client)
	fmt.Fprintf(os.Stderr, "TDX Finance MCP v1.0.0 (stdio): %d 工具 + 45 投资技能\n", toolCount())

	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Fprintf(os.Stderr, "MCP 服务错误: %v\n", err)
		os.Exit(1)
	}
}

func runSSE(client *tdx.UnifiedClient, cfg *Config) {
	mcpServer := buildMCPServer(client)
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.WebPort)

	sseServer := server.NewSSEServer(mcpServer,
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
	)

	fmt.Fprintf(os.Stderr, "TDX Finance MCP v1.0.0 (SSE): %d 工具 + 45 投资技能\n", toolCount())
	fmt.Fprintf(os.Stderr, "  SSE 端点: http://0.0.0.0:%d/sse\n", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "  消息端点: http://0.0.0.0:%d/message\n", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "  使用 --port=%d 可更改端口\n", cfg.WebPort)

	if err := sseServer.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "SSE 服务错误: %v\n", err)
		os.Exit(1)
	}
}

func runStreamableHTTP(client *tdx.UnifiedClient, cfg *Config) {
	mcpServer := buildMCPServer(client)
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.WebPort)

	httpServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
	)

	fmt.Fprintf(os.Stderr, "TDX Finance MCP v1.0.0 (Streamable HTTP): %d 工具 + 45 投资技能\n", toolCount())
	fmt.Fprintf(os.Stderr, "  端点: http://0.0.0.0:%d/mcp\n", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "  使用 --port=%d 可更改端口\n", cfg.WebPort)

	if err := httpServer.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Streamable HTTP 服务错误: %v\n", err)
		os.Exit(1)
	}
}

func runCombined(client *tdx.UnifiedClient, cfg *Config) {
	mcpServer := buildMCPServer(client)
	httpMCP := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
	)

	webServer := web.NewServer(client, fmt.Sprintf("0.0.0.0:%d", cfg.WebPort))

	rootMux := http.NewServeMux()
	rootMux.Handle("/mcp", httpMCP)
	rootMux.Handle("/", webServer.Handler())

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "TDX Finance Combined v1.0.0: %d 工具 + 45 投资技能\n", toolCount())
	fmt.Fprintf(os.Stderr, "  MCP Streamable HTTP: http://0.0.0.0:%d/mcp\n", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "  REST API 文档:      http://0.0.0.0:%d/\n", cfg.WebPort)
	fmt.Fprintf(os.Stderr, "  健康检查:           http://0.0.0.0:%d/api/v1/health\n", cfg.WebPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: rootMux,
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Combined 服务错误: %v\n", err)
		os.Exit(1)
	}
}
