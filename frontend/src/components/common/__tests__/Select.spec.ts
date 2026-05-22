import { describe, expect, it, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'

import Select from '../Select.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const options = [
  { value: '', label: '全部' },
  { value: 'openai', label: 'OpenAI' }
]

describe('Select', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('uses the selected label as the default accessible name', () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: '',
        options
      }
    })

    expect(wrapper.find('.select-trigger').attributes('aria-label')).toBe('全部')
  })

  it('keeps the dropdown inside a narrow viewport', async () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(430)

    const wrapper = mount(Select, {
      props: {
        modelValue: '',
        options
      },
      attachTo: document.body
    })

    vi.spyOn(wrapper.element, 'getBoundingClientRect').mockReturnValue({
      left: 300,
      right: 380,
      top: 120,
      bottom: 154,
      width: 80,
      height: 34,
      x: 300,
      y: 120,
      toJSON: () => ({})
    } as DOMRect)

    await wrapper.find('.select-trigger').trigger('click')
    await wrapper.vm.$nextTick()

    const dropdown = document.body.querySelector('.select-dropdown-portal') as HTMLElement | null
    expect(dropdown?.style.left).toBe('218px')
    expect(dropdown?.style.width).toBe('200px')

    wrapper.unmount()
  })
})
