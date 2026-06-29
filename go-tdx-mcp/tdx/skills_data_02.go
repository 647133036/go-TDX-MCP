package tdx

// SkillGGTZLJYJ 个股投资逻辑研究
func SkillGGTZLJYJ() SkillInfo {
	return SkillInfo{
		ID:          "tdx-ggtzljyj",
		Name:        "个股投资逻辑研究",
		Description: "用于个股核心投资逻辑深度研究",
		Markdown: `**Skill 分类**: 个股研究 / 基本面分析 / 机构研究

**适用场景**: 用户输入股票名称后，快速形成接近券商研究风格的结构化个股分析。

**常用工具**: tdx_quotes, tdx_api_data, tdx_kline, wenda_report_query

**分析框架**:
1. 确认公司是什么（主营业务、收入利润来源、行业赛道、产业链位置）
2. 明确核心投资逻辑（成长驱动/周期反转/价值修复/事件催化/产业趋势受益/竞争格局改善）
3. 看行业和竞争格局（景气度、竞争格局稳定性、龙头还是跟随者）
4. 验证财务质量（收入利润增速、毛利率净利率、现金流质量、资产负债、ROE、费用率、存货应收变化）
5. 理解市场当前交易什么（业绩超预期/新产品放量/景气反转/估值切换/政策驱动/情绪催化）
6. 评估估值是否匹配逻辑（成长股看PE/PEG/PS/EVEBITDA，价值股看PE/PB/股息率，周期股看PB/周期中枢估值）
7. 列出风险点（短期/中期/长期）
8. 给出投资结论

**输出模板**: 1.公司概况 2.核心投资逻辑 3.行业与竞争格局 4.财务质量验证 5.当前市场预期 6.估值与位置判断 7.风险分析 8.投资结论`,
	}
}

// SkillGGWDZK 个股问答总控
func SkillGGWDZK() SkillInfo {
	return SkillInfo{
		ID:          "tdx-ggwdzk",
		Name:        "个股问答总控",
		Description: "用于个股相关问题的总控路由，自动识别问题类型并调度到最合适的分析框架",
		Markdown: `**Skill 分类**: 总控路由 / 个股研究中枢 / 决策支持

**适用场景**: 用户输入一句问题如"这个票短期怎么看？"，负责识别问题类型、调取合适数据，路由到最合适的分析框架。

**常用工具**: tdx_quotes, tdx_kline, tdx_api_data, tdx_screener

**路由逻辑**:
- 交易相关 → tdx-event-driven-short-term-catalyst 技能
- 基本面相关 → tdx-ggtzljyj 技能
- 估值相关 → tdx-valuation-pricing-framework 技能
- 机构持仓 → tdx-jgccgdfx 技能
- 消息驱动 → wenda_news_query + tdx-event-driven 技能
- 技术分析 → tdx_kline + tdx-trade-plan 技能

先识别问题核心诉求，再选择合适的子技能框架展开分析。`,
	}
}

// SkillGGYCBFX 公告与财报分析
func SkillGGYCBFX() SkillInfo {
	return SkillInfo{
		ID:          "tdx-ggycbfx",
		Name:        "公告与财报分析",
		Description: "用于解读公告、年报、季报、业绩预告，判断其对股价的影响",
		Markdown: `**Skill 分类**: 公告解读 / 财报分析 / 基本面验证

**适用场景**: 用户上传或粘贴公告、年报、季报、业绩预告，判断其对股价的影响。

**分析框架**:
1. 判断公告性质（正面/负面/中性/复杂）
2. 评估影响力级别（重大/中等/轻微/噪音）
3. 聚焦三个关键问题：
   - 对现有投资逻辑的验证或修正
   - 市场预期的变化方向
   - 可能的股价反应路径（短期冲击/中期重估）
4. 结合当前市场环境和股价位置判断公告是被高估还是低估了影响
5. 事后跟踪建议

**数据获取**: 使用 tdx_api_data 工具进行事实校验和扩展查询。`,
	}
}

// SkillGSZDDF 公司质地打分
func SkillGSZDDF() SkillInfo {
	return SkillInfo{
		ID:          "tdx-gszddf",
		Name:        "公司质地打分",
		Description: "用于对公司质地进行结构化多维度打分，评估是否是好公司",
		Markdown: `**Skill 分类**: 基本面研究 / 公司质量评估 / 长线选股

**适用场景**: 不是只看故事，而是对公司"质地"有结构化判断：是不是好公司，壁垒强不强，成长质量怎么样。

**分析框架**:
多维度打分体系（每个维度1-10分）：
1. 商业模式（护城河、可复制性、规模效应、客户粘性）
2. 成长质量（收入增速稳定性、利润增速稳定性、内生or外延增长）
3. 盈利能力（ROE、毛利率、净利率趋势）
4. 财务健康（资产负债率、现金流质量、应收账款质量）
5. 公司治理（管理层能力、股权结构、小股东保护、信息披露）
6. 行业地位（市占率、品牌力、定价权、行业景气度位置）
7. 估值合理性（当前估值 vs 历史估值 vs 同行估值）

**输出**: 总分及自评表，明确是好公司（推荐）、一般公司（跟踪）或差公司（回避）。`,
	}
}

// SkillHotTopic 查询个股热点题材
func SkillHotTopic() SkillInfo {
	return SkillInfo{
		ID:          "tdx-hot-topic",
		Name:        "查询个股热点题材",
		Description: "使用tdx_api_data查询个股热点题材、板块族谱、主题库和事件驱动标签",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询个股热点题材信息。

按股票代码获取热点题材板块族谱、主题库、事件驱动和信息面概览。

**Entry**: TdxSharePCCW.tdxf10_gg_zxts

**查询类型**: 热点题材板块族谱、主题库分类、事件驱动标签、信息面概览

当用户想看某只股票属于哪些热点题材或概念板块时使用。`,
	}
}

// SkillIndustryChain 查询行业产业链
func SkillIndustryChain() SkillInfo {
	return SkillInfo{
		ID:          "tdx-industry-chain",
		Name:        "查询行业产业链",
		Description: "使用tdx_api_data查询行业产业链结构和重要事件",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询行业产业链和行业重要事件。

**Entry**: TdxSharePCCW.tdxf10_gg_cycl

**适用场景**: 当用户想看某个行业的产业链结构、上下游关系、重要事件时使用。

返回产业链全景图、上下游公司列表、行业重要事件和关键数据。`,
	}
}

// SkillIndustryChainMapping 查询行业产业链映射
func SkillIndustryChainMapping() SkillInfo {
	return SkillInfo{
		ID:          "tdx-industry-chain-mapping",
		Name:        "查询行业产业链映射",
		Description: "用于行业趋势方向的产业链映射分析，识别核心受益标的",
		Markdown: `**Skill 分类**: 产业研究 / 主题投资 / 行业挖掘

**适用场景**: 用户输入一个行业趋势、技术方向或政策方向，快速得到完整产业链结构、核心环节、最受益公司和映射逻辑。

**分析框架**:
1. 识别输入方向属于哪个产业赛道
2. 梳理完整产业链结构（上游→中游→下游→终端）
3. 识别核心环节（技术壁垒最高、利润最集中、弹性最大的环节）
4. 映射到A股标的（区分核心受益、次受益和题材跟风）
5. 评估行业催化剂和兑现节奏

**常用工具**: tdx_api_data, wenda_news_query, tdx-quotes`,
	}
}

// SkillJGCCGDFX 机构持仓股东分析
func SkillJGCCGDFX() SkillInfo {
	return SkillInfo{
		ID:          "tdx-jgccgdfx",
		Name:        "机构持仓股东分析",
		Description: "通过十大股东、机构持仓变化判断资金进出方向和筹码稳定程度",
		Markdown: `**Skill 分类**: 股东结构 / 机构动向 / 个股深度研究

**适用场景**: 通过十大股东、机构持仓、持股集中度变化，判断资金进出方向、筹码稳定程度。

**分析框架**:
1. 十大股东变动趋势（增持/减持/新进/退出）
2. 机构持仓变化（公募/私募/社保/保险/QFII）
3. 持股集中度分析（前十大占比变化、股东人数变化）
4. 机构认可度判断（是否机构重仓、是否有核心机构锁定）
5. 筹码稳定性评估
6. 对中长期逻辑的验证

**常用工具**: tdx-main-position, tdx_quotes, wenda_news_query

**输出模板**: 表格形式，包含持股变动、机构类型、变动股数、占总股本比例、变动方向`,
	}
}

// SkillJJZCYJD 基金重仓拥挤度
func SkillJJZCYJD() SkillInfo {
	return SkillInfo{
		ID:          "tdx-jjzcyjd",
		Name:        "基金重仓拥挤度",
		Description: "判断基金重仓股是机构共识优质资产还是已过度拥挤",
		Markdown: `**Skill 分类**: 机构行为 / 拥挤交易 / 风险管理 / 组合研究

**适用场景**: 判断某只基金重仓股到底是机构共识优质资产，还是已过度拥挤。

**分析框架**:
1. 基金持仓占比分析
2. 持仓变动趋势（持续增加还是开始减少）
3. 重仓基金数量和集中度
4. 与历史拥挤度对比
5. 拥挤度风险评级（低/中低/中/中高/高）
6. 如果过度拥挤，识别可能触发解体的扳机事件

**常用工具**: tdx-main-position, tdx-shareholder-research, tdx_quotes, tdx_kline, tdx_api_data`,
	}
}

// SkillLHBXWFG 查询龙虎榜席位风格
func SkillLHBXWFG() SkillInfo {
	return SkillInfo{
		ID:          "tdx-lhbxwfg",
		Name:        "查询龙虎榜席位风格",
		Description: "分析龙虎榜席位买卖结构和博弈判断",
		Markdown: `**Skill 分类**: 龙虎榜席位风格 / 资金行为 / 游资研究 / 短线交易

**适用场景**: 回答"这只票龙虎榜怎么看""哪些席位在主导""更像接力还是兑现"等问题。

**优先工具**: tdx_api_data, tdx_lookup_stock, tdx_quotes, tdx_kline, 必要时补充 tdx_screener, tdx_indicator_select

**分析框架**:
1. 席位识别与分类（知名游资/机构席位/量化席位/券商自营/散户席位）
2. 席位买卖结构（买方和卖方的净额、集中度）
3. 席位风格匹配（该席位历史操作模式）
4. 博弈结构判断（接力/低吸/兑现/对倒）
5. 席位持续性和退出信号

**输出模板**: 席位买卖明细表、净买卖汇总、博弈判断、后续观察重点`,
	}
}

// SkillMainPosition 主力资金
func SkillMainPosition() SkillInfo {
	return SkillInfo{
		ID:          "tdx-main-position",
		Name:        "主力资金",
		Description: "使用tdx_api_data查询机构持股、北向资金和持仓对比数据",
		Markdown: `**F10技能**: 使用 tdx_api_data 查询机构持股、北向资金和持仓对比数据。

**Entry**: TdxSharePCCW.tdxf10_gg_zlcc

**查询类型**:
- 机构持股: 查看机构持仓变动
- 北向资金 (rtype=bszj): 查看北向资金持仓变化
- 持仓对比: 不同机构类型持仓对比
- 主力持仓趋势: 时间序列上的持仓变化

当用户想看机构持仓、北向资金流入流出或主力资金动向时使用。`,
	}
}

// SkillMRTYJB 每日投研简报
func SkillMRTYJB() SkillInfo {
	return SkillInfo{
		ID:          "tdx-mrtyjb",
		Name:        "每日投研简报",
		Description: "每日生成精炼聚焦的市场投研梳理简报",
		Markdown: `**Skill 分类**: 投研信息整合 / 日报 / 决策支持

**适用场景**: 每日获得一份"精炼、聚焦"的市场投研梳理，帮助构建当日观察思路。

**信息获取**: tdx_quotes, wenda_report_query, wenda_news_query

**输出结构**:
1. 盘面概况（主要指数涨跌、成交额、涨跌家数）
2. 热点板块与资金流向（今日最强/最弱板块、资金流入/流出方向）
3. 重要消息与研报观点（政策、宏观、行业、公司重点消息）
4. 龙虎榜与资金行为（异动个股、席位动向）
5. 明日关注（关键事件、重点标的、风险提示）
6. 一句话投研小结`,
	}
}

// SkillPositionDecision 仓位决策
func SkillPositionDecision() SkillInfo {
	return SkillInfo{
		ID:          "tdx-position-decision",
		Name:        "仓位决策",
		Description: "解决仓位管理问题，给出当前市场环境下的仓位建议",
		Markdown: `**Skill 分类**: 仓位管理 / 风险控制 / 交易决策

**适用场景**: 解决用户最核心的问题"现在到底该上仓位还是降仓位"。

**分析框架**:
1. 当前市场风险收益比评估
2. 仓位区间建议（进攻/中性/防御/避险）+ 对应仓位百分比
3. 不同仓位下的操作逻辑
4. 触发仓位调整的条件（什么情况下加仓/减仓）
5. 仓位与当前市场环境的匹配度分析

**仓位级别**: 重仓(80-100%)、偏重(60-80%)、中性(40-60%)、偏轻(20-40%)、轻仓(0-20%)

**输入**: 用户持仓情况、风险偏好、当前市场环境`,
	}
}

// SkillQuantLocal 通达信TQ-Local
func SkillQuantLocal() SkillInfo {
	return SkillInfo{
		ID:          "tdx-quant-local",
		Name:        "通达信TQ-Local",
		Description: "通过本地通达信客户端的HTTP服务直接调用tqcenter接口进行量化投研",
		Markdown: `**Skill 分类**: 量化投研 / 本地数据

**说明**: TdxQuant是由通达信软件提供的证券行情分析和量化投研平台。本技能使用本地通达信客户端的HTTP服务直接调用tqcenter接口，不再生成Python文件。

**前提条件**: 需要本地安装并运行通达信客户端，且tqcenter HTTP服务已开启。

**适用场景**: 通过本地通达信客户端获取行情数据、进行量化策略回测。`,
	}
}

// SkillQuantPython 通达信TQ-Python
func SkillQuantPython() SkillInfo {
	return SkillInfo{
		ID:          "tdx-quant-python",
		Name:        "通达信TQ-Python",
		Description: "通过Python代码与tqcenter接口交互，进行量化策略开发和数据批量获取",
		Markdown: `**Skill 分类**: 量化投研 / Python接口

**说明**: TdxQuant是由通达信软件提供的证券行情分析和量化投研平台，专注于为证券投资者提供行情信息获取、数据分析、策略研究、投资决策和智能交易的全流程解决方案。

本技能支持Python代码通过tqcenter接口与本地的通达信客户端交互。

**前提条件**: 本地通达信客户端运行中，且Python环境中已安装tqcenter相关依赖。

**适用场景**: Python量化策略开发、数据批量获取、回测框架集成。`,
	}
}
