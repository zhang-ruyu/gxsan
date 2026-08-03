// 智能刷新辅助：判断页面是否可见、A股是否处于交易时段。
// 自动刷新前先做这两道判断，避免在标签页隐藏或非交易时段无意义地打行情接口。

// isPageVisible 标签页是否可见（隐藏时不应发起刷新，省流量/CPU）。
export function isPageVisible() {
  return typeof document === 'undefined' || document.visibilityState !== 'hidden'
}

// isMarketOpen 当前是否处于 A 股交易时段（周一~周五 9:30-11:30 / 13:00-15:00）。
// 说明：未包含法定节假日；闭市后、周末、午休一律视为非交易时段，暂停自动刷新。
export function isMarketOpen() {
  const now = new Date()
  const day = now.getDay() // 0=周日 6=周六
  if (day === 0 || day === 6) return false
  const mins = now.getHours() * 60 + now.getMinutes()
  const morningStart = 9 * 60 + 30
  const morningEnd = 11 * 60 + 30
  const afternoonStart = 13 * 60
  const afternoonEnd = 15 * 60
  return (mins >= morningStart && mins <= morningEnd) ||
         (mins >= afternoonStart && mins <= afternoonEnd)
}

// refreshReason 返回当前自动刷新状态的可读描述，用于在界面上提示用户。
export function refreshReason() {
  if (!isPageVisible()) return '已暂停（页面隐藏）'
  if (!isMarketOpen()) return '已暂停（非交易时段）'
  return '交易中（每30秒自动刷新）'
}
