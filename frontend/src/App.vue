<script setup>
import AdminPage from './pages/AdminPage.vue'
import BattleFxGalleryPage from './pages/BattleFxGalleryPage.vue'
import PublicPage from './pages/PublicPage.vue'
import TalentBuffDemoPage from './pages/TalentBuffDemoPage.vue'

const currentPath = window.location.pathname
const currentView = currentPath.startsWith('/admin')
    ? AdminPage
    : currentPath.startsWith('/internal/battle-fx-gallery')
        ? BattleFxGalleryPage
        : currentPath.startsWith('/__talent-buff-demo')
            ? TalentBuffDemoPage
            : PublicPage
</script>

<template>
  <component :is="currentView"/>
</template>
