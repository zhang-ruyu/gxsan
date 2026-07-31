<template>
  <span class="signal-badge" :class="signalClass">
    {{ signalText }}
  </span>
</template>

<script>
export default {
  name: 'SignalBadge',
  props: {
    signal: {
      type: String,
      default: 'WATCH'
    }
  },
  computed: {
    signalClass() {
      const map = {
        'BUY': 'signal-buy',
        'HOLD': 'signal-hold',
        'WATCH': 'signal-watch',
        'SELL': 'signal-sell'
      }
      return map[this.signal] || 'signal-watch'
    },
    signalText() {
      const map = {
        'BUY': '买入',
        'HOLD': '持有',
        'WATCH': '观望',
        'SELL': '卖出'
      }
      return map[this.signal] || '观望'
    }
  }
}
</script>
