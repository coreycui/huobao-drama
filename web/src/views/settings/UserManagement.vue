<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <PageHeader
        :title="$t('userManagement.title')"
        :subtitle="$t('userManagement.subtitle')"
        :show-back="true"
        :back-text="$t('common.back')"
      >
        <template #actions>
          <el-button @click="loadUsers">
            <el-icon><Refresh /></el-icon>
            <span>{{ $t('common.refresh') }}</span>
          </el-button>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span>{{ $t('userManagement.createUser') }}</span>
          </el-button>
        </template>
      </PageHeader>

      <div v-loading="loading" class="user-table-wrapper">
        <el-table :data="users" border stripe>
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="username" :label="$t('auth.username')" min-width="220" />
          <el-table-column prop="role" :label="$t('auth.role')" width="140">
            <template #default="{ row }">
              <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">
                {{ row.role === 'admin' ? $t('auth.admin') : $t('auth.user') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.actions')" width="220" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="openEditDialog(row)">{{ $t('common.edit') }}</el-button>
              <el-button
                size="small"
                type="danger"
                :disabled="row.id === authStore.user?.id"
                @click="handleDelete(row)"
              >
                {{ $t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-dialog
        v-model="dialogVisible"
        :title="isEdit ? $t('userManagement.editUser') : $t('userManagement.createUser')"
        width="480px"
        :close-on-click-modal="false"
      >
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
          <el-form-item :label="$t('auth.username')" prop="username">
            <el-input v-model="form.username" :placeholder="$t('auth.usernamePlaceholder')" />
          </el-form-item>

          <el-form-item :label="$t('auth.role')" prop="role">
            <el-select v-model="form.role" style="width: 100%">
              <el-option :label="$t('auth.admin')" value="admin" />
              <el-option :label="$t('auth.user')" value="user" />
            </el-select>
          </el-form-item>

          <el-form-item
            :label="isEdit ? $t('userManagement.resetPassword') : $t('auth.password')"
            prop="password"
          >
            <el-input
              v-model="form.password"
              type="password"
              show-password
              :placeholder="isEdit ? $t('userManagement.passwordOptional') : $t('auth.passwordPlaceholder')"
            />
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? $t('common.save') : $t('common.create') }}
          </el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { PageHeader } from '@/components/common'
import { userAPI, type UserItem, type UserRole } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const users = ref<UserItem[]>([])
const formRef = ref<FormInstance>()

const createDefaultForm = () => ({
  username: '',
  password: '',
  role: 'user' as UserRole
})

const form = reactive(createDefaultForm())

const rules: FormRules = {
  username: [{ required: true, message: t('auth.usernameRequired'), trigger: 'blur' }],
  role: [{ required: true, message: t('auth.roleRequired'), trigger: 'change' }],
  password: [
    {
      validator: (_rule, value, callback) => {
        if (!isEdit.value && !value) {
          callback(new Error(t('auth.passwordRequired')))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const resetForm = () => {
  Object.assign(form, createDefaultForm())
  formRef.value?.clearValidate()
}

const loadUsers = async () => {
  loading.value = true
  try {
    users.value = await userAPI.list()
  } catch (error: any) {
    ElMessage.error(error?.message || t('userManagement.loadFailed'))
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (user: UserItem) => {
  isEdit.value = true
  editingId.value = user.id
  Object.assign(form, {
    username: user.username,
    password: '',
    role: user.role === 'admin' ? 'admin' : 'user'
  })
  formRef.value?.clearValidate()
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) {
    return
  }

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    return
  }

  submitting.value = true
  try {
    if (isEdit.value && editingId.value !== null) {
      const payload = {
        username: form.username,
        role: form.role,
        ...(form.password ? { password: form.password } : {})
      }
      await userAPI.update(editingId.value, payload)
      ElMessage.success(t('userManagement.updateSuccess'))
    } else {
      await userAPI.create({
        username: form.username,
        password: form.password,
        role: form.role
      })
      ElMessage.success(t('userManagement.createSuccess'))
    }

    dialogVisible.value = false
    await loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.message || t(isEdit.value ? 'userManagement.updateFailed' : 'userManagement.createFailed'))
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (user: UserItem) => {
  try {
    await ElMessageBox.confirm(
      t('userManagement.deleteConfirm', { username: user.username }),
      t('common.confirmDelete'),
      {
        type: 'warning',
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel')
      }
    )
    await userAPI.delete(user.id)
    ElMessage.success(t('userManagement.deleteSuccess'))
    await loadUsers()
  } catch (error: any) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    ElMessage.error(error?.message || t('userManagement.deleteFailed'))
  }
}

onMounted(loadUsers)
</script>

<style scoped>
.user-table-wrapper {
  margin-top: 24px;
}
</style>
