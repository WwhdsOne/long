<script setup>
import AdminPage from './pages/AdminPage.vue'
import PublicPage from './pages/PublicPage.vue'

const currentPath = window.location.pathname
const currentView = currentPath.startsWith('/admin') ? AdminPage : PublicPage
</script>

<template>
  <component :is="currentView" />
</template>
