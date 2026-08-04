<template>
  <div>
    <div class="page-header">
      <h2>持仓管理</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">
          刷新数据
        </button>
        <button class="btn btn-primary btn-sm" @click="openAddModal">添加持仓</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在加载持仓数据...</p>
    </div>

    <div v-else>
      <!-- 持仓列表 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">持仓（{{ holdings.length }}只 · 总市值 ¥{{ formatNumber(totalMarketValue) }}）</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>代码</th>
              <th>名称</th>
              <th>持股数</th>
              <th>成本价</th>
              <th>现价</th>
              <th>市值</th>
              <th>盈亏</th>
              <th>收益率</th>
              <th>成本股息率</th>
              <th>行动提示</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="stock in holdings" :key="stock.code">
              <td class="font-bold">{{ stock.code }}</td>
              <td>{{ stock.name }}</td>
              <td>{{ stock.shares }}</td>
              <td>¥{{ fmtCost(stock.avg_cost) }}</td>
              <td>¥{{ stock.price.toFixed(2) }}</td>
              <td>¥{{ formatNumber(stock.market_value) }}</td>
              <td :class="stock.profit >= 0 ? 'text-green' : 'text-red'">
                ¥{{ formatNumber(stock.profit) }}
              </td>
              <td :class="stock.profit_pct >= 0 ? 'text-green' : 'text-red'">
                {{ stock.profit_pct.toFixed(2) }}%
              </td>
              <td :class="stock.yield_on_cost > 0 ? 'text-orange' : ''">
                {{ stock.yield_on_cost > 0 ? stock.yield_on_cost.toFixed(2) + '%' : '-' }}
              </td>
              <td class="advice-cell">
                <template v-if="adviceMap[stock.code]">
                  <SignalBadge :signal="adviceMap[stock.code].action" />
                  <div class="advice-detail">
                    <span v-if="adviceMap[stock.code].action === 'BUY'" class="text-orange">
                      投入 ¥{{ formatNumber(adviceMap[stock.code].suggested_buy_amount) }}
                    </span>
                    <span v-else-if="adviceMap[stock.code].action === 'SELL'" class="text-red">
                      卖 {{ adviceMap[stock.code].suggested_sell_shares }} 股
                    </span>
                    <span class="advice-reason">{{ adviceMap[stock.code].reason }}</span>
                    <span v-if="adviceMap[stock.code].cost_correctable" class="cost-tip">
                      ⚠️ 成本可能失真，可一键修正
                    </span>
                  </div>
                </template>
                <span v-else class="text-muted">—</span>
              </td>
              <td>
                <button class="btn btn-sm btn-secondary" @click="editHolding(stock)">编辑</button>
                <button class="btn btn-sm btn-danger" @click="removeHolding(stock.code)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="holdings.length === 0" class="empty-state">
        <p>暂无持仓</p>
      </div>
    </div>

    <!-- 添加/编辑持仓弹窗 -->
    <div class="modal" v-if="showModal">
      <div class="modal-content">
        <h3>{{ isEditing ? '编辑持仓' : '添加持仓' }}</h3>
        <div class="form-group">
          <label class="form-label">股票代码</label>
          <input class="form-input" v-model="formData.code" :disabled="isEditing" placeholder="例如: 601398">
        </div>
        <div class="form-group">
          <label class="form-label">股票名称</label>
          <input class="form-input" v-model="formData.name" :disabled="isEditing" placeholder="例如: 工商银行">
        </div>
        <div class="form-group">
          <label class="form-label">持股数量</label>
          <input class="form-input" type="number" v-model.number="formData.shares" placeholder="1000">
        </div>
        <div class="form-group">
          <label class="form-label">成本价 (元)</label>
          <input class="form-input" type="number" step="0.0001" v-model.number="formData.avg_cost" placeholder="5.0000">
        </div>
        <div class="form-group">
          <label class="form-label">真实投入总额 (元)</label>
          <input class="form-input" type="number" step="0.01" v-model.number="formData.total_cost" placeholder="含手续费的实际花费，如 12345.67">
          <span class="form-hint">填此项可一键推导精确每股成本（总价÷股数），避免手动录入被四舍五入</span>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveHolding">{{ isEditing ? '保存' : '添加' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetPortfolio, AddHolding, RemoveHolding, GetActionAdvice, CorrectHoldingCost } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'
import SignalBadge from '../components/SignalBadge.vue'

export default {
  name: 'Portfolio',
  components: { SignalBadge },
  data() {
    return {
      loading: true,
      holdings: [],
      advices: [],
      showModal: false,
      isEditing: false,
      formData: {
        code: '',
        name: '',
        shares: 0,
        avg_cost: 0,
        total_cost: 0
      },
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    totalMarketValue() {
      return this.holdings.reduce((sum, s) => sum + (s.market_value || 0), 0)
    },
    adviceMap() {
      const m = {}
      for (const a of this.advices) {
        m[a.code] = a
      }
      return m
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num || 0).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    // 成本价按每股精度展示：最多4位小数、去尾零（A股总价/股数常产生4位小数）
    fmtCost(num) {
      const v = Number(num || 0)
      return parseFloat(v.toFixed(4)).toString()
    },
    async refreshData() {
      try {
        this.refreshState = refreshReason()
        console.log('开始刷新持仓数据')
        const [portfolioStr, adviceStr] = await Promise.all([
          GetPortfolio(),
          GetActionAdvice()
        ])
        console.log('获取到持仓:', portfolioStr)
        this.holdings = parseJSON(portfolioStr)
        this.advices = parseJSON(adviceStr)
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取持仓数据失败:', error)
        alert('刷新失败: ' + error.message)
      }
      this.loading = false
    },
    openAddModal() {
      this.isEditing = false
      this.formData = { code: '', name: '', shares: 0, avg_cost: 0, total_cost: 0 }
      this.showModal = true
    },
    editHolding(stock) {
      this.isEditing = true
      this.formData = {
        code: stock.code,
        name: stock.name,
        shares: stock.shares,
        avg_cost: stock.avg_cost,
        total_cost: 0
      }
      this.showModal = true
    },
    closeModal() {
      this.showModal = false
    },
    async saveHolding() {
      try {
        const { code, name, shares, avg_cost, total_cost } = this.formData
        await AddHolding(code, name, shares, Number(avg_cost) || 0)
        // 若填了真实投入总额，用其推导精确每股成本（一键修正）
        if (Number(total_cost) > 0) {
          await CorrectHoldingCost(code, Number(total_cost))
        }
        this.closeModal()
        await this.refreshData()
      } catch (error) {
        alert('保存失败: ' + error.message)
      }
    },
    async removeHolding(code) {
      if (confirm('确定删除该持仓吗？')) {
        try {
          await RemoveHolding(code)
          await this.refreshData()
        } catch (error) {
          alert('删除失败: ' + error.message)
        }
      }
    },
    startAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
      }
      this.refreshTimer = setInterval(() => {
        // 智能刷新：页面隐藏或非交易时段暂停
        this.refreshState = refreshReason()
        if (!isPageVisible() || !isMarketOpen()) {
          console.log('自动刷新持仓跳过:', this.refreshState)
          return
        }
        console.log('自动刷新持仓触发')
        this.refreshData()
      }, 30000)
    },
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    }
  },
  mounted() {
    console.log('Portfolio组件挂载')
    this.refreshData()
    this.startAutoRefresh()
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.last-update {
  font-size: 12px;
  color: var(--text-muted);
}

.refresh-state {
  font-size: 12px;
  color: var(--text-muted);
  padding: 2px 8px;
  border-radius: 10px;
  background: #f1f5f9;
}

.text-green { color: var(--success-color); }
.text-red { color: var(--danger-color); }
.text-orange { color: #e8590c; }
.text-muted { color: var(--text-muted); }

.advice-cell {
  max-width: 240px;
}

.advice-detail {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 4px;
  font-size: 12px;
}

.advice-reason {
  color: var(--text-muted);
  line-height: 1.4;
}

.cost-tip {
  color: #e8590c;
}

.form-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-muted);
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 24px;
  border-radius: 8px;
  width: 420px;
  max-width: 90%;
}

.modal-content h3 {
  margin-bottom: 20px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}
</style>
