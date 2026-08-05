package model

// TrackingStock 跟踪推荐股票（来自 skill 推荐池，数据截止 2026.7.25）
type TrackingStock struct {
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	SkillYield       string  `json:"skill_yield"`        // skill 中的静态股息率描述
	AssetType        string  `json:"asset_type"`         // 弱周期 / 顺周期
	CoreLogic        string  `json:"core_logic"`         // 核心逻辑
	KeyRisk          string  `json:"key_risk"`           // 关键风险
	Advice           string  `json:"advice"`             // 操作建议
	// 以下为运行时实时填充
	Price            float64 `json:"price"`              // 实时价格
	CurrentYield     float64 `json:"current_yield"`     // 实时静态股息率
	DividendPerShare float64 `json:"dividend_per_share"` // 每股分红
}

// TrackingCategory 跟踪推荐分类
type TrackingCategory struct {
	Name   string           `json:"name"`   // 分类名称
	Stocks []TrackingStock  `json:"stocks"`
}

// TrackingCategories 跟踪推荐股票池（源自 A 股分红投资 skill）
// 数据来源：微信公众号「分红养老之路」（森哥）实战周记
// 最后更新：2026年7月25日
var TrackingCategories = []TrackingCategory{
	{
		Name: "银行/保险",
		Stocks: []TrackingStock{
			{Code: "601328", Name: "交通银行", SkillYield: "4.66%", AssetType: "弱周期", CoreLogic: "六大行股息率最高", KeyRisk: "增资摊薄(工/农)", Advice: "攒股区5-7%积极低吸"},
			{Code: "601658", Name: "邮储银行", SkillYield: "4.30%", AssetType: "弱周期", CoreLogic: "网点下沉护城河", KeyRisk: "增资摊薄", Advice: "合理区4-4.5%持有"},
			{Code: "601398", Name: "工商银行", SkillYield: "4.10%", AssetType: "弱周期", CoreLogic: "全球最大银行", KeyRisk: "特别国债增资", Advice: "减仓区<3.5%网格卖"},
			{Code: "601939", Name: "建设银行", SkillYield: "4.09%", AssetType: "弱周期", CoreLogic: "房贷+基建龙头", KeyRisk: "特别国债增资", Advice: "GJD压盘≠基本面恶化"},
			{Code: "601988", Name: "中国银行", SkillYield: "3.99%", AssetType: "弱周期", CoreLogic: "国际化程度高", KeyRisk: "特别国债增资", Advice: "国家不倒，分红不少"},
			{Code: "601288", Name: "农业银行", SkillYield: "3.84%", AssetType: "弱周期", CoreLogic: "县域金融壁垒", KeyRisk: "特别国债增资", Advice: "等科技降温，红利估值修复"},
			{Code: "600036", Name: "招商银行", SkillYield: "~5.5%", AssetType: "弱周期", CoreLogic: "零售之王、确定性最强", KeyRisk: "经济下行不良率", Advice: "银行板块首选底仓"},
			{Code: "601318", Name: "中国平安", SkillYield: "~5%+", AssetType: "弱周期", CoreLogic: "PEV 0.68、14年分红增长", KeyRisk: "牛市弹性大但波动也大", Advice: "不适合养老底仓，适合弹性补充"},
		},
	},
	{
		Name: "水电/油气",
		Stocks: []TrackingStock{
			{Code: "600900", Name: "长江电力", SkillYield: "~3.7%", AssetType: "弱周期", CoreLogic: "70%分红至2030，波动极低", KeyRisk: "来水偏枯", Advice: "压舱石首选，底仓长期不动"},
			{Code: "600886", Name: "国投电力", SkillYield: "~3-4%", AssetType: "弱周期", CoreLogic: "水火并济，分红稳定增长", KeyRisk: "火电扭亏不确定", Advice: "水电辅仓"},
			{Code: "600938", Name: "中国海油", SkillYield: "~4-5%", AssetType: "顺周期", CoreLogic: "成本<30美元/桶", KeyRisk: "油价回落业绩分红双杀", Advice: "高景气等6-7%股息率；A等32元(4.5%+)低吸"},
		},
	},
	{
		Name: "煤炭/资源/有色",
		Stocks: []TrackingStock{
			{Code: "601088", Name: "中国神华", SkillYield: "~5-6%", AssetType: "顺周期", CoreLogic: "十年均值7.3%，≥65%分红三年", KeyRisk: "煤价下行周期", Advice: "50元启动减仓，大跌大买"},
			{Code: "601225", Name: "陕西煤业", SkillYield: "~5-7%", AssetType: "顺周期", CoreLogic: "高ROE、低成本高弹性", KeyRisk: "煤价下行周期", Advice: "28元启动减仓"},
			{Code: "601899", Name: "紫金矿业", SkillYield: "~2-3%", AssetType: "顺周期", CoreLogic: "金铜长期波浪上涨，成长>股息", KeyRisk: "金铜暴跌", Advice: "34→32→30→28→26→24网格，目标成本28"},
			{Code: "000408", Name: "藏格矿业", SkillYield: "~3-4%", AssetType: "顺周期", CoreLogic: "钾锂双资源", KeyRisk: "锂价波动", Advice: "小仓位立足收息"},
			{Code: "000933", Name: "神火股份", SkillYield: "~7%+", AssetType: "顺周期", CoreLogic: "铝价+5%→业绩+14%", KeyRisk: "铝价下跌", Advice: "等分红率提至50%→冲刺10%"},
			{Code: "000807", Name: "云铝股份", SkillYield: "~7%+", AssetType: "顺周期", CoreLogic: "安全边际高", KeyRisk: "铝价下跌", Advice: "等分红率提至50%→冲刺10%"},
		},
	},
	{
		Name: "通信/运营商",
		Stocks: []TrackingStock{
			{Code: "600941", Name: "中国移动", SkillYield: "~4-5%", AssetType: "弱周期", CoreLogic: "≥75%分红承诺，弱周期高息品种", KeyRisk: "历史弹性大(跌30%+→翻倍)", Advice: "愿承担波动选移动，求稳选长电"},
			{Code: "601728", Name: "中国电信", SkillYield: "~4-5%", AssetType: "弱周期", CoreLogic: "派息稳定、低估值", KeyRisk: "算力资本开支压力", Advice: "子女教育账户持有"},
		},
	},
	{
		Name: "制造/消费",
		Stocks: []TrackingStock{
			{Code: "000333", Name: "美的集团", SkillYield: "~4-5%", AssetType: "顺周期", CoreLogic: "出海+分红回购，等效分红率8%+", KeyRisk: "汇率+高基数短期承压", Advice: "底仓重仓，逢高网格减不超15%"},
			{Code: "000651", Name: "格力电器", SkillYield: "~5%+", AssetType: "顺周期", CoreLogic: "百亿回购+分红给力", KeyRisk: "内需依赖度高", Advice: "高股息区间攒股"},
			{Code: "600690", Name: "海尔智家", SkillYield: "~2-3%", AssetType: "顺周期", CoreLogic: "出海龙头，等效8%+", KeyRisk: "股息率偏低", Advice: "20以内建仓"},
			{Code: "600887", Name: "伊利股份", SkillYield: "~4-5%", AssetType: "顺周期", CoreLogic: "奶粉双位数增长", KeyRisk: "消费疲软", Advice: "攒股池标的，等更好价格"},
			{Code: "600066", Name: "宇通客车", SkillYield: "~5-7%", AssetType: "顺周期", CoreLogic: "出口高增长", KeyRisk: "周期性订单", Advice: "年报季低吸"},
		},
	},
	{
		Name: "白酒/医药",
		Stocks: []TrackingStock{
			{Code: "600519", Name: "贵州茅台", SkillYield: "~3.6%", AssetType: "顺周期", CoreLogic: "每股分红52元，只能吃息", KeyRisk: "业绩低增速、批价未回暖", Advice: "非底仓，等消费复苏信号"},
			{Code: "000858", Name: "五粮液", SkillYield: "~6%", AssetType: "顺周期", CoreLogic: "股息率已具吸引力", KeyRisk: "需求疲软", Advice: "小仓位吃息，等戴维斯双击"},
			{Code: "000568", Name: "泸州老窖", SkillYield: "~6%", AssetType: "顺周期", CoreLogic: "股息率已具吸引力", KeyRisk: "需求疲软", Advice: "小仓位吃息，等戴维斯双击"},
			{Code: "000538", Name: "云南白药", SkillYield: "~2-3%", AssetType: "顺周期", CoreLogic: "中药龙头+消费属性", KeyRisk: "股息率偏低", Advice: "银发经济长期持有"},
			{Code: "000423", Name: "东阿阿胶", SkillYield: "~5.5-6%", AssetType: "顺周期", CoreLogic: "滋补龙头，历史最高股息率", KeyRisk: "消费悲观持续", Advice: "攒股窗口，每年复投"},
			{Code: "600750", Name: "江中药业", SkillYield: "~5%+", AssetType: "顺周期", CoreLogic: "中药消费化", KeyRisk: "消费悲观持续", Advice: "攒股窗口，每年复投"},
		},
	},
	{
		Name: "红利ETF",
		Stocks: []TrackingStock{
			{Code: "512890", Name: "红利低波ETF", SkillYield: "~4.5-5.5%", AssetType: "弱周期", CoreLogic: "不研究个股者，养老底仓50-60%", KeyRisk: "系统性回撤", Advice: "一键配置红利组合"},
		},
	},
}

// TrackingCategoryNames 返回所有分类名称
func TrackingCategoryNames() []string {
	names := make([]string, len(TrackingCategories))
	for i, c := range TrackingCategories {
		names[i] = c.Name
	}
	return names
}

// TrackingAllCodes 返回所有跟踪推荐股票的代码（去重）
func TrackingAllCodes() []string {
	seen := map[string]bool{}
	var codes []string
	for _, cat := range TrackingCategories {
		for _, s := range cat.Stocks {
			if !seen[s.Code] {
				seen[s.Code] = true
				codes = append(codes, s.Code)
			}
		}
	}
	return codes
}
