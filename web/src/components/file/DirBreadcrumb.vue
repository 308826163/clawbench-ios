<template>
  <div class="dir-breadcrumb-wrap">
    <!-- 盘符选择按钮（在面包屑外面，避免 overflow 裁剪弹窗） -->
    <div class="drive-switcher-wrap">
      <span class="crumb drive-trigger" @click.stop="driveOpen = !driveOpen">
        <HardDrive :size="14" />
      </span>
      <!-- 盘符卡片弹窗 -->
      <div v-if="driveOpen" class="drive-dropdown" @click.stop>
        <div
          v-for="drive in driveList"
          :key="drive.letter"
          class="drive-item"
          :class="{ active: drive.root === currentRoot }"
          @click="onDriveClick(drive.root)"
        >
          <span class="drive-letter">
            <HardDrive :size="14" />
            {{ drive.letter }}
          </span>
          <span class="drive-path">{{ truncatePath(drive.root) }}</span>
        </div>
      </div>
    </div>

    <!-- 面包屑路径（可滚动，默认滚到最右） -->
    <div class="dir-breadcrumb" ref="breadcrumbRef">
      <template v-for="(part, i) in parts" :key="i">
        <span class="crumb-sep">›</span>
        <span
          class="crumb"
          :class="{ current: i === parts.length - 1 }"
          @click="i < parts.length - 1 && $emit('navigate', reconstructPath(parts.slice(0, i + 1)))"
        >{{ part }}</span>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { HardDrive, ChevronDown, FolderOpen } from 'lucide-vue-next'
import { splitPath } from '@/utils/path.ts'
import { store } from '@/stores/app.ts'

const props = defineProps({
  path: { type: String, default: '' },
})

const emit = defineEmits(['navigate'])

const driveOpen = ref(false)

// 从 store 获取可用盘符和项目路径
const rootPaths = computed(() => store.state.rootPaths || [])
const projectRoot = computed(() => store.state.projectRoot || '')

// 当前所在盘符根路径
const currentRoot = computed(() => {
  for (const root of rootPaths.value) {
    const cleanRoot = root.replace(/[/\\]$/, '')
    const cleanPath = (props.path || '').replace(/[/\\]$/, '')
    if (cleanPath === cleanRoot || cleanPath.startsWith(cleanRoot + '\\') || cleanPath.startsWith(cleanRoot + '/')) {
      return root
    }
  }
  return rootPaths.value[0] || ''
})

// 构建盘符列表
const driveList = computed(() => {
  return rootPaths.value.map(root => {
    const match = root.match(/^([A-Za-z]):/)
    return {
      letter: match ? `${match[1]}:` : root,
      root: root,
    }
  })
})

// 截断过长路径
function truncatePath(p) {
  if (!p) return ''
  const maxLen = 30
  if (p.length <= maxLen) return p
  return '...' + p.slice(p.length - maxLen + 3)
}

// 点击盘符
function onDriveClick(root) {
  driveOpen.value = false

  // 如果是C盘，跳转到C:\Users\23611
  if (root && root.match(/^[A-Za-z]:\\?$/)) {
    const driveLetter = root.charAt(0).toUpperCase()
    if (driveLetter === 'C') {
      emit('navigate', 'C:\\Users\\23611')
      return
    }
  }

  emit('navigate', root)
}

// 点击外部关闭
function onDocClick() {
  driveOpen.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))

// Reconstruct an absolute path from breadcrumb segments
function reconstructPath(segments) {
  if (segments.length === 0) return ''
  if (/^[A-Za-z]:\\$/.test(segments[0])) {
    const path = segments[0] + segments.slice(1).join('\\')

    // 如果是C盘，限制只能导航到C:\Users及以下的路径
    const driveLetter = segments[0].charAt(0).toUpperCase()
    if (driveLetter === 'C') {
      if (!path.startsWith('C:\\Users') && path !== 'C:\\') {
        // 如果尝试导航到C:\Users以上，返回C:\Users
        return 'C:\\Users'
      }
    }

    return path
  }
  return '/' + segments.join('/')
}

const breadcrumbRef = ref(null)

// 路径变化时滚动到最右边，显示最新（最深）的路径段
watch(() => props.path, async () => {
  await nextTick()
  if (breadcrumbRef.value) {
    breadcrumbRef.value.scrollLeft = breadcrumbRef.value.scrollWidth
  }
})

const parts = computed(() => {
  if (!props.path || props.path === '.') return []
  const segments = splitPath(props.path).filter(p => p !== '')
  if (segments.length > 0 && /^[A-Za-z]:$/.test(segments[0])) {
    segments[0] = segments[0] + '\\'
  }
  return segments
})
</script>

<style scoped>
.dir-breadcrumb-wrap {
  display: flex;
  align-items: center;
  gap: 0;
}
.dir-breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-muted, #999);
  scrollbar-width: none;
}
.dir-breadcrumb::-webkit-scrollbar {
  display: none;
}

.crumb {
  padding: 3px 6px;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s;
  display: inline-flex;
  align-items: center;
}

.crumb:hover {
  background: var(--bg-secondary, #e0e0e0);
  color: var(--accent-color, var(--accent-color));
}

.crumb.current {
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  cursor: default;
}

.crumb.current:hover {
  background: none;
  color: var(--text-primary, #1a1a1a);
}

.crumb-sep {
  color: var(--text-muted, #999);
  font-size: 11px;
}

/* ── 盘符切换器 ── */
.drive-switcher-wrap {
  position: relative;
}

.drive-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 50%;
  cursor: pointer;
  color: var(--text-muted, #999);
  transition: all 0.15s;
}

.drive-trigger:hover {
  background: var(--bg-secondary, #e0e0e0);
  color: var(--accent-color, var(--accent-color));
}

.drive-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  z-index: 200;
  min-width: 200px;
  max-width: 85vw;
  background: var(--bg-primary, #fff);
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 6px;
}

.drive-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 13px;
}

.drive-item:hover {
  background: var(--bg-tertiary, #f0f0f0);
}

.drive-item.active {
  background: color-mix(in srgb, var(--accent-color, var(--accent-color)) 10%, transparent);
}

.drive-item.drive-project {
  border-top: 1px solid var(--border-color, #e5e5e5);
  margin-top: 4px;
  padding-top: 10px;
}

.drive-letter {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  flex-shrink: 0;
  color: var(--text-primary, #1a1a1a);
}

.drive-item.active .drive-letter {
  color: var(--accent-color, var(--accent-color));
}

.drive-path {
  font-size: 11px;
  color: var(--text-muted, #999);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
  max-width: 160px;
}
</style>
