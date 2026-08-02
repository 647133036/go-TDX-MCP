# User Instruction Memory

This file records user instructions, preferences, and teachings for reference in future interactions.

## Format

### User Instruction Entry
User instruction entries should follow this format:

[User Instruction Summary]
- Date: [YYYY-MM-DD]
- Context: [Mentioned scenario or time]
- Instructions:
  - [Content of user teaching or instruction, described line by line]

### Project Knowledge Entry
Entries discovered by the Agent during task execution should follow this format:

[Project Knowledge Summary]
- Date: [YYYY-MM-DD]
- Context: Discovered by Agent while performing [specific task description]
- Category: [Operations & Deployment|Build Methods|Testing Methods|Troubleshooting & Debugging|Workflow & Collaboration|Environment Configuration]
- Instructions:
  - [Specific knowledge points, described line by line]

## Deduplication Strategy
- Before adding a new entry, check for similar or identical instructions.
- If a duplicate is found, skip the new entry or merge it with the existing one.
- When merging, update the context or date information.
- This helps avoid redundant entries and keeps the memory file tidy.

## Entries

[Project Knowledge Summary]
- Date: 2026-08-02
- Context: Discovered by Agent while troubleshooting EastMoney data source connectivity issues
- Category: Troubleshooting & Debugging
- Instructions:
  - EastMoney data source availability: `push2delay.eastmoney.com` (HTTP/HTTPS) and `push2ex.eastmoney.com` (HTTPS) are available; `push2his.eastmoney.com` and `push2.eastmoney.com` are completely blocked (HTTP 000)
  - For K-line data, use `tdx_kline` (TDX TCP) instead of `tdx_eastmoney_kline_history` (which requires push2his)
  - EastMoney datacenter API (`datacenter-web.eastmoney.com`) works for financial data queries with correct report names (e.g., `RPT_LICO_FN_CPD` with columns like `SECURITY_CODE`, `REPORTDATE`, `BASIC_EPS`, etc.)
  - EastMoney F10 CompanySurvey API (`emweb.securities.eastmoney.com/PC_HSF10/CompanySurvey/CompanySurveyAjax`) works for company profile data using `code=SZ000001` or `code=SH600519` format
  - For board data, `tdx_eastmoney_sector_boards` with `board_type` of `industry`/`concept`/`region` is the working alternative to the broken TDX TCP board tools
  - TDX TCP server (`tdxhub.icfqs.com:7615`) has F10/board/capital flow features unregistered (503), but K-line, quotes, and basic data work
  - Build command: `cd /workspace/go-tdx-mcp && go build -o /tmp/tdx-mcp .`
  - Server is started via background terminal on port 8796 with `--web` flag