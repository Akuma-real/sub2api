import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AccountTestModal from '../AccountTestModal.vue'

const { getAvailableModels, getById, copyToClipboard } = vi.hoisted(() => ({
  getAvailableModels: vi.fn(),
  getById: vi.fn(),
  copyToClipboard: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getAvailableModels,
      getById
    }
  }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.accounts.imagePromptDefault': 'Generate a cute orange cat astronaut sticker on a clean pastel background.'
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (key === 'admin.accounts.imageReceived' && params?.count) {
          return `received-${params.count}`
        }
        return messages[key] || key
      }
    })
  }
})

function createStreamResponse(lines: string[]) {
  const encoder = new TextEncoder()
  const chunks = lines.map((line) => encoder.encode(line))
  let index = 0

  return {
    ok: true,
    body: {
      getReader: () => ({
        read: vi.fn().mockImplementation(async () => {
          if (index < chunks.length) {
            return { done: false, value: chunks[index++] }
          }
          return { done: true, value: undefined }
        })
      })
    }
  } as Response
}

type MountAccount = {
  id: number
  name: string
  platform: string
  type: string
  status: string
  error_message?: string | null
}

function mountModal(account?: Partial<MountAccount>) {
  return mount(AccountTestModal, {
    props: {
      show: false,
      account: {
        id: 42,
        name: 'Gemini Image Test',
        platform: 'gemini',
        type: 'apikey',
        status: 'error',
        error_message: 'previous failure',
        ...account
      }
    } as any,
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
        Select: {
          props: ['modelValue', 'options', 'valueKey', 'labelKey'],
          emits: ['update:modelValue'],
          template: `
            <div class="select-stub">
              <button
                v-for="option in options"
                :key="String(option[valueKey || 'value'])"
                class="select-option-stub"
                type="button"
                @click="$emit('update:modelValue', option[valueKey || 'value'])"
              >
                {{ option[labelKey || 'label'] }}
              </button>
            </div>
          `
        },
        TextArea: {
          props: ['modelValue'],
          emits: ['update:modelValue'],
          template: '<textarea class="textarea-stub" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
        },
        Icon: true
      }
    }
  })
}

describe('AccountTestModal', () => {
  beforeEach(() => {
    getAvailableModels.mockResolvedValue([
      { id: 'gemini-2.0-flash', display_name: 'Gemini 2.0 Flash' },
      { id: 'gemini-2.5-flash-image', display_name: 'Gemini 2.5 Flash Image' },
      { id: 'gemini-3.1-flash-image', display_name: 'Gemini 3.1 Flash Image' }
    ])
    getById.mockResolvedValue({
      id: 42,
      name: 'Gemini Image Test',
      platform: 'gemini',
      type: 'apikey',
      status: 'active',
      error_message: null
    })
    copyToClipboard.mockReset()
    Object.defineProperty(globalThis, 'localStorage', {
      value: {
        getItem: vi.fn((key: string) => (key === 'auth_token' ? 'test-token' : null)),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn()
      },
      configurable: true
    })
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gemini-2.5-flash-image"}\n',
        'data: {"type":"image","image_url":"data:image/png;base64,QUJD","mime_type":"image/png"}\n',
        'data: {"type":"test_complete","success":true,"duration_ms":2000,"first_token_ms":400,"generation_ms":1600,"input_tokens":12,"output_tokens":8,"total_tokens":20,"output_chars":18,"image_count":1}\n'
      ])
    ) as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('gemini 图片模型测试会携带提示词并渲染图片预览', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const promptInput = wrapper.find('textarea.textarea-stub')
    expect(promptInput.exists()).toBe(true)
    await promptInput.setValue('draw a tiny orange cat astronaut')

    const buttons = wrapper.findAll('button')
    const startButton = buttons.find((button) => button.text().includes('admin.accounts.startTest'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await flushPromises()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toEqual({
      model_id: 'gemini-3.1-flash-image',
      prompt: 'draw a tiny orange cat astronaut',
      mode: 'default'
    })

    const preview = wrapper.find('img[alt="test-image-1"]')
    expect(preview.exists()).toBe(true)
    expect(preview.attributes('src')).toBe('data:image/png;base64,QUJD')
    expect(wrapper.text()).toContain('admin.accounts.testPassedThisRun')
    expect(wrapper.text()).toContain('5.00 admin.accounts.testMetrics.tokensPerSecondUnit')
    expect(wrapper.text()).toContain('20')
    expect(getById).toHaveBeenCalledWith(42)
    expect(wrapper.emitted('tested')?.[0]?.[0]).toMatchObject({ id: 42, status: 'active' })
  })

  it('速度测试模式会发送 speed mode 并优先展示后端吞吐率', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT 5.4' }
    ])
    getById.mockResolvedValue({
      id: 42,
      name: 'OpenAI Speed Test',
      platform: 'openai',
      type: 'apikey',
      status: 'active',
      error_message: null
    })
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gpt-5.4","mode":"speed"}\n',
        'data: {"type":"content","text":"benchmark output"}\n',
        'data: {"type":"test_complete","success":true,"duration_ms":3000,"first_token_ms":500,"generation_ms":2500,"input_tokens":180,"output_tokens":750,"total_tokens":930,"output_chars":3200,"output_tokens_per_second":300,"output_chars_per_second":1280}\n'
      ])
    ) as any

    const wrapper = mountModal({
      name: 'OpenAI Speed Test',
      platform: 'openai',
      status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    const speedOption = wrapper.findAll('button.select-option-stub')
      .find((button) => button.text().includes('admin.accounts.openai.testModeSpeed'))
    expect(speedOption).toBeTruthy()
    await speedOption!.trigger('click')

    const startButton = wrapper.findAll('button')
      .find((button) => button.text().includes('admin.accounts.startTest'))
    expect(startButton).toBeTruthy()
    await startButton!.trigger('click')
    await flushPromises()
    await flushPromises()

    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toEqual({
      model_id: 'gpt-5.4',
      prompt: '',
      mode: 'speed'
    })
    expect(wrapper.text()).toContain('admin.accounts.sendingSpeedTestMessage')
    expect(wrapper.text()).toContain('300.0 admin.accounts.testMetrics.tokensPerSecondUnit')
    expect(wrapper.text()).toContain('admin.accounts.speedTestMode')
  })
})
