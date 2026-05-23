<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- Account Info Card -->
      <div
        v-if="displayAccount"
        :class="[
          'flex items-center justify-between rounded-lg border p-3 transition-colors',
          status === 'success'
            ? 'border-success/30 bg-success/10'
            : displayAccount.status === 'error'
              ? 'border-error/25 bg-surface-soft'
              : 'border-hairline bg-surface-soft'
        ]"
      >
        <div class="flex items-center gap-3">
          <div
            :class="[
              'flex h-10 w-10 items-center justify-center rounded-lg',
              accountIconClass
            ]"
          >
            <Icon
              :name="accountIconName"
              size="md"
              class="text-on-primary"
              :class="status === 'connecting' && 'animate-spin'"
              :stroke-width="2"
            />
          </div>
          <div>
            <div class="font-semibold text-ink">{{ displayAccount.name }}</div>
            <div class="flex items-center gap-1.5 text-xs text-muted">
              <span
                class="rounded bg-hairline-soft px-1.5 py-0.5 text-[10px] font-medium uppercase"
              >
                {{ displayAccount.type }}
              </span>
              <span>{{ t('admin.accounts.account') }}</span>
            </div>
          </div>
        </div>
        <span
          :class="accountStatusBadgeClass"
        >
          {{ accountStatusLabel }}
        </span>
      </div>

      <div class="space-y-1.5">
        <label class="text-sm font-medium text-body">
          {{ t('admin.accounts.selectTestModel') }}
        </label>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? t('common.loading') + '...' : t('admin.accounts.selectTestModel')"
        />
      </div>

      <div v-if="isOpenAIAccount" class="space-y-1.5">
        <label class="text-sm font-medium text-body">
          {{ t('admin.accounts.openai.testMode') }}
        </label>
        <Select
          v-model="testMode"
          :options="openAITestModeOptions"
          :disabled="status === 'connecting'"
        />
      </div>

      <div v-if="supportsImageTest" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.imagePromptLabel')"
          :placeholder="t('admin.accounts.imagePromptPlaceholder')"
          :hint="t('admin.accounts.imageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>

      <!-- Terminal Output -->
      <div class="group relative">
        <div
          ref="terminalRef"
          class="max-h-[240px] min-h-[120px] overflow-y-auto rounded-lg border border-hairline-soft bg-surface-dark p-4 font-mono text-sm"
        >
          <!-- Status Line -->
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-muted">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.readyToTest') }}</span>
          </div>
          <div v-else-if="status === 'connecting'" class="flex items-center gap-2 text-warning">
            <Icon name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            <span>{{ t('admin.accounts.connectingToApi') }}</span>
          </div>

          <!-- Output Lines -->
          <div v-for="(line, index) in outputLines" :key="index" :class="line.class">
            {{ line.text }}
          </div>

          <!-- Streaming Content -->
          <div v-if="streamingContent" class="text-success">
            {{ streamingContent }}<span class="animate-pulse">_</span>
          </div>

          <!-- Result Status -->
          <div
            v-if="status === 'success'"
            class="mt-3 flex items-center gap-2 border-t border-hairline-soft pt-3 text-success"
          >
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.testCompleted') }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="mt-3 flex items-center gap-2 border-t border-hairline-soft pt-3 text-error"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Copy Button -->
        <button
          v-if="outputLines.length > 0"
          @click="copyOutput"
          class="absolute right-2 top-2 rounded-lg bg-surface-dark-elevated/80 p-1.5 text-muted-soft opacity-0 transition-all hover:bg-surface-dark-soft hover:text-on-dark group-hover:opacity-100"
          :title="t('admin.accounts.copyOutput')"
        >
          <Icon name="link" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div
        v-if="metrics"
        class="grid grid-cols-2 gap-2 sm:grid-cols-4"
      >
        <div
          v-for="item in primaryMetricItems"
          :key="item.key"
          class="rounded-lg border border-hairline bg-surface-soft p-3"
        >
          <div class="mb-1 flex items-center gap-1.5 text-xs font-medium text-muted">
            <Icon :name="item.icon" size="xs" :stroke-width="2" />
            <span>{{ item.label }}</span>
          </div>
          <div class="text-base font-semibold text-ink">{{ item.value }}</div>
        </div>
      </div>

      <div
        v-if="metrics && tokenMetricItems.length > 0"
        class="grid grid-cols-2 gap-2 sm:grid-cols-3"
      >
        <div
          v-for="item in tokenMetricItems"
          :key="item.key"
          class="rounded-lg border border-hairline bg-canvas px-3 py-2"
        >
          <div class="text-xs font-medium text-muted">{{ item.label }}</div>
          <div class="mt-0.5 text-sm font-semibold text-body-strong">{{ item.value }}</div>
        </div>
      </div>

      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="text-xs font-medium text-body">
          {{ t('admin.accounts.imagePreview') }}
        </div>
        <div class="flex flex-wrap justify-center gap-3">
          <div
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            class="group/img relative cursor-pointer overflow-hidden rounded-lg border border-hairline bg-canvas transition hover:border-primary-300"
            @click="previewImageUrl = image.url"
          >
            <img :src="image.url" :alt="`test-image-${index + 1}`" class="max-h-[360px] w-full object-contain" />
            <div class="absolute inset-0 flex items-center justify-center bg-ink/0 transition-colors group-hover/img:bg-ink/20">
              <Icon name="eye" size="lg" class="text-on-primary opacity-0 transition-opacity group-hover/img:opacity-100" :stroke-width="2" />
            </div>
            <div class="border-t border-hairline-soft px-3 py-1.5 text-xs text-muted">
              {{ image.mimeType || 'image/*' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Image Lightbox -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="previewImageUrl"
            class="fixed inset-0 z-[100] flex items-center justify-center bg-ink/80 p-4"
            @click.self="previewImageUrl = ''"
          >
            <button
              class="absolute right-4 top-4 rounded-full bg-ink/50 p-2 text-on-dark transition-colors hover:bg-ink/70"
              @click="previewImageUrl = ''"
            >
              <Icon name="x" size="lg" :stroke-width="2" />
            </button>
            <img
              :src="previewImageUrl"
              alt="preview"
              class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-2xl"
            />
          </div>
        </Transition>
      </Teleport>

      <!-- Test Info -->
      <div class="flex items-center justify-between px-1 text-xs text-muted">
        <div class="flex items-center gap-3">
          <span class="flex items-center gap-1">
            <Icon name="grid" size="sm" :stroke-width="2" />
            {{ t('admin.accounts.testModel') }}
          </span>
        </div>
        <span class="flex items-center gap-1">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{
            supportsImageTest
              ? t('admin.accounts.imageTestMode')
              : t('admin.accounts.testPrompt')
          }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleClose"
          class="rounded-lg bg-surface-card px-4 py-2 text-sm font-medium text-body transition-colors hover:bg-hairline-soft"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="startTest"
          :disabled="status === 'connecting' || loadingModels || !selectedModelId"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
            status === 'connecting' || loadingModels || !selectedModelId
              ? 'cursor-not-allowed bg-hairline text-muted'
              : status === 'success'
                ? 'bg-success/100 text-on-primary hover:bg-success/80'
                : status === 'error'
                  ? 'bg-accent-amber text-on-primary hover:bg-accent-amber/80'
                  : 'bg-primary-500 text-on-primary hover:bg-primary-600'
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('admin.accounts.testing')
                : status === 'idle'
                  ? t('admin.accounts.startTest')
                  : t('admin.accounts.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

interface OutputLine {
  text: string
  class: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

interface TestMetrics {
  duration_ms?: number
  first_token_ms?: number
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
  cache_creation_tokens?: number
  cache_read_tokens?: number
  image_output_tokens?: number
  output_chars?: number
  image_count?: number
}

type TestEvent = TestMetrics & {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}

type MetricItem = {
  key: string
  label: string
  value: string
  icon: 'clock' | 'bolt' | 'chart' | 'chat'
}

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'tested', account: Account): void
}>()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
let abortController: AbortController | null = null
const generatedImages = ref<PreviewImage[]>([])
const refreshedAccount = ref<Account | null>(null)
const metrics = ref<TestMetrics | null>(null)
const testMode = ref<'default' | 'compact'>('default')
const previewImageUrl = ref('')
const prioritizedGeminiModels = ['gemini-3.1-flash-image', 'gemini-2.5-flash-image', 'gemini-3.5-flash', 'gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-3-flash-preview', 'gemini-3-pro-preview', 'gemini-2.0-flash']
const isOpenAIAccount = computed(() => props.account?.platform === 'openai')
const openAITestModeOptions = computed<SelectOption[]>(() => [
  { value: 'default', label: t('admin.accounts.openai.testModeDefault') },
  { value: 'compact', label: t('admin.accounts.openai.testModeCompact') }
])
const supportsGeminiImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gemini-') || !modelID.includes('-image')) return false

  return props.account?.platform === 'gemini' || (props.account?.platform === 'antigravity' && props.account?.type === 'apikey')
})

const supportsOpenAIImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gpt-image-')) return false
  return props.account?.platform === 'openai'
})

const supportsImageTest = computed(() => supportsGeminiImageTest.value || supportsOpenAIImageTest.value)
const displayAccount = computed(() => refreshedAccount.value || props.account)
const testedAsHealthy = computed(() => status.value === 'success')
const accountIconName = computed(() => {
  if (status.value === 'connecting') return 'refresh'
  if (status.value === 'success') return 'check'
  if (displayAccount.value?.status === 'error' || status.value === 'error') return 'exclamationTriangle'
  return 'play'
})
const accountIconClass = computed(() => {
  if (status.value === 'success') return 'bg-success'
  if (displayAccount.value?.status === 'error' || status.value === 'error') return 'bg-error'
  return 'bg-primary-500'
})
const accountStatusLabel = computed(() => {
  if (testedAsHealthy.value) return t('admin.accounts.testPassedThisRun')
  const accountStatus = displayAccount.value?.status || 'inactive'
  return t(`admin.accounts.status.${accountStatus}`)
})
const accountStatusBadgeClass = computed(() => {
  const base = 'rounded-full px-2.5 py-1 text-xs font-semibold'
  if (testedAsHealthy.value) return `${base} bg-success/15 text-success`
  if (displayAccount.value?.status === 'active') return `${base} bg-success/15 text-success`
  if (displayAccount.value?.status === 'error') return `${base} bg-error/15 text-error`
  return `${base} bg-surface-card text-body`
})
const hasUsageTokens = computed(() => {
  const data = metrics.value
  if (!data) return false
  return Number(data.input_tokens || 0) > 0 ||
    Number(data.output_tokens || 0) > 0 ||
    Number(data.total_tokens || 0) > 0 ||
    Number(data.cache_creation_tokens || 0) > 0 ||
    Number(data.cache_read_tokens || 0) > 0 ||
    Number(data.image_output_tokens || 0) > 0
})
const primaryMetricItems = computed<MetricItem[]>(() => {
  const data = metrics.value
  if (!data) return []
  const durationMs = Number(data.duration_ms || 0)
  const outputTokens = Number(data.output_tokens || 0)
  const outputChars = Number(data.output_chars || 0)
  const images = Number(data.image_count || 0)
  const throughputValue = outputTokens > 0
    ? formatPerSecond(outputTokens, durationMs, t('admin.accounts.testMetrics.tokensPerSecondUnit'))
    : formatPerSecond(outputChars, durationMs, t('admin.accounts.testMetrics.charsPerSecondUnit'))

  return [
    {
      key: 'duration',
      label: t('admin.accounts.testMetrics.duration'),
      value: formatDurationMs(durationMs),
      icon: 'clock'
    },
    {
      key: 'first-token',
      label: t('admin.accounts.testMetrics.firstToken'),
      value: data.first_token_ms != null ? formatDurationMs(Number(data.first_token_ms)) : '-',
      icon: 'bolt'
    },
    {
      key: 'throughput',
      label: outputTokens > 0
        ? t('admin.accounts.testMetrics.tokenThroughput')
        : t('admin.accounts.testMetrics.charThroughput'),
      value: throughputValue,
      icon: 'chart'
    },
    {
      key: 'output',
      label: images > 0 ? t('admin.accounts.testMetrics.images') : t('admin.accounts.testMetrics.output'),
      value: images > 0 ? String(images) : formatNumber(outputTokens || outputChars),
      icon: 'chat'
    }
  ]
})
const tokenMetricItems = computed(() => {
  const data = metrics.value
  if (!data || !hasUsageTokens.value) return []
  return [
    { key: 'input', label: t('admin.accounts.testMetrics.inputTokens'), value: formatNumber(data.input_tokens || 0) },
    { key: 'output', label: t('admin.accounts.testMetrics.outputTokens'), value: formatNumber(data.output_tokens || 0) },
    { key: 'total', label: t('admin.accounts.testMetrics.totalTokens'), value: formatNumber(data.total_tokens || 0) },
    { key: 'cache-create', label: t('admin.accounts.testMetrics.cacheCreationTokens'), value: formatNumber(data.cache_creation_tokens || 0) },
    { key: 'cache-read', label: t('admin.accounts.testMetrics.cacheReadTokens'), value: formatNumber(data.cache_read_tokens || 0) },
    { key: 'image-tokens', label: t('admin.accounts.testMetrics.imageTokens'), value: formatNumber(data.image_output_tokens || 0) }
  ].filter(item => item.value !== '0' || item.key === 'total')
})

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((a, b) => {
    const aPriority = priorityMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
    const bPriority = priorityMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
    if (aPriority !== bPriority) return aPriority - bPriority
    return 0
  })
}

const formatNumber = (value: number) => {
  return new Intl.NumberFormat().format(Math.max(0, Math.round(value)))
}

const formatDurationMs = (value: number) => {
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1000) return `${Math.round(value)}ms`
  return `${(value / 1000).toFixed(value < 10000 ? 2 : 1)}s`
}

const formatPerSecond = (count: number, durationMs: number, unit: string) => {
  if (!Number.isFinite(count) || count <= 0 || !Number.isFinite(durationMs) || durationMs <= 0) {
    return '-'
  }
  const perSecond = count / (durationMs / 1000)
  return `${perSecond.toFixed(perSecond >= 10 ? 1 : 2)} ${unit}`
}

// Load available models when modal opens
watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      testPrompt.value = ''
      testMode.value = 'default'
      refreshedAccount.value = null
      resetState()
      await loadAvailableModels()
    } else {
      abortStream()
    }
  }
)

watch(selectedModelId, () => {
  if (supportsImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('admin.accounts.imagePromptDefault')
  }
})

const loadAvailableModels = async () => {
  if (!props.account) return

  loadingModels.value = true
  selectedModelId.value = '' // Reset selection before loading
  try {
    const models = await adminAPI.accounts.getAvailableModels(props.account.id)
    availableModels.value = props.account.platform === 'gemini' || props.account.platform === 'antigravity'
      ? sortTestModels(models)
      : models
    // Default selection by platform
    if (availableModels.value.length > 0) {
      if (props.account.platform === 'gemini') {
        selectedModelId.value = availableModels.value[0].id
      } else {
        // Try to select Sonnet as default, otherwise use first model
        const sonnetModel = availableModels.value.find((m) => m.id.includes('sonnet'))
        selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
      }
    }
  } catch (error) {
    console.error('Failed to load available models:', error)
    // Fallback to empty list
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  metrics.value = null
  previewImageUrl.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

const addLine = (text: string, className: string = 'text-muted-soft') => {
  outputLines.value.push({ text, class: className })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

const startTest = async () => {
  if (!props.account || !selectedModelId.value) return

  resetState()
  status.value = 'connecting'
  addLine(t('admin.accounts.startingTestForAccount', { name: props.account.name }), 'text-primary-600')
  addLine(t('admin.accounts.testAccountTypeLabel', { type: props.account.type }), 'text-muted-soft')
  addLine('', 'text-muted-soft')

  abortStream()

  abortController = new AbortController()

  try {
    // Create EventSource for SSE
    const url = `/api/v1/admin/accounts/${props.account.id}/test`

    // Use fetch with streaming for SSE since EventSource doesn't support POST
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model_id: selectedModelId.value,
        prompt: supportsImageTest.value ? testPrompt.value.trim() : '',
        mode: isOpenAIAccount.value ? testMode.value : 'default'
      }),
      signal: abortController.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const jsonStr = line.slice(6).trim()
          if (jsonStr) {
            try {
              const event = JSON.parse(jsonStr)
              handleEvent(event)
            } catch (e) {
              console.error('Failed to parse SSE event:', e)
            }
          }
        }
      }
    }
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      status.value = 'idle'
      return
    }
    status.value = 'error'
    const msg = error instanceof Error ? error.message : 'Unknown error'
    errorMessage.value = msg
    addLine(`Error: ${msg}`, 'text-error')
  }
}

const refreshTestedAccount = async () => {
  if (!props.account) return
  try {
    const latest = await adminAPI.accounts.getById(props.account.id)
    refreshedAccount.value = latest
    emit('tested', latest)
  } catch (error) {
    console.error('Failed to refresh tested account:', error)
  }
}

const extractMetrics = (event: TestEvent): TestMetrics => ({
  duration_ms: event.duration_ms,
  first_token_ms: event.first_token_ms,
  input_tokens: event.input_tokens,
  output_tokens: event.output_tokens,
  total_tokens: event.total_tokens,
  cache_creation_tokens: event.cache_creation_tokens,
  cache_read_tokens: event.cache_read_tokens,
  image_output_tokens: event.image_output_tokens,
  output_chars: event.output_chars,
  image_count: event.image_count
})

const handleEvent = (event: TestEvent) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'text-success')
      if (event.model) {
        addLine(t('admin.accounts.usingModel', { model: event.model }), 'text-accent-teal')
      }
      addLine(
        supportsImageTest.value
            ? t('admin.accounts.sendingImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'text-muted-soft'
      )
      addLine('', 'text-muted-soft')
      addLine(t('admin.accounts.response'), 'text-warning')
      break

    case 'content':
      if (event.text) {
        streamingContent.value += event.text
        scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.imageReceived', { count: generatedImages.value.length }), 'text-accent-amber')
      }
      break

    case 'test_complete':
      // Move streaming content to output lines
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-success')
        streamingContent.value = ''
      }
      metrics.value = extractMetrics(event)
      if (event.success) {
        status.value = 'success'
        refreshTestedAccount()
      } else {
        status.value = 'error'
        errorMessage.value = event.error || 'Test failed'
      }
      break

    case 'error':
      status.value = 'error'
      errorMessage.value = event.error || 'Unknown error'
      if (streamingContent.value) {
        addLine(streamingContent.value, 'text-success')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const text = outputLines.value.map((l) => l.text).join('\n')
  copyToClipboard(text, t('admin.accounts.outputCopied'))
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
