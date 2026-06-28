import { describe, expect, it } from 'vitest'
import {
  buildMessageSnapshot,
  parseMessages,
} from '@/utils/chatSessionUtils.ts'

describe('buildMessageSnapshot', () => {
  it('creates fingerprint from message properties', () => {
    const msgs = [
      { id: '1', role: 'user', content: 'hello', createdAt: '2026-01-01T00:00:00Z', streaming: false },
    ]
    const snapshot = buildMessageSnapshot(msgs)
    expect(snapshot).toBe('1:user:5:2026-01-01T00:00:00Z:0')
  })

  it('handles missing id', () => {
    const msgs = [
      { role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false },
    ]
    const snapshot = buildMessageSnapshot(msgs)
    expect(snapshot).toBe(':user:2:2026-01-01:0')
  })

  it('handles empty content', () => {
    const msgs = [
      { id: '2', role: 'assistant', content: '', createdAt: '', streaming: true },
    ]
    const snapshot = buildMessageSnapshot(msgs)
    expect(snapshot).toBe('2:assistant:0::1')
  })

  it('handles multiple messages', () => {
    const msgs = [
      { id: '1', role: 'user', content: 'hello', createdAt: '2026-01-01', streaming: false },
      { id: '2', role: 'assistant', content: 'world', createdAt: '2026-01-01', streaming: false },
    ]
    const snapshot = buildMessageSnapshot(msgs)
    expect(snapshot).toBe('1:user:5:2026-01-01:0|2:assistant:5:2026-01-01:0')
  })

  it('returns empty for empty array', () => {
    expect(buildMessageSnapshot([])).toBe('')
  })

  it('detects changes in content length', () => {
    const msgs1 = [{ id: '1', role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false }]
    const msgs2 = [{ id: '1', role: 'user', content: 'hello', createdAt: '2026-01-01', streaming: false }]
    expect(buildMessageSnapshot(msgs1)).not.toBe(buildMessageSnapshot(msgs2))
  })

  it('detects streaming flag change', () => {
    const msgs1 = [{ id: '1', role: 'assistant', content: '', createdAt: '', streaming: false }]
    const msgs2 = [{ id: '1', role: 'assistant', content: '', createdAt: '', streaming: true }]
    expect(buildMessageSnapshot(msgs1)).not.toBe(buildMessageSnapshot(msgs2))
  })

  it('detects role change', () => {
    const msgs1 = [{ id: '1', role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false }]
    const msgs2 = [{ id: '1', role: 'assistant', content: 'hi', createdAt: '2026-01-01', streaming: false }]
    expect(buildMessageSnapshot(msgs1)).not.toBe(buildMessageSnapshot(msgs2))
  })

  it('detects createdAt change', () => {
    const msgs1 = [{ id: '1', role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false }]
    const msgs2 = [{ id: '1', role: 'user', content: 'hi', createdAt: '2026-01-02', streaming: false }]
    expect(buildMessageSnapshot(msgs1)).not.toBe(buildMessageSnapshot(msgs2))
  })

  it('produces stable output for identical input', () => {
    const msgs = [{ id: '1', role: 'user', content: 'hello', createdAt: '2026-01-01', streaming: false }]
    expect(buildMessageSnapshot(msgs)).toBe(buildMessageSnapshot(msgs))
  })

  it('handles message count change', () => {
    const msgs1 = [{ id: '1', role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false }]
    const msgs2 = [
      { id: '1', role: 'user', content: 'hi', createdAt: '2026-01-01', streaming: false },
      { id: '2', role: 'assistant', content: 'reply', createdAt: '2026-01-01', streaming: false },
    ]
    expect(buildMessageSnapshot(msgs1)).not.toBe(buildMessageSnapshot(msgs2))
  })
})

describe('parseMessages', () => {
  const mockParseAssistantContent = (content: string) => {
    if (!content) return { blocks: [], metadata: null, cancelled: false }
    try {
      const parsed = JSON.parse(content)
      if (parsed.blocks) return { blocks: parsed.blocks, metadata: parsed.metadata || null, cancelled: parsed.cancelled || false }
    } catch {}
    return { blocks: [{ type: 'text', text: content }], metadata: null, cancelled: false }
  }

  it('parses assistant messages with blocks', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Hello' }] }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].blocks).toEqual([{ type: 'text', text: 'Hello' }])
  })

  it('parses user messages into text blocks', () => {
    const msgs = [
      { role: 'user', content: 'Hello AI' },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].blocks).toEqual([{ type: 'text', text: 'Hello AI' }])
  })

  it('creates empty blocks for user messages with no content', () => {
    const msgs = [
      { role: 'user', content: '' },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].blocks).toEqual([])
  })

  it('preserves user blocks if already present', () => {
    const msgs = [
      { role: 'user', content: 'Hello', blocks: [{ type: 'text', text: 'Hello' }] },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].blocks).toEqual([{ type: 'text', text: 'Hello' }])
  })

  it('marks streaming assistant messages as fromDB', () => {
    const msgs = [
      { role: 'assistant', content: '', streaming: true },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].fromDB).toBe(true)
    expect(result[0].streaming).toBe(true)
  })

  it('does not mark non-streaming messages as fromDB', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Done' }] }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].fromDB).toBeUndefined()
  })

  it('handles mixed user and assistant messages', () => {
    const msgs = [
      { role: 'user', content: 'Question' },
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Answer' }] }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result).toHaveLength(2)
    expect(result[0].blocks[0].text).toBe('Question')
    expect(result[1].blocks[0].text).toBe('Answer')
  })

  it('extracts metadata from assistant content', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Hi' }], metadata: { tokens: 50 } }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].metadata).toEqual({ tokens: 50 })
  })

  it('extracts cancelled flag from assistant content', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'partial' }], cancelled: true }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].cancelled).toBe(true)
  })

  it('does not set metadata when null from parser', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Hi' }] }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    // metadata is null from parser — condition `if (metadata)` is false
    expect(result[0].metadata).toBeUndefined()
  })

  it('does not set cancelled when false from parser', () => {
    const msgs = [
      { role: 'assistant', content: JSON.stringify({ blocks: [{ type: 'text', text: 'Hi' }] }) },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    // cancelled is false from parser — condition `if (cancelled)` is false
    expect(result[0].cancelled).toBeUndefined()
  })

  it('handles empty array', () => {
    const result = parseMessages([], mockParseAssistantContent)
    expect(result).toEqual([])
  })

  it('preserves other message properties', () => {
    const msgs = [
      { role: 'user', content: 'Hello', id: 'msg-1', createdAt: '2026-01-01' },
    ]
    const result = parseMessages(msgs, mockParseAssistantContent)
    expect(result[0].id).toBe('msg-1')
    expect(result[0].createdAt).toBe('2026-01-01')
  })
})
