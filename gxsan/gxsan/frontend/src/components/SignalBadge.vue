<template>
  <span class="signal-badge" :class="signalClass">
    {{ signalText }}
  </span>
</template>

<script>
import { signalClass, signalText } from '../utils/signal'

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
      return signalClass(this.signal)
    },
    signalText() {
      return signalText(this.signal)
    }
  }
}
</script>
