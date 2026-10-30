<script setup lang="ts">
import { computed, ref } from 'vue';
import { fetchSettings, resetSettings, saveSettings, testLdapBind } from '@/service/api/settings';
import { $t } from '@/locales';

const emptySettings: Api.Settings.RuntimeSettings = {
  rustdeskWeb: { server: '', relay: '', key: '' },
  webClient: { url: '/web/' },
  timezone: 'Europe/Rome',
  ldap: {
    enabled: false,
    host: '',
    port: 389,
    use_tls: false,
    start_tls: true,
    bind_dn: '',
    bind_password: '',
    base_dn: '',
    user_filter: '(sAMAccountName=%s)',
    insecure_skip_verify: false,
    active_directory: false
  }
};

const loading = ref(true);
const saving = ref(false);
const settings = ref<Api.Settings.RuntimeSettings>(emptySettings);
const yamlDefaults = ref<Api.Settings.RuntimeSettings>(emptySettings);
const timezones = ref<string[]>([]);
const webClientAvailable = ref(false);
const serverTime = ref('');
const savedAt = ref('');
const yamlFile = ref('');
const testUsername = ref('');
const testResult = ref<Api.Settings.LdapTestResult | null>(null);

const timezoneOptions = computed(() => {
  const list = [...timezones.value];

  if (settings.value.timezone && !list.includes(settings.value.timezone)) {
    list.push(settings.value.timezone);
  }

  return list.map(timezone => ({ label: timezone, value: timezone }));
});

async function loadSettings() {
  loading.value = true;

  const res = await fetchSettings();

  if (!res.error) {
    settings.value = res.data.settings;
    yamlDefaults.value = res.data.defaults;
    timezones.value = res.data.timezones;
    webClientAvailable.value = res.data.webClient.available;
    serverTime.value = res.data.serverTime;
    savedAt.value = res.data.savedAt;
    yamlFile.value = res.data.yamlFile;
  }

  loading.value = false;
}

async function checkWebClient() {
  try {
    const response = await fetch('/web/status');
    const data = await response.json();
    webClientAvailable.value = Boolean(data.available);
  } catch {
    webClientAvailable.value = false;
  }
}

async function handleSave() {
  saving.value = true;

  const res = await saveSettings(settings.value);

  if (res.error) {
    window.$message?.error(res.message || $t('page.settings.saveFailed'));
  } else {
    window.$message?.success($t('page.settings.saved'));
    await loadSettings();
  }

  saving.value = false;
}

async function handleReset() {
  const res = await resetSettings();

  if (res.error) {
    window.$message?.error(res.message || $t('page.settings.resetFailed'));
    return;
  }

  settings.value = res.data;
  testResult.value = null;
  window.$message?.success($t('page.settings.resetSuccess'));
  await loadSettings();
}

async function handleTestLdap() {
  testResult.value = null;

  const res = await testLdapBind({
    ...settings.value.ldap,
    test_username: testUsername.value
  });

  if (res.error) {
    window.$message?.error(res.message || $t('page.settings.testFailed'));
    return;
  }

  testResult.value = res.data;

  if (res.data.error) {
    window.$message?.error($t('page.settings.testFailed'));
  } else {
    window.$message?.success($t('page.settings.testSuccess'));
  }
}

loadSettings();
checkWebClient();
</script>

<template>
  <NSpace vertical :size="16">
    <NCard :title="$t('page.settings.title')" :bordered="false">
      <NSpace vertical :size="12">
        <div class="text-13px text-#646464">{{ $t('page.settings.intro') }}</div>
        <NSpace align="center" :size="8">
          <NTag type="info" size="small">{{ $t('page.settings.yamlFile') }}: {{ yamlFile }}</NTag>
          <NTag type="success" size="small">{{ $t('page.settings.serverTime') }}: {{ serverTime }}</NTag>
          <NTag v-if="savedAt" type="warning" size="small">{{ $t('page.settings.savedAt') }}: {{ savedAt }}</NTag>
        </NSpace>
        <NSpace align="center" :size="8">
          <NButton type="primary" :loading="saving" @click="handleSave">
            {{ $t('page.settings.save') }}
          </NButton>
          <NButton @click="handleReset">{{ $t('page.settings.reset') }}</NButton>
        </NSpace>
      </NSpace>
    </NCard>

    <NCard :title="$t('page.settings.rustdeskWeb')" :bordered="false">
      <NSpace vertical :size="14">
        <NAlert :type="webClientAvailable ? 'success' : 'warning'" :title="$t('page.settings.webClientStatus')">
          <NSpace align="center" :size="8">
            <span>
              {{ webClientAvailable ? $t('page.settings.webClientReady') : $t('page.settings.webClientMissing') }}
            </span>
            <a
              v-if="webClientAvailable"
              href="/web/"
              target="_blank"
              rel="noopener noreferrer"
              class="text-13px text-primary underline"
            >
              {{ $t('page.settings.openClient') }}
            </a>
          </NSpace>
        </NAlert>

        <NGrid :x-gap="16" :y-gap="14" responsive="screen" item-responsive>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.server') }}</div>
            <NInput v-model:value="settings.rustdeskWeb.server" :placeholder="yamlDefaults.rustdeskWeb.server" />
          </NGi>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.relay') }}</div>
            <NInput v-model:value="settings.rustdeskWeb.relay" :placeholder="yamlDefaults.rustdeskWeb.relay" />
          </NGi>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.key') }}</div>
            <NInput v-model:value="settings.rustdeskWeb.key" :placeholder="yamlDefaults.rustdeskWeb.key" />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.webClientUrl') }}</div>
            <NInput v-model:value="settings.webClient.url" :placeholder="yamlDefaults.webClient.url" />
          </NGi>
        </NGrid>

        <div class="text-12px text-#646464">{{ $t('page.settings.rustdeskWebHint') }}</div>
      </NSpace>
    </NCard>

    <NCard :title="$t('page.settings.timezone')" :bordered="false">
      <NSpace vertical :size="14">
        <NGrid :x-gap="16" :y-gap="14" responsive="screen" item-responsive>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.timezone') }}</div>
            <NSelect v-model:value="settings.timezone" :options="timezoneOptions" />
          </NGi>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.timezoneCustom') }}</div>
            <NInput v-model:value="settings.timezone" :placeholder="yamlDefaults.timezone" />
          </NGi>
        </NGrid>
        <div class="text-12px text-#646464">{{ $t('page.settings.timezoneHint') }}</div>
      </NSpace>
    </NCard>

    <NCard :title="$t('page.settings.ldap')" :bordered="false">
      <NSpace vertical :size="14">
        <NSpace align="center" :size="18">
          <NSpace align="center" :size="8">
            <NSwitch v-model:value="settings.ldap.enabled" />
            <span class="text-13px">{{ $t('page.settings.ldapEnabled') }}</span>
          </NSpace>
          <NSpace align="center" :size="8">
            <NSwitch v-model:value="settings.ldap.active_directory" />
            <span class="text-13px">{{ $t('page.settings.activeDirectory') }}</span>
          </NSpace>
        </NSpace>

        <NGrid :x-gap="16" :y-gap="14" responsive="screen" item-responsive>
          <NGi span="24 s:24 m:8">
            <div class="text-13px mb-6px">{{ $t('page.settings.host') }}</div>
            <NInput v-model:value="settings.ldap.host" :placeholder="yamlDefaults.ldap.host" />
          </NGi>
          <NGi span="24 s:24 m:4">
            <div class="text-13px mb-6px">{{ $t('page.settings.port') }}</div>
            <NInputNumber v-model:value="settings.ldap.port" placeholder="389" />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.bindDn') }}</div>
            <NInput v-model:value="settings.ldap.bind_dn" :placeholder="yamlDefaults.ldap.bind_dn" />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.bindPassword') }}</div>
            <NInput
              v-model:value="settings.ldap.bind_password"
              type="password"
              :placeholder="yamlDefaults.ldap.bind_password"
            />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.baseDn') }}</div>
            <NInput v-model:value="settings.ldap.base_dn" :placeholder="yamlDefaults.ldap.base_dn" />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.userFilter') }}</div>
            <NInput v-model:value="settings.ldap.user_filter" placeholder="(sAMAccountName=%s)" />
          </NGi>
        </NGrid>

        <NSpace align="center" :size="18">
          <NSpace align="center" :size="8">
            <NSwitch v-model:value="settings.ldap.use_tls" />
            <span class="text-13px">{{ $t('page.settings.useTls') }}</span>
          </NSpace>
          <NSpace align="center" :size="8">
            <NSwitch v-model:value="settings.ldap.start_tls" />
            <span class="text-13px">{{ $t('page.settings.startTls') }}</span>
          </NSpace>
          <NSpace align="center" :size="8">
            <NSwitch v-model:value="settings.ldap.insecure_skip_verify" />
            <span class="text-13px">{{ $t('page.settings.insecureSkipVerify') }}</span>
          </NSpace>
        </NSpace>

        <NGrid :x-gap="16" :y-gap="14" responsive="screen" item-responsive>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">{{ $t('page.settings.testUsername') }}</div>
            <NInput v-model:value="testUsername" :placeholder="$t('page.settings.testUsernamePlaceholder')" />
          </NGi>
          <NGi span="24 s:24 m:12">
            <div class="text-13px mb-6px">&nbsp;</div>
            <NButton type="primary" @click="handleTestLdap">{{ $t('page.settings.testBind') }}</NButton>
          </NGi>
        </NGrid>

        <div class="text-12px text-#646464">{{ $t('page.settings.ldapHint') }}</div>

        <NAlert
          v-if="testResult"
          :type="testResult.error ? 'error' : 'success'"
          :title="testResult.error ? $t('page.settings.testFailed') : $t('page.settings.testSuccess')"
        >
          <NSpace vertical :size="4">
            <NText v-for="step in testResult.steps" :key="step" class="text-12px">{{ step }}</NText>
            <NText v-if="testResult.activeDirectoryDetected" class="text-12px">
              {{ $t('page.settings.adDetected') }}
            </NText>
            <NText v-if="testResult.firstDN" class="text-12px">{{ testResult.firstDN }}</NText>
          </NSpace>
        </NAlert>
      </NSpace>
    </NCard>
  </NSpace>
</template>

<style scoped></style>
