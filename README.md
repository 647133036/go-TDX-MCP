# TDX Finance MCP

TDX Finance MCP 是一个基于 Go 语言实现的 MCP（Model Context Protocol）服务器，提供 A 股全量行情、K 线、技术因子、回测、缠论分析、基本面数据、宏观数据、港美股行情、Web 数据抓取等一站式金融服务能力，并附带一个投资数据爬虫子项目（investing-scrapers）。

- **版本**: v1.0.3
- **语言**: Go 1.26.0
- **协议**: MCP（Stdio / SSE / Streamable HTTP）
- **MCP 工具**: 212 个
- **投资技能 Prompts**: 45 个
- **REST API**: 50+ 端点

---

## 目录

- [快速开始](#快速开始)
- [运行模式](#运行模式)
- [配置](#配置)
- [数据源架构](#数据源架构)
- [MCP 工具列表](#mcp-工具列表)
- [REST API](#rest-api)
- [TDX 公式引擎](#tdx-公式引擎)
- [回测引擎](#回测引擎)
- [缠论分析](#缠论分析)
- [因子分析](#因子分析)
- [离线数据同步](#离线数据同步)
- [投资数据爬虫](#投资数据爬虫)
- [技术栈](#技术栈)
- [开发测试](#开发测试)

---

## 快速开始

### 安装依赖

```bash
cd go-tdx-mcp
go mod download
```

### 编译

```bash
go build -o go-tdx-mcp .
```

### 运行（Stdio 模式）

```bash
./go-tdx-mcp
```

通过 Stdio 模式接入 Claude Desktop、Cursor、Continue 等 MCP 客户端。在 MCP 客户端配置中添加：

```json
{
  "mcpServers": {
    "tdx-finance": {
      "command": "./go-tdx-mcp",
      "args": [],
      "cwd": "/path/to/go-tdx-mcp"
    }
  }
}
```

### 运行（HTTP 模式）

```bash
# SSE 模式
./go-tdx-mcp --sse

# Streamable HTTP 模式
./go-tdx-mcp --streamable-http

# 综合模式（MCP + REST API）
./go-tdx-mcp --web
```

HTTP 模式默认监听 `0.0.0.0:8000`，可通过 `--port=<port>` 指定端口。

---

## 运行模式

| 模式 | 命令 | 端口 | 说明 |
|------|------|------|------|
| Stdio | `./go-tdx-mcp` | - | MCP 标准 IO 协议，供 MCP 客户端调用 |
| SSE | `./go-tdx-mcp --sse` | 8000 | SSE Server-Sent Events，`/sse` + `/message` |
| Streamable HTTP | `./go-tdx-mcp --streamable-http` | 8000 | MCP Streamable HTTP 协议，`/mcp` |
| Combined | `./go-tdx-mcp --web` | 8000 | MCP `/mcp` + REST API `/` |

Combined 模式下同时提供：
- `http://localhost:8000/mcp` — MCP Streamable HTTP 端点
- `http://localhost:8000/` — REST API 文档入口
- `http://localhost:8000/api/v1/health` — 健康检查

---

## 配置

### 配置文件

```json
// config.json
{
  "token": "your-TDX_TOKEN",
  "timeout": 30,
  "web_port": 8000,
  "tdx_host": "",
  "tdx_port": 0
}
```

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `token` | TQLEX HTTP API 认证 Token | 空（仅用 TCP） |
| `timeout` | HTTP 请求超时（秒） | 30 |
| `web_port` | HTTP 服务端口 | 8000 |
| `tdx_host` | 自定义 TDX TCP 服务器地址 | 空（使用默认服务器池） |
| `tdx_port` | 自定义 TDX TCP 端口 | 0（使用默认 7709） |

启动时支持通过环境变量 `TDX_TOKEN` 传递 Token。

### 数据源连接

| 数据源 | 地址 | 说明 |
|--------|------|------|
| TQLEX HTTP | `http://tdxhub.icfqs.com:7615/TQLEX` | 需 Token，用于 F10、资金流等 |
| TDX TCP 默认 | `218.75.122.92:7709` | gotdx 协议，行情/K线/板块 |
| TDX TCP 备用池 | 多服务器自动切换 | 健康检查 + 故障转移 |
| RAG 服务 | `https://ai.icfqs.com:8965/v1/rag-entity-retrieve` | 实体检索 |
| 东方财富 | `push2delay.eastmoney.com` 等 | 免费行情/板块/财务/宏观 |
| 新浪财经 | `vipfx.sh.stockfinance.sina.com.cn` | K线备选/融资融券/大宗交易 |

---

## 数据源架构

```
┌─────────────────────────────────────────────────────┐
│                  TDX Finance MCP Server              │
├─────────────────────────────────────────────────────┤
│  UnifiedClient                                       │
│  ┌─────────────────────────────────────────────────┐ │
│  │  TQLEX HTTP Client  (tdxhub.icfqs.com:7615)     │ │
│  │  gotdx TCP Client   (218.75.122.92:7709)        │ │
│  │  EastMoney Scraper  (板块/财务/宏观/新闻)         │ │
│  │  Sina Finance      (K线备选/融资融券)             │ │
│  │  问财/同花顺爬虫   (表格数据 chromedp)           │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  本地计算引擎                                      │ │
│  │  Formula Engine  |  Indicator  |  Backtest       │ │
│  │  Chanlun         |  Factor     |  Portfolio      │ │
│  │  Offline Sync    |  Screen     |  Finance        │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  Web Server (REST API)                           │ │
│  │  WebSocket 实时行情推送                           │ │
│  └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

**UnifiedClient 数据流策略**：
1. **行情/K线**：优先 TCP 协议直连（最快），TCP 不可用时回退 Sina HTTP（无需 Token）
2. **板块数据**：优先 TCP，回退东方财富
3. **资金流/新闻/财务/宏观**：直接调用东方财富 HTTP（免费，无需 Token）
4. **F10 深度数据**：优先 TCP，回退 TQLEX HTTP（需 Token）
5. **NLP 选股**：TQLEX HTTP（需 Token），备选本地公式引擎/因子引擎

---

## MCP 工具列表

### 核心工具（6 个）

| 工具名 | 说明 |
|--------|------|
| `tdx_quotes` | 获取 A 股实时行情：报价、五档盘口、涨跌幅、成交量 |
| `tdx_kline` | 获取 K 线历史数据，支持日/周/月/分钟线及复权 |
| `tdx_lookup_stock` | 股票代码/名称模糊搜索 |
| `tdx_screener` | 通达信公式全市场选股 |
| `tdx_indicator_select` | 通达信指标选股（别名） |
| `tdx_api_data` | 统一 F10 API 调用：公司概况、盈利预测、龙虎榜等 |

### 扩展工具（64 个）

**行情与交易**
| 工具名 | 说明 |
|--------|------|
| `tdx_tick` | 分时行情 |
| `tdx_transaction` | 逐笔成交明细 |
| `tdx_quote_realtime` | 实时报价增强 |
| `tdx_quote_list_extended` | 报价列表增强 |

**K 线数据**
| 工具名 | 说明 |
|--------|------|
| `tdx_kline_extended` | K 线增强（多周期） |
| `tdx_daily_line` | 日线 |
| `tdx_week_line` | 周线 |
| `tdx_month_line` | 月线 |
| `tdx_5min_line` | 5 分钟线 |
| `tdx_15min_line` | 15 分钟线 |
| `tdx_30min_line` | 30 分钟线 |
| `tdx_60min_line` | 60 分钟线 |

**板块与概念**
| 工具名 | 说明 |
|--------|------|
| `tdx_board_list` | 行业/概念板块列表 |
| `tdx_board_members` | 板块成分股 |
| `tdx_belong_board` | 个股所属板块 |
| `tdx_board_ranking` | 板块涨跌幅排名 |
| `tdx_sector_ranking` | 行业板块排名 |
| `tdx_industry_ranking` | 概念板块排名 |

**资金流与异动**
| 工具名 | 说明 |
|--------|------|
| `tdx_capital_flow` | 主力资金流向 |
| `tdx_auction` | 集合竞价数据 |
| `tdx_unusual` | 异动监控 |
| `tdx_market_stat` | 市场统计概览 |
| `tdx_top_gainers` | 涨幅榜 |
| `tdx_top_losers` | 跌幅榜 |

**技术指标**
| 工具名 | 说明 |
|--------|------|
| `tdx_indicator_compute` | 批量技术指标计算 |
| `tdx_MACD` | MACD 指标 |
| `tdx_KDJ` | KDJ 指标 |
| `tdx_RSI` | RSI 指标 |
| `tdx_WR` | 威廉指标 |
| `tdx_BOLL` | 布林带 |
| `tdx_EMA` | 指数移动平均 |
| `tdx_DMA` | 差离指标 |
| `tdx_ASI` | 振动升降指标 |
| `tdx_VR` | 成交量比率 |
| `tdx_ROC` | 变动率指标 |
| `tdx_OBV` | 能量潮 |
| `tdx_MFI` | 资金流量指标 |
| `tdx_ADX` | 平均趋向指标 |
| `tdx_ARBR` | 人气意愿指标 |
| `tdx_CCI` | 顺势指标 |
| `tdx_DMI` | 趋向指标 |
| `tdx_technical_indicator` | 自定义技术指标 |

**缠论与回测**
| 工具名 | 说明 |
|--------|------|
| `tdx_chanlun` | 缠论分析（分型/笔/中枢/买卖点） |
| `tdx_backtest` | 策略回测 |

**信息公告**
| 工具名 | 说明 |
|--------|------|
| `tdx_announcement` | 公司公告（巨潮资讯网） |
| `tdx_financial` | 财务报表（利润表/资产负债表/现金流量表） |
| `tdx_stock_profile` | 个股画像 |

**港美股与期货**
| 工具名 | 说明 |
|--------|------|
| `tdx_ex_markets` | 扩展市场列表 |
| `tdx_ex_kline` | 扩展市场 K 线 |
| `tdx_ex_quote` | 扩展市场报价 |
| `tdx_ex_quote_list` | 扩展市场股票列表 |
| `tdx_ex_tick` | 扩展市场分时 |

**离线数据**
| 工具名 | 说明 |
|--------|------|
| `tdx_offline_home` | 检测 TDX 安装目录 |
| `tdx_offline_daily` | 读取本地日线文件 |
| `tdx_offline_min` | 读取本地分钟线文件 |
| `tdx_offline_gbbq` | 读取股本变迁数据 |
| `tdx_offline_blocks` | 读取本地板块数据 |
| `tdx_offline_ex_files` | 扩展市场文件列表 |
| `tdx_offline_ex_daily` | 读取扩展市场日线 |
| `tdx_offline_financial` | 读取本地财务数据 |
| `tdx_offline_sync_daily` | 同步单只股票日线 |
| `tdx_offline_sync_all` | 批量同步日线数据 |

**服务器信息**
| 工具名 | 说明 |
|--------|------|
| `tdx_server_info` | TDX 服务器状态 |
| `tdx_symbol_info` | 股票基本信息 |

### V3 工具（8 个）

| 工具名 | 说明 |
|--------|------|
| `tdx_market_overview` | 全市场涨跌家数统计、涨停/跌停/炸板数 |
| `tdx_sector_flow` | 板块资金流向（东方财富，无需 Token） |
| `tdx_top_gainers_losers` | 涨跌幅排行 + 量比/振幅/换手异动 |
| `tdx_financial_metrics` | 个股核心财务指标提取（EPS/ROE/营收/净利） |
| `tdx_macro_data` | 宏观经济数据（CPI/PMI/GDP） |
| `wenda_macro_query` | 自然语言宏观问答（自动采集指标 + 推理路径） |
| `tdx_news_sentiment` | 财经新闻情感分析 |
| `tdx_table_scraper` | 同花顺问财/通达信问小达/东方财富表格抓取 |

### 新增工具（20+ 个）

**TDX 协议数据工具**
| 工具名 | 说明 |
|--------|------|
| `tdx_quote_list` | 按市场获取报价列表 |
| `tdx_quote_batch` | 批量报价查询 |
| `tdx_kline_data` | K 线数据查询 |
| `tdx_fs_minute` | 分时数据 |
| `tdx_transaction_data` | 逐笔数据 |
| `tdx_security_filter` | 证券筛选 |
| `tdx_stock_basic_info` | 个股基本信息 |
| `tdx_stock_dividend_info` | 分红数据 |
| `tdx_stock_split_info` | 送转数据 |
| `tdx_ipo_calendar` | IPO 日历 |
| `tdx_stock_list_by_market` | 按市场获取股票列表 |
| `tdx_stock_list_by_sector` | 按板块获取股票列表 |
| `tdx_stock_list_by_industry` | 按行业获取股票列表 |
| `tdx_stock_list_by_exchange` | 按交易所获取股票列表 |
| `tdx_stock_list_by_status` | 按状态筛选股票 |
| `tdx_index_constituent_list` | 指数成分股 |
| `tdx_etf_list` | ETF 列表 |
| `tdx_etf_info` | ETF 详情 |
| `tdx_etf_holdings` | ETF 持仓 |
| `tdx_etf_net_value` | ETF 净值 |

**基本面与量化**
| 工具名 | 说明 |
|--------|------|
| `tdx_fundamental_filter` | 基本面筛选（PE/PB/ROE/营收增长） |
| `tdx_pe_percentile` | PE 百分位 |
| `tdx_pb_percentile` | PB 百分位 |
| `tdx_revenue_growth_rank` | 营收增长排名 |
| `tdx_profit_growth_rank` | 利润增长排名 |
| `tdx_roe_rank` | ROE 排名 |
| `tdx_debt_ratio_rank` | 资产负债率排名 |
| `tdx_insider_trading` | 内幕交易监测 |
| `tdx_shareholder_change` | 股东变更 |
| `tdx_margin_detail` | 融资融券明细 |
| `tdx_northbound_detail` | 北向持股明细 |
| `tdx_block_trade_detail` | 大宗交易明细 |
| `tdx_sector_rotation` | 板块轮动数据 |
| `tdx_market_breadth` | 市场广度 |
| `tdx_volume_price_analysis` | 量价分析 |

**缠论/因子/增强回测**
| 工具名 | 说明 |
|--------|------|
| `tdx_factor_info` | 因子信息 |
| `tdx_factor_report` | 因子分析报告 |
| `tdx_factor_forward_returns` | 因子前向收益 |
| `tdx_chanlun_fenxing` | 缠论分型 |
| `tdx_chanlun_bi` | 缠论笔 |
| `tdx_chanlun_zhongshu` | 缠论中枢 |
| `tdx_chanlun_maimaidian` | 缠论买卖点 |
| `tdx_chanlun_merge_klines` | 缠论 K 线合并 |
| `tdx_enhanced_backtest` | 增强回测（组合策略） |

**TE 金融数据/基金/融资融券/龙虎榜/可转债/期货**
| 工具名 | 说明 |
|--------|------|
| `tdx_te_financial` | TE 金融数据 |
| `tdx_fund_nav` | 基金净值 |
| `tdx_fund_nav_history` | 基金净值历史 |
| `tdx_fund_holding` | 基金持仓 |
| `tdx_margin_trading` | 融资融券数据 |
| `tdx_margin_trading_sina` | 融资融券（新浪） |
| `tdx_lh_bang` | 龙虎榜数据 |
| `tdx_convertible_bond` | 可转债数据 |
| `tdx_futures_quote` | 期货行情 |
| `tdx_futures_kline` | 期货 K 线 |

**新浪财经行情**
| 工具名 | 说明 |
|--------|------|
| `tdx_sina_quote` | 新浪个股行情 |
| `tdx_sina_index` | 新浪指数行情 |
| `tdx_sina_fund` | 新浪基金行情 |

**公式引擎**
| 工具名 | 说明 |
|--------|------|
| `tdx_formula_parse` | 解析通达信公式源码，返回 AST |
| `tdx_formula_execute` | 执行通达信公式，返回输出序列和信号 |
| `tdx_formula_list` | 列出公式引擎支持的所有内置函数 |

**RAG 增强工具**
| 工具名 | 说明 |
|--------|------|
| `tdx_rag_query` | 自然语言金融问答（RAG + 数据工具自动路由） |

### K 线周期代码

| 代码 | 周期 |
|------|------|
| 3 | 60 分钟 |
| 4 | 日线 |
| 5 | 周线 |
| 6 | 月线 |
| 9 | 1 分钟 |
| 10 | 5 分钟 |
| 11 | 15 分钟 |
| 12 | 30 分钟 |

### 市场代码

| 代码 | 市场 |
|------|------|
| 0 | 深圳 |
| 1 | 上海 |
| 2 | 北交所 |

---

## REST API

Combined 模式（`--web`）下提供以下 REST API 端点：

### 健康检查

```
GET /api/v1/health
```

### 行情数据

```
GET /api/v1/quotes?code=000001&market=0
GET /api/v1/quotes/batch?codes=000001,600000
```

### K 线数据

```
GET /api/v1/kline?code=000001&market=0&period=day&count=200&fq_type=0
```

### 技术指标

```
GET /api/v1/indicator/list
GET /api/v1/indicator/compute-all?code=000001&market=0&indicators=MACD,KDJ,RSI
POST /api/v1/indicator/compute  { "data": [...], "indicators": ["MACD"] }
```

### 缠论分析

```
GET /api/v1/chanlun?code=000001&market=0&period=day&count=200
```

### 策略回测

```
GET /api/v1/backtest?code=000001&market=0&strategy=ma_cross&count=2000
```

### 通达信公式

```
POST /api/v1/formula/parse    { "formula": "MA(CLOSE,5)" }
POST /api/v1/formula/execute  { "formula": "MACD", "code": "000001" }
GET /api/v1/formula/functions
```

### 财务数据

```
GET /api/v1/financial?code=000001&type=lrb
```

### 公司公告

```
GET /api/v1/announcement?code=000001&count=30
```

### 市场概览与宏观

```
GET /api/v1/market-overview
GET /api/v1/macro?indicator=CPI&count=12
```

### 爬虫工具

```
GET /api/v1/scraper?query=ROE%3E15%25&source=all
GET /api/v1/scraper/sector-boards?board_type=HY
GET /api/v1/scraper/northbound-flow?days=5
GET /api/v1/scraper/northbound-stocks?count=10
GET /api/v1/scraper/fund-nav?code=000001
GET /api/v1/scraper/margin-trade
GET /api/v1/scraper/crypto?symbols=bitcoin,ethereum
```

### 离线数据

```
GET /api/v1/offline/home
GET /api/v1/offline/daily?market=sh&code=600000
GET /api/v1/offline/min?market=sh&code=600000&min_type=lc5
```

### WebSocket 实时行情

```
WS /ws/realtime/SZ000001
WS /ws/realtime/SH600000
```

每 3 秒推送一次东方财富实时报价数据。

### 板块数据

```
GET /api/v1/board/list?board_type=HY&count=50
GET /api/v1/board/members?board_symbol=BK0634
GET /api/v1/board/ranking?board_type=HY&top_n=10
```

---

## TDX 公式引擎

内置完整的通达信公式词法/语法分析器，支持：

### 核心能力

- **词法分析**（lexer）：中文字符串、数字、标识符、运算符、括号等
- **语法分析**（parser）：递归下降解析，生成 AST
- **求值引擎**（eval）：支持 100+ 内置函数
- **绘图支持**：LINESTICK、STICKLINE、DRAWTEXT、DRAWICON 等
- **BKCOLOR** 变色支持
- **交易信号**：CROSS、BACKSET 等信号识别
- **NaN 边界处理**

### 内置函数

支持 MA、EMA、SMA、WMA、DMA、STDDEV、HV、HIGH、LOW、CLOSE、OPEN、VOL、AMOUNT、DATE、TIME、YEAR、MONTH、DAY 等基础函数，以及 MACD、KDJ、RSI、BOLL、DMI、WR、CCI、ARBR、OBV、SAR、TRIX 等复合指标函数。

### 使用方式

通过 MCP 工具 `tdx_formula_parse` / `tdx_formula_execute` / `tdx_formula_list` 调用，或 REST API `POST /api/v1/formula/parse` 和 `POST /api/v1/formula/execute`。

### 测试

```bash
cd go-tdx-mcp
go test ./formula/... -v
```

包含：词法测试、语法测试、内置函数测试、复合函数测试、性能基准测试、NaN 边界测试、bug 回归测试。

---

## 回测引擎

基于 `backtest` 包，提供完整的策略回测框架：

### 核心组件

- **Engine**：回测引擎，管理资金、持仓、交易
- **Strategy**：策略接口（`MA_Cross`、`MACD_Cross`、`RSI_Reversal`、`Bollinger_Breakout` 等内置策略）
- **Execution**：交易执行（市场价/限价单）
- **Slippage**：滑点模型（固定/百分比/随机）
- **Attribution**：收益归因分析
- **Combo**：组合策略回测
- **Risk**：风险指标计算（夏普比率、最大回撤、年化波动率等）

### 性能指标

返回 `total_return`、`total_return_pct`、`cagr`、`max_drawdown`、`sharpe_ratio`、`annual_volatility`、`win_rate`、`profit_factor`、`avg_win`、`avg_loss`、`total_trades` 等。

---

## 缠论分析

基于 `chanlun` 包，实现缠论核心分析：

- **分型**（FenXing）：顶分型/底分型识别
- **笔**（Bi）：上下笔连接
- **中枢**（ZhongShu）：中枢区间定位
- **线段**（XianDuan）：线段划分
- **买卖点**（MaiMaiDian）：一类/二类买卖点
- **背驰**（BeiChi）：趋势背驰判断
- **K 线合并**：K 线级别递归合并

---

## 因子分析

基于 `factor` 包，提供量化因子分析框架：

- 因子信息注册与查询（`factor.Get(name)`）
- 因子分析报告（`factor.FactorReport`）
- 前向收益计算（Forward Returns）
- 通过 MCP 工具 `tdx_factor_info`、`tdx_factor_report`、`tdx_factor_forward_returns` 调用

---

## 离线数据同步

基于 `offline` 包，支持 TDX 本地数据文件读写：

- **日线文件**（`.day`）：读取/写入 A 股日线 K 线
- **分钟线文件**（`.lc1`/`.lc5`/`.lc15`/`.lc30`/`.lc60`）：分钟级 K 线
- **股本变迁**（`gbbq`）：送转/配股/增发数据
- **板块文件**（`blocknew/`）：本地板块成分股
- **外部数据**（`vipdoc/ds/`）：港股/美股/期货/外汇/指数
- **财务数据**（`gpcw*.dat`）：通达信财务数据包
- **同步工具**：从 TQLEX/TCP 下载最新数据并写入本地 `.day` 文件

---

## 投资数据爬虫

`investing-scrapers` 是一个独立的 Go CLI 项目，从 Investing.com 抓取全球投资数据。

### 安装

```bash
cd investing-scrapers
go build -o investing-scrapers ./cmd/scrape
```

### 命令列表（11 个）

| 命令 | 说明 |
|------|------|
| `investing-scrapers currencies` | 全球货币汇率列表 |
| `investing-scrapers commodities` | 大宗商品列表 |
| `investing-scrapers commodity-quote` | 大宗商品详情报价 |
| `investing-scrapers index-quote` | 全球指数详情报价 |
| `investing-scrapers indices` | 全球指数列表 |
| `investing-scrapers search` | 搜索金融标的 |
| `investing-scrapers funds` | 基金列表 |
| `investing-scrapers crypto` | 加密货币行情 |
| `investing-scrapers calendar` | 经济日历（已发布事件） |
| `investing-scrapers calendar-published` | 已发布经济事件 |
| `investing-scrapers calendar-upcoming` | 即将发布经济事件 |

### 经济日历

每日约 85 条经济事件，每条包含 29 个字段（日期/时间/货币/国家/事件名称/实际值/预测值/前值 等）。

---

## 技术栈

### 核心依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `mark3labs/mcp-go` | v0.55.0 | MCP 协议实现 |
| `bensema/gotdx` | v0.0.0-20260625 | TDX TCP 协议客户端 |
| `chromedp/chromedp` | v0.15.1 | 无头浏览器（表格爬虫） |
| `gorilla/websocket` | v1.5.3 | WebSocket 实时行情 |
| `PuerkitoBio/goquery` | v1.12.0 | HTML 解析（基金净值等） |
| `golang.org/x/net` | v0.56.0 | 网络工具库 |

### 项目结构

```
go-tdx-mcp/
├── main.go                    # 程序入口（MCP Server / HTTP Server）
├── go.mod                     # 模块定义
├── config.example.json        # 配置模板
├── tdx/                       # MCP 工具实现层
│   ├── client.go              # Client 接口定义
│   ├── types.go               # 数据类型定义
│   ├── unified_client.go      # 统一客户端（HTTP + TCP 双通道）
│   ├── unified_bridge.go      # TCP/HTTP 桥接方法
│   ├── tcp_client.go          # gotdx TCP 客户端封装
│   ├── async_client.go        # 异步客户端
│   ├── collector.go           # 多服务器聚合器
│   ├── health.go              # 服务器健康检查
│   ├── tools.go               # 核心 6 个 MCP 工具
│   ├── tools_expanded.go      # 64 个扩展工具
│   ├── tools_v3.go            # 8 个 V3 工具
│   ├── tools_new.go           # 20+ 个新工具
│   ├── skills.go              # 投资技能 Prompts 定义
│   ├── skills_data_*.go       # 投资技能数据
│   ├── strategies.go          # 策略数据
│   ├── data_demo_test.go      # 数据演示测试
│   ├── e2e_protocol_test.go   # 7 个 E2E 协议测试
│   ├── test_comprehensive_test.go  # 212 工具全量测试
│   ├── integration_tcp_test.go     # TCP 直连集成测试
│   ├── tools_expanded_test.go      # 扩展工具测试
│   ├── tools_new_test.go           # 新工具测试
│   └── tdx_protocol_tools_test.go  # 协议工具测试
├── formula/                   # TDX 公式引擎（14 个文件）
│   ├── lexer.go               # 词法分析器
│   ├── parser.go              # 语法分析器
│   ├── ast.go                 # 抽象语法树
│   ├── eval.go                # 求值引擎
│   ├── builtin.go             # 内置函数
│   ├── builtin_complex.go     # 复合函数
│   ├── engine.go              # 公式引擎入口
│   ├── token.go               # Token 定义
│   ├── errors.go              # 错误类型
│   └── *_test.go              # 测试文件
├── indicator/                 # 技术指标计算
├── backtest/                  # 策略回测引擎（6 个文件）
│   ├── engine.go              # 回测引擎
│   ├── strategies.go          # 内置策略
│   ├── execution.go           # 交易执行
│   ├── slippage.go            # 滑点模型
│   ├── attribution.go         # 收益归因
│   └── combo.go               # 组合回测
├── chanlun/                   # 缠论分析
├── factor/                    # 因子分析
├── scraper/                   # Web 数据抓取（16 个文件）
│   ├── scraper.go             # chromedp 表格爬虫
│   ├── eastmoney_enhanced.go  # 东方财富增强抓取
│   ├── sina.go                # 新浪财经
│   ├── macro.go               # 宏观经济数据
│   ├── northbound.go          # 北向资金
│   ├── fund_nav_web.go        # 基金净值
│   ├── fund_holding.go        # 基金持仓
│   ├── margin_trade_web.go    # 融资融券
│   ├── block_trade.go         # 大宗交易
│   ├── hk_us.go               # 港美股
│   ├── tableparser.go         # 表格解析
│   ├── ocr.go                 # OCR 识别
│   ├── antiban.go             # 反爬机制
│   ├── imagepreprocess.go     # 图片预处理
│   └── tradingeconomics.go    # 宏观数据
├── screen/                    # 选股模块
├── offline/                   # 离线数据同步
├── portfolio/                 # 组合优化
├── finance/                   # 财报数据
└── web/                       # REST API 服务器
    ├── server.go              # HTTP 路由和处理器
    └── server_test.go         # 服务器测试
```

---

## 开发测试

### 运行全部测试

```bash
cd go-tdx-mcp
go test ./... -v
```

### 运行特定包测试

```bash
# 公式引擎测试
go test ./formula/... -v

# 综合工具测试（212 个工具）
go test ./tdx -run TestComprehensiveAllTools -v

# 协议 E2E 测试（7 个测试）
go test ./tdx -run TestE2E -v

# TCP 集成测试
go test ./tdx -run TestIntegration -v
```

### 运行性能测试

```bash
cd go-tdx-mcp
go test ./formula/... -run=Performance -bench=.
```

### 编译二进制

```bash
cd go-tdx-mcp
go build -o go-tdx-mcp .
```

---

## 许可证

MIT
