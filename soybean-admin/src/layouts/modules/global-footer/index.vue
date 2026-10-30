<script setup lang="ts">
defineOptions({
  name: 'GlobalFooter'
});
</script>

<template>
  <DarkModeContainer class="h-full flex-center gap-8px">
    <a href="https://www.etitech.net" target="_blank" rel="noopener noreferrer">
      <img src="@/assets/imgs/etitech-logo.png" class="h-22px" alt="Etitech" />
    </a>
    <span class="text-12px">Copyright © 2026 E.T.I. s.r.l.</span>
    <span class="text-12px">·</span>
    <a
      href="https://github.com/lantongxue/rustdesk-api-server-pro"
      target="_blank"
      rel="noopener noreferrer"
    >
      based on rustdesk-api-server-pro (MIT)
    </a>
    <span class="text-12px">·</span>
    <a href="https://oxygenwireless.net" target="_blank" rel="noopener noreferrer">
      <img src="@/assets/imgs/oxygen-logo.png" class="h-24px" alt="Oxygen Wireless" />
    </a>
  </DarkModeContainer>
</template>

<style scoped></style>
