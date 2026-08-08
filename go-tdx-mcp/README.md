# TDX Finance MCP v1.0.2

通达信金融数据 MCP 服务器，提供 A 股、港股、美股、加密货币、期货等多市场金融数据服务。

**215+ 个 MCP 工具** + **45+ 个投资技能**，覆盖实时行情、K 线、技术指标、缠论分析、量化回测、资金流向、板块分析、基金数据、宏观数据、新闻情感等全场景。

## 重要说明：同代码标的区分

同一代码 `000001` 可对应多个不同标的，必须通过 `setcode`/`market` 参数指定市场：

| 代码 | setcode/market | 标的名称 |
|------|---------------|----------|
| 000001 | **0（默认/深市）** | 平安银行（股票） |
| 000001 | **1（沪市）** | 上证指数（指数） |
| 600000 | 1（沪市） | 浦发银行（股票） |
| 600000 | 0（深市） | 无效（沪深代码段不同） |

**各接口的市场参数名：**

| 接口类型 | 参数名 | 取值 | 说明 |
|----------|--------|------|------|
| MCP 工具 | `setcode` | `0` / `1` | 不传默认 `0`（深市） |
| REST API | `market` | `0` / `1` | 不传默认 `0`（深市） |
| WebSocket | 前缀 `SH` / `SZ` | `SH600000` / `SZ000001` | 无前缀默认深市 |

> **经验法则**：A 股股票代码段 `000001`~`000009` 是深市老股票（平安银行等），`600000`~`609999` 是沪市老股票。指数代码如 `000001`（上证综指）需要设 `setcode=1`。

## 运行模式

| 模式 | 命令 | 说明 |
|------|------|------|
| MCP Stdio | `./go-tdx-mcp` | 标准 MCP 协议，与 Claude Desktop / Cursor / Windsurf 集成 |
| MCP SSE | `./go-tdx-mcp --sse --port=8000` | SSE 流式模式，远程 HTTP 调用 |
| MCP Streamable | `./go-tdx-mcp --streamable-http --port=8000` | Streamable HTTP 模式 |
| **混合模式** | `./go-tdx-mcp --web --port=8000` | MCP + REST API + WebSocket 三合一 |

混合模式启动后：
- API 文档首页：`http://localhost:8000/`
- MCP Streamable HTTP：`http://localhost:8000/mcp`
- 健康检查：`http://localhost:8000/api/v1/health`
- 实时行情推送：`ws://localhost:8000/ws/realtime/000001`

## 快速开始

```bash
git clone https://github.com/647133036/go-TDX-MCP.git
cd go-TDX-MCP
go build -o go-tdx-mcp .
```

### 配置

创建 `config.json`（所有字段可选，**无需 token 即可运行**）：

```json
{
  "timeout": 30
}
```

或通过环境变量：

| 变量 | 说明 |
|------|------|
| `TDX_HOST` | 通达信 TCP 服务器地址（可选，不配则自动选择最快节点） |
| `TDX_PORT` | 通达信 TCP 服务器端口（可选） |
| `TDX_TOKEN` | 通达信 HTTP API Token（可选，仅用于高级功能如 RAG 问答/自然语言选股） |

环境变量优先级高于配置文件。

### 集成 Claude Desktop

在 `claude_desktop_config.json` 中添加：

```json
{
  "mcpServers": {
    "tdx-finance": {
      "command": "/path/to/go-tdx-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

> 无需配置 `TDX_TOKEN`。默认使用通达信 TCP 直连（自动选择最快节点），覆盖实时行情、K 线、板块、缠论、指标等全部功能。如需 RAG 问答/自然语言选股，可添加 `TDX_TOKEN`。

## 工具分类

| 类别 | 数量 | 核心能力 |
|------|------|----------|
| Core（核心） | 17 | 实时报价、K 线、股票信息、选股、指标选择 |
| Expanded（扩展） | 132 | 板块行情、资金流、缠论、回测、财务、公告、离线数据 |
| V3（高级） | 18 | 市场概览、板块资金流、涨跌停、财务指标、宏观、舆情、爬虫 |
| New（新增） | 276 | 加密货币、基金净值、融资融券、龙虎榜、可转债、期货、因子计算、选股扫描、量化回测组合、投资组合优化、RAG 查询、缠论全套、AI 分析、表格解析、新闻情感、东财增强数据 |
| **合计** | **443** | |

### 代表性工具

```
tdx_kline             A 股 K 线（周期/除权/数量可选）
tdx_quotes            实时报价（买盘/卖盘/五档）
tdx_kline_data        K 线数据（多周期/复权）
tdx_capital_flow      个股/板块资金流向
tdx_factor_compute    因子计算（MA/MACD/RSI/ATR/布林带等）
tdx_factor_analyze    因子分析
tdx_screen_scan       选股扫描（策略组合）
tdx_enhanced_backtest 量化回测（单策略/多策略/组合模式）
tdx_chanlun_merge_klines   缠论分型/K 线合并
tdx_chanlun_build_bi       缠论笔构建
tdx_chanlun_build_zhong_shu 缠论中枢
tdx_chanlun_find_mai_mai_dian 缠论买卖点
tdx_tecrypto_data    加密货币行情
tdx_tecrypto_kline   加密货币 K 线
tdx_fund_nav         基金净值
tdx_margin_trade     融资融券
tdx_dragon_tiger     龙虎榜
tdx_convertible_bond 可转债
tdx_futures_quote    期货行情
tdx_northbound_flow  北向资金
tdx_northbound_top10 北向持股 Top10
tdx_market_overview   市场概览
tdx_security_filter  证券筛选
tdx_stock_basic_info 个股基本信息
tdx_news_search      新闻搜索
tdx_eastmoney_*      东方财富增强数据系列
tdx_table_parser_*   表格解析工具系列
tdx_ocr_recognize    OCR 图像识别
```

## REST API

混合模式下所有 MCP 工具均有对应的 REST 端点，位于 `/api/v1/` 路径下。

### 行情数据

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/quotes` | GET | `codes` | 实时报价（支持多代码） |
| `/api/v1/bars` | GET | `code`, `market`, `period`, `count` | K 线数据 |
| `/api/v1/symbol-info` | GET | `code`, `market` | 标的基本信息 |
| `/api/v1/quote-list` | GET | `count`, `sort_type` | 行情列表 |
| `/api/v1/market-stat` | GET | - | 市场统计 |

### 技术指标

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/indicator/list` | GET | - | 指标列表（34 个） |
| `/api/v1/indicator/compute` | POST | data, indicators | 指标计算 |
| `/api/v1/indicator/compute_all` | GET | `code`, `market`, `indicators` | 批量计算 |

### 缠论与回测

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/chanlun/analyze` | GET | `code`, `market`, `period`, `count` | 缠论分析（分型/笔/中枢/买卖点） |
| `/api/v1/backtest/run` | GET | `code`, `market`, `strategy`, `count` | 量化回测 |

### 财务与公告

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/financial/report` | GET | `code`, `type` | 财务报表（利润表/资产负债表/现金流量表） |
| `/api/v1/announcements` | GET | `code`, `count` | 公告列表 |

### 资金与板块

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/capital-flow` | GET | `code`, `market` | 资金流向 |
| `/api/v1/auction` | GET | `code`, `market` | 集合竞价 |
| `/api/v1/unusual` | GET | `market`, `count` | 异动监控 |
| `/api/v1/board/list` | GET | `board_type`, `top_n` | 板块列表 |
| `/api/v1/board/members` | GET | `board_symbol`, `count` | 板块成分股 |
| `/api/v1/block` | GET | `filename` | 板块文件数据 |
| `/api/v1/market-overview` | GET | - | 市场概览 |

### 扩展市场

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/ex/markets` | GET | - | 扩展市场列表 |
| `/api/v1/ex/quote` | GET | `ex_market`, `code` | 扩展市场报价 |
| `/api/v1/ex/bars` | GET | `ex_market`, `code` | 扩展市场 K 线 |

### 宏观数据

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/macro-data` | GET | `indicator`, `count` | 宏观经济数据（CPI/GDP/PMI/M2/SHIBOR 等） |

### 综合爬虫

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/scraper` | GET | - | 全量股票列表 |
| `/api/v1/scraper/sector-boards` | GET | `board_type` | 板块数据 |
| `/api/v1/scraper/northbound-flow` | GET | `days` | 北向资金流向 |
| `/api/v1/scraper/northbound-stocks` | GET | `count` | 北向持股明细 |
| `/api/v1/scraper/northbound-holders` | GET | `count` | 北向持仓机构 |
| `/api/v1/scraper/margin-trade` | GET | - | 融资融券余额 |
| `/api/v1/scraper/fund-search` | GET | `keyword`, `page_size` | 基金搜索 |
| `/api/v1/scraper/fund-nav` | GET | `code`, `count` | 基金净值 |
| `/api/v1/scraper/fund-holding` | GET | `code` | 基金持仓 |
| `/api/v1/scraper/hkus-quote` | GET | `code`, `market` | 港股行情 |
| `/api/v1/scraper/crypto` | GET | `symbols` | 加密货币行情 |

### 离线数据

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/offline/home` | GET | - | 检测 TDX 数据目录 |
| `/api/v1/offline/daily` | GET | `code`, `market` | 日线数据 |
| `/api/v1/offline/min` | GET | `code`, `market` | 分钟线数据 |

### WebSocket 实时推送

| 端点 | 协议 | 说明 |
|------|------|------|
| `/ws/realtime/{symbol}` | WebSocket | 3 秒轮询推送实时行情 |

用法：`ws://host/ws/realtime/000001`（深市，默认）或 `ws://host/ws/realtime/SH600000`（沪市）/ `ws://host/ws/realtime/SZ000001`（深市显式指定）

> 注意：`000001` 不加前缀默认为深市平安银行，查上证指数需使用 `SH000001`。

### 系统信息

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/server-info` | GET | 服务器信息 |
| `/` | GET | API 文档首页 |

## 数据源

| 数据源 | 用途 | 状态 |
|--------|------|------|
| **TDX TCP** | 通达信行情服务器（实时行情、K 线、财务、缠论数据） | 可用，**无需 token** |
| **TDX HTTP (TQLEX)** | 通达信 HTTP 网关（tdxhub.icfqs.com:7615，RAG 问答、自然语言选股） | **需 TDX_TOKEN** |
| **东方财富 (push2delay)** | 实时行情、板块、资金流（延迟数据） | 可用 |
| **东方财富 (datacenter)** | 财务数据、宏观数据、公告 | 可用 |
| **新浪财经** | 行情数据补充 | 可用 |
| **腾讯证券** | 融资融券数据 | 可用 |
| **本地 TDX 数据** | vipdoc 目录下的日线/分钟线/板块文件 | 可选 |
| **Binance API** | 加密货币行情 | 可用 |
| **CoinGecko API** | 加密货币数据备用源 | 可用 |

> 注意：`push2his.eastmoney.com` 已被完全阻断，不可用。所有历史 K 线数据通过 TDX 数据源获取。

## 技术栈

- **语言**：Go 1.26+
- **MCP 协议**：mark3labs/mcp-go
- **HTTP/WebSocket**：gorilla/mux + gorilla/websocket
- **TDX 协议**：通达信二进制协议（TCP 直连）+ HTTP TQLEX 网关
- **爬虫**：net/http、goquery
- **计算引擎**：自研因子计算、缠论分析、量化回测引擎

## 项目结构

```
go-TDX-MCP/
├── main.go               # 入口：MCP/Web 双模式
├── config.json           # 配置文件
├── web/
│   ├── server.go         # Web API + WebSocket 服务器
│   └── server_test.go    # 回归测试
├── tdx/
│   ├── client.go         # HTTP TQLEX 客户端
│   ├── tcp_client.go     # TDX TCP 客户端
│   ├── unified_client.go # 统一客户端（TCP + HTTP 智能路由）
│   ├── unified_bridge.go # 桥接层（TCP ↔ HTTP 格式转换）
│   ├── collector.go      # 多主机数据采集器
│   ├── tools.go          # 核心工具集（17 个）
│   ├── tools_expanded.go # 扩展工具集（132 个）
│   ├── tools_v3.go       # V3 工具集（18 个）
│   ├── tools_new.go      # 新增工具集（276 个）
│   └── strategies.go     # 量化策略定义
├── indicator/            # 技术指标计算引擎（34 个指标）
├── factor/               # 因子计算引擎
├── backtest/             # 量化回测引擎
├── chanlun/              # 缠论分析引擎（分型/笔/中枢/买卖点）
├── finance/              # 财务报表解析
├── offline/              # 离线数据读写（TDX vipdoc）
├── scraper/              # 网页数据爬虫（东方财富/新浪/腾讯）
├── portfolio/            # 投资组合优化
├── screen/               # 选股扫描引擎
└── vipdoc/               # 本地 TDX 数据目录（可选）
```

## 测试

```bash
go build ./...
go test ./...
go vet ./...
```

## Changelog

### v1.0.2（2026-08-04）
- 修复 K 线 HTTP ListItem 解析字段映射错位：Close/High/Low 与 Vol/Amount 索引颠倒
- 修复 web 层 K 线周期编码：`klinePeriodToCode` 与 `tdx.PeriodToCode` 不一致导致 day/week/month 返回错误周期
- 修复 `fetchKlines` 无法解析 TQLEX ListItem 响应，始终 fallback 到被阻断的 eastmoney 的问题
- 修复 `tdx/tools_expanded.go` / `tdx/unified_bridge.go` 中 5 处 `len(fields)<6` 却访问 `fields[8]` 的越界 panic 风险
- 修复 M2 宏观数据字段名错误：API 实际字段为 `BASIC_CURRENCY`（非 `M2`），`CURRENCY`（非 `M1`）
- 新增 `web/server_test.go`、`tdx/listitem_parse_test.go` 回归测试
- 全部数据采集接口实测验证通过（行情/K 线/财务/板块/缠论/指标/回测/资金流/港股/基金/宏观/公告/加密/北向/WebSocket）

### v1.0.1（2026-08-02）
- 统一 K 线数据解析路径：TCP SecurityBar ↔ HTTP ListHead/ListItem 双格式自动适配
- 修复 `tdx_kline_data` 的 HTTP fallback 使用不存在的 `TdxShare.PBQuotes` entry
- 修复 `KlineQuery` 在 TCP 不可用时跳过 HTTP fallback 的逻辑缺陷
- 新增 WebSocket 实时行情推送（`/ws/realtime/{symbol}`，3 秒轮询）
- 修复 `parseKlineBars` / `parseKlineBarsToDayBars` / `parseChanlunKlines` 缺 HTTP 格式支持
- 修复 `parseBarsFromResponse` 缺 HTTP ListHead/ListItem 格式支持
- 旧版 `HandleKline`（tools.go）补 TCP 格式处理

### v1.0.0
- 初始版本：215+ 个 MCP 工具 + 45+ 个投资技能
- 支持 Stdio / SSE / Streamable HTTP / Web 混合模式
- 覆盖 A 股、港股、美股、加密货币、期货、基金全市场数据

## 许可证

MIT