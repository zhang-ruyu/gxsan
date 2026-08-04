package model

// LifecycleStageName 投资生命周期阶段中文名（启动/滚雪球/自由/收获）
// 用于养老现金流测算的阶段化策略提示，与持仓账户体系无关。
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
