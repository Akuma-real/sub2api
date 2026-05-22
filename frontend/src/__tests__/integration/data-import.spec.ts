import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'
import { adminAPI } from '@/api/admin'

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData: vi.fn()
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('ImportDataModal', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    vi.mocked(adminAPI.accounts.importData).mockReset()
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    await wrapper.find('form').trigger('submit')
    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('导入时携带选择的默认分组', async () => {
    vi.mocked(adminAPI.accounts.importData).mockResolvedValue({
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      account_created: 1,
      account_failed: 0
    })

    const wrapper = mount(ImportDataModal, {
      props: {
        show: true,
        groups: [
          { id: 42, name: 'openai-default', platform: 'openai', status: 'active' } as any
        ]
      },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          Select: {
            props: ['modelValue', 'options', 'disabled'],
            emits: ['update:modelValue'],
            template: `
              <select
                data-test="default-group-select"
                :disabled="disabled"
                :value="modelValue === null || modelValue === undefined ? '' : String(modelValue)"
                @change="$emit('update:modelValue', $event.target.value === '' ? null : Number($event.target.value))"
              >
                <option
                  v-for="option in options"
                  :key="String(option.value)"
                  :value="option.value === null ? '' : String(option.value)"
                >
                  {{ option.label }}
                </option>
              </select>
            `
          }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File([
      JSON.stringify({ type: 'sub2api-data', version: 1, exported_at: '2026-05-22T00:00:00Z', proxies: [], accounts: [] })
    ], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve(JSON.stringify({ type: 'sub2api-data', version: 1, exported_at: '2026-05-22T00:00:00Z', proxies: [], accounts: [] }))
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.get('[data-test="default-group-select"]').setValue('42')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(adminAPI.accounts.importData).toHaveBeenCalledWith({
      data: {
        type: 'sub2api-data',
        version: 1,
        exported_at: '2026-05-22T00:00:00Z',
        proxies: [],
        accounts: []
      },
      group_ids: [42],
      skip_default_group_bind: true
    })
  })

  it('支持拖拽选择导入文件', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const file = new File(['{}'], 'dragged.json', { type: 'application/json' })
    await wrapper.find('[data-test="data-import-dropzone"]').trigger('drop', {
      dataTransfer: { files: [file] }
    })

    expect(wrapper.text()).toContain('dragged.json')
  })
})
