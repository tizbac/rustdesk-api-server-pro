<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { GLOBAL_SIDER_MENU_ID } from '@/constants/app';
import { router } from '@/router';
import { $t } from '@/locales';
import SvgIcon from '@/components/custom/svg-icon.vue';
import GlobalLogo from '../global-logo/index.vue';

defineOptions({
  name: 'GlobalSider'
});

const appStore = useAppStore();
const themeStore = useThemeStore();

const isVerticalMix = computed(() => themeStore.layout.mode === 'vertical-mix');
const isHorizontalMix = computed(() => themeStore.layout.mode === 'horizontal-mix');
const darkMenu = computed(() => !themeStore.darkMode && !isHorizontalMix.value && themeStore.sider.inverted);
const showLogo = computed(() => !isVerticalMix.value && !isHorizontalMix.value);
const menuWrapperClass = computed(() => (showLogo.value ? 'flex-1-hidden' : 'h-full'));
</script>

<template>
  <DarkModeContainer class="size-full flex-col-stretch shadow-sider" :inverted="darkMenu">
    <GlobalLogo
      v-if="showLogo"
      :show-title="!appStore.siderCollapse"
      :style="{ height: themeStore.header.height + 'px' }"
    />
    <div :id="GLOBAL_SIDER_MENU_ID" :class="menuWrapperClass"></div>
    <div class="flex-end-center px-12px pb-10px">
      <NButton text size="small" :title="$t('route.settings')" @click="router.push('/settings')">
        <SvgIcon icon="ph:settings" class="text-18px" />
      </NButton>
    </div>
  </DarkModeContainer>
</template>

<style scoped></style>
