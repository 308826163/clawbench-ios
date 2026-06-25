/**
 * Chat context composable for managing attached files and quote data.
 *
 * This composable provides a unified interface for managing:
 * - Attached files (files attached to chat messages)
 * - Quote data (text selected for quoting in chat)
 *
 * It replaces the separate useFileUpload.ts and useQuoteQuestion.ts
 * implementations with a single shared context.
 */

import { ref, computed } from 'vue'

/** Quote data structure */
export interface QuoteData {
    text: string
    filePath?: string
    language?: string
    startLine?: number
    endLine?: number
}

// Global shared state (all components share the same instance)
const attachedFiles = ref<string[]>([])
const quoteData = ref<QuoteData | null>(null)

/**
 * Chat context composable for managing attached files and quote data.
 */
export function useChatContext() {
    /**
     * Add a file to the attached files list.
     * @param path File path to attach
     */
    function addAttachedFile(path: string) {
        if (path && !attachedFiles.value.includes(path)) {
            attachedFiles.value = [...attachedFiles.value, path]
        }
    }

    /**
     * Remove a file from the attached files list by index.
     * @param index Index of the file to remove
     */
    function removeAttachedFile(index: number) {
        const newFiles = [...attachedFiles.value]
        newFiles.splice(index, 1)
        attachedFiles.value = newFiles
    }

    /**
     * Remove a file from the attached files list by path.
     * @param path File path to remove
     */
    function removeAttachedFileByPath(path: string) {
        const index = attachedFiles.value.indexOf(path)
        if (index >= 0) {
            const newFiles = [...attachedFiles.value]
            newFiles.splice(index, 1)
            attachedFiles.value = newFiles
        }
    }

    /**
     * Toggle a file in the attached files list.
     * @param path File path to toggle
     */
    function toggleAttachedFile(path: string) {
        if (!path) return
        const index = attachedFiles.value.indexOf(path)
        if (index >= 0) {
            removeAttachedFile(index)
        } else {
            addAttachedFile(path)
        }
    }

    /**
     * Check if a file is in the attached files list.
     * @param path File path to check
     * @returns True if the file is attached
     */
    function hasAttachedFile(path: string): boolean {
        return attachedFiles.value.includes(path)
    }

    /**
     * Set the quote data.
     * @param data Quote data to set
     */
    function setQuoteData(data: QuoteData | null) {
        quoteData.value = data
    }

    /**
     * Clear all attached files and quote data.
     */
    function clearAll() {
        attachedFiles.value = []
        quoteData.value = null
    }

    /**
     * Number of attached files.
     */
    const attachmentCount = computed(() => attachedFiles.value.length)

    /**
     * Whether there are any attached files.
     */
    const hasAttachments = computed(() => attachedFiles.value.length > 0)

    return {
        // State
        attachedFiles,
        quoteData,
        attachmentCount,
        hasAttachments,

        // Methods
        addAttachedFile,
        removeAttachedFile,
        removeAttachedFileByPath,
        toggleAttachedFile,
        hasAttachedFile,
        setQuoteData,
        clearAll,
    }
}
