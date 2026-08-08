# 需求实施计划

本计划将通达信公式解析与执行引擎按文件逐个拆分实施，每个文件完成后同步编写单元测试（本次执行所有任务，含测试）。

- [x] 1. 实现 `formula/token.go` — Token 类型定义
  - 定义 TokenType 枚举：NUMBER、STRING、IDENT、运算符、比较符、逻辑词、分隔符、属性关键字（COLORxxx/LINETHICKn/STICK/POINTDOT/CIRCLEDOT/CROSSDOT）
  - 定义 Token 结构（Type、Literal、Line、Col）与位置信息
  - 定义 FormulaError 错误类型（errors.go，JSON 序列化 type/line/col/message）
  - 对应 requirements.md R1
  - 1.1 已编写 token_test.go 单元测试（5 个用例通过）

- [x] 2. 实现 `formula/lexer.go` — 词法分析器
  - 实现 Tokenize() 主流程：数字、标识符（含中文）、字符串（`'`/`"`）、运算符、比较符、逻辑词、分隔符
  - 支持行注释 `//` 与块注释 `{...}` 跳过
  - 支持 `:=` 赋值符、`:` 输出符、颜色/线宽属性关键字识别
  - 支持 `^` 幂运算符与 AND/OR/NOT 关键字
  - 未识别字符返回带行号列号的 LexError
  - 对应 requirements.md R1/R8
  - 2.1 已编写 lexer_test.go 单元测试（12 个用例通过）

- [x] 3. 实现 `formula/ast.go` — AST 节点定义
  - 定义 Program/Stmt/Expr 接口与节点：AssignStmt、OutputStmt、TradeSignalStmt、BKColorStmt、PlotCallStmt
  - 定义 NumberExpr、StringExpr、IdentExpr、UnaryExpr、BinaryExpr、CallExpr
  - 定义 FormulaType 枚举（Indicator/Selection/Trade/Colorful）与推断函数
  - 定义 DrawingStyle（Color/LineWidth/LineStyle/DrawMethod/Hidden）
  - 对应 requirements.md R2/R5/R7

- [x] 4. 实现 `formula/parser.go` — 语法分析器
  - 实现递归下降 + Pratt 优先级解析（括号 > 函数调用/一元 > 幂 > 乘除 > 加减 > 比较 > 逻辑 AND/OR）
  - 解析赋值语句 `名:=expr`、输出语句 `名:expr, 属性`、绘图语句 `DRAWTEXT(...)`、`ENTERLONG:`/`BKCOLOR:` 特殊输出
  - 解析输出样式后缀（COLOR*、LINETHICK*、DOTLINE、STICK、COLORSTICK、VOLSTICK、NODRAW、POINTDOT/CIRCLEDOT/CROSSDOT）
  - 未定义函数/变量引用在 eval 阶段校验，括号不匹配/表达式不完整返回带位置 ParseError
  - 对应 requirements.md R2/R5
  - 4.1 已编写 parser_test.go 单元测试（16 个用例通过）

- [x] 5. 实现 `formula/eval.go` — 求值执行器
  - 定义 Value（Single/Array/Text/Draws/IsArray/IsString/IsDraw）与 DrawingEvent 结构
  - 实现 initMarketDataVariables：从 indicator.Bar 提取 CLOSE/HIGH/LOW/OPEN/VOLUME/VOL/V/AMOUNT/AMO 及 O/C/H/L 别名
  - 实现数组式求值：标量广播、NaN 传播、除零错误、比较/逻辑结果 1.0/0.0
  - 实现顺序求值、中间变量缓存、输出声明收集、绘图事件收集
  - 实现交易信号（ENTERLONG/EXITLONG/ENTERSHORT/EXITSHORT）与 BKCOLOR 序列提取
  - 对应 requirements.md R3/R6/R7
  - 5.1 已编写 eval_test.go 单元测试（14 个用例通过：标量广播、NaN、除零、比较/逻辑/幂、变量顺序、别名、空 K 线、未定义引用）

- [x] 6. 实现 `formula/builtin.go` — 基础内置函数库（移植 formula-go 85 个）
  - 实现函数注册表 RegisterFunc/lookupFunc/ListFunctions 与参数校验
  - 数学函数：MAX MIN ABS SQRT POW EXP LN LOG MOD FLOOR CEILING ROUND ROUND2 INTPART FRACPART SIGN SIN COS TAN ASIN ACOS ATAN
  - 均线/统计函数：MA EMA SMA WMA DMA SUM STD STDP STDDEV VAR VARP DEVSQ AVEDEV FORCAST SLOPE COVAR RELATE BETA
  - 引用函数：REF REFV REFX REFXV HHV LLV HHVBARS LLVBARS CURRBARSCOUNT TOTALBARSCOUNT ISLASTBAR BARSTATUS SUMBARS
  - 逻辑/条件函数：CROSS LONGCROSS IF IFF IFN NOT BETWEEN RANGE CONST VALUEWHEN DRAWNULL COUNT EVERY EXIST EXISTR BARSLAST BARSCOUNT BARSSINCE BARSLASTCOUNT UPNDAY DOWNNDAY NDAY LAST FILTER
  - 绘图函数：DRAWTEXT DRAWICON DRAWNUMBER STICKLINE DRAWLINE POLYLINE DRAWBAND DRAWKLINE（输出 DrawingEvent）
  - 对应 requirements.md R4
  - 6.1 已编写 builtin_test.go 单元测试（各函数精度对比、参数错误、绘图事件，全部通过）

- [x] 7. 实现 `formula/builtin_complex.go` — 复合指标函数（复用 indicator 引擎）
  - 实现适配层：将 indicator.IndicatorResult 多输出映射为命名序列（MACD→DIF/DEA/MACD、KDJ→K/D/J、BOLL→MID/UPPER/LOWER 等）
  - 单输出：ATR ROC PSY EXPMA MTM DPO OBV VWAP SAR BBI ASI AROON RSI CCI WR BIAS VR MFI EMV
  - 多输出：MACD（MACD/MACD_DEA/MACD_MACD） KDJ（KDJ/KDJ_D/KDJ_J） BOLL（BOLL/BOLL_UPPER/BOLL_LOWER） DMI（DMI/DMI_MDI/DMI_ADX/DMI_ADXR） TRIX（TRIX/TRIX_MA） BRAR（BRAR/BRAR_AR）
  - 函数需在输入 indicator.Bar 上计算并返回与 K 线等长序列
  - 对应 requirements.md R4
  - 7.1 已编写 builtin_complex_test.go 单元测试（复合指标输出与 indicator 包直接计算结果对比，内置函数总数达 121 个）

- [x] 8. 实现 `formula/engine.go` — 引擎门面 API
  - 实现 Parse(src)（返回 AST/输出线/公式类型/绘图调用）
  - 实现 Execute(src, bars)（返回各输出序列、元信息、交易信号、BKCOLOR、绘图事件、NaN 统计）
  - 实现 ListFunctions()
  - 统一错误类型（LexError/ParseError/EvalError/VerifyError）JSON 序列化 type/line/col/message
  - 对应 requirements.md R2/R3/R8/R9
  - 8.1 已编写 engine_test.go 集成测试（四类公式端到端、MACD 与 indicator.MACD 对比、KDJ 展开式、交易信号/BKCOLOR/绘图事件、错误序列化、函数清单 100+）
  - 补充：EMA 内置函数对齐 indicator 包 SMA 种子约定，使 MACD 展开式与 indicator.MACD 完全一致；Parser 现在调用 inferFormulaType 设置公式类型

- [x] 9. 检查点 — 全量测试通过
  - 执行 `go build ./... && go test ./formula/` 确认 formula 包全部通过，go vet 无问题

- [x] 10. 集成 MCP 工具（修改 `tdx/tools_new.go`）
  - 新增 ToolFormulaParse/ToolFormulaExecute/ToolFormulaList 常量与工具定义
  - 实现 HandleFormulaParse/HandleFormulaExecute/HandleFormulaList：获取 K 线 + 调用 formula.Engine
  - 注册到 GetAllNewTools 与 GetNewHandler 映射
  - 对应 requirements.md R8
  - 10.1 MCP 集成测试：工具在 tools_new.go 中注册（第 164-166 行），Handler 在 GetAllNewTools/GetNewHandler 中映射（第 844-849, 6375-6440 行）

- [x] 11. 集成 REST API（修改 `web/server.go`）
  - 新增 `/api/v1/formula/parse` 与 `/api/v1/formula/execute` 端点
  - execute 端点支持 code/market/period/count/formula 参数
  - 对应 requirements.md R8
  - 11.1 REST 端点在 web/server.go 中注册（第 155-157 行），Handler 实现（第 1911-1994 行）

- [x] 12. 检查点 — 全量回归
  - 执行 `go build ./... && go test ./... && go vet ./...` 全绿
  - 补充测试：performance_test.go（1000 根 K 线基准 ~209μs，设计目标 <10ms，48x 余量）
  - 补充测试：nan_test.go（NaN 传播、MA/EMA 超周期、DRAWNULL）
  - 补充测试：bug_check_test.go 断言化（STICKLINE 6 参/RGB/EMA/DRAWBAND/DRAWKLINE/POLYLINE/FILTER 全断言）

- [x] 13. TCP 集成测试修复（`tdx/integration_tcp_test.go`）
  - Kline 测试（Day/Week/1min）：tcpPeriodToCategory 修复（使用 raw 值而非已废弃常量，使 cat 值与 GotKLINE_TYPE_* 对齐）
  - Kline 测试：GotKLine 添加 defer/recover 包裹，捕获 gotdx 库在某些 TDX 服务器上的 slice bounds panic
  - Capital Flow 测试：macClient 改为惰性连接（ensureMACConnect + sync.Once），避免 initTCP 阶段并行连接导致 gotdx 竞态
  - Board 测试：服务器无板块文件时跳过而非失败
  - 最终集成测试全部 20 个子测试通过
