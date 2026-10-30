<script setup lang="ts">
import { $t } from '@/locales';
import sponsorships from '@/assets/imgs/sponsorships.png';

defineOptions({
  name: 'CreativityBanner'
});
</script>

<template>
  <NCard :title="$t('page.home.friendlySponsorship')" :bordered="false" size="small" class="h-full card-wrapper">
    <template #footer>
      <div>{{ $t('page.home.thankYou') }}</div>
    </template>
    <div class="h-full flex-col-center gap-12px">
      <NSpace x-gap="16px" align="center">
        <span class="text-12px text-#646464">{{ $t('system.sponsors') }}</span>
        <a href="https://www.etitech.net" target="_blank" rel="noopener noreferrer">
          <img src="@/assets/imgs/etitech-logo.png" class="h-44px" alt="Etitech" />
        </a>
        <a href="https://oxygenwireless.net" target="_blank" rel="noopener noreferrer">
          <img src="@/assets/imgs/oxygen-logo.png" class="h-52px" alt="Oxygen Wireless" />
        </a>
      </NSpace>
      <div class="text-12px text-#646464">
        {{ $t('page.home.cupOfCoffee') }}
        (<a href="https://github.com/lantongxue" target="_blank" rel="noopener noreferrer">lantongxue</a>)
      </div>
      <NSpace>
        <NImage :src="sponsorships"></NImage>
      </NSpace>
    </div>
  </NCard>
</template>

<style scoped></style>
