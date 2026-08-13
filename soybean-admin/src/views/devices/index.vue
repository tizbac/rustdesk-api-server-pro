<script setup lang="tsx">
import { NTag } from 'naive-ui';
import { fetchDevicesList } from '@/service/api/devices';
import { $t } from '@/locales';
import { useAppStore } from '@/store/modules/app';
import { useTable } from '@/hooks/common/table';
import { AuditTypeOptions } from '@/constants/business';
import TableHeader from './components/table-header.vue';
import AuditBaseLogsSearch from './components/search.vue';
import { fetchConfig } from '@/service/api/config';

import { onMounted, ref, h } from 'vue';

const appStore = useAppStore();

const auditTypes: any = {
  '0': '',
  '1': 'error',
  '2': 'success'
};

// Fetch web client URL from config
const webClientUrl = ref<string>('');

const fetchWebClientUrl = async () => {
  try {
    const res = await fetchConfig();
    webClientUrl.value = res.data.webClientUrl;
  } catch (error) {
    console.error('Failed to fetch web client URL:', error);
  }
};

// Fetch config on component mount
onMounted(() => {
  fetchWebClientUrl();
});

const {
  columns,
  columnChecks,
  data,
  getData,
  getDataByPage,
  loading,
  mobilePagination,
  searchParams,
  resetSearchParams
} = useTable({
  apiFn: fetchDevicesList,
  showTotal: true,
  apiParams: {
    current: 1,
    size: 10,
    // if you want to use the searchParams in Form, you need to define the following properties, and the value is null
    // the value can not be undefined, otherwise the property in Form will not be reactive
    hostname: null,
    username: null,
    rustdesk_id: null
  },
  columns: () => [
    {
      key: 'id',
      title: 'ID',
      align: 'center'
    },
    {
      key: 'rustdesk_id',
      title: $t('dataMap.device.rustdesk_id'),
      align: 'center',
      render: (row) => {
        if (!webClientUrl.value) return row.rustdesk_id;
        return h(
          'a',
          {
            href: `${webClientUrl.value}/#/?id=${row.rustdesk_id}`,
            target: '_blank',
            style: {
              color: 'blue',
              textDecoration: 'underline'
            }
          },
          row.rustdesk_id
        );
      }
    },
    {
      key: 'hostname',
      title: $t('dataMap.device.hostname'),
      align: 'center'
    },
    {
      key: 'username',
      title: $t('dataMap.device.username'),
      align: 'center'
    },
    {
      key: 'version',
      title: $t('dataMap.device.version'),
      align: 'center'
    },
    {
      key: 'os',
      title: $t('dataMap.device.os'),
      align: 'center'
    },
    {
      key: 'uumemoryid',
      title: $t('dataMap.device.memory'),
      align: 'center'
    },
    {
      key: 'created_at',
      title: $t('dataMap.audit.created_at'),
      align: 'center'
    }
  ]
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <AuditBaseLogsSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />

    <NCard :title="$t('route.devices')" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <TableHeader v-model:columns="columnChecks" :loading="loading" @refresh="getData" />
      </template>
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="962"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
