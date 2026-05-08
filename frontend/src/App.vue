<script setup>
import {defineAsyncComponent} from 'vue'

const pageMap = {
    admin: () => import('./pages/AdminPage.vue'),
    battleFxGallery: () => import('./pages/BattleFxGalleryPage.vue'),
    talentBuffDemo: () => import('./pages/TalentBuffDemoPage.vue'),
    public: () => import('./pages/PublicPage.vue'),
}

const currentPath = window.location.pathname
const pageKey = currentPath.startsWith('/admin')
    ? 'admin'
    : currentPath.startsWith('/internal/battle-fx-gallery')
        ? 'battleFxGallery'
        : currentPath.startsWith('/__talent-buff-demo')
            ? 'talentBuffDemo'
            : 'public'

const currentView = defineAsyncComponent(pageMap[pageKey])
</script>

<template>
  <component :is="currentView"/>
</template>
