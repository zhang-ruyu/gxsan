<template>
  <div>
    <div class="page-header">
      <h2>养老现金流测算</h2>
      <div class="header-actions">
        <span class="lifecycle-badge">生命周期：{{ lifecycleStage }}</span>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在测算...</p>
    </div>

    <div v-else>
      <!-- 输入参数 -->
      <div class="card">
        <div class="card-header"><span class="card-title">测算参数</span></div>
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">目标月分红 (元)</label>
            <input class="form-input" type="number" v-model.number="params.monthly" />
          </div>
          <div class="form-group">
            <label class="form-label">每月定投 (元)</label>
            <input class="form-input" type="number" v-model.number="params.invest" />
          </div>
          <div class="form-group">
            <label class="form-label">定投年数</label>
            <input class="form-input" type="number" v-model.number="params.years" />
          </div>
          <div class="form-group">
            <label class="form-label">年化收益率 (%)</label>
            <input class="form-input" type="number" step="0.5" v-model.number="params.rate" />
          </div>
        </div>
        <button class="btn btn-primary" @click="calculate">开始测算</button>
      </div>

      <!-- 目标倒推 -->
      <div class="card" v-if="result">
        <div class="card-header"><span class="card-title">目标倒推：所需本金</span></div>
        <p class="hint">所需本金 = 目标年分红 ÷ 成本股息率（年分红 = 月分红 × 12 = ¥{{ formatNumber(result.monthly * 12) }}）</p>
        <table class="table">
          <thead>
            <tr><th>成本股息率</th><th>所需本金</th></tr>
          </thead>
          <tbody>
            <tr v-for="row in result.plan" :key="row.cost_yield">
              <td>{{ (row.cost_yield * 100).toFixed(0) }}%</td>
              <td>¥{{ formatNumber(row.required_principal) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 定投复利模拟 -->
      <div class="card" v-if="result">
        <div class="card-header"><span class="card-title">定投复利模拟（分红100%复投）</span></div>
        <table class="table">
          <thead>
            <tr><th>年份</th><th>总资产</th><th>年股息</th><th>月均股息</th></tr>
          </thead>
          <tbody>
            <tr v-for="s in result.schedule" :key="s.year"
                :class="{ 'highlight': s.year === 1 || s.year % 5 === 0 }">
              <td>{{ s.year }}</td>
              <td>¥{{ formatNumber(s.total_asset) }}</td>
              <td>¥{{ formatNumber(s.annual_dividend) }}</td>
              <td>¥{{ formatNumber(s.monthly_dividend) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 退休提款策略 -->
      <div class="card" v-if="result">
        <div class="card-header"><span class="card-title">退休提款策略</span></div>
        <p class="retirement-note">{{ result.retirement_note }}</p>
      </div>
    </div>
  </div>
</template>

<script>
import { GetPension } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'

export default {
  name: 'Pension',
  data() {
    return {
      loading: false,
      lifecycleStage: '',
      params: { monthly: 5000, invest: 3000, years: 20, rate: 5 },
      result: null
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num).toFixed(0).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    async calculate() {
      this.loading = true
      try {
        const raw = await GetPension(
          this.params.monthly, this.params.invest, this.params.years, this.params.rate
        )
        const data = parseJSON(raw)
        this.result = data
        this.lifecycleStage = data.lifecycle_stage
      } catch (e) {
        alert('测算失败: ' + e.message)
      }
      this.loading = false
    }
  },
  mounted() {
    this.calculate()
  }
}
</script>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}
.hint { color: var(--text-muted); font-size: 13px; margin-bottom: 12px; }
.highlight { background: #f0f7ff; font-weight: 600; }
.retirement-note {
  background: #f6fff6; border-left: 4px solid var(--success-color);
  padding: 12px 16px; border-radius: 4px; line-height: 1.6;
}
.lifecycle-badge {
  font-size: 12px; padding: 4px 10px; border-radius: 12px;
  background: #eef2ff; color: #4f46e5;
}
.header-actions { display: flex; align-items: center; gap: 16px; }
</style>
