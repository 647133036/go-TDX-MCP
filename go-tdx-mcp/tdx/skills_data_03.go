package tdx

// SkillReportRating 查询个股研报评级一致预期
func SkillReportRating() SkillInfo {
	return SkillInfo{
		ID:          "tdx-report-rating",
		Name:        "查询个股研报评级一致预期",
		Description: "使用tdx_api_data查询A股个股的研报评级一致预期",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询A股个股的研报评级一致预期。

涵盖券商评级、目标价、盈利预测的一致预期数据。

当用户想看某只股票的机构评级、目标价或盈利预测时使用。

**Entry**: TdxSharePCCW.tdxf9_ag_cwsj_yjyj`,
	}
}

// SkillShareCapital 查询股本信息
func SkillShareCapital() SkillInfo {
	return SkillInfo{
		ID:          "tdx-share-capital",
		Name:        "查询股本信息",
		Description: "查询个股股本结构、股本变动、限售解禁、股票回购等数据",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询个股股本信息。

**查询类型**:
- 股本结构: 总股本、流通股本、限售股、港股股本
- 股本变动: 历史股本变动记录
- 限售解禁: 未来解禁计划、解禁数量
- 股票回购: 回购计划和实施情况

**Entry**: TdxSharePCCW.tdxf10_gg_gbjg

当用户想看总股本、流通股本、限售解禁或股票回购信息时使用。`,
	}
}

// SkillShareholderResearch 查询股东信息
func SkillShareholderResearch() SkillInfo {
	return SkillInfo{
		ID:          "tdx-shareholder-research",
		Name:        "查询股东信息",
		Description: "查询个股股东研究数据，包括控股股东、股东人数、股东排名等",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询个股股东研究数据。

**查询类型**:
- 控股股东: 实际控制人和控股股东信息
- 股东人数: 历史股东人数变化趋势
- 股东排名: 前十大股东排名和变动
- 前十大股东: 详细的前十大股东信息

**Entry**: TdxSharePCCW.tdxf10_gg_gdyj

当用户想看股东结构、股东人数变化或前十大股东信息时使用。`,
	}
}

// SkillStockEvents 查询股票事件信息
func SkillStockEvents() SkillInfo {
	return SkillInfo{
		ID:          "tdx-stock-events",
		Name:        "查询股票事件信息",
		Description: "查询股票交易日历、公司事件、重大事项等信息",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询股票事件信息。

**查询类型**:
- 交易日历: 股票停牌、复牌、除权除息日
- 公司事件: 股东大会、分红实施日、业绩发布日期
- 重大事项: 资产重组、要约收购、退市风险警示

**Entry**: TdxSharePCCW.tdxf10_gg_comreq

当用户想看股票重要日期、公司事件或重大事项时使用。`,
	}
}

// SkillTCZQCXX 查询题材生命周期与持续性
func SkillTCZQCXX() SkillInfo {
	return SkillInfo{
		ID:          "tdx-tczqcxx",
		Name:        "查询题材生命周期与持续性",
		Description: "判断某个题材当前处于发酵、扩散、主升、分歧、退潮还是尾声阶段",
		Markdown: `**Skill 分类**: 题材交易 / 市场周期 / 主题投资

**适用场景**: 判断某个题材当前处于发酵、扩散、主升、分歧、退潮还是尾声阶段。

**分析框架**:
1. 题材阶段定位（发酵→扩散→主升→分歧→退潮→尾声）
2. 阶段特征识别：
   - 发酵期：消息开始传播，认知度低
   - 扩散期：更多参与者加入，涨停家数增加
   - 主升期：龙头加速、板块集体爆发、赚钱效应最强
   - 分歧期：高位标的开始松动、低位补涨出现、龙虎榜席位变杂
   - 退潮期：高标大幅回落、首板减少、资金流出
   - 尾声期：热度大降、仅剩少数活口
3. 持续性评估（产业逻辑是否支撑/资金是否认可/催化是否持续）
4. 核心锚点识别（谁是情绪龙头、谁是趋势中军）
5. 明日观察清单（看回流/扩散/分歧转一致还是退潮确认）

**用户关键词**: 题材生命周期、主线持续性、热点轮动、题材退潮、题材强度、主升阶段`,
	}
}

// SkillTradePlan 生成交易计划
func SkillTradePlan() SkillInfo {
	return SkillInfo{
		ID:          "tdx-trade-plan",
		Name:        "生成交易计划",
		Description: "为已选股票制定可执行的交易计划，包括入场、加仓、止损、止盈策略",
		Markdown: `**Skill 分类**: 交易决策 / 技术与资金共振 / 计划管理

**适用场景**: 用户已选好股票但不知道怎么做计划：在哪买、在哪加减仓、何时止盈止损。

**分析框架**:
1. 多周期技术位置分析（周线/日线/60分钟/30分钟关键支撑阻力位）
2. 量价关系与资金信号识别
3. 入场计划：
   - 最佳入场价位或区间
   - 入场条件（突破确认/回踩确认/量能配合）
   - 初始仓位建议
4. 加仓条件（什么情况下可以加仓）
5. 止损设置：
   - 技术止损位（跌破关键支撑）
   - 时间止损（震荡多久不涨则离场）
   - 逻辑止损（买入逻辑被证伪）
6. 止盈策略：
   - 第一目标位（分批止盈）
   - 第二目标位（趋势持仓）
   - 移动止盈规则
7. 风险收益比计算

**核心原则**: 计划必须可执行，每个条件都是明确的、可验证的。`,
	}
}

// SkillTradingInfo 查询个股交易相关数据
func SkillTradingInfo() SkillInfo {
	return SkillInfo{
		ID:          "tdx-trading-info",
		Name:        "查询个股交易相关数据",
		Description: "查询个股龙虎榜、大宗交易、资金流向、融资融券、异动信息等交易数据",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询交易相关数据。

**查询类型**:
- 龙虎榜: 个股龙虎榜数据
- 大宗交易: 大宗交易记录和折溢价
- 资金流向: 主力资金净流入流出
- 融资融券: 融资余额和融券余量
- 异动信息: 涨跌幅异动、成交量异动、换手率异动

**Entry**: TdxSharePCCW.tdxf10_gg_jyds

当用户想看交易数据、龙虎榜、大宗交易或资金流向时使用。`,
	}
}

// SkillValuationPricing 估值与定价框架分析
func SkillValuationPricing() SkillInfo {
	return SkillInfo{
		ID:          "tdx-valuation-pricing-framework",
		Name:        "估值与定价框架分析",
		Description: "选择合适的估值方法，判断当前估值是否合理，测算目标估值与目标价",
		Markdown: `**Skill 分类**: 估值分析 / 定价逻辑 / 投资决策

**适用场景**: 判断一家公司该用什么估值方法、当前估值是否合理、市场在交易什么预期。

**分析框架**:
1. 选择合适的估值方法：
   - 成长型公司：PE/PEG/PS/EVEBITDA/DCF
   - 价值型公司：PE/PB/股息率/DDM
   - 周期型公司：PB/周期中枢估值/盈利拐点分析
   - 资产型公司：PB/NAV/重置成本
2. 当前估值水平分析：
   - 与历史估值区间对比
   - 与同行业估值对比
   - 与国际可比公司估值对比
3. 估值驱动因素拆解（当前估值隐含的增长预期）
4. 重估/杀估值的触发条件
5. 目标估值与目标价测算

**常用工具**: tdx_quotes, tdx_api_data, tdx_indicator_select, wenda_report_query`,
	}
}

// SkillWXDA 问小达选A股
func SkillWXDA() SkillInfo {
	return SkillInfo{
		ID:          "tdx-wxd-a",
		Name:        "问小达选A股",
		Description: "通过自然语言查询进行A股股票筛选，支持行情、技术、财务等多条件组合筛选",
		Markdown: `**Skill 分类**: 智能选股 / A股筛选

**适用场景**: 通过自然语言查询进行A股股票筛选，支持行情指标、技术形态、财务指标、行业概念等多条件组合筛选。

**数据获取**: 调用 tdx_screener 进行选股查询。

**选股维度**:
- 行情指标：涨跌幅、成交量、换手率、市值
- 技术形态：MACD金叉死叉、均线排列、突破形态
- 财务指标：PE、PB、ROE、营收增速、净利润增速
- 行业概念：所属行业、概念板块、主题标签
- 资金行为：主力净流入、北向资金、机构持仓

当用户询问A股股票筛选问题时，必须使用此技能。`,
	}
}

// SkillWXDBK 问小达选板块
func SkillWXDBK() SkillInfo {
	return SkillInfo{
		ID:          "tdx-wxd-bk",
		Name:        "问小达选板块",
		Description: "通过行业估值、资金流向、涨跌幅、板块类型等多条件筛选市场板块",
		Markdown: `**Skill 分类**: 智能选板块

**适用场景**: 通过行业估值、资金流向、涨跌幅、板块类型等多条件筛选市场板块。

**数据获取**: 使用 tdx_screener 查询数据。

**选板块维度**:
- 行业估值：PE、PB分位数
- 资金流向：板块资金净流入流出
- 涨跌幅：板块阶段涨跌表现
- 板块类型：行业板块、概念板块、风格板块

当用户询问板块筛选问题时，必须使用此技能。`,
	}
}

// SkillWXDETF 问小达选ETF
func SkillWXDETF() SkillInfo {
	return SkillInfo{
		ID:          "tdx-wxd-etf",
		Name:        "问小达选ETF",
		Description: "根据行情、跟踪指数基本面、规模、风格类型等条件筛选ETF",
		Markdown: `**Skill 分类**: 智能选ETF

**适用场景**: 根据行情、跟踪指数基本面、规模、风格类型等条件筛选ETF。

**数据获取**: 使用 tdx_screener 获取数据。

**选ETF维度**:
- 跟踪指数：沪深300、中证500、科创50、行业指数等
- 规模与流动性：基金规模、日均成交额
- 费率：管理费率、托管费率
- 折溢价：当前折溢价率
- 风格：宽基、行业、主题、策略、跨境

当用户询问ETF筛选问题时，必须使用此技能。`,
	}
}

// SkillWXDJJ 问小达选基金
func SkillWXDJJ() SkillInfo {
	return SkillInfo{
		ID:          "tdx-wxd-jj",
		Name:        "问小达选基金",
		Description: "根据基金类型、业绩、基金经理、风险、持仓等维度筛选公募基金",
		Markdown: `**Skill 分类**: 智能选基金

**适用场景**: 根据基金类型、业绩、基金经理、风险、持仓等维度筛选公募基金。

**数据获取**: 使用 tdx_screener 查询数据。

**选基金维度**:
- 基金类型：股票型、混合型、债券型、货币型、QDII
- 业绩：近1月/3月/6月/1年/3年收益率
- 基金经理：管理年限、历史业绩、管理规模
- 风险指标：最大回撤、夏普比率、波动率
- 资产配置：股票仓位、行业集中度、重仓股

当用户询问基金筛选问题时，必须使用此技能。`,
	}
}

// SkillYJYGBY 业绩预告博弈
func SkillYJYGBY() SkillInfo {
	return SkillInfo{
		ID:          "tdx-yjygby",
		Name:        "业绩预告博弈",
		Description: "分析业绩预告发布前后的市场交易策略，判断超预期或低于预期的博弈机会",
		Markdown: `**Skill 分类**: 事件驱动 / 财报博弈 / 短中期交易

**适用场景**: 想知道某个业绩预告、快报、正式财报发布前后市场会怎么交易。

**分析框架**:
1. 业绩预告类型（预增/预减/扭亏/首亏/续亏）
2. 与市场预期的偏差分析：
   - 卖方一致预期 vs 实际预告
   - 市场隐含预期 vs 实际数据
3. 历史业绩预告后的股价反应规律
4. 当前股价是否已反映业绩预期
5. 博弈策略：
   - 超预期+低位：关注介入机会
   - 超预期+高位：注意兑现风险
   - 低于预期+高位：回避
   - 低于预期+低位：关注利空出尽机会
6. 关键跟踪时间节点（正式报告发布日期）

**数据获取**: 使用 wenda_notice_query 查询业绩预告数据。`,
	}
}

// SkillZJFTJYTL 专家访谈纪要提炼
func SkillZJFTJYTL() SkillInfo {
	return SkillInfo{
		ID:          "tdx-zjftjytl",
		Name:        "专家访谈纪要提炼",
		Description: "将专家访谈、渠道反馈、供应链调研等内容提炼成可用于投资判断的结构化结论",
		Markdown: `**Skill 分类**: 一手信息加工 / 行业研究 / 纪要提炼 / 投资洞察

**适用场景**: 将专家访谈、渠道反馈、供应链调研、电话会纪要等内容提炼成可用于投资判断的结构化结论。

**分析框架**:
1. 信息源评估（信息可靠性、时效性、覆盖范围）
2. 核心论点提取（3-5个最重要的发现）
3. 信息分类（需求端/供给端/竞争格局/政策/技术创新）
4. 与已有认知的差异（哪些是增量信息，哪些验证已有判断）
5. 对投资决策的影响（正面/负面/中性）
6. 后续验证方向

**数据追踪**: 配合 tdx_api_data、wenda_news_query、wenda_report_query 做事实校验。`,
	}
}

// SkillZTLTBY 查询龙头博弈分析
func SkillZTLTBY() SkillInfo {
	return SkillInfo{
		ID:          "tdx-ztltby",
		Name:        "查询龙头博弈分析",
		Description: "分析涨停结构、连板梯队、情绪周期，判断个股龙头博弈价值",
		Markdown: `**Skill 分类**: 短线交易 / 情绪周期 / 龙头战法

**适用场景**: 判断今天市场在打什么板、连板高度和梯队是否健康、谁是情绪龙头、某只股票是否具备龙头博弈价值。

**分析框架**:
1. 涨停结构全景：
   - 涨停家数、跌停家数、炸板率
   - 连板高度分布（首板/2连板/3连板/4连及以上）
   - 封板时间分布（早盘/午盘/尾盘）
2. 梯队分析：
   - 总龙头：最高辨识度、全市场情绪锚点
   - 分支龙头：各题材方向的核心标的
   - 补涨标的：跟随龙头的低位活跃股
   - 中军标的：趋势稳定的容量股
3. 情绪周期定位（冰点/修复/主升/高位震荡/退潮）
4. 个股博弈价值：
   - 弱转强信号识别
   - 分歧转一致信号
   - 卡位和晋级概率
   - 止损与离场条件
5. 明日博弈推演：
   - 核心观察标的
   - 博弈情景推演（3种情景）
   - 各情景下的应对策略

**用户关键词**: 涨停结构、连板、首板、换手板、封板质量、炸板、卡位、打板、接力、龙头战法、情绪冰点/修复/主升/退潮`,
	}
}
