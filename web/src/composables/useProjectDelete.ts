import { ref } from 'vue'
import { apiDelete } from '@/utils/api'
import { useToast } from '@/composables/useToast'

export function useProjectDelete() {
  const toast = useToast()
  const deleting = ref(false)

  /**
   * Delete a project completely (including chat sessions and history).
   * Returns true if the project was the current project and needs switching.
   */
  async function deleteProject(path: string): Promise<boolean> {
    deleting.value = true
    try {
      await apiDelete('/api/recent-projects', {
        body: {
          path: path,
          completely: true,
        },
      })
      toast.show('项目已删除', { icon: '✓', type: 'success', duration: 3000 })
      return true
    } catch (err) {
      console.error('Failed to delete project:', err)
      toast.show('删除项目失败', { icon: '⚠️', type: 'error', duration: 3000 })
      return false
    } finally {
      deleting.value = false
    }
  }

  /**
   * Remove a project from recent list only (without deleting chat data).
   */
  async function removeFromRecent(path: string): Promise<boolean> {
    deleting.value = true
    try {
      await apiDelete('/api/recent-projects', {
        body: {
          path: path,
          completely: false,
        },
      })
      toast.show('已从列表中移除', { icon: '✓', type: 'success', duration: 3000 })
      return true
    } catch (err) {
      console.error('Failed to remove project:', err)
      toast.show('移除失败', { icon: '⚠️', type: 'error', duration: 3000 })
      return false
    } finally {
      deleting.value = false
    }
  }

  return {
    deleting,
    deleteProject,
    removeFromRecent,
  }
}
