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
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="stock in group" :key="stock.code">
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
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveHolding">{{ isEditing ? '保存' : '添加' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetPortfolio, AddHolding, RemoveHolding } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'

export default {
  name: 'Portfolio',
  data() {
    return {
      loading: true,
      holdings: [],
      showModal: false,
      isEditing: false,
      formData: {
        code: '',
        name: '',
        shares: 0,
        avg_cost: 0
      },
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    totalMarketValue() {
      return this.holdings.reduce((sum, s) => sum + (s.market_value || 0), 0)
    }
  },
  methods: {
    formatNumber(num) {
      return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
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
        const portfolioStr = await GetPortfolio()
        console.log('获取到持仓:', portfolioStr)
        this.holdings = parseJSON(portfolioStr)
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取持仓数据失败:', error)
        alert('刷新失败: ' + error.message)
      }
      this.loading = false
    },
    openAddModal() {
      this.isEditing = false
      this.formData = { code: '', name: '', shares: 0, avg_cost: 0 }
      this.showModal = true
    },
    editHolding(stock) {
      this.isEditing = true
      this.formData = {
        code: stock.code,
        name: stock.name,
        shares: stock.shares,
        avg_cost: stock.avg_cost
      }
      this.showModal = true
    },
    closeModal() {
      this.showModal = false
    },
    async saveHolding() {
      try {
        if (this.isEditing) {
          await AddHolding(this.formData.code, this.formData.name, this.formData.shares, this.formData.avg_cost)
        } else {
          await AddHolding(this.formData.code, this.formData.name, this.formData.shares, this.formData.avg_cost)
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
  width: 400px;
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
