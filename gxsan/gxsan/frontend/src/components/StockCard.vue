<template>
  <div class="card stock-card" @click="$emit('click')">
    <div class="stock-info">
      <div>
        <div class="stock-name">{{ stock.name }}</div>
        <div class="stock-code">{{ stock.code }}</div>
      </div>
      <div style="text-align: right">
        <div class="stock-price" :class="stock.change >= 0 ? 'price-up' : 'price-down'">
          ¥{{ stock.price.toFixed(2) }}
        </div>
        <div class="stock-change" :class="stock.change >= 0 ? 'price-up' : 'price-down'">
          {{ stock.change >= 0 ? '+' : '' }}{{ stock.change.toFixed(2) }}%
        </div>
      </div>
    </div>
    <div class="stock-metrics">
      <div class="metric">
        <div class="metric-label">股息率</div>
        <div class="metric-value" :class="stock.dividend_yield >= stock.target_yield ? 'price-up' : 'price-down'">
          {{ stock.dividend_yield.toFixed(2) }}%
        </div>
      </div>
      <div class="metric">
        <div class="metric-label">目标股息率</div>
        <div class="metric-value">{{ stock.target_yield.toFixed(2) }}%</div>
      </div>
      <div class="metric">
        <div class="metric-label">信号</div>
        <div class="metric-value">
          <span class="signal-badge" :class="signalClass">{{ signalText }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { signalClass, signalText } from '../utils/signal'

export default {
  name: 'StockCard',
  props: {
    stock: {
      type: Object,
      required: true
    }
  },
  emits: ['click'],
  computed: {
    signalClass() {
      return signalClass(this.stock.signal)
    },
    signalText() {
      return signalText(this.stock.signal)
    }
  }
}
</script>
