# Requirements Document

## Introduction

本特性为 TDX Finance MCP 项目增加通达信公式（TDX Formula）解析与执行引擎，使用户能够直接输入通达信软件风格的公式文本（技术指标公式、条件选股公式），在任意标的的 K 线数据上完成解析、校验与执行计算，并获得与通达信软件一致的序列输出。该能力可被现有选股（screen）、回测（backtest）、因子（factor）引擎复用。

## Glossary

- **TDX Formula**：通达信软件使用的公式语言，基于分号分隔的语句，支持序列数据、内置函数、赋值输出与绘图语句。
- **AST（Abstract Syntax Tree）**：公式文本解析后生成的抽象语法树，作为执行引擎的输入。
- **序列（Series）**：与 K 线长度一致的浮点数组，是公式执行的基本数据类型。
- **常量（Scalar）**：单个浮点数值，执行时可自动广播为与序列同长度的数组。
- **中间变量**：以 `:=` 赋值的变量，参与计算但不输出。
- **输出线（Output Line）**：以 `:` 或 `:` 后跟绘图属性赋值的变量，执行结果对外可见。
- **K 线（Bar）**：单根行情记录，包含 Open/High/Low/Close/Vol/Amount 字段。
- **内置函数（Builtin Function）**：引擎预定义的函数，如 MA、EMA、HHV、LLV、REF、CROSS、MACD 等。
- **交易系统公式（Trade Formula）**：以 `ENTERLONG:`、`EXITLONG:`、`ENTERSHORT:`、`EXITSHORT:` 输出买卖信号序列的公式类型。
- **五彩 K 线公式（Colorful K-Line Formula）**：以 `BKCOLOR:` 输出每根 K 线着色值（0-8）的公式类型。

## Requirements

### Requirement 1: 词法分析

**User Story:** AS 量化用户，I want 将通达信公式文本切分为 Token 流，so that 后续语法分析可以基于标准 Token 进行。

#### Acceptance Criteria

1. WHEN 公式文本包含数字、标识符、运算符、括号、分号、冒号、比较符、逻辑关键字，系统 SHALL 将其切分为类型正确的 Token 序列。
2. WHEN 公式文本包含无法识别的字符，系统 SHALL 返回包含行号与列号位置的语法错误。
3. WHEN 公式文本包含行注释 `//` 或块注释 `{...}`，系统 SHALL 跳过注释内容。
4. WHEN 公式文本包含字符串字面量（双引号包裹），系统 SHALL 将其识别为字符串 Token。

### Requirement 2: 语法分析与 AST 构建

**User Story:** AS 量化用户，I want 将 Token 流解析为 AST，so that 表达式结构可被程序化处理与执行。

#### Acceptance Criteria

1. WHEN 输入合法的公式语句，系统 SHALL 构建包含赋值语句、表达式、函数调用、输出声明的 AST。
2. WHEN 公式包含 `名称:表达式;` 形式，系统 SHALL 将其标记为输出语句。
3. WHEN 公式包含 `名称:=表达式;` 形式，系统 SHALL 将其标记为中间变量赋值。
4. WHEN 公式运算符存在优先级（括号、函数调用、乘除、加减、比较、逻辑），系统 SHALL 按标准优先级构建 AST。
5. WHEN 公式语法不完整或存在括号不匹配，系统 SHALL 返回带位置信息的语法错误。
6. WHEN 公式包含未定义函数或未定义变量引用，系统 SHALL 在解析或校验阶段报告错误。

### Requirement 3: 序列执行引擎

**User Story:** AS 量化用户，I want 在 K 线数据上执行 AST，so that 得到与通达信一致的指标序列输出。

#### Acceptance Criteria

1. WHEN 执行 AST，系统 SHALL 对每个输出语句计算长度与输入 K 线一致的序列。
2. WHEN 表达式同时包含序列与常量，系统 SHALL 将常量广播为序列后计算。
3. WHEN 公式引用收盘价、开盘价、最高价、最低价、成交量、成交额，系统 SHALL 将其绑定到对应 K 线字段。
4. WHEN 公式中的中间变量被后续语句引用，系统 SHALL 按语句顺序计算并缓存中间变量。
5. WHEN 计算周期不足（如 MA 前期窗口不满），系统 SHALL 对应位置输出 NaN 或 0（与通达信一致），且不影响后续有效值。
6. WHILE 同一公式被执行多次且输入 K 线不同，系统 SHALL 保持无状态，每次执行结果仅依赖输入。

### Requirement 4: 内置函数库

**User Story:** AS 量化用户，I want 引擎内置尽可能全量的通达信函数，so that 公式无需自定义即可完成常见指标计算。

#### Acceptance Criteria

1. WHEN 公式调用 MA、EMA、SMA、WMA、HHV、LLV、REF、SUM、COUNT、STD、VAR，系统 SHALL 计算正确序列。
2. WHEN 公式调用 CROSS、IF、MAX、MIN、ABS、SQRT、POW、EXP、LN、MOD、SIGN、BETWEEN，系统 SHALL 计算正确结果。
3. WHEN 公式调用 MACD、KDJ、RSI、BOLL、CCI、ROC、WR、BIAS、TRIX、DMI、EXPMA、ATR，系统 SHALL 复用 indicator 引擎实现并返回与通达信一致的序列。
4. WHEN 公式调用 BARSLAST、BARSSINCE、FILTER、BACKSET、VALUEWHEN、EXIST、EVERY、LAST、FINDHIGH、FINDLOW 等引用/逻辑函数，系统 SHALL 计算正确序列。
5. WHEN 函数参数数量或类型不匹配，系统 SHALL 返回带函数名与参数位置的错误。
6. WHEN 公式调用财务类函数（如 FINANCE、CAPITAL、TOTALCAP 等常量类），系统 SHALL 返回对应常数值或明确的支持状态。

### Requirement 5: 绘图语句与属性解析

**User Story:** AS 量化用户，I want 公式中的绘图语句与输出属性被解析和保留，so that 输出线附带样式信息。

#### Acceptance Criteria

1. WHEN 输出语句后跟随 `COLORxxx`、`LINETHICKn`、`STICK`、`POINTDOT` 属性，系统 SHALL 解析并保留在输出结果元信息中。
2. WHEN 公式包含 DRAWTEXT、DRAWICON、STICKLINE、POLYLINE、DRAWLINE、DRAWBAND 绘图语句，系统 SHALL 解析参数并在执行时生成对应绘制指令数据。
3. WHEN 绘图语句条件为假，系统 SHALL 对应位置输出空标记而非报错。

### Requirement 6: 条件选股公式执行

**User Story:** AS 量化用户，I want 执行选股类公式得到每只股票最后一个 K 线的判定结果，so that 可用于批量选股。

#### Acceptance Criteria

1. WHEN 公式为条件选股类型（无输出线或最终判定为布尔表达式），系统 SHALL 返回布尔判定结果与最后有效值。
2. WHEN 对多只股票依次执行同一选股公式，系统 SHALL 输出每只股票的命中/未命中判定与最近信号日期。

### Requirement 7: 交易系统公式与五彩 K 线执行

**User Story:** AS 量化用户，I want 执行交易系统公式与五彩 K 线公式，so that 可直接驱动回测与 K 线着色。

#### Acceptance Criteria

1. WHEN 公式包含 `ENTERLONG:`/`EXITLONG:`/`ENTERSHORT:`/`EXITSHORT:` 输出，系统 SHALL 检测为交易系统公式并计算各买卖信号序列。
2. WHEN 交易系统公式被执行，系统 SHALL 输出每个信号的最后一根触发 K 线索引与价格。
3. WHEN 公式包含 `BKCOLOR:` 输出，系统 SHALL 检测为五彩 K 线公式并计算每根 K 线的着色值（0-8）。
4. WHEN 一个公式同时包含技术指标输出线与交易信号输出，系统 SHALL 兼容执行并分别归类输出。

### Requirement 8: MCP 工具与 REST API 集成

**User Story:** AS MCP 用户，I want 通过 MCP 工具和 REST API 调用公式引擎，so that 无需修改代码即可解析与执行公式。

#### Acceptance Criteria

1. WHEN MCP 客户端调用公式解析工具，系统 SHALL 返回 AST 结构、输出线列表与校验错误。
2. WHEN MCP 客户端调用公式执行工具并传入代码/市场/周期/公式文本，系统 SHALL 返回各输出线序列及元信息。
3. WHEN REST API 请求 `/api/v1/formula/parse` 与 `/api/v1/formula/execute`，系统 SHALL 返回与 MCP 工具一致的 JSON 结果。
4. WHEN 公式执行工具被传入无效公式，系统 SHALL 返回包含行号、列号与原因的错误信息。

### Requirement 9: 错误处理与诊断

**User Story:** AS 量化用户，I want 清晰的错误诊断信息，so that 能快速定位公式问题。

#### Acceptance Criteria

1. WHEN 词法或语法错误发生，系统 SHALL 返回错误类型、行号、列号与附近上下文。
2. WHEN 执行期错误发生（如除零、非法周期参数），系统 SHALL 返回包含语句名或函数名的诊断信息。
3. WHEN 公式执行成功但包含 NaN 输出，系统 SHALL 在结果中标记无效值数量。

## Out of Scope

- 不实现通达信公式编辑器的完整图形渲染（图形化 K 线叠加绘制）。
- 不实现与通达信桌面端公式文件的二进制格式兼容。
- 不实现跨股票聚合公式（如 INDEXC 等市场级序列）的完整语义。
