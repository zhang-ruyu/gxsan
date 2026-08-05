<template>
  <div>
    <div class="page-header">
      <h2>网格策略</h2>
      <div class="header-actions">
        <span class="text-muted small">为每只持仓设定「股息率到多少 → 加仓/减仓多少」的纪律，打开软件即可看到当前该买还是该卖。</span>
      </div>
    </div>

    <div v-if="holdings.length === 0" class="empty-state">
      <p>还没有持仓</p>
      <p class="text-muted">去「持仓管理」添加股票后，这里可以逐只设置网格策略。</p>
      <router-link to="/portfolio" class="btn btn-primary btn-sm">去添加持仓</router-link>
    </div>

    <div v-else class="grid-layout">
      <!-- 左：持仓选择 + 当前状态 -->
      <div class="grid-left">
        <div class="card">
          <div class="card-header"><span class="card-title">选择持仓</span></div>
          <select v-model="selectedCode" class="select" @change="loadGrid">
            <option v-for="h in holdings" :key="h.code" :value="h.code">
              {{ h.name }} ({{ h.code }})
            </option>
          </select>

          <div v-if="grid" class="status-panel" :class="'status-' + grid.action.toLowerCase()">
            <div class="status-yield">
              当前股息率 <b>{{ grid.current_yield > 0 ? grid.current_yield.toFixed(2) + '%' : '—' }}</b>
            </div>
            <div class="status-action">
              <span class="action-badge" :class="'badge-' + grid.action.toLowerCase()">{{ actionText(grid.action) }}</span>
            </div>
            <div class="status-reason text-muted small">{{ grid.reason }}</div>
            <div class="status-detail" v-if="grid.action === 'BUY'">
              建议投入 <b class="text-orange">¥{{ formatNumber(grid.buy_amount) }}</b>
              （第 {{ grid.buy_level }} 档）
            </div>
            <div class="status-detail" v-else-if="grid.action === 'SELL'">
              建议卖出 <b class="text-red">{{ grid.sell_percent }}%</b>
              （第 {{ grid.sell_level }} 档）
            </div>
          </div>
        </div>
      </div>

      <!-- 右：网格编辑 -->
      <div class="grid-right">
        <div class="card">
          <div class="card-header"><span class="card-title">买入网格（股息率越高越便宜，越该加仓）</span></div>
          <table class="grid-table">
            <thead>
              <tr><th>档位</th><th>股息率 ≥ (%)</th><th>加仓金额 (¥)</th><th>当前</th></tr>
            </thead>
            <tbody>
              <tr v-for="(g, i) in buyGrids" :key="'b' + i">
                <td>{{ i + 1 }}</td>
                <td><input type="number" step="0.1" v-model.number="g.yield" class="input num" /></td>
                <td><input type="number" step="100" v-model.number="g.amount" class="input num" /></td>
                <td class="cell-state" :class="buyState(i)">{{ buyState(i) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="card">
          <div class="card-header"><span class="card-title">卖出网格（股息率越低越贵，越该减仓）</span></div>
          <table class="grid-table">
            <thead>
              <tr><th>档位</th><th>股息率 ≤ (%)</th><th>卖出比例 (%)</th><th>当前</th></tr>
            </thead>
            <tbody>
              <tr v-for="(g, i) in sellGrids" :key="'s' + i">
                <td>{{ i + 1 }}</td>
                <td><input type="number" step="0.1" v-model.number="g.yield" class="input num" /></td>
                <td><input type="number" step="5" v-model.number="g.amount" class="input num" /></td>
                <td class="cell-state" :class="sellState(i)">{{ sellState(i) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="actions">
          <button class="btn btn-primary" @click="save" :disabled="saving">
            {{ saving ? '保存中...' : '保存网格策略' }}
          </button>
          <button class="btn btn-secondary" @click="loadGrid">重置</button>
          <span v-if="savedTip" class="saved-tip text-green">✓ 已保存</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetPortfolio, GetHoldingGrid, SaveHoldingGrid } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'

export default {
  name: 'Grid',
  data() {
    return {
      holdings: [],
      selectedCode: '',
      grid: null,
      buyGrids: this.emptyGrids(),
      sellGrids: this.emptyGrids(),
      saving: false,
      savedTip: false
    }
  },
  methods: {
    emptyGrids() {
      return [
        { yield: 0, amount: 0 },
        { yield: 0, amount: 0 },
        { yield: 0, amount: 0 },
        { yield: 0, amount: 0 },
        { yield: 0, amount: 0 }
      ]
    },
    formatNumber(num) {
      return Number(num || 0).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    actionText(a) {
      return { BUY: '加仓', SELL: '减仓', HOLD: '持有', NONE: '未设置' }[a] || '持有'
    },
    async loadHoldings() {
      try {
        const str = await GetPortfolio()
        this.holdings = parseJSON(str) || []
        if (this.holdings.length > 0 && !this.selectedCode) {
          this.selectedCode = this.holdings[0].code
        }
        if (this.selectedCode) this.loadGrid()
      } catch (e) {
        console.error('加载持仓失败', e)
      }
    },
    async loadGrid() {
      if (!this.selectedCode) return
      try {
        const str = await GetHoldingGrid(this.selectedCode)
        const g = parseJSON(str)
        if (!g || !g.code) return
        this.grid = g
        // 复制网格（避免直接改源对象）
        this.buyGrids = (g.buy_grids || this.emptyGrids()).map(x => ({ yield: x.yield, amount: x.amount }))
        this.sellGrids = (g.sell_grids || this.emptyGrids()).map(x => ({ yield: x.yield, amount: x.amount }))
        this.savedTip = false
      } catch (e) {
        console.error('加载网格失败', e)
      }
    },
    buyState(i) {
      if (!this.grid || this.grid.action !== 'BUY') return ''
      return this.grid.buy_level === i + 1 ? 'now' : (this.grid.buy_level > i + 1 ? 'passed' : '')
    },
    sellState(i) {
      if (!this.grid || this.grid.action !== 'SELL') return ''
      return this.grid.sell_level === i + 1 ? 'now' : (this.grid.sell_level > i + 1 ? 'passed' : '')
    },
    async save() {
      this.saving = true
      this.savedTip = false
      try {
        await SaveHoldingGrid(this.selectedCode, JSON.stringify(this.buyGrids), JSON.stringify(this.sellGrids))
        await this.loadGrid()
        this.savedTip = true
        setTimeout(() => { this.savedTip = false }, 2500)
      } catch (e) {
        console.error('保存失败', e)
        alert('保存失败：' + (e.message || e))
      }
      this.saving = false
    }
  },
  mounted() {
    this.loadHoldings()
  }
}
</script>

<style scoped>
.grid-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 16px;
  align-items: start;
}
@media (max-width: 900px) {
  .grid-layout { grid-template-columns: 1fr; }
}

.select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #d6dce5;
  border-radius: 8px;
  font-size: 14px;
  background: #fff;
}

.status-panel {
  margin-top: 16px;
  padding: 14px;
  border-radius: 10px;
  background: #f8fafc;
  border-left: 4px solid #94a3b8;
}
.status-panel.status-buy { border-left-color: #22c55e; }
.status-panel.status-sell { border-left-color: #ef4444; }
.status-panel.status-hold { border-left-color: #94a3b8; }
.status-panel.status-none { border-left-color: #cbd5e1; background: #f8fafc; }

.status-yield { font-size: 15px; }
.status-action { margin: 8px 0; }
.action-badge {
  display: inline-block;
  padding: 2px 12px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}
.badge-buy { background: #22c55e; }
.badge-sell { background: #ef4444; }
.badge-hold { background: #94a3b8; }
.status-detail { margin-top: 6px; font-size: 14px; }

.grid-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 8px;
}
.grid-table th {
  text-align: left;
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
  padding: 6px 8px;
  border-bottom: 1px solid #eef1f5;
}
.grid-table td {
  padding: 6px 8px;
  border-bottom: 1px solid #f4f6f9;
}
.input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #d6dce5;
  border-radius: 6px;
  font-size: 14px;
}
.input.num { max-width: 140px; }

.cell-state {
  font-size: 12px;
  color: var(--text-muted);
}
.cell-state.now { color: #fff; background: #4f46e5; border-radius: 6px; text-align: center; }
.cell-state.passed { color: #16a34a; }

.actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.saved-tip { font-size: 13px; }

.text-orange { color: #e8590c; }
.text-red { color: var(--danger-color); }
.text-green { color: var(--success-color); }
.text-muted { color: var(--text-muted); }
.small { font-size: 12px; }

.btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
  font-size: 14px;
}
.btn-primary { background: #4f46e5; color: #fff; }
.btn-primary:disabled { opacity: 0.6; cursor: default; }
.btn-secondary { background: #fff; border-color: #d6dce5; color: #475569; }

.empty-state {
  padding: 40px;
  text-align: center;
  background: #fff;
  border-radius: 12px;
}
</style>
