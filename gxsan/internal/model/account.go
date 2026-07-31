package model

// Account 资金账户（对应技能「多账户体系」）
// 将资金按用途物理隔离，各自目标/策略/风险偏好不同。
type Account struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // 养老/教��/港美股/娱乐/打新
}

// 账户类型常量
const (
	AccountTypePension = "养老"
	AccountTypeEdu     = "教育"
	AccountTypeGlobal  = "港美股"
	AccountTypeFun     = "娱乐"
	AccountTypeIPO     = "打新"
)

// DefaultAccounts 返回默认账户集合
func DefaultAccounts() []Account {
	return []Account{
		{Name: "养老账户", Type: AccountTypePension},
		{Name: "教育基金", Type: AccountTypeEdu},
		{Name: "港美股账户", Type: AccountTypeGlobal},
		{Name: "娱乐账户", Type: AccountTypeFun},
		{Name: "打新账户", Type: AccountTypeIPO},
	}
}

// LifecycleStageName 生命周期阶段中文名（启动/滚雪球/自由/收获）
func LifecycleStageName(stage int) string {
	switch stage {
	case 1:
		return "启动期"
	case 2:
		return "滚雪球期"
	case 3:
		return "自由期"
	case 4:
		return "收获期"
	default:
		return "滚雪球期"
	}
}
