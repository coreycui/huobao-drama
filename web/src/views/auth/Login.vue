<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-brand">🎬 HuoBao Drama</div>
      <h1 class="login-title">{{ $t('auth.loginTitle') }}</h1>
      <p class="login-subtitle">{{ $t('auth.loginSubtitle') }}</p>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        label-position="top"
        @keyup.enter="handleLogin"
      >
        <el-form-item :label="$t('auth.username')" prop="username">
          <el-input
            v-model="form.username"
            :placeholder="$t('auth.usernamePlaceholder')"
            size="large"
          />
        </el-form-item>

        <el-form-item :label="$t('auth.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="$t('auth.passwordPlaceholder')"
            size="large"
          />
        </el-form-item>

        <el-button
          type="primary"
          class="login-submit"
          size="large"
          :loading="submitting"
          @click="handleLogin"
        >
          {{ $t('auth.login') }}
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [{ required: true, message: t('auth.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('auth.passwordRequired'), trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) {
    return
  }

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    return
  }

  submitting.value = true
  try {
    await authStore.login({
      username: form.username,
      password: form.password
    })
    ElMessage.success(t('auth.loginSuccess'))
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.replace(redirect || '/')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || t('auth.loginFailed'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.16), transparent 38%),
    radial-gradient(circle at bottom right, rgba(6, 182, 212, 0.18), transparent 42%),
    linear-gradient(180deg, var(--bg-primary) 0%, var(--bg-secondary) 100%);
}

.login-card {
  width: min(100%, 420px);
  padding: 36px 32px;
  border: 1px solid var(--border-primary);
  border-radius: 24px;
  background: var(--bg-card);
  box-shadow: 0 18px 60px rgba(15, 23, 42, 0.12);
}

.login-brand {
  margin-bottom: 12px;
  font-size: 1rem;
  font-weight: 700;
  color: var(--accent);
}

.login-title {
  margin: 0;
  font-size: 1.75rem;
  color: var(--text-primary);
}

.login-subtitle {
  margin: 8px 0 24px;
  color: var(--text-secondary);
}

.login-form {
  display: flex;
  flex-direction: column;
}

.login-submit {
  width: 100%;
  margin-top: 8px;
}
</style>
