<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { $t } from '@/locales';
import { useNaiveForm } from '@/hooks/common/form';
import { useAuthStore } from '@/store/modules/auth';
import { fetchCaptcha, fetchAuthLoginConfig } from '@/service/api/auth';

defineOptions({
  name: 'PwdLogin'
});

const authStore = useAuthStore();
const { formRef, validate } = useNaiveForm();

const model: Api.Form.LoginForm = reactive({
  username: '',
  password: '',
  code: '',
  captchaId: ''
});

const captcha: Api.Auth.Captcha = reactive({
  id: '',
  img: ''
});

const ldapEnabled = ref(false);

const isLdapLogin = computed(() => ldapEnabled.value);

const rules = computed<Record<keyof Api.Form.LoginForm, App.Global.FormRule[]>>(() => {
  const baseRules: Record<keyof Api.Form.LoginForm, App.Global.FormRule[]> = {
    username: [
      {
        required: true,
        message: '账号不能为空'
      }
    ],
    password: [
      {
        required: true,
        message: '密码不能为空'
      }
    ]
  };

  if (!isLdapLogin.value) {
    baseRules.code = [
      {
        required: true,
        message: '验证码不能为空'
      }
    ];
    baseRules.captchaId = [
      {
        required: true,
        message: '验证码不能为空'
      }
    ];
  }

  return baseRules;
});

async function handleSubmit() {
  await validate();
  const err = await authStore.login(model);
  if (!isLdapLogin.value && err?.response?.data.message === 'CaptchaError') {
    handleCaptcha();
  }
}

async function handleCaptcha() {
  const c = await fetchCaptcha();
  captcha.id = c.data?.id || '';
  captcha.img = c.data?.img || '';
  model.captchaId = captcha.id || '';
}

async function fetchLoginConfig() {
  try {
    const res = await fetchAuthLoginConfig();
    ldapEnabled.value = res.data?.ldapEnabled || false;
  } catch {
    ldapEnabled.value = false;
  }
}

onMounted(() => {
  fetchLoginConfig();
  if (!isLdapLogin.value) {
    handleCaptcha();
  }
});
</script>

<template>
  <NForm ref="formRef" :model="model" :rules="rules" size="large" :show-label="false">
    <NFormItem path="userName">
      <NInput v-model:value="model.username" :placeholder="$t('page.login.common.userNamePlaceholder')" />
    </NFormItem>
    <NFormItem path="password">
      <NInput
        v-model:value="model.password"
        type="password"
        show-password-on="click"
        :placeholder="$t('page.login.common.passwordPlaceholder')"
      />
    </NFormItem>
    <NFormItem v-if="!isLdapLogin" path="code">
      <NInput v-model:value="model.code" :clearable="true" :placeholder="$t('page.login.common.codePlaceholder')" />
      <div class="pl-8px">
        <img width="152" height="40" class="cursor-pointer" :src="captcha.img" @click="handleCaptcha" />
      </div>
    </NFormItem>
    <NSpace vertical :size="24">
      <div class="flex-y-center justify-between">
        <NCheckbox>{{ $t('page.login.pwdLogin.rememberMe') }}</NCheckbox>
      </div>
      <NButton
        attr-type="submit"
        type="primary"
        size="large"
        round
        block
        :loading="authStore.loginLoading"
        @click="handleSubmit"
      >
        {{ $t('common.confirm') }}
      </NButton>
    </NSpace>
  </NForm>
</template>

<style scoped></style>
