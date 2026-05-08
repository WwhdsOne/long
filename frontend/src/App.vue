<script setup>
import AdminPage from './pages/AdminPage.vue'
import BattleFxGalleryPage from './pages/BattleFxGalleryPage.vue'
import TalentBuffDemoPage from './pages/TalentBuffDemoPage.vue'
import PublicPage from './pages/PublicPage.vue'

const pageMap = {
    admin: AdminPage,
    battleFxGallery: BattleFxGalleryPage,
    talentBuffDemo: TalentBuffDemoPage,
    public: PublicPage,
}

const currentPath = window.location.pathname
const pageKey = currentPath.startsWith('/admin')
    ? 'admin'
    : currentPath.startsWith('/internal/battle-fx-gallery')
        ? 'battleFxGallery'
        : currentPath.startsWith('/__talent-buff-demo')
            ? 'talentBuffDemo'
            : 'public'

const currentView = pageMap[pageKey]
</script>

<template>
  <component :is="currentView"/>
</template>
