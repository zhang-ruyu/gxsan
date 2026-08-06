<template>
  <div class="app-container">
    <TitleBar />
    <Sidebar />
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script>
import Sidebar from './components/Sidebar.vue'
import TitleBar from './components/TitleBar.vue'

export default {
  name: 'App',
  components: {
    Sidebar,
    TitleBar
  }
}
</script>
