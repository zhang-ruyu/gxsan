<template>
  <div>
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else>
      <!-- 基础设置 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">基础设置</span>
        </div>
        <div class="form-group">
          <label class="form-label">默认目标股息率 (%)</label>
          <input class="form-input" type="number" step="0.1" v-model.number="config.default_target_yield" @change="saveConfig('default_target_yield', config.default_target_yield)">
        </div>
        <div class="form-group">
          <label class="form-label">最少分红年数</label>
          <input class="form-input" type="number" v-model.number="config.min_dividend_years" @change="saveConfig('min_dividend_years', config.min_dividend_years)">
        </div>
        <div class="form-group">
          <label class="form-label">便宜价折扣 (%)</label>
          <input class="form-input" type="number" step="0.01" v-model.number="config.cheap_discount" @change="saveConfig('cheap_discount', config.cheap_discount)">
        </div>
        <div class="form-group">
          <label class="form-label">昂贵价溢价 (%)</label>
          <input class="form-input" type="number" step="0.01" v-model.number="config.expensive_premium" @change="saveConfig('expensive_premium', config.expensive_premium)">
        </div>
      </div>

      <!-- 资金设置 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">资金设置</span>
        </div>
        <div class="form-group">
          <label class="form-label">可用资金 (¥)</label>
          <input class="form-input" type="number" v-model.number="fundInfo.available_fund" @change="saveFund">
        </div>
        <div class="form-group">
          <label class="form-label">单只股票最大仓位 (%)</label>
          <input class="form-input" type="number" step="0.1" v-model.number="fundInfo.max_position_pct" @change="saveFund">
        </div>
      </div>

      <!-- 监控列表 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">监控列表</span>
          <button class="btn btn-primary btn-sm" @click="openAddModal">添加股票</button>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>代码</th>
              <th>名称</th>
              <th>目标股息率</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in watchlist" :key="item.code">
              <td>{{ item.code }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.target_yield }}%</td>
              <td>
                <button class="btn btn-sm btn-secondary" @click="openEditModal(item)">编辑</button>
                <button class="btn btn-sm btn-danger" @click="removeWatchlist(item.code)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="watchlist.length === 0" class="empty-state">
          <p>暂无监控股票</p>
        </div>
      </div>
    </div>

    <!-- 添加/编辑股票弹窗 -->
    <div class="modal" v-if="showModal">
      <div class="modal-content">
        <h3>{{ isEditing ? '编辑监控股票' : '添加监控股票' }}</h3>
        <div class="form-group">
          <label class="form-label">股票代码</label>
          <input class="form-input" v-model="formData.code" :disabled="isEditing" placeholder="例如: 601398">
        </div>
        <div class="form-group">
          <label class="form-label">股票名称</label>
          <input class="form-input" v-model="formData.name" placeholder="例如: 工商银行">
        </div>
        <div class="form-group">
          <label class="form-label">目标股息率 (%)</label>
          <input class="form-input" type="number" step="0.1" v-model.number="formData.target_yield" placeholder="5.0">
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveWatchlist">{{ isEditing ? '保存' : '添加' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetConfig, SetConfig, GetFundInfo, SetAvailableFund, GetWatchlist, AddWatchlist, RemoveWatchlist, UpdateWatchlist } from '../../wailsjs/go/main/App'

export default {
  name: 'Settings',
  data() {
    return {
      loading: true,
      config: {
        default_target_yield: 4,
        min_dividend_years: 3,
        cheap_discount: 0.8,
        expensive_premium: 1.2
      },
      fundInfo: {
        available_fund: 50000,
        max_position_pct: 30
      },
      watchlist: [],
      showModal: false,
      isEditing: false,
      editingCode: '',
      formData: {
        code: '',
        name: '',
        target_yield: 5
      }
    }
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const configStr = await GetConfig()
        const configData = JSON.parse(configStr)
        this.config = {
          default_target_yield: configData.default_target_yield || 4,
          min_dividend_years: configData.min_dividend_years || 3,
          cheap_discount: configData.cheap_discount || 0.8,
          expensive_premium: configData.expensive_premium || 1.2
        }

        const fundStr = await GetFundInfo()
        this.fundInfo = JSON.parse(fundStr)

        const watchlistStr = await GetWatchlist()
        this.watchlist = JSON.parse(watchlistStr)
      } catch (error) {
        console.error('加载配置失败:', error)
      }
      this.loading = false
    },
    async saveConfig(key, value) {
      try {
        await SetConfig(key, String(value))
      } catch (error) {
        alert('保存失败: ' + error.message)
      }
    },
    async saveFund() {
      try {
        await SetAvailableFund(this.fundInfo.available_fund)
        await SetConfig('max_position_pct', String(this.fundInfo.max_position_pct))
      } catch (error) {
        alert('保存失败: ' + error.message)
      }
    },
    openAddModal() {
      this.isEditing = false
      this.editingCode = ''
      this.formData = { code: '', name: '', target_yield: 5 }
      this.showModal = true
    },
    openEditModal(item) {
      this.isEditing = true
      this.editingCode = item.code
      this.formData = {
        code: item.code,
        name: item.name,
        target_yield: item.target_yield
      }
      this.showModal = true
    },
    closeModal() {
      this.showModal = false
      this.isEditing = false
      this.editingCode = ''
      this.formData = { code: '', name: '', target_yield: 5 }
    },
    async saveWatchlist() {
      try {
        if (this.isEditing) {
          await UpdateWatchlist(this.formData.code, this.formData.name, this.formData.target_yield)
        } else {
          await AddWatchlist(this.formData.code, this.formData.name, this.formData.target_yield)
        }
        this.closeModal()
        await this.loadData()
      } catch (error) {
        alert('保存失败: ' + error.message)
      }
    },
    async removeWatchlist(code) {
      if (confirm('确定删除该监控股票吗？')) {
        try {
          await RemoveWatchlist(code)
          await this.loadData()
        } catch (error) {
          alert('删除失败: ' + error.message)
        }
      }
    }
  },
  mounted() {
    this.loadData()
  }
}
</script>

<style scoped>
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

.empty-state {
  text-align: center;
  padding: 20px;
  color: #999;
}
</style>
