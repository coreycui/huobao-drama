<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <PageHeader
        :title="$t('styleManagement.title')"
        :show-back="true"
        :back-text="$t('common.back')"
      >
        <template #actions>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            <span>{{ $t("styleManagement.addStyle") }}</span>
          </el-button>
        </template>
      </PageHeader>

      <div v-loading="loading" class="style-table-wrapper">
        <el-table :data="styles" border stripe>
          <el-table-column prop="sort_order" :label="$t('styleManagement.sortOrder')" width="80" />
          <el-table-column prop="style_value" :label="$t('styleManagement.styleValue')" width="140" />
          <el-table-column prop="name_zh" :label="$t('styleManagement.nameZh')" width="140" />
          <el-table-column prop="name_en" :label="$t('styleManagement.nameEn')" />
          <el-table-column prop="is_active" :label="$t('common.status')" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'info'">
                {{ row.is_active ? $t("styleManagement.active") : $t("styleManagement.inactive") }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.edit')" width="160" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="handleEdit(row)">{{ $t("common.edit") }}</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{ $t("common.delete") }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Create/Edit Dialog -->
      <el-dialog
        v-model="dialogVisible"
        :title="isEdit ? $t('styleManagement.editStyle') : $t('styleManagement.addStyle')"
        width="480px"
        :close-on-click-modal="false"
      >
        <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
          <el-form-item :label="$t('styleManagement.styleValue')" prop="style_value">
            <el-input v-model="form.style_value" :placeholder="$t('styleManagement.styleValuePlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('styleManagement.nameZh')" prop="name_zh">
            <el-input v-model="form.name_zh" :placeholder="$t('styleManagement.nameZhPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('styleManagement.nameEn')" prop="name_en">
            <el-input v-model="form.name_en" :placeholder="$t('styleManagement.nameEnPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('styleManagement.sortOrder')">
            <el-input-number v-model="form.sort_order" :min="0" :max="9999" style="width: 100%" />
          </el-form-item>
          <el-form-item :label="$t('common.status')">
            <el-switch v-model="form.is_active" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">{{ $t("common.cancel") }}</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t("common.save") }}</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { styleAPI, type ImageStyle } from "@/api/style";
import PageHeader from "@/components/common/PageHeader.vue";

const loading = ref(false);
const submitting = ref(false);
const styles = ref<ImageStyle[]>([]);
const dialogVisible = ref(false);
const isEdit = ref(false);
const editId = ref<number | null>(null);
const formRef = ref<FormInstance>();

const defaultForm = () => ({
  style_value: "",
  name_zh: "",
  name_en: "",
  sort_order: 0,
  is_active: true,
});

const form = reactive(defaultForm());

const rules: FormRules = {
  style_value: [{ required: true, message: "请输入风格值", trigger: "blur" }],
  name_zh: [{ required: true, message: "请输入中文名称", trigger: "blur" }],
  name_en: [{ required: true, message: "请输入英文名称", trigger: "blur" }],
};

async function loadStyles() {
  loading.value = true;
  try {
    styles.value = await styleAPI.list(true);
  } catch {
    ElMessage.error("加载风格列表失败");
  } finally {
    loading.value = false;
  }
}

function showCreateDialog() {
  isEdit.value = false;
  editId.value = null;
  Object.assign(form, defaultForm());
  dialogVisible.value = true;
}

function handleEdit(row: ImageStyle) {
  isEdit.value = true;
  editId.value = row.id;
  Object.assign(form, {
    style_value: row.style_value,
    name_zh: row.name_zh,
    name_en: row.name_en,
    sort_order: row.sort_order,
    is_active: row.is_active,
  });
  dialogVisible.value = true;
}

async function handleDelete(row: ImageStyle) {
  try {
    await ElMessageBox.confirm(`确定要删除风格"${row.name_zh}"吗？`, "删除确认", {
      type: "warning",
      confirmButtonText: "确定",
      cancelButtonText: "取消",
    });
    await styleAPI.delete(row.id);
    ElMessage.success("删除成功");
    await loadStyles();
  } catch {
    // cancelled or error
  }
}

async function handleSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid) => {
    if (!valid) return;
    submitting.value = true;
    try {
      if (isEdit.value && editId.value !== null) {
        await styleAPI.update(editId.value, { ...form });
        ElMessage.success("更新成功");
      } else {
        await styleAPI.create({ ...form });
        ElMessage.success("创建成功");
      }
      dialogVisible.value = false;
      await loadStyles();
    } catch {
      ElMessage.error(isEdit.value ? "更新失败" : "创建失败");
    } finally {
      submitting.value = false;
    }
  });
}

onMounted(loadStyles);
</script>

<style scoped>
.style-table-wrapper {
  margin-top: 1.5rem;
}
</style>
