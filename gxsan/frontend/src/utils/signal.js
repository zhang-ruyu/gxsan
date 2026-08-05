// 信号 → 样式/文案 的统一映射。
// SignalBadge 组件与 Detail.vue 等共用，避免各页面重复实现同一套映射逻辑。

const SIGNAL_MAP = {
  BUY: { cls: 'signal-buy', text: '建议买入' },
  HOLD: { cls: 'signal-hold', text: '持有' },
  SELL: { cls: 'signal-sell', text: '建议卖出' },
  WATCH: { cls: 'signal-watch', text: '关注中' }
}

// signalClass 返回信号对应的 CSS 类名（默认关注中）。
export function signalClass(signal) {
  const m = SIGNAL_MAP[signal]
  return m ? m.cls : 'signal-watch'
}

// signalText 返回信号对应的中文文案（默认关注中）。
export function signalText(signal) {
  const m = SIGNAL_MAP[signal]
  return m ? m.text : '关注中'
}
