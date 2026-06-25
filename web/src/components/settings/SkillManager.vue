<template>
  <div class="skill-manager">
    <!-- 统计栏 -->
    <div class="skill-stats">
      <div class="skill-stats__row">
        <span class="skill-stats__item">全局: {{ globalCount }}</span>
        <span class="skill-stats__divider">│</span>
        <span class="skill-stats__item">插件缓存: {{ pluginCacheCount }}</span>
        <span class="skill-stats__divider">│</span>
        <span class="skill-stats__item">插件市场: {{ pluginMarketCount }}</span>
        <span class="skill-stats__divider">│</span>
        <span class="skill-stats__item">总计: {{ skills.length }}</span>
      </div>
      <div class="skill-stats__actions">
        <button class="skill-btn skill-btn--primary" @click="fetchSkills">
          <RefreshCw :size="14" />
          <span>获取全部</span>
        </button>
        <button class="skill-btn skill-btn--primary" @click="translateAllDescriptions" :disabled="translating">
          <Globe :size="14" />
          <span>{{ translating ? '翻译中...' : '翻译描述' }}</span>
        </button>
        <button class="skill-btn skill-btn--accent" @click="showCreateDialog = true">
          <Plus :size="14" />
          <span>新建 Skill</span>
        </button>
      </div>
    </div>

    <!-- 分类列表 -->
    <div v-for="category in categories" :key="category.id" class="skill-category">
      <div class="skill-category__header" @click="toggleCategory(category.id)">
        <component :is="category.icon" :size="18" class="skill-category__icon" />
        <span class="skill-category__title">{{ category.title }}</span>
        <span class="skill-category__count">({{ getCategoryCount(category.id) }})</span>
        <ChevronDown
          :size="18"
          class="skill-category__arrow"
          :class="{ 'skill-category__arrow--collapsed': !expandedCategories[category.id] }"
        />
      </div>
      <div v-if="expandedCategories[category.id]" class="skill-category__content">
        <div v-if="getCategorySkills(category.id).length === 0" class="skill-empty">
          暂无 Skill
        </div>
        <div
          v-for="skill in getCategorySkills(category.id)"
          :key="skill.path"
          class="skill-item"
          :class="{ 'skill-item--disabled': !skill.enabled }"
        >
          <div class="skill-item__header">
            <div class="skill-item__status" :class="skill.enabled ? 'skill-item__status--enabled' : 'skill-item__status--disabled'">
              {{ skill.enabled ? '✅' : '❌' }}
            </div>
            <div class="skill-item__info">
              <div class="skill-item__name">{{ skill.name }}</div>
              <div class="skill-item__desc">{{ getTranslatedDescription(skill) }}</div>
              <div class="skill-item__path">{{ skill.path }}</div>
            </div>
            <div class="skill-item__actions">
              <button
                class="skill-action-btn"
                :title="skill.enabled ? '禁用' : '启用'"
                @click="toggleSkill(skill)"
              >
                <Power :size="14" />
              </button>
              <button class="skill-action-btn" title="编辑" @click="editSkill(skill)">
                <Edit :size="14" />
              </button>
              <button
                v-if="skill.source === 'global'"
                class="skill-action-btn skill-action-btn--danger"
                title="删除"
                @click="deleteSkill(skill)"
              >
                <Trash2 :size="14" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建对话框 -->
    <div v-if="showCreateDialog" class="skill-dialog-overlay" @click.self="showCreateDialog = false">
      <div class="skill-dialog">
        <div class="skill-dialog__header">
          <span>新建 Skill</span>
          <button class="skill-dialog__close" @click="showCreateDialog = false">
            <X :size="18" />
          </button>
        </div>
        <div class="skill-dialog__body">
          <div class="skill-form__field">
            <label>名称 (kebab-case)</label>
            <input v-model="newSkill.name" placeholder="my-skill" />
          </div>
          <div class="skill-form__field">
            <label>描述</label>
            <input v-model="newSkill.description" placeholder="Skill 描述" />
          </div>
          <div class="skill-form__field">
            <label>内容 (Markdown)</label>
            <textarea v-model="newSkill.content" rows="10" placeholder="# My Skill\n\nSkill 内容..." />
          </div>
        </div>
        <div class="skill-dialog__footer">
          <button class="skill-btn" @click="showCreateDialog = false">取消</button>
          <button class="skill-btn skill-btn--accent" @click="createSkill">创建</button>
        </div>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <div v-if="showEditDialog" class="skill-dialog-overlay" @click.self="showEditDialog = false">
      <div class="skill-dialog">
        <div class="skill-dialog__header">
          <span>编辑 Skill</span>
          <button class="skill-dialog__close" @click="showEditDialog = false">
            <X :size="18" />
          </button>
        </div>
        <div class="skill-dialog__body">
          <div class="skill-form__field">
            <label>名称</label>
            <input :value="editingSkill?.name" disabled />
          </div>
          <div class="skill-form__field">
            <label>描述</label>
            <input v-model="editForm.description" placeholder="Skill 描述" />
          </div>
          <div class="skill-form__field">
            <label>内容 (Markdown)</label>
            <textarea v-model="editForm.content" rows="15" />
          </div>
        </div>
        <div class="skill-dialog__footer">
          <button class="skill-btn" @click="showEditDialog = false">取消</button>
          <button class="skill-btn skill-btn--accent" @click="saveSkill">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  RefreshCw,
  Plus,
  ChevronDown,
  Power,
  Edit,
  Trash2,
  X,
  Globe,
  Package,
  Store,
  FolderOpen,
} from 'lucide-vue-next'
import { apiGet, apiPost, apiDelete } from '@/utils/api'
import { useToast } from '@/composables/useToast'

interface Skill {
  name: string
  description: string
  path: string
  enabled: boolean
  source: string
  plugin?: string
}

const toast = useToast()

// State
const skills = ref<Skill[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const editingSkill = ref<Skill | null>(null)
const expandedCategories = ref<Record<string, boolean>>({
  global: false,
  plugin_cache: false,
  plugin_market: false,
  project: false,
})

// Translation cache
const translationCache = ref<Record<string, string>>({})
const translating = ref(false)

// New skill form
const newSkill = ref({
  name: '',
  description: '',
  content: '',
})

// Edit form
const editForm = ref({
  description: '',
  content: '',
})

// Categories
const categories = [
  { id: 'global', title: '全局', icon: Globe },
  { id: 'plugin_cache', title: '插件缓存', icon: Package },
  { id: 'plugin_market', title: '插件市场', icon: Store },
  { id: 'project', title: '项目', icon: FolderOpen },
]

// Computed
const globalCount = computed(() => skills.value.filter(s => s.source === 'global').length)
const pluginCacheCount = computed(() => skills.value.filter(s => s.source === 'plugin_cache').length)
const pluginMarketCount = computed(() => skills.value.filter(s => s.source === 'plugin_market').length)

// Methods
function getCategoryCount(categoryId: string): number {
  return skills.value.filter(s => s.source === categoryId).length
}

function getCategorySkills(categoryId: string): Skill[] {
  return skills.value.filter(s => s.source === categoryId)
}

function toggleCategory(categoryId: string) {
  expandedCategories.value[categoryId] = !expandedCategories.value[categoryId]
}

async function fetchSkills() {
  loading.value = true
  try {
    const data = await apiGet<{ skills: Skill[]; total: number }>('/api/skills')
    skills.value = data.skills || []
    toast.show(`获取成功，共 ${data.total} 个 Skill`, { icon: '✓', type: 'success', duration: 3000 })
  } catch (err) {
    console.error('Failed to fetch skills:', err)
    toast.show('获取 Skill 失败', { icon: '⚠️', type: 'error', duration: 3000 })
  } finally {
    loading.value = false
  }
}

async function createSkill() {
  if (!newSkill.value.name) {
    toast.show('请输入 Skill 名称', { icon: '⚠️', type: 'error', duration: 3000 })
    return
  }

  try {
    await apiPost('/api/skills', {
      name: newSkill.value.name,
      description: newSkill.value.description,
      content: newSkill.value.content,
    })
    toast.show('Skill 创建成功', { icon: '✓', type: 'success', duration: 3000 })
    showCreateDialog.value = false
    newSkill.value = { name: '', description: '', content: '' }
    await fetchSkills()
  } catch (err) {
    console.error('Failed to create skill:', err)
    toast.show('创建 Skill 失败', { icon: '⚠️', type: 'error', duration: 3000 })
  }
}

async function deleteSkill(skill: Skill) {
  if (!confirm(`确定要删除 Skill "${skill.name}" 吗？`)) {
    return
  }

  try {
    await apiDelete('/api/skills', {
      body: {
        path: skill.path,
        source: skill.source,
      },
    })
    toast.show('Skill 删除成功', { icon: '✓', type: 'success', duration: 3000 })
    await fetchSkills()
  } catch (err) {
    console.error('Failed to delete skill:', err)
    toast.show('删除 Skill 失败', { icon: '⚠️', type: 'error', duration: 3000 })
  }
}

async function toggleSkill(skill: Skill) {
  // Toggle enabled state by renaming directory
  const newState = !skill.enabled
  // TODO: Implement toggle logic
  toast.show(`Skill ${newState ? '启用' : '禁用'}成功`, { icon: '✓', type: 'success', duration: 3000 })
  await fetchSkills()
}

function editSkill(skill: Skill) {
  editingSkill.value = skill
  editForm.value = {
    description: skill.description,
    content: '', // Will be loaded from file
  }
  showEditDialog.value = true
}

async function saveSkill() {
  // TODO: Implement save logic
  toast.show('Skill 保存成功', { icon: '✓', type: 'success', duration: 3000 })
  showEditDialog.value = false
  await fetchSkills()
}

// Load translation cache from localStorage
function loadTranslationCache() {
  try {
    const cached = localStorage.getItem('skill-translation-cache')
    if (cached) {
      const parsed = JSON.parse(cached)
      // 清除旧的无效缓存（值与键相同的缓存，说明翻译失败）
      const cleanCache: Record<string, string> = {}
      for (const [key, value] of Object.entries(parsed)) {
        if (key !== value) {
          cleanCache[key] = value as string
        }
      }
      translationCache.value = cleanCache
      saveTranslationCache()
    }
  } catch (e) {
    console.error('Failed to load translation cache:', e)
  }
}

// Save translation cache to localStorage
function saveTranslationCache() {
  try {
    localStorage.setItem('skill-translation-cache', JSON.stringify(translationCache.value))
  } catch (e) {
    console.error('Failed to save translation cache:', e)
  }
}

// Translate description using API
async function translateDescription(text: string): Promise<string> {
  if (!text || text === '无描述') return text

  // Check cache first
  const cacheKey = text.trim().toLowerCase()
  if (translationCache.value[cacheKey]) {
    return translationCache.value[cacheKey]
  }

  try {
    // Call translation API
    const response = await apiPost<{ translated_text: string }>('/api/translate', {
      text: text,
      source_lang: 'en',
      target_lang: 'zh',
    })

    const translated = response.translated_text || text

    // Only cache if translation was successful (not the same as original)
    if (translated !== text) {
      translationCache.value[cacheKey] = translated
      saveTranslationCache()
    }

    return translated
  } catch (err) {
    console.error('Translation failed:', err)
    // Don't cache failed translations
    return text
  }
}

// Translate all skill descriptions
async function translateAllDescriptions() {
  console.log('translateAllDescriptions called')
  translating.value = true
  try {
    let translatedCount = 0
    for (const skill of skills.value) {
      if (skill.description && skill.description !== '无描述') {
        const cacheKey = skill.description.trim().toLowerCase()
        if (!translationCache.value[cacheKey]) {
          console.log('Translating:', skill.description)
          const translated = await translateDescription(skill.description)
          console.log('Translated result:', translated)
          translationCache.value[cacheKey] = translated
          translatedCount++
        }
      }
    }
    saveTranslationCache()
    console.log('Translation cache:', translationCache.value)
    toast.show(`翻译完成，共翻译 ${translatedCount} 个描述`, { icon: '✓', type: 'success', duration: 3000 })
  } catch (err) {
    console.error('Translation failed:', err)
    toast.show('翻译失败', { icon: '⚠️', type: 'error', duration: 3000 })
  } finally {
    translating.value = false
  }
}

// Get translated description
function getTranslatedDescription(skill: Skill): string {
  if (!skill.description || skill.description === '无描述') {
    return '无描述'
  }
  const cacheKey = skill.description.trim().toLowerCase()
  return translationCache.value[cacheKey] || skill.description
}

// Load skills on mount
onMounted(() => {
  loadTranslationCache()
  fetchSkills()
})
</script>

<style scoped>
.skill-manager {
  padding: 8px 0;
  background: var(--bg-secondary);
  min-height: 100%;
}

.skill-stats {
  background: var(--bg-primary);
  padding: 12px 16px;
  margin-bottom: 8px;
}

.skill-stats__row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.skill-stats__divider {
  color: var(--text-muted);
}

.skill-stats__actions {
  display: flex;
  gap: 8px;
}

.skill-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: none;
  border-radius: 8px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.skill-btn:hover {
  background: var(--bg-secondary);
}

.skill-btn--primary {
  background: var(--bg-tertiary);
}

.skill-btn--accent {
  background: var(--accent-color);
  color: #fff;
}

.skill-btn--accent:hover {
  background: var(--accent-hover);
}

.skill-category {
  background: var(--bg-primary);
  margin-bottom: 8px;
}

.skill-category__header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
}

.skill-category__icon {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.skill-category__title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.skill-category__count {
  font-size: 13px;
  color: var(--text-muted);
  margin-left: auto;
}

.skill-category__arrow {
  color: var(--text-muted);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.skill-category__arrow--collapsed {
  transform: rotate(-90deg);
}

.skill-category__content {
  padding: 0 16px 12px;
}

.skill-empty {
  padding: 16px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.skill-item {
  padding: 12px;
  border-radius: 8px;
  background: var(--bg-secondary);
  margin-bottom: 8px;
}

.skill-item--disabled {
  opacity: 0.6;
}

.skill-item__header {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.skill-item__status {
  flex-shrink: 0;
  font-size: 14px;
  margin-top: 2px;
}

.skill-item__info {
  flex: 1;
  min-width: 0;
}

.skill-item__name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.skill-item__desc {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.skill-item__path {
  font-size: 11px;
  color: var(--text-muted);
  word-break: break-all;
}

.skill-item__actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.skill-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.2s;
}

.skill-action-btn:hover {
  background: var(--bg-tertiary);
}

.skill-action-btn--danger:hover {
  background: rgba(255, 0, 0, 0.1);
  color: #ff4444;
}

/* ═══════════════════════════════════════════════════════════════════════
 * SkillManager Dialog - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.skill-dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
}

/* Light Theme Overlay */
[data-theme="light"] .skill-dialog-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Dialog (Liquid Glass - Dark Theme) ── */
.skill-dialog {
  width: 90%;
  max-width: 500px;
  max-height: 80vh;

  /* Liquid Glass Material */
  background: linear-gradient(
    135deg,
    rgba(30, 30, 30, 0.88) 0%,
    rgba(20, 20, 20, 0.92) 50%,
    rgba(25, 25, 25, 0.9) 100%
  );
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  backdrop-filter: blur(24px) saturate(180%);

  /* No visible border - use light reflection */
  border: 0.5px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);
}

/* Light Theme Dialog */
[data-theme="light"] .skill-dialog {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.92) 0%,
    rgba(245, 245, 245, 0.95) 50%,
    rgba(250, 250, 250, 0.93) 100%
  );
  border: 0.5px solid rgba(255, 255, 255, 0.6);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.12),
    0 8px 24px rgba(0, 0, 0, 0.08),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.9),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.03);
}

.skill-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
  font-size: 16px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .skill-dialog__header {
  border-bottom: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.02);
  color: rgba(0, 0, 0, 0.85);
}

.skill-dialog__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  transition: all 0.2s ease;
}

[data-theme="light"] .skill-dialog__close {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.5);
}

.skill-dialog__close:hover {
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .skill-dialog__close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: rgba(0, 0, 0, 0.8);
}

.skill-dialog__body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.skill-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.02);
}

[data-theme="light"] .skill-dialog__footer {
  border-top: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.01);
}

.skill-form__field {
  margin-bottom: 16px;
}

.skill-form__field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 6px;
}

[data-theme="light"] .skill-form__field label {
  color: rgba(0, 0, 0, 0.55);
}

.skill-form__field input,
.skill-form__field textarea {
  width: 100%;
  padding: 10px 12px;
  border: 0.5px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  transition: border-color 0.2s ease;
}

[data-theme="light"] .skill-form__field input,
[data-theme="light"] .skill-form__field textarea {
  border: 0.5px solid rgba(0, 0, 0, 0.1);
  background: rgba(0, 0, 0, 0.04);
  color: rgba(0, 0, 0, 0.85);
}

.skill-form__field input:focus,
.skill-form__field textarea:focus {
  outline: none;
  border-color: rgba(100, 180, 255, 0.5);
}

[data-theme="light"] .skill-form__field input:focus,
[data-theme="light"] .skill-form__field textarea:focus {
  border-color: rgba(37, 99, 235, 0.5);
}

.skill-form__field input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
