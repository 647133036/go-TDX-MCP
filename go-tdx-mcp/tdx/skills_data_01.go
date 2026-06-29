package tdx

// SkillAGZXSB A股市场主线识别
func SkillAGZXSB() SkillInfo {
	return SkillInfo{
		ID:          "tdx-agzxsb",
		Name:        "A股市场主线识别",
		Description: "识别当前A股市场真正的交易主线，分析市场结构、题材周期、资金行为",
		Markdown: "**Skill 分类**\n市场结构 / 题材周期 / 资金行为\n\n" +
			"**适用人群**\n短线交易者、波段交易者、职业投资者、市场研究员\n\n" +
			"**适用场景**\n每天开盘后、午盘、收盘后，用户都需要快速知道：\n" +
			"今天市场到底在交易什么，真正的主线是什么，情绪在什么位置，明天应该盯哪里。\n\n" +
			"**系统提示**\n你是一名中国A股顶级市场结构研究员。基于当前盘面数据识别A股市场真正的交易主线。\n\n" +
			"**分析步骤**\n" +
			"1. 判断市场整体环境（强势/震荡/弱势/退潮）\n" +
			"2. 识别主线与伪主线（从板块涨幅、题材热度、成交额集中度、涨停股数量、连板高度等维度）\n" +
			"3. 识别龙头、中军、补涨和后排（区分情绪龙头、趋势中军、补涨标的、跟风后排）\n" +
			"4. 判断情绪周期（冰点/修复/主升/高位震荡/退潮）\n" +
			"5. 评估主线持续性（产业逻辑、事件催化、资金合力三维度打分）\n" +
			"6. 给出下一交易日观察重点\n\n" +
			"**输出模板**\n【1.市场环境】【2.当前主线】【3.次级热点】【4.核心锚点个股】【5.情绪周期】【6.主线持续性评估】【7.明日观察重点】【8.一句话交易结论】\n\n" +
			"**常用工具**: " + "`tdx_screener`" + ", " + "`tdx_quotes`" + "",
	}
}

// SkillBKBJ 板块比较
func SkillBKBJ() SkillInfo {
	return SkillInfo{
		ID:          "tdx-bkbj",
		Name:        "板块比较",
		Description: "在两个或多个板块之间进行深度对比分析，帮助投资者做出板块选择决策",
		Markdown: "**Skill 分类**\n行业比较 / 轮动研判 / 配置选择\n\n" +
			"**适用人群**\n行业研究员、中短线投资者、主题投资参与者\n\n" +
			"**适用场景**\n用户在两个或多个板块之间难以抉择时使用，如算力 vs CPO、机器人 vs AI应用。\n\n" +
			"**分析框架**\n" +
			"1. 提炼各板块核心逻辑\n" +
			"2. 对比当前交易位置（低位预热/启动确认/主升演绎/高位拥挤/退潮修复）\n" +
			"3. 对比催化强度（硬催化/软催化/存量叙事/预期透支）\n" +
			"4. 对比资金偏好（成交额、龙头强度、中军承接力）\n" +
			"5. 对比拥挤度与估值压力\n" +
			"6. 形成排序（短期/中期优先级）\n\n" +
			"**输出模板**\n" +
			"1.核心逻辑对比表 2.交易位置对比表 3.催化强度对比表 4.资金偏好对比表 5.估值与拥挤度对比表 6.风险收益比排序表 7.配置建议表\n\n" +
			"**常用工具**: " + "`tdx-board-cpbd`" + ", " + "`tdx_kline`" + ", " + "`wenda_news_query`" + ", " + "`wenda_notice_query`" + "",
	}
}

// SkillBoardCPBD 板块操盘必读
func SkillBoardCPBD() SkillInfo {
	return SkillInfo{
		ID:          "tdx-board-cpbd",
		Name:        "板块操盘必读",
		Description: "查询板块基础资料、阶段表现、市场统计等操盘必读数据",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询板块操盘必读数据。\n\n" +
			"**Entry**: " + "`TdxSharePCCW.skef10_bk_cpbd_jczl`" + "\n\n" +
			"**查询类型**:\n" +
			"- basic_info (branch=001): 总市值、PE、PB、成分股数量\n" +
			"- detail (branch=002): 板块分类、解析和关联证券\n" +
			"- stage_return (branch=003, timeType=1m): 阶段涨幅、板块排名\n" +
			"- market_stats (branch=004): 成交额、涨跌家数、市场统计\n\n" +
			"**F10技能**: 当用户想看板块基础资料、阶段表现或市场统计时使用。",
	}
}

// SkillBoardValuation 板块估值
func SkillBoardValuation() SkillInfo {
	return SkillInfo{
		ID:          "tdx-board-valuation",
		Name:        "板块估值",
		Description: "查询个股在板块中的估值对比排名以及板块历史估值走势",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询板块估值数据。\n\n" +
			"**Entry**: " + "`TdxSharePCCW.skef10_hy_hydw_gzsppm`" + "\n\n" +
			"**查询类型**:\n" +
			"- relative_valuation (queryType=01): 个股在板块中的估值对比、排名、与沪深300对比\n" +
			"- history_valuation (queryType=02): 板块或指数历史估值走势\n\n" +
			"**F10技能**: 当用户想看个股在板块中的估值定位或板块历史估值走势时使用。",
	}
}

// SkillBXZJXW 北向资金行为
func SkillBXZJXW() SkillInfo {
	return SkillInfo{
		ID:          "tdx-bxzjxw",
		Name:        "北向资金行为",
		Description: "分析北向资金的流向、偏好方向和交易性质，判断外资对市场风格的指示意义",
		Markdown: "**Skill 分类**\n资金行为 / 市场风格 / 机构偏好分析\n\n" +
			"**适用人群**\n中短线投资者、趋势交易者、机构跟踪用户\n\n" +
			"**适用场景**\n用户想知道北向资金到底在买什么、卖什么，是趋势性信号还是短期扰动。\n\n" +
			"**系统提示**\n你是一名中国A股外资行为研究专家，熟悉北向资金在不同市场阶段的行业偏好和交易特征。\n\n" +
			"**分析框架**\n" +
			"1. 看总体，不只看净值（连续性、集中度、与指数关系、与市场情绪是否共振）\n" +
			"2. 看偏好方向（哪些行业、哪些龙头、哪些风格）\n" +
			"3. 判断行为性质（趋势性布局/事件性交易/被动配置/短期避险）\n" +
			"4. 分析对市场风格的指示意义\n" +
			"5. 提炼跟踪线索\n\n" +
			"**数据获取**\n使用 " + "`tdx_api_data`" + " 查询主力持仓-北向资金：\n" +
			"- entry: TdxSharePCCW.tdxf10_gg_zlcc\n" +
			"- rtype: bszj\n\n" +
			"**输出模板**\n【1.北向资金总体行为】【2.重点偏好方向】【3.行为性质判断】【4.对市场风格的指示意义】【5.后续跟踪线索】【6.风险提示】【7.综合结论】",
	}
}

// SkillCHLTZ 出海链投资
func SkillCHLTZ() SkillInfo {
	return SkillInfo{
		ID:          "tdx-chltz",
		Name:        "出海链投资",
		Description: "分析中国企业出海类型、竞争力、兑现路径和关键风险，判断出海链投资价值",
		Markdown: "**Skill 分类**\n全球化投资 / 制造业研究 / 行业比较优势 / 中长期成长\n\n" +
			"**适用场景**\n用户关注中国企业出海、海外扩张、全球份额提升时。\n\n" +
			"**分析框架**\n" +
			"1. 定义出海逻辑类型（产品出口扩张/海外产能布局/品牌全球化/渠道全球化/技术或服务输出）\n" +
			"2. 判断海外成长驱动\n" +
			"3. 分析竞争力（成本/技术/产品力/响应速度/客户资源/本地化能力）\n" +
			"4. 拆解兑现路径（订单→收入→利润）\n" +
			"5. 识别关键风险（关税与地缘政治/汇率波动/海外本地化失败/渠道费用）\n" +
			"6. 标的分层（全球化龙头/正在突破的二线成长股/交易性主题股/讲故事型弱兑现标的）\n\n" +
			"**数据获取**: 使用 " + "`tdx_api_data`" + " 查询经营分析-主营构成（entry: TdxSharePCCW.tdxf10_gg_jyfx, rtype: zygc）",
	}
}

// SkillCompanyInfo 查询公司信息
func SkillCompanyInfo() SkillInfo {
	return SkillInfo{
		ID:          "tdx-company-info",
		Name:        "查询公司信息",
		Description: "查询公司概要、主营业务、基础资料、董监高、参控股公司等信息",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询公司信息。\n\n" +
			"**查询类型**:\n" +
			"- overview (entry: TdxSharePCCW.tdxf10_gg_zxts, fixedTag: gsgy): 公司概要、主营业务、关联主题\n" +
			"- basic_info (entry: TdxSharePCCW.tdxf10_gg_gsgk, fixedTag: 0): 基础资料、业务分类、ESG报告\n" +
			"- issuance_trading (fixedTag: 8): 上市发行、募资历史\n" +
			"- executives (fixedTag: 20): 董监高名单\n" +
			"- affiliates (fixedTag: 3): 参控股公司\n\n" +
			"**F10技能**: 当用户想看公司概要、主营业务、基础资料或董监高信息时使用。",
	}
}

// SkillCZZDXFXJS 持仓诊断与风险检视
func SkillCZZDXFXJS() SkillInfo {
	return SkillInfo{
		ID:          "tdx-czzdxfxjs",
		Name:        "持仓诊断与风险检视",
		Description: "对用户持仓组合进行结构化诊断，识别风险暴露、结构性问题和调整方向",
		Markdown: "**Skill 分类**\n组合管理 / 风险控制 / 投顾服务\n\n" +
			"**适用场景**\n用户将当前持仓发送后，希望了解组合是否存在隐患、风险敞口在哪里。\n\n" +
			"**分析框架**\n" +
			"1. 组合概况（行业分布、风格分布、仓位水平、持仓数量、集中度）\n" +
			"2. 识别主要风险暴露（单一行业占比过重、高度集中于同一题材、高相关性标的重复配置）\n" +
			"3. 识别结构性问题（龙头标的缺失、跟风品种过多、逻辑主线混乱）\n" +
			"4. 区分核心仓与问题仓（进攻核心、稳定器、观察仓、拖累项）\n" +
			"5. 给出仓位建议（偏高/合理/偏低 + 增仓或减仓的优先方向）\n" +
			"6. 给出调整顺序（先处理什么，后处理什么）\n\n" +
			"**常用工具**: " + "`tdx_indicator_select`" + ", " + "`tdx_quotes`" + ", " + "`tdx_kline`" + ", " + "`wenda_news_query`" + "\n\n" +
			"**输出模板**\n【1.组合总体画像】【2.主要风险暴露】【3.结构性问题】【4.核心仓/观察仓/问题仓】【5.仓位与风格建议】【6.调整优先级】【7.总体结论】",
	}
}

// SkillDividendFinancing 查询分红融资
func SkillDividendFinancing() SkillInfo {
	return SkillInfo{
		ID:          "tdx-dividend-financing",
		Name:        "查询分红融资",
		Description: "查询个股分红、募资、股息率、派现融资比、配股、增发等数据",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询个股分红融资数据。\n\n" +
			"**主入口**: TdxSharePCCW.tdxf10_gg_fhrz\n\n" +
			"**fixedTag映射**:\n" +
			"- pxmz: 分红与募资概览\n" +
			"- fh: 分红图\n" +
			"- fhlszs_glzfl: 分红历史走势-股利支付率\n" +
			"- fhlszs_gxl: 分红历史走势-股息率\n" +
			"- fhpm_glzfl: 分红排名-股利支付率\n" +
			"- fhpm_gxl: 分红排名-股息率\n" +
			"- fhpm_pxrzb: 分红排名-派现融资比\n" +
			"- zf: 增发方案\n" +
			"- zfpg: 增发获配明细\n" +
			"- pf: 配股方案\n\n" +
			"**F10技能**: 当用户提到分红、派现融资比、股息率、配股、增发等场景时使用。",
	}
}

// SkillDragonTiger 查询个股龙虎榜
func SkillDragonTiger() SkillInfo {
	return SkillInfo{
		ID:          "tdx-dragon-tiger",
		Name:        "查询个股龙虎榜",
		Description: "查询指定个股的龙虎榜可用日期和指定日期的席位买卖明细",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询个股龙虎榜数据。\n\n" +
			"**查询类型**:\n" +
			"- dates (entry: TdxSharePCCW.tdxf10_gg_comreq, fixedTag: jglhb): 龙虎榜可用日期列表\n" +
			"- list (entry: TdxSharePCCW.tdxf10_gg_jyds, fixedTag: jglhb, extra: <date>): 指定日期龙虎榜明细、席位画像\n\n" +
			"**F10技能**: 当用户想看某只股票的龙虎榜可用日期或指定日期明细、席位买卖额时使用。\n\n" +
			"**调用方式**:\n先查dates获取可用日期 → 再用list查询具体日期明细",
	}
}

// SkillEarningsWarning 查询个股业绩预警
func SkillEarningsWarning() SkillInfo {
	return SkillInfo{
		ID:          "tdx-earnings-warning",
		Name:        "查询个股业绩预警",
		Description: "查询个股业绩预警数据，包括预告类型、利润变动幅度等",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询个股业绩预警数据。\n\n" +
			"**Entry**: TdxSharePCCW.tdxf9_ag_cwsj_yjyj\n\n" +
			"**参数**: code(6位股票代码) + extra(证券id，如gsSz0000526)\n\n" +
			"**返回字段**: reportPeriod, forecastType, forecastProfit10k, profitChangePct, changeLowerPct, changeUpperPct, isWarning, latestForecastDate\n\n" +
			"**已知状态**: 当前上游返回\"功能未注册\"，需确认服务端已注册对应功能。",
	}
}

// SkillEventDrivenShortTerm 事件驱动与短线催化分析
func SkillEventDrivenShortTerm() SkillInfo {
	return SkillInfo{
		ID:          "tdx-event-driven-short-term-catalyst",
		Name:        "事件驱动与短线催化分析",
		Description: "判断个股在未来3-10个交易日内的交易价值、催化强弱、持续性与风险",
		Markdown: "**Skill 分类**\n事件驱动 / 短线交易 / 预期差博弈\n\n" +
			"**适用人群**\n短线交易者、题材交易者、波段投资者\n\n" +
			"**适用场景**\n判断某只股票在未来3-10个交易日内是否存在交易价值、催化强弱、持续性与风险。\n\n" +
			"**必查数据**\n" +
			"- " + "`tdx_quotes`" + ": 行情、估值、成交、盘口\n" +
			"- " + "`tdx_kline`" + ": K线、趋势位置、波动区间（period=4, wantNum=5）\n" +
			"- " + "`tdx_screener`" + ": 观察涨停分布与梯队高度\n" +
			"- " + "`wenda_news_query`" + ": 事件验证\n" +
			"- " + "`tdx_api_data`" + ": 公司公告、事件标签\n\n" +
			"**分析框架**\n" +
			"1. 判断催化类型（政策/产业/业绩/公告/海外映射/情绪催化）\n" +
			"2. 判断催化强度（强/中/弱/噪音）\n" +
			"3. 判断预期差（市场是否已知、股价是否提前交易、龙头还是后排）\n" +
			"4. 看资金和板块是否支持（强势位置、扩散能力、梯队形成、资金行为）\n" +
			"5. 做路径推演（最强/中性/弱势三种情景）\n" +
			"6. 给出跟踪点与操作建议\n\n" +
			"**数据准确性**: Amount单位是\"元\"，除以1亿得\"亿元\"；Volume单位是\"手\"；Lead已是%；ZSZ/LTGB单位需换算\n\n" +
			"**输出模板**: 1.事件概览 2.催化强度判断 3.预期差分析 4.资金与板块联动 5.三种情景推演 6.关键观察信号 7.交易建议与风险提示",
	}
}

// SkillFHGDHB 分红与股东回报
func SkillFHGDHB() SkillInfo {
	return SkillInfo{
		ID:          "tdx-fhgdhb",
		Name:        "分红与股东回报",
		Description: "评估公司分红质量、资本配置能力和红利资产属性",
		Markdown: "**Skill 分类**\n股东回报 / 价值投资 / 资本配置 / 长线研究\n\n" +
			"**适用人群**\n价值投资者、红利策略用户、中长线配置资金\n\n" +
			"**适用场景**\n判断一家公司分红高不高、是否稳定、是不是\"真红利\"。\n\n" +
			"**分析框架**\n" +
			"1. 构建股东回报画像（分红、回购、再融资、资本开支、自由现金流）\n" +
			"2. 评估分红质量（持续性、比例稳定性、现金流支撑）\n" +
			"3. 分析回购与分红的协同\n" +
			"4. 评估资本配置能力（审慎/激进/低效）\n" +
			"5. 判断是否具备红利资产属性（现金流稳定、商业模式成熟、资本开支可控）\n" +
			"6. 提示风险（周期顶部高分红、可持续性弱、再融资频繁稀释）\n\n" +
			"**数据获取**: 使用" + "`tdx_api_data`" + "查询分红融资模板\n" +
			"- dividend_financing_dividend_and_fundraising: 分红与募资\n" +
			"- dividend_history_trend_dividend_yield: 股息率走势\n" +
			"- dividend_ranking_dividend_yield: 股息率排名",
	}
}

// SkillFinancials 查询财务分析数据
func SkillFinancials() SkillInfo {
	return SkillInfo{
		ID:          "tdx-financials",
		Name:        "查询财务分析数据",
		Description: "查询个股财务摘要、资产负债表、利润表、现金流量表等核心财务数据",
		Markdown: "使用 " + "`tdx_api_data`" + " 查询个股财务分析数据。\n\n" +
			"**Entry**: TdxSharePCCW.tdxf10_gg_zxts\n\n" +
			"**查询类型映射**:\n" +
			"- 财务摘要: 查看公司营收、利润、现金流等核心财务指标\n" +
			"- 资产负债表: 查看资产负债结构\n" +
			"- 利润表: 查看收入成本利润\n" +
			"- 现金流量表: 查看经营活动现金流\n\n" +
			"**F10技能**: 当用户想看公司财务数据、营收利润、资产负债等时使用。",
	}
}

// SkillFSXYPMBS 反身性与泡沫识别
func SkillFSXYPMBS() SkillInfo {
	return SkillInfo{
		ID:          "tdx-fsxypmsb",
		Name:        "反身性与泡沫识别",
		Description: "判断热门板块或个股是否已进入预期自我强化的泡沫阶段，识别基本面驱动与预期驱动",
		Markdown: "**Skill 分类**\n市场行为 / 预期管理 / 风险识别 / 高阶交易研究\n\n" +
			"**适用场景**\n判断一个热门板块或个股是否已进入预期自我强化的泡沫阶段。\n\n" +
			"**分析框架**\n" +
			"分析基本面→预期→价格→行为的正反馈循环是否正在形成或已被打破。\n\n" +
			"使用TDX技能组合自动获取全维度数据：\n" +
			"- tdx-financials: 财务基本面\n" +
			"- tdx-trading-info: 交易信息\n" +
			"- tdx-dragon-tiger: 龙虎榜\n" +
			"- tdx-hot-topic: 热点题材\n" +
			"- tdx-board-valuation: 板块估值\n" +
			"- tdx-report-rating: 研报评级\n" +
			"- tdx-main-position: 主力资金\n" +
			"- tdx-company-info: 公司信息\n" +
			"- tdx-shareholder-research: 股东研究\n" +
			"- tdx-industry-chain: 产业链\n" +
			"- tdx-stock-events: 股票事件\n" +
			"- tdx-board-cpbd: 板块操盘必读\n" +
			"- " + "`tdx_kline`" + ", " + "`tdx_quotes`" + ": 行情K线\n\n" +
			"**输出要点**\n判断是基本面驱动还是预期驱动，识别索罗斯式反身性是否在形成。",
	}
}
