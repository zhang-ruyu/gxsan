<template>
  <div>
    <div class="page-header"><h2>资产转换</h2></div>

    <!-- 模式切换 -->
    <div class="mode-switch">
      <button class="btn btn-sm" :class="mode === 'compare' ? 'btn-primary' : 'btn-secondary'"
              @click="mode = 'compare'">个股对比</button>
      <button class="btn btn-sm" :class="mode === 'realestate' ? 'btn-primary' : 'btn-secondary'"
              @click="mode = 'realestate'">房产 → 股权</button>
    </div>

    <!-- 个股对比 -->
    <div class="card" v-if="mode === 'compare'">
      <div class="card-header"><span class="card-title">个股多维对比（股息率 / 估值 / 分红稳定性）</span></div>
      <div class="form-grid">
        <div class="form-group">
          <label class="form-label">股票A代码</label>
          <input class="form-input" v-model="compare.codeA" placeholder="如 600036" />
        </div>
        <div class="form-group">
          <label class="form-label">股票B代码</label>
          <input class="form-input" v-model="compare.codeB" placeholder="如 600938" />
        </div>
      </div>
      <button class="btn btn-primary" @click="doCompare">对比</button>

      <div v-if="compareResult" class="compare-result">
        <table class="table">
          <thead>
            <tr><th>维度</th><th>{{ compareResult.a.name }}</th><th>{{ compareResult.b.name }}</th></tr>
          </thead>
          <tbody>
            <tr><td>当前股息率</td><td>{{ compareResult.a.current_yield.toFixed(2) }}%</td><td>{{ compareResult.b.current_yield.toFixed(2) }}%</td></tr>
            <tr><td>PE</td><td>{{ compareResult.a.pe.toFixed(2) }}</td><td>{{ compareResult.b.pe.toFixed(2) }}</td></tr>
            <tr><td>PB</td><td>{{ compareResult.a.pb.toFixed(2) }}</td><td>{{ compareResult.b.pb.toFixed(2) }}</td></tr>
            <tr><td>连续分红</td><td>{{ compareResult.a.dividend_years }}年</td><td>{{ compareResult.b.dividend_years }}年</td></tr>
            <tr v-if="compareResult.a.yield_on_cost || compareResult.b.yield_on_cost">
              <td>成本股息率</td><td>{{ compareResult.a.yield_on_cost.toFixed(2) }}%</td><td>{{ compareResult.b.yield_on_cost.toFixed(2) }}%</td></tr>
          </tbody>
        </table>
        <div class="conclusion" :class="betterClass">
          结论：{{ betterText }} 更优 — {{ compareResult.reason }}
        </div>
      </div>
    </div>

    <!-- 房产转股权 -->
    <div class="card" v-if="mode === 'realestate'">
      <div class="card-header"><span class="card-title">房产 → 红利股权 现金流对比</span></div>
      <div class="form-grid">
        <div class="form-group">
          <label class="form-label">本金 (元)</label>
          <input class="form-input" type="number" v-model.number="re.principal" />
        </div>
        <div class="form-group">
          <label class="form-label">租售比 (%)</label>
          <input class="form-input" type="number" step="0.1" v-model.number="re.reYield" />
        </div>
        <div class="form-group">
          <label class="form-label">红利股息率 (%)</label>
          <input class="form-input" type="number" step="0.1" v-model.number="re.eqYield" />
        </div>
      </div>
      <button class="btn btn-primary" @click="doRealEstate">计算</button>

      <div v-if="reResult" class="compare-result">
        <table class="table">
          <thead><tr><th>指标</th><th>房产(租售)</th><th>红利股权</th></tr></thead>
          <tbody>
            <tr><td>年现金流</td><td>¥{{ formatNumber(reResult.real_estate_income) }}</td><td>¥{{ formatNumber(reResult.equity_income) }}</td></tr>
            <tr><td>现金流倍数</td><td colspan="2" class="text-center">股权 / 房产 = {{ reResult.income_multiple.toFixed(2) }} 倍</td></tr>
          </tbody>
        </table>
        <p class="hint">转换后现金流翻 {{ reResult.income_multiple.toFixed(1) }} 倍（且 T+1 变现、零维护）</p>
      </div>
    </div>
  </div>
</template>

<script>
import { CompareStocks, RealEstateToEquity } from '../../wailsjs/go/main/App'

export default {
  name: 'Conversion',
  data() {
    return {
      mode: 'compare',
      compare: { codeA: '', codeB: '' },
      compareResult: null,
      re: { principal: 1000000, reYield: 1.5, eqYield: 5 },
      reResult: null
    }
  },
  computed: {
    betterClass() {
      if (!this.compareResult) return ''
      return this.compareResult.better === 'A' ? 'better-a'
          : this.compareResult.better === 'B' ? 'better-b' : 'better-flat'
    },
    betterText() {
      if (!this.compareResult) return ''
      return this.compareResult.better === 'A' ? this.compareResult.a.name
          : this.compareResult.better === 'B' ? this.compareResult.b.name : '两者'
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num).toFixed(0).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    async doCompare() {
      if (!this.compare.codeA || !this.compare.codeB) { alert('请输入两只股票代码'); return }
      try {
        const raw = await CompareStocks(this.compare.codeA, this.compare.codeB)
        this.compareResult = JSON.parse(raw)
      } catch (e) { alert('对比失败: ' + e.message) }
    },
    async doRealEstate() {
      try {
        const raw = await RealEstateToEquity(this.re.principal, this.re.reYield, this.re.eqYield)
        this.reResult = JSON.parse(raw)
      } catch (e) { alert('计算失败: ' + e.message) }
    }
  }
}
</script>

<style scoped>
.mode-switch { display: flex; gap: 12px; margin-bottom: 16px; }
.form-grid {
  display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px; margin-bottom: 16px;
}
.compare-result { margin-top: 16px; }
.conclusion {
  margin-top: 12px; padding: 12px 16px; border-radius: 4px; line-height: 1.6; font-weight: 600;
}
.better-a { background: #eef7ff; border-left: 4px solid #3b82f6; }
.better-b { background: #f0fff4; border-left: 4px solid var(--success-color); }
.better-flat { background: #f5f5f5; border-left: 4px solid #999; }
.hint { color: var(--text-muted); font-size: 13px; margin-top: 8px; }
.text-center { text-align: center; }
</style>
