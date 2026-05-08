<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, FolderOpened, Refresh, Search } from '@element-plus/icons-vue'
import { QueryGitLogs, ScanRepositories, SelectDirectory } from '../wailsjs/go/main/App'
import { model } from '../wailsjs/go/models'
import { ClipboardSetText } from '../wailsjs/runtime/runtime'
import type { GitCommit, GitLogResponse, RepoError, RepoInfo } from './types/git'
import { parseRepoNames } from './utils/formatter'

const SETTINGS_STORAGE_KEY = 'git-worklog-extractor:settings'

interface PersistedSettings {
  rootPath: string
  excludeRepoText: string
  maxDepth: number
  dateRange: string[]
  since?: string
  until?: string
  author: string
  noMerges: boolean
}

const rootPath = ref('')
const excludeRepoText = ref('')
const maxDepth = ref(2)
const dateRange = ref<[string, string] | []>([])
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

const selectedRepoSummary = computed(() => `${selectedRepos.value.length} / ${repos.value.length} 已选择`)

const commitCopyText = computed(() => formatCommitForCopy(commits.value))

function normalizeDateRange(value: unknown): [string, string] | [] {
  if (!Array.isArray(value) || value.length < 2) {
    return []
  }

  const start = typeof value[0] === 'string' ? value[0] : ''
  const end = typeof value[1] === 'string' ? value[1] : ''
  return start || end ? [start, end] : []
}

function normalizeSettings(settings: Partial<PersistedSettings>): PersistedSettings {
  const nextMaxDepth = Number(settings.maxDepth)
  const legacyDateRange = settings.since || settings.until ? [settings.since || '', settings.until || ''] : []

  return {
    rootPath: typeof settings.rootPath === 'string' ? settings.rootPath : '',
    excludeRepoText: typeof settings.excludeRepoText === 'string' ? settings.excludeRepoText : '',
    maxDepth: Number.isFinite(nextMaxDepth) ? nextMaxDepth : 2,
    dateRange: Array.isArray(settings.dateRange) ? settings.dateRange : legacyDateRange,
    author: typeof settings.author === 'string' ? settings.author : '',
    noMerges: typeof settings.noMerges === 'boolean' ? settings.noMerges : true,
  }
}

function loadSettings() {
  try {
    const rawSettings = localStorage.getItem(SETTINGS_STORAGE_KEY)
    if (!rawSettings) {
      return
    }

    const settings = normalizeSettings(JSON.parse(rawSettings) as Partial<PersistedSettings>)
    rootPath.value = settings.rootPath
    excludeRepoText.value = settings.excludeRepoText
    maxDepth.value = settings.maxDepth
    dateRange.value = normalizeDateRange(settings.dateRange)
    author.value = settings.author
    noMerges.value = settings.noMerges
  } catch {
    // Ignore invalid persisted settings so the app can still start normally.
  }
}

function saveSettings() {
  const settings: PersistedSettings = {
    rootPath: rootPath.value,
    excludeRepoText: excludeRepoText.value,
    maxDepth: Number(maxDepth.value),
    dateRange: [...dateRange.value],
    author: author.value,
    noMerges: Boolean(noMerges.value),
  }

  try {
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(settings))
  } catch {
    // localStorage may be unavailable or full; this should not block normal usage.
  }
}

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

  const [rangeSince, rangeUntil] = dateRange.value

  querying.value = true
  repoErrors.value = []
  try {
    const result = (await QueryGitLogs(model.GitLogRequest.createFrom({
      repos: selectedRepos.value,
      since: rangeSince || '',
      until: rangeUntil ? `${rangeUntil} 23:59:59` : '',
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

function handleSelectAllRepos() {
  selectedRepoPaths.value = repos.value.map((repo) => repo.path)
}

function handleClearSelectedRepos() {
  selectedRepoPaths.value = []
}

function formatCommitForCopy(commitList: GitCommit[]): string {
  return commitList
    .map((commit) => {
      const files = commit.files.length > 0 ? commit.files.join('\n') : '无文件变更记录'

      return [
        `项目 [${commit.repoName}]`,
        `提交时间: ${commit.commitTime}`,
        `作者: ${commit.authorName}`,
        `说明: ${commit.message}`,
        '',
        '变更文件:',
        files,
        '----------------------------------------',
      ].join('\n')
    })
    .join('\n\n')
}

async function copyTextToClipboard(text: string) {
  try {
    await ClipboardSetText(text)
    return true
  } catch {}

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch {}

  try {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.setAttribute('readonly', 'true')
    textarea.style.position = 'fixed'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(textarea)
    return ok
  } catch {
    return false
  }
}

async function handleCopyAllCommits() {
  const text = commitCopyText.value.trim()
  if (!text) {
    ElMessage.warning('暂无可复制的提交记录')
    return
  }

  const ok = await copyTextToClipboard(text)
  if (ok) {
    ElMessage.success('已复制提交记录文本')
  } else {
    ElMessage.error('复制失败，请手动选择文本复制')
  }
}

onMounted(loadSettings)

watch(
  [rootPath, excludeRepoText, maxDepth, dateRange, author, noMerges],
  saveSettings,
)
</script>

<template>
  <main class="app-shell">
    <div class="main-layout">
      <section class="toolbar glass-card">
        <div class="title-block">
          <h1>Git Worklog Extractor</h1>
          <span>本地 Git 工作记录提取器</span>
        </div>
        <div class="toolbar-actions">
          <el-button :icon="Refresh" :loading="scanning" type="primary" @click="handleScanRepositories">
            扫描仓库
          </el-button>
          <el-button :icon="Search" :loading="querying" type="success" @click="handleQueryGitLogs">
            获取提交记录
          </el-button>
        </div>
      </section>

      <section class="filter-panel glass-card">
        <el-form label-position="top">
          <div class="filter-grid">
            <el-form-item label="根目录" class="root-path-field">
              <div class="path-row">
                <el-input v-model="rootPath" placeholder="请选择或填写 Git 项目根目录" clearable />
                <el-button :icon="FolderOpened" @click="handleSelectDirectory">选择目录</el-button>
              </div>
            </el-form-item>

            <el-form-item label="日期范围" class="date-range-field">
              <el-date-picker
                v-model="dateRange"
                type="daterange"
                value-format="YYYY-MM-DD"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                range-separator="至"
              />
            </el-form-item>

            <el-form-item label="扫描深度" class="depth-field">
              <el-input-number v-model="maxDepth" :min="0" :max="12" />
            </el-form-item>

            <el-form-item label="提交类型" class="merge-field">
              <el-checkbox v-model="noMerges">排除 Merge 提交</el-checkbox>
            </el-form-item>

            <el-form-item label="排除项目名称" class="exclude-field">
              <el-input
                v-model="excludeRepoText"
                placeholder="多个项目用逗号分隔"
                clearable
              />
            </el-form-item>

            <el-form-item label="作者" class="author-field">
              <el-input v-model="author" placeholder="姓名或邮箱关键字" clearable />
            </el-form-item>
          </div>
        </el-form>
      </section>

      <section v-if="repoErrors.length > 0" class="error-section glass-card">
        <el-alert
          v-for="error in repoErrors"
          :key="`${error.repoPath}-${error.message}`"
          :title="`${error.repoName || error.repoPath}: ${error.message}`"
          type="warning"
          show-icon
          :closable="false"
        />
      </section>

      <section class="content-grid">
        <section class="repo-section glass-card">
          <div class="section-header">
            <div>
              <h2>仓库列表</h2>
              <span>{{ selectedRepoSummary }}</span>
            </div>
            <div class="section-actions">
              <el-button size="small" @click="handleSelectAllRepos">全选</el-button>
              <el-button size="small" @click="handleClearSelectedRepos">清空</el-button>
            </div>
          </div>

          <el-table
            v-if="repos.length > 0"
            :data="repos"
            class="repo-table"
            height="100%"
            row-key="path"
          >
            <el-table-column label="" width="46">
              <template #default="{ row }">
                <el-checkbox v-model="selectedRepoPaths" :label="row.path">
                  <span class="sr-only">选择仓库</span>
                </el-checkbox>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="项目名" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tooltip :content="row.path" placement="top" effect="dark">
                  <span class="repo-name">{{ row.name }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else class="fill-empty" description="尚未扫描到仓库" />
        </section>

        <section class="commit-section glass-card">
          <div class="section-header">
            <div>
              <h2>提交记录</h2>
              <span>{{ commits.length }} 条</span>
            </div>
            <el-button :icon="DocumentCopy" size="small" type="primary" plain @click="handleCopyAllCommits">
              复制当前列表文本
            </el-button>
          </div>

          <div class="commit-preview">
            <pre v-if="commitCopyText" class="commit-text">{{ commitCopyText }}</pre>
            <el-empty v-else class="commit-empty" description="暂无提交记录，请先扫描并获取提交记录" />
          </div>
        </section>
      </section>
    </div>
  </main>
</template>

<style scoped>
.app-shell {
  height: 100vh;
  overflow: hidden;
  box-sizing: border-box;
  background: linear-gradient(135deg, #f5f7fb 0%, #eef3f8 100%);
  color: #1f2937;
  text-align: left;
}

.main-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 0;
  padding: 18px 22px;
  box-sizing: border-box;
  gap: 14px;
}

.glass-card {
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 18px 50px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(18px);
}

.toolbar,
.toolbar-actions,
.section-header,
.section-actions,
.path-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar,
.section-header {
  justify-content: space-between;
}

.toolbar {
  flex: 0 0 auto;
  padding: 16px 18px;
}

.title-block h1,
.section-header h2 {
  margin: 0;
  letter-spacing: 0;
}

.title-block h1 {
  color: #111827;
  font-size: 24px;
  font-weight: 700;
  line-height: 32px;
}

.title-block span,
.section-header span {
  color: #64748b;
  font-size: 13px;
}

.toolbar :deep(.el-button),
.filter-panel :deep(.el-button),
.section-actions :deep(.el-button),
.commit-section :deep(.el-button) {
  border-radius: 999px;
}

.filter-panel {
  flex: 0 0 auto;
  padding: 12px 16px 2px;
}

.filter-panel :deep(.el-form-item) {
  margin-bottom: 10px;
}

.filter-panel :deep(.el-form-item__label) {
  margin-bottom: 4px;
  color: #475569;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}

.filter-panel :deep(.el-date-editor.el-input),
.filter-panel :deep(.el-date-editor.el-input__wrapper),
.filter-panel :deep(.el-input-number) {
  width: 100%;
}

.filter-panel :deep(.el-input__wrapper),
.filter-panel :deep(.el-textarea__inner) {
  border-radius: 12px;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.28) inset;
}

.filter-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(260px, 1.35fr) 170px 118px 150px;
  gap: 10px 14px;
  align-items: end;
}

.root-path-field {
  grid-column: span 2;
}

.date-range-field {
  grid-column: span 1;
}

.exclude-field {
  grid-column: span 3;
}

.author-field {
  grid-column: span 2;
}

.depth-field,
.merge-field {
  grid-column: span 1;
}

.merge-field {
  align-self: end;
}

.merge-field :deep(.el-form-item__content) {
  min-height: 32px;
  align-items: center;
}

.merge-field :deep(.el-checkbox) {
  height: 32px;
}

.path-row {
  width: 100%;
}

.path-row .el-input {
  min-width: 0;
  flex: 1;
}

.error-section {
  display: grid;
  flex: 0 0 auto;
  gap: 8px;
  padding: 10px 14px;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(280px, 0.32fr) minmax(560px, 0.68fr);
  flex: 1;
  min-height: 0;
  gap: 16px;
}

.repo-section,
.commit-section {
  display: flex;
  min-width: 0;
  min-height: 0;
  padding: 14px 16px 16px;
  box-sizing: border-box;
  flex-direction: column;
}

.section-header {
  flex: 0 0 auto;
  margin-bottom: 12px;
}

.section-header > div:first-child {
  min-width: 0;
}

.section-header h2 {
  color: #111827;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
}

.repo-table {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border-radius: 12px;
}

.repo-table :deep(.el-table__inner-wrapper::before) {
  display: none;
}

.repo-table :deep(.el-table__cell) {
  padding: 7px 0;
}

.repo-table :deep(.el-table__body-wrapper) {
  overflow-x: hidden;
}

.repo-name {
  display: block;
  max-width: 100%;
  overflow: hidden;
  color: #1f2937;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fill-empty,
.commit-empty {
  flex: 1;
}

.commit-preview {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.82);
}

.commit-text {
  min-height: 100%;
  margin: 0;
  padding: 16px 18px;
  box-sizing: border-box;
  color: #1f2937;
  font-family: Consolas, "SFMono-Regular", "Liberation Mono", "Courier New", monospace;
  font-size: 13px;
  line-height: 22px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
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

@media (max-width: 800px) {
  .main-layout {
    padding: 14px;
  }

  .toolbar,
  .toolbar-actions,
  .path-row {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-grid,
  .content-grid {
    grid-template-columns: 1fr;
  }

  .root-path-field,
  .date-range-field,
  .exclude-field,
  .author-field,
  .depth-field,
  .merge-field {
    grid-column: auto;
  }

  .content-grid {
    overflow: auto;
  }

  .repo-section,
  .commit-section {
    min-height: 360px;
  }
}
</style>
