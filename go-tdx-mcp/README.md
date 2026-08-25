# TDX Finance MCP v1.0.5

通达信金融数据 MCP 服务器，提供 A 股、港股、美股、加密货币、期货、基金等多市场金融数据服务。

**215 个 MCP 工具** + **45+ 个投资技能**，覆盖实时行情、K 线、技术指标、缠论分析、量化回测、资金流向、板块分析、基金数据、宏观数据、新闻情感等全场景。混合模式下另提供约 100 个 REST API 端点与 WebSocket 实时推送。

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
cd go-TDX-MCP/go-tdx-mcp
go build -o go-tdx-mcp .
```

### 配置

创建 `config.json`（完整字段）：

```json
{
  "token": "your_tdx_token",
  "timeout": 30,
  "web_port": 8000,
  "tdx_host": "",
  "tdx_port": 0,
  "db_path": ""
}
```

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `token` | 通达信 HTTP API Token（可选，部分免费数据源不需要） | 空 |
| `timeout` | 请求超时秒数 | 30 |
| `web_port` | Web/混合模式监听端口 | 8000 |
| `tdx_host` | 通达信 TCP 服务器地址（留空自动选择最快主机） | 空 |
| `tdx_port` | 通达信 TCP 服务器端口 | 0 |
| `db_path` | 策略存储数据库路径 | `~/.tdx-mcp/strategies.db` |

环境变量：

| 变量 | 说明 |
|------|------|
| `TDX_TOKEN` | 通达信 HTTP API Token，优先级高于配置文件 |

> 加密货币（Binance）、基金净值、融资融券、龙虎榜、可转债、期货等数据源免费无需 Token。

### 集成 Claude Desktop

在 `claude_desktop_config.json` 中添加：

```json
{
  "mcpServers": {
    "tdx-finance": {
      "command": "/path/to/go-tdx-mcp",
      "args": [],
      "env": {
        "TDX_TOKEN": "your_token"
      }
    }
  }
}
```

## 工具分类

| 类别 | 数量 | 核心能力 |
|------|------|----------|
| Core（核心） | 6 | 实时报价、K 线、自然语言检索、智能选股、指标选择、F10 内部 API |
| Expanded（扩展） | 64 | 分时/逐笔、板块行情、资金流、技术指标计算、缠论、回测、财务、公告、离线数据、扩展市场（港美股期货） |
| V3（高级） | 8 | 市场概览、板块资金流、涨跌停、财务指标、宏观、舆情、网页表格爬虫、自然语言问答 |
| New（新增） | 137 | 加密货币、基金、融资融券、龙虎榜、可转债、期货、北向资金、因子计算、选股扫描、增强回测、投资组合、缠论全套、表格解析、OCR、RAG、东财增强数据系列 |
| **合计** | **215** | 全部工具均有对应处理器，注册成功率 100% |

### 代表性工具

```
tdx_quotes            A 股实时行情（报价/五档盘口/涨跌幅）
tdx_kline             A 股 K 线（多周期/前复权）
tdx_kline_data        K 线数据（日/周/月/分钟）
tdx_quote_realtime    实时报价（单只/多只）
tdx_capital_flow      个股主力资金流向
tdx_factor_compute    因子计算（momentum/technical/volume 等）
tdx_factor_analyze    因子分析（IC/分层收益/多空）
tdx_screen_scan       技术信号扫描（MACD金叉/KDJ金叉/放量等）
tdx_enhanced_backtest 增强回测（16 策略+滑点+执行模拟）
tdx_chanlun_merge_klines   缠论 K 线合并
tdx_chanlun_build_bi       缠论笔构建
tdx_chanlun_build_zhongshu 缠论中枢
tdx_chanlun_find_maimaidian 缠论买卖点
tdx_tecrypto_data     加密货币行情（Binance）
tdx_tecrypto_kline    加密货币 K 线
tdx_fund_nav          基金净值
tdx_margin_trade      融资融券
tdx_dragon_tiger      龙虎榜
tdx_convertible_bond  可转债
tdx_futures_quote     期货报价
tdx_northbound_flow   北向资金实时净流入
tdx_northbound_top10  北向资金十大成交
tdx_market_overview   市场概览（涨跌家数/涨停/跌停）
tdx_security_filter   证券筛选
tdx_stock_basic_info  个股基本信息
tdx_news_search       新闻搜索
tdx_news_sentiment    新闻情感分析
tdx_eastmoney_realtime_quote  东财实时报价
tdx_eastmoney_capital_flow   东财资金流向
tdx_table_parser_url  表格解析（URL/HTML→CSV/JSON）
tdx_ocr_recognize     OCR 图像识别
tdx_rag_query         RAG 智能问答
wenda_macro_query     自然语言宏观/策略问答
tdx_macro_data        宏观经济数据（CPI/GDP/PMI/M2）
```

## REST API

混合模式下约 100 个端点，位于 `/api/v1/` 路径下，均与 MCP 工具对应。

### 行情与 K 线

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/v1/quotes` | GET | `codes` | 实时报价（多代码） |
| `/api/v1/bars` | GET | `code`, `market`, `period`, `count` | K 线数据 |
| `/api/v1/bars/index` | GET | `code`, `market`, `period`, `count` | 指数 K 线 |
| `/api/v1/symbol-info` | GET | `code`, `market` | 标的基本信息 |
| `/api/v1/quote-list` | GET | `count`, `sort_type` | 行情列表 |
| `/api/v1/market-stat` | GET | - | 市场统计 |
| `/api/v1/market-overview` | GET | - | 市场概览 |
| `/api/v1/market/strength` | GET | - | 市场强度 |
| `/api/v1/minute`、`/api/v1/minute/history`、`/api/v1/minute/multi` | GET | `code`, `market` | 分时数据 |
| `/api/v1/transaction`、`/api/v1/transaction/history` | GET | `code`, `market` | 逐笔成交 |

### 技术指标与因子

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/indicator/list` | GET | 指标列表 |
| `/api/v1/indicator/compute` | GET/POST | 指标计算 |
| `/api/v1/indicator/compute_all` | GET | 批量计算 |
| `/api/v1/factor/list` | GET | 因子列表 |
| `/api/v1/factor/compute` | GET | 因子计算 |
| `/api/v1/factor/analyze` | GET | 因子分析 |
| `/api/v1/volume-profile` | GET | 量价分析 |

### 缠论与回测

| 端点 | 说明 |
|------|------|
| `/api/v1/chanlun/analyze`、`/api/v1/chanlun/multi` | 缠论分析（分型/笔/中枢/买卖点） |
| `/api/v1/backtest/run`、`/run/async`、`/run-all` | 回测执行 |
| `/api/v1/backtest/tasks`、`/tasks/{id}` | 回测任务状态 |
| `/api/v1/backtest/optimize`、`/portfolio`、`/multi-strategy` | 组合回测 |
| `/api/v1/backtest/signal-scan`、`/signal-rank` | 信号扫描/排名 |
| `/api/v1/backtest/strategies`、`/api/v1/strategies` | 策略管理（SQLite 存储） |

### 财务与公告

| 端点 | 说明 |
|------|------|
| `/api/v1/financial/report`、`/file-list`、`/records`、`/api/v1/finance` | 财务报表与文件 |
| `/api/v1/announcements` | 公告列表 |
| `/api/v1/company/category`、`/content` | 公司资料 |

### 资金与板块

| 端点 | 说明 |
|------|------|
| `/api/v1/capital-flow`、`/fund-flow`、`/fund-flow/history` | 资金流向 |
| `/api/v1/auction` | 集合竞价 |
| `/api/v1/unusual` | 异动监控 |
| `/api/v1/board/list`、`/members`、`/ranking`、`/change-ranking`、`/summary` | 板块数据 |
| `/api/v1/belong-board`、`/block` | 个股板块归属/板块文件 |

### 扩展市场（港美股/期货）

| 端点 | 说明 |
|------|------|
| `/api/v1/ex/markets` | 扩展市场列表 |
| `/api/v1/ex/quote`、`/quotes`、`/quotes-list` | 扩展市场报价 |
| `/api/v1/ex/bars`、`/minute`、`/transaction`、`/transaction-all`、`/chart-sampling` | 扩展市场 K 线/分时 |
| `/api/v1/ex/list` | 扩展市场标的列表 |

### 爬虫与宏观数据

| 端点 | 说明 |
|------|------|
| `/api/v1/scraper` | 全量股票列表 |
| `/api/v1/scraper/sector-boards` | 板块数据 |
| `/api/v1/scraper/northbound-flow`、`/northbound-stocks`、`/northbound-holders` | 北向资金 |
| `/api/v1/scraper/fund-nav`、`/fund-holding`、`/fund-search` | 基金数据 |
| `/api/v1/scraper/margin-trade` | 融资融券 |
| `/api/v1/scraper/hkus-quote`、`/crypto` | 港美股/加密货币 |
| `/api/v1/dragon-tiger` | 龙虎榜 |
| `/api/v1/convertible-bond` | 可转债列表 |
| `/api/v1/futures-quote` | 期货行情（新浪） |
| `/api/v1/macro-data`、`/api/v1/news-sentiment` | 宏观/舆情 |

### 离线数据（依赖本地通达信数据）

| 端点 | 说明 |
|------|------|
| `/api/v1/offline/home` | 检测 TDX 数据目录 |
| `/api/v1/offline/daily`、`/min` | 日线/分钟线 |
| `/api/v1/offline/gbbq` | 股本变迁 |
| `/api/v1/offline/blocks` | 板块文件（白名单 4 个文件名） |
| `/api/v1/offline/ex-files`、`/ex-daily` | 扩展市场离线数据 |

> 离线端点依赖本地通达信客户端数据（`vipdoc` 目录），无 TDX 安装时返回 404。

### 证券与系统

| 端点 | 说明 |
|------|------|
| `/api/v1/security/list`、`/list-all`、`/security-count` | 证券列表/数量 |
| `/api/v1/xdxr`、`/history-orders`、`/index/info`、`/index/momentum` | 除权除息/历史/指数 |
| `/api/v1/server/hosts`、`/test`、`/switch`、`/server-info` | 服务器管理（`/test` 为白名单模式） |

### WebSocket 实时推送

| 端点 | 协议 | 说明 |
|------|------|------|
| `/ws/realtime/{symbol}` | WebSocket | 3 秒轮询推送实时行情 |

用法：`ws://host/ws/realtime/000001`（深市，默认）或 `ws://host/ws/realtime/SH600000`（沪市）。

## 安全

- **路径穿越防护**：`offline/daily`、`offline/min`、`offline/gbbq`、`offline/blocks`、`offline/ex-daily` 全面移除可绕过校验的自由参数（`path`/`vipdoc`），改为白名单校验 + `filepath.Join` 前缀限制；`offline/gbbq` 强制 `code` 6 位数字；`offline/blocks` 限定 4 个合法文件名（`block_zs.dat`/`block_gn.dat`/`block_fz.dat`/`block_fy.dat`）。
- **SSRF 防护**：`server/test` 和 `fetch-urls` 改用预设白名单 host，拒绝非白名单目标，阻止内网探测和云元数据访问。

## 数据源

| 数据源 | 用途 | 是否需 Token |
|--------|------|--------------|
| **TDX TCP** | 通达信行情服务器（实时行情、K 线、财务、缠论数据），自动选择最快主机 | 否 |
| **TDX HTTP (TQLEX)** | 通达信 HTTP 网关（tdxhub.icfqs.com:7615，行情、K 线、基础数据） | 是 |
| **东方财富 (push2delay)** | 实时行情、板块、资金流（延迟数据） | 否 |
| **东方财富 (datacenter)** | 财务、宏观、龙虎榜、可转债、北向、涨跌停、大宗交易 | 否 |
| **新浪财经** | A 股/港股/美股行情、财务报表、期货行情、融资融券、大宗交易 | 否 |
| **腾讯证券** | 融资融券 | 否 |
| **Binance API** | 加密货币行情与 K 线 | 否 |
| **CoinGecko API** | 加密货币数据备用源 | 否 |
| **巨潮资讯网** | 公司公告 | 否 |
| **天天基金网** | 基金净值/持仓 | 否 |
| **同花顺问财 (iwencai)** | 自然语言选股/板块/资讯 | 否 |
| **本地 TDX 数据** | vipdoc 目录下的日线/分钟线/板块文件 | 否 |

> 注意：`push2his.eastmoney.com` 已被阻断不可用，历史 K 线通过 TDX 数据源获取。

## 技术栈

- **语言**：Go 1.26+
- **MCP 协议**：mark3labs/mcp-go
- **HTTP/WebSocket**：gorilla/websocket + 标准库 net/http
- **TDX 协议**：通达信二进制协议（TCP 直连）+ HTTP TQLEX 网关
- **爬虫**：net/http、goquery
- **计算引擎**：自研因子计算、缠论分析、量化回测、投资组合优化引擎
- **存储**：SQLite（策略持久化）

## 项目结构

```
go-tdx-mcp/
├── main.go               # 入口：MCP/Web/SSE/Streamable 多模式
├── config.example.json   # 配置示例
├── test_all_api.py       # 功能测试脚本
├── web/
│   ├── server.go         # Web API + WebSocket 服务器（约 100 路由）
│   ├── handlers_new.go   # 扩展端点处理器
│   ├── handlers_missing.go
│   ├── strategy_store.go # 策略 SQLite 存储
│   └── server_test.go    # 回归测试
├── tdx/
│   ├── client.go         # HTTP TQLEX 客户端
│   ├── tcp_client.go     # TDX TCP 客户端（K 线获取/自动重连）
│   ├── unified_client.go # 统一客户端（TCP + HTTP 智能路由）
│   ├── unified_bridge.go # 桥接层（TCP ↔ HTTP 格式转换）
│   ├── collector.go      # 多主机数据采集器
│   ├── tools.go          # 核心工具集（6 个）
│   ├── tools_expanded.go # 扩展工具集（64 个）
│   ├── tools_v3.go       # V3 工具集（8 个）
│   ├── tools_new.go      # 新增工具集（137 个）
│   ├── strategies.go     # 量化策略定义
│   └── types.go          # 请求/响应类型定义
├── indicator/            # 技术指标计算引擎（30+ 指标）
├── factor/               # 因子计算引擎
├── backtest/             # 量化回测引擎
├── chanlun/              # 缠论分析引擎（分型/笔/中枢/买卖点）
├── finance/              # 财务报表解析
├── offline/              # 离线数据读写（TDX vipdoc）
├── scraper/              # 网页数据爬虫（东方财富/新浪/腾讯/同花顺）
├── portfolio/            # 投资组合优化与风险
├── screen/               # 选股扫描引擎
├── vipdoc/               # 本地 TDX 数据目录（可选，运行时生成）
└── cmd/stress_test/      # 压力测试工具
```

## 测试

```bash
go build ./...
go test ./...
go vet ./...
```

功能测试（混合模式启动后）：

```bash
python3 test_all_api.py
```

单元测试覆盖 tdx/indicator/factor/backtest/chanlun/finance/scraper/web 等包，共 167 个测试函数。

## Changelog

### v1.0.5（2026-08-24）
- 修复财务报表字段名 GBK 乱码：引入 golang.org/x/text/encoding/simplifiedchinese 真正 GBK→UTF8 转码
- 修复财务报表请求超时：原 `Timeout: 15`（15 纳秒）改为 30 秒 + 失败重试 2 次 + UA/Referer
- 修复龙虎榜数据源：`RPT_BILLBOARD_DAILYDETAIL` 报表不存在 → 改用 `RPT_BILLBOARD_LIST`，字段映射修正（BOARD_TYPE/TOTAL_BUY_AMT/TOTAL_SELL_AMT，净买入=买-卖）
- 修复可转债字段映射错位：正股代码误取债券代码 → 改取 `CONVERT_STOCK_CODE`；发行规模/转股价/起息/到期/转股开始 6 字段全部修正
- 修复期货行情数据源：腾讯 qt.gtimg.cn 不支持国内期货合约 → 改用新浪 hq.sinajs.cn（AU0/CU0 等主力连续），名称 GBK 转码 + high/low 兜底
- 修复公告接口失效：cninfo `hisAnnouncement/query` 返回空 → 改用 `topSearch/detailOfQuery`
- 修复 `HandleFuturesQuote` 循环变量 `err` 污染外层导致丢弃已采集数据
- `HandleFinancialMetrics` 从仅查利润表改为综合三表（利润/资产负债/现金流量）合并
- 新增 3 个 REST 端点：`/api/v1/dragon-tiger`、`/api/v1/convertible-bond`、`/api/v1/futures-quote`
- 实测 5 类端点均返回干净数据（go test 验证通过）

### v1.0.4（2026-08-23）
- 文档修正：README 工具数量更正为实际值（215 个，原误标 443）；分类表更新为 Core 6 / Expanded 64 / V3 8 / New 137
- 文档修正：REST API 端点表与项目结构对齐 `web/server.go` 实际路由
- 版本同步：main.go 版本号从 1.0.0 更新至 1.0.4
- 数据源表补充：新增同花顺问财、巨潮、天天基金等免费源并标注是否需 Token

### v1.0.3（2026-08-23）
- 安全加固：`offline/gbbq` 移除 `path` 参数，强制 `code` 6 位数字 + `filepath.Join` 前缀校验
- 安全加固：`server/test` 改用白名单模式，消除 SSRF 漏洞
- 安全加固：`offline/blocks` 限定 4 个合法文件名白名单
- 安全加固：`offline/daily`/`offline/min`/`offline/ex-daily` 移除 `vipdoc` 参数，增加白名单校验
- 修复：`fund-flow` 500 错误（MAC 服务器 broken pipe）→ `reconnectMAC` 先 `Disconnect()` 再重连
- 修复：`reconnectMain`/`reconnectEx` 重连不彻底 → 先 `Disconnect()` 再重连
- 修复：`ensureMACConnected` 缺少互斥锁保护 → 新增 `macMu` 锁
- `fetchKlines` 增加 TCP 原生 K 线获取回退（三级数据源：TCP → 离线 → HTTP）

### v1.0.2（2026-08-04）
- 修复 K 线 HTTP ListItem 解析字段映射错位：Close/High/Low 与 Vol/Amount 索引颠倒
- 修复 web 层 K 线周期编码与 `tdx.PeriodToCode` 不一致
- 修复 `fetchKlines` 无法解析 TQLEX ListItem 响应
- 修复 `tools_expanded.go`/`unified_bridge.go` 中 5 处 `len(fields)<6` 访问 `fields[8]` 越界 panic
- 修复 M2 宏观数据字段名：实际为 `BASIC_CURRENCY`（非 `M2`）
- 新增 `web/server_test.go`、`tdx/listitem_parse_test.go` 回归测试

### v1.0.1（2026-08-02）
- 统一 K 线数据解析路径：TCP SecurityBar ↔ HTTP ListHead/ListItem 双格式自动适配
- 修复 `tdx_kline_data` HTTP fallback 使用不存在的 `TdxShare.PBQuotes` entry
- 修复 `KlineQuery` 在 TCP 不可用时跳过 HTTP fallback 的逻辑缺陷
- 新增 WebSocket 实时行情推送
- 修复多个 K 线解析函数缺 HTTP 格式支持

### v1.0.0
- 初始版本：215 个 MCP 工具 + 45+ 个投资技能
- 支持 Stdio / SSE / Streamable HTTP / Web 混合模式
- 覆盖 A 股、港股、美股、加密货币、期货、基金全市场数据

## 许可证

MIT
