<script lang="ts" setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { FolderOpened, Refresh, Search } from '@element-plus/icons-vue'
import { QueryGitLogs, ScanRepositories, SelectDirectory } from '../wailsjs/go/main/App'
import { model } from '../wailsjs/go/models'
import type { GitCommit, GitLogResponse, RepoError, RepoInfo } from './types/git'
import { formatFiles, parseRepoNames } from './utils/formatter'

const rootPath = ref('')
const excludeRepoText = ref('')
const maxDepth = ref(2)
const since = ref('')
const until = ref('')
const author = ref('')
const noMerges = ref(true)
const repos = ref<RepoInfo[]>([])
const selectedRepoPaths = ref<string[]>([])
const commits = ref<GitCommit[]>([])
const repoErrors = ref<RepoError[]>([])
const scanning = ref(false)
const querying = ref(false)

const selectedRepos = computed(() => {
  const selected = new Set(selectedRepoPaths.value)
  return repos.value.filter((repo) => selected.has(repo.path))
})

async function handleSelectDirectory() {
  const path = await SelectDirectory()
  if (path) {
    rootPath.value = path
  }
}

async function handleScanRepositories() {
  if (!rootPath.value.trim()) {
    ElMessage.warning('请选择或填写根目录')
    return
  }

  scanning.value = true
  commits.value = []
  repoErrors.value = []
  try {
    const result = await ScanRepositories(model.ScanRequest.createFrom({
      rootPath: rootPath.value.trim(),
      maxDepth: maxDepth.value,
      excludeRepoNames: parseRepoNames(excludeRepoText.value),
    }))
    repos.value = result || []
    selectedRepoPaths.value = repos.value.map((repo) => repo.path)
    ElMessage.success(`扫描到 ${repos.value.length} 个 Git 仓库`)
  } catch (error) {
    ElMessage.error(String(error))
  } finally {
    scanning.value = false
  }
}

async function handleQueryGitLogs() {
  if (selectedRepos.value.length === 0) {
    ElMessage.warning('请至少勾选一个仓库')
    return
  }

  querying.value = true
  repoErrors.value = []
  try {
    const result = (await QueryGitLogs(model.GitLogRequest.createFrom({
      repos: selectedRepos.value,
      since: since.value || '',
      until: until.value ? `${until.value} 23:59:59` : '',
      author: author.value.trim(),
      noMerges: noMerges.value,
    }))) as GitLogResponse
    commits.value = result?.commits || []
    repoErrors.value = result?.errors || []
    if (repoErrors.value.length > 0) {
      ElMessage.warning(`已获取 ${commits.value.length} 条记录，${repoErrors.value.length} 个仓库失败`)
    } else {
      ElMessage.success(`已获取 ${commits.value.length} 条提交记录`)
    }
  } catch (error) {
    ElMessage.error(String(error))
  } finally {
    querying.value = false
  }
}
</script>

<template>
  <main class="app-shell">
    <section class="toolbar">
      <div class="title-block">
        <h1>Git Worklog Extractor</h1>
        <span>扫描本地仓库并提取提交记录</span>
      </div>
      <el-button :icon="Refresh" :loading="scanning" type="primary" @click="handleScanRepositories">
        扫描仓库
      </el-button>
    </section>

    <section class="filter-panel">
      <el-form label-position="top">
        <div class="form-grid">
          <el-form-item label="根目录">
            <div class="path-row">
              <el-input v-model="rootPath" placeholder="请选择或填写 Git 项目根目录" clearable />
              <el-button :icon="FolderOpened" @click="handleSelectDirectory">选择目录</el-button>
            </div>
          </el-form-item>

          <el-form-item label="扫描深度">
            <el-input-number v-model="maxDepth" :min="0" :max="12" />
          </el-form-item>

          <el-form-item label="排除项目名称">
            <el-input
              v-model="excludeRepoText"
              :autosize="{ minRows: 2, maxRows: 4 }"
              placeholder="多个项目用逗号或换行分隔"
              type="textarea"
            />
          </el-form-item>

          <el-form-item label="作者">
            <el-input v-model="author" placeholder="支持姓名或邮箱关键词" clearable />
          </el-form-item>

          <el-form-item label="开始时间">
            <el-date-picker v-model="since" type="date" value-format="YYYY-MM-DD" placeholder="选择开始日期" />
          </el-form-item>

          <el-form-item label="结束时间">
            <el-date-picker v-model="until" type="date" value-format="YYYY-MM-DD" placeholder="选择结束日期" />
          </el-form-item>
        </div>
        <div class="action-row">
          <el-checkbox v-model="noMerges">排除 Merge 提交</el-checkbox>
          <el-button :icon="Search" :loading="querying" type="success" @click="handleQueryGitLogs">
            获取提交记录
          </el-button>
        </div>
      </el-form>
    </section>

    <section class="repo-section">
      <div class="section-header">
        <h2>仓库列表</h2>
        <span>{{ selectedRepos.length }} / {{ repos.length }} 已选择</span>
      </div>
      <el-table
        v-if="repos.length > 0"
        :data="repos"
        height="220"
        row-key="path"
      >
        <el-table-column label="" width="48">
          <template #default="{ row }">
            <el-checkbox v-model="selectedRepoPaths" :label="row.path">
              <span class="sr-only">选择仓库</span>
            </el-checkbox>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="项目名" min-width="180" />
        <el-table-column prop="path" label="路径" min-width="420" show-overflow-tooltip />
      </el-table>
      <el-empty v-else description="尚未扫描到仓库" />
    </section>

    <section v-if="repoErrors.length > 0" class="error-section">
      <el-alert
        v-for="error in repoErrors"
        :key="`${error.repoPath}-${error.message}`"
        :title="`${error.repoName || error.repoPath}：${error.message}`"
        type="warning"
        show-icon
        :closable="false"
      />
    </section>

    <section class="commit-section">
      <div class="section-header">
        <h2>提交记录</h2>
        <span>{{ commits.length }} 条</span>
      </div>
      <el-table :data="commits" height="360" stripe>
        <el-table-column prop="repoName" label="项目名" width="180" show-overflow-tooltip />
        <el-table-column prop="commitTime" label="提交时间" width="190" />
        <el-table-column prop="authorName" label="作者" width="150" show-overflow-tooltip />
        <el-table-column prop="message" label="提交说明" min-width="260" show-overflow-tooltip />
        <el-table-column label="变更文件" min-width="320">
          <template #default="{ row }">
            <pre class="file-list">{{ formatFiles(row.files) }}</pre>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </main>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  padding: 24px;
  box-sizing: border-box;
  background: #f5f7fb;
  color: #1f2937;
  text-align: left;
}

.toolbar,
.section-header,
.action-row,
.path-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar,
.section-header,
.action-row {
  justify-content: space-between;
}

.title-block h1,
.section-header h2 {
  margin: 0;
  letter-spacing: 0;
}

.title-block h1 {
  font-size: 24px;
  line-height: 32px;
}

.title-block span,
.section-header span {
  color: #6b7280;
  font-size: 13px;
}

.filter-panel,
.repo-section,
.commit-section,
.error-section {
  margin-top: 16px;
  padding: 16px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #ffffff;
}

.form-grid {
  display: grid;
  grid-template-columns: minmax(320px, 2fr) 160px minmax(240px, 1fr);
  gap: 12px 16px;
}

.path-row {
  width: 100%;
}

.path-row .el-input {
  flex: 1;
}

.error-section {
  display: grid;
  gap: 8px;
}

.file-list {
  max-height: 96px;
  margin: 0;
  overflow: auto;
  color: #374151;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 18px;
  white-space: pre-wrap;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 900px) {
  .toolbar,
  .action-row,
  .path-row {
    align-items: stretch;
    flex-direction: column;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
