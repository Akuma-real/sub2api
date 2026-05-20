<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div
          class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"
        ></div>
      </div>

      <template v-else>
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
        >
          <div>
            <h1 class="text-2xl font-semibold text-ink">
              {{ t("admin.riskControl.title") }}
            </h1>
            <p class="mt-1 text-sm text-muted">
              {{ t("admin.riskControl.description") }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary inline-flex items-center gap-2"
              :disabled="statusLoading"
              @click="loadStatus(false)"
            >
              <Icon
                name="refresh"
                size="sm"
                :class="statusLoading ? 'animate-spin' : ''"
              />
              {{ t("admin.riskControl.refreshStatus") }}
            </button>
            <button
              type="button"
              class="btn btn-primary inline-flex items-center gap-2"
              @click="openSettings"
            >
              <Icon name="cog" size="sm" />
              {{ t("admin.riskControl.openSettings") }}
            </button>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="item in overviewItems"
            :key="item.key"
            class="rounded-lg border border-hairline-soft bg-canvas px-4 py-3 shadow-sm"
          >
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg"
                :class="item.iconClass"
              >
                <Icon :name="item.icon" size="sm" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex min-w-0 items-center justify-between gap-2">
                  <p class="truncate text-xs font-medium text-muted">
                    {{ item.label }}
                  </p>
                  <span
                    v-if="item.badge"
                    class="inline-flex flex-shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="item.badgeClass"
                  >
                    {{ item.badge }}
                  </span>
                </div>
                <div class="mt-1 flex min-w-0 items-baseline gap-2">
                  <p class="truncate text-xl font-semibold leading-7 text-ink">
                    {{ item.value }}
                  </p>
                  <p v-if="item.meta" class="truncate text-xs text-muted">
                    {{ item.meta }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div
            class="flex flex-col gap-4 border-b border-hairline-soft px-6 py-4 lg:flex-row lg:items-center lg:justify-between"
          >
            <div>
              <h2 class="text-lg font-semibold text-ink">
                {{ t("admin.riskControl.workerStatus") }}
              </h2>
              <p class="mt-1 text-sm text-muted">
                {{ t("admin.riskControl.workerStatusHint") }}
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-2 text-sm text-muted">
              <span>{{ t("admin.riskControl.autoRefresh") }}</span>
              <span v-if="status?.last_cleanup_at">
                {{
                  t("admin.riskControl.lastCleanup", {
                    time: formatDateTime(status.last_cleanup_at),
                  })
                }}
              </span>
            </div>
          </div>

          <div
            class="grid grid-cols-1 gap-6 p-6 xl:grid-cols-[minmax(0,360px)_1fr]"
          >
            <div class="space-y-4">
              <div class="rounded-lg border border-hairline-soft p-4">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <p class="text-sm font-medium text-ink">
                      {{ t("admin.riskControl.queueUsage") }}
                    </p>
                    <p class="mt-1 text-xs text-muted">
                      {{ formatNumber(status?.queue_length ?? 0) }} /
                      {{
                        formatNumber(
                          status?.queue_size ?? configForm.queue_size,
                        )
                      }}
                    </p>
                  </div>
                  <span class="text-sm font-semibold text-ink">{{
                    queueUsagePercent
                  }}</span>
                </div>
                <div
                  class="mt-4 h-2 overflow-hidden rounded-full bg-surface-card"
                >
                  <div
                    class="h-full rounded-full bg-primary-500 transition-all duration-300"
                    :style="queueUsageStyle"
                  ></div>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-lg bg-surface-soft p-4">
                  <p class="text-xs text-muted">
                    {{ t("admin.riskControl.activeWorkers") }}
                  </p>
                  <p class="mt-2 text-2xl font-semibold text-ink">
                    {{ status?.active_workers ?? 0 }}
                  </p>
                </div>
                <div class="rounded-lg bg-success/15 p-4">
                  <p class="text-xs text-muted">
                    {{ t("admin.riskControl.idleWorkers") }}
                  </p>
                  <p class="mt-2 text-2xl font-semibold text-success">
                    {{ status?.idle_workers ?? configForm.worker_count }}
                  </p>
                </div>
                <div class="rounded-lg bg-surface-soft p-4">
                  <p class="text-xs text-muted">
                    {{ t("admin.riskControl.processed") }}
                  </p>
                  <p class="mt-2 text-2xl font-semibold text-ink">
                    {{ formatNumber(status?.processed ?? 0) }}
                  </p>
                </div>
                <div class="rounded-lg bg-surface-soft p-4">
                  <p class="text-xs text-muted">
                    {{ t("admin.riskControl.droppedErrors") }}
                  </p>
                  <p class="mt-2 text-2xl font-semibold text-ink">
                    {{
                      formatNumber(
                        (status?.dropped ?? 0) + (status?.errors ?? 0),
                      )
                    }}
                  </p>
                </div>
              </div>
            </div>

            <div>
              <div class="mb-3 flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-ink">
                    {{ t("admin.riskControl.workerPool") }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    {{
                      t("admin.riskControl.workerPoolMeta", {
                        active: status?.active_workers ?? 0,
                        idle: status?.idle_workers ?? configForm.worker_count,
                        total: status?.worker_count ?? configForm.worker_count,
                      })
                    }}
                  </p>
                </div>
                <span
                  class="inline-flex items-center rounded-full bg-surface-card px-2.5 py-1 text-xs font-medium text-body"
                >
                  {{ modeLabel(status?.mode ?? configForm.mode) }}
                </span>
              </div>
              <div
                class="grid grid-cols-2 gap-2 sm:grid-cols-4 md:grid-cols-6 xl:grid-cols-8 2xl:grid-cols-10"
              >
                <div
                  v-for="worker in workerSlots"
                  :key="worker.id"
                  class="flex h-12 items-center justify-between rounded-lg border px-3 transition-colors"
                  :class="workerSlotClass(worker.state)"
                  :title="worker.label"
                >
                  <span class="text-sm font-semibold">#{{ worker.id }}</span>
                  <span
                    class="h-2.5 w-2.5 rounded-full"
                    :class="workerDotClass(worker.state)"
                  ></span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div
            class="flex flex-col gap-4 border-b border-hairline-soft px-6 py-4"
          >
            <div
              class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between"
            >
              <div>
                <h2 class="text-lg font-semibold text-ink">
                  {{ t("admin.riskControl.records") }}
                </h2>
                <p class="mt-1 text-sm text-muted">
                  {{ t("admin.riskControl.recordsHint") }}
                </p>
              </div>
              <button
                type="button"
                class="btn btn-secondary inline-flex items-center gap-2"
                :disabled="logsLoading"
                @click="loadLogs"
              >
                <Icon
                  name="refresh"
                  size="sm"
                  :class="logsLoading ? 'animate-spin' : ''"
                />
                {{ t("admin.riskControl.refresh") }}
              </button>
            </div>

            <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-6">
              <Select
                v-model="filters.result"
                :options="resultOptions"
                @change="reloadLogsFromFirstPage"
              />
              <Select
                v-model="filters.group_id"
                :options="groupFilterOptions"
                @change="reloadLogsFromFirstPage"
              />
              <Select
                v-model="filters.endpoint"
                :options="endpointOptions"
                @change="reloadLogsFromFirstPage"
              />
              <input
                v-model.trim="filters.search"
                type="search"
                class="input"
                :placeholder="t('admin.riskControl.filters.search')"
                @keyup.enter="reloadLogsFromFirstPage"
              />
              <input
                v-model="filters.from"
                type="datetime-local"
                class="input"
                :title="t('admin.riskControl.filters.from ')"
                @change="reloadLogsFromFirstPage"
              />
              <input
                v-model="filters.to"
                type="datetime-local"
                class="input"
                :title="t('admin.riskControl.filters.to')"
                @change="reloadLogsFromFirstPage"
              />
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-hairline">
              <thead class="bg-surface-soft">
                <tr>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.time") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.group") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.user") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.apiKey") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.endpoint") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.result") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.highest") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.actionMeta") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.latency") }}
                  </th>
                  <th
                    class="px-5 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted"
                  >
                    {{ t("admin.riskControl.table.input") }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-hairline-soft bg-canvas">
                <tr v-if="logsLoading">
                  <td
                    colspan="10"
                    class="px-5 py-12 text-center text-sm text-muted"
                  >
                    {{ t("common.loading") }}
                  </td>
                </tr>
                <tr v-else-if="logs.length === 0">
                  <td
                    colspan="10"
                    class="px-5 py-12 text-center text-sm text-muted"
                  >
                    {{ t("admin.riskControl.emptyLogs") }}
                  </td>
                </tr>
                <template v-else>
                  <tr
                    v-for="row in logs"
                    :key="row.id"
                    class="hover:bg-surface-soft"
                  >
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      {{ formatDateTime(row.created_at) }}
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      {{ row.group_name || "-" }}
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      <div>{{ row.user_email || "-" }}</div>
                      <div v-if="row.user_id" class="text-xs text-muted-soft">
                        UID {{ row.user_id }}
                      </div>
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      {{ row.api_key_name || "-" }}
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      <div>{{ row.endpoint || "-" }}</div>
                      <div class="text-xs text-muted-soft">
                        {{ row.provider || "-" }} / {{ row.model || "-" }}
                      </div>
                    </td>
                    <td class="whitespace-nowrap px-5 py-4">
                      <span
                        class="inline-flex rounded-md px-2 py-1 text-xs font-medium"
                        :class="resultBadgeClass(row)"
                      >
                        {{ resultLabel(row) }}
                      </span>
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      <div>{{ row.highest_category || "-" }}</div>
                      <div class="text-xs text-muted-soft">
                        {{ percent(row.highest_score) }}
                      </div>
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      <div>{{ violationCountText(row) }}</div>
                      <div class="text-xs text-muted-soft">
                        {{
                          row.email_sent
                            ? t("admin.riskControl.emailSent")
                            : t("admin.riskControl.emailNotSent")
                        }}
                        <span v-if="row.auto_banned">
                          / {{ t("admin.riskControl.autoBanned") }}</span
                        >
                      </div>
                      <button
                        v-if="canUnbanRow(row)"
                        type="button"
                        class="mt-2 inline-flex items-center gap-1 rounded-md border border-success/25 bg-success/15 px-2 py-1 text-xs font-medium text-success transition-colors hover:bg-success/15 disabled:cursor-not-allowed disabled:opacity-60"
                        :disabled="unbanningUserID === row.user_id"
                        @click="unbanUser(row)"
                      >
                        <Icon
                          name="checkCircle"
                          size="xs"
                          :class="
                            unbanningUserID === row.user_id
                              ? 'animate-spin'
                              : ''
                          "
                        />
                        {{
                          unbanningUserID === row.user_id
                            ? t("common.processing")
                            : t("admin.riskControl.unbanUser")
                        }}
                      </button>
                    </td>
                    <td class="whitespace-nowrap px-5 py-4 text-sm text-body">
                      <div>{{ latencyText(row.upstream_latency_ms) }}</div>
                      <div
                        v-if="
                          row.queue_delay_ms !== null &&
                          row.queue_delay_ms !== undefined
                        "
                        class="text-xs text-muted-soft"
                      >
                        {{
                          t("admin.riskControl.queueDelay", {
                            ms: row.queue_delay_ms,
                          })
                        }}
                      </div>
                    </td>
                    <td class="w-[320px] max-w-sm px-5 py-4 text-sm text-body">
                      <button
                        type="button"
                        class="group flex w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left transition-colors hover:bg-surface-card"
                        :title="inputSummaryText(row)"
                        @click="openInputDetail(row)"
                      >
                        <span class="min-w-0 flex-1 truncate">{{
                          inputSummaryText(row)
                        }}</span>
                        <Icon
                          name="eye"
                          size="xs"
                          class="flex-shrink-0 text-muted-soft transition-colors group-hover:text-primary-500"
                        />
                      </button>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>

          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="onPageChange"
            @update:pageSize="onPageSizeChange"
          />
        </div>
      </template>

      <BaseDialog
        :show="settingsOpen"
        :title="t('admin.riskControl.settingsTitle')"
        width="extra-wide"
        @close="settingsOpen = false"
      >
        <div class="space-y-6">
          <div
            class="flex gap-2 overflow-x-auto border-b border-hairline-soft pb-3"
          >
            <button
              v-for="tab in settingsTabs"
              :key="tab.id"
              type="button"
              class="inline-flex whitespace-nowrap rounded-lg px-3 py-2 text-sm font-medium transition-colors"
              :class="
                activeSettingsTab === tab.id
                  ? 'bg-primary-50 text-primary-700 '
                  : 'text-muted hover:bg-surface-soft hover:text-ink '
              "
              @click="activeSettingsTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>

          <div v-if="activeSettingsTab === 'basic'" class="space-y-5">
            <div class="grid grid-cols-1 gap-5 lg:grid-cols-2">
              <div
                class="flex items-center justify-between rounded-lg border border-hairline-soft p-4"
              >
                <div>
                  <p class="text-sm font-medium text-ink">
                    {{ t("admin.riskControl.enabled") }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    {{ t("admin.riskControl.enabledHint") }}
                  </p>
                </div>
                <Toggle v-model="configForm.enabled" />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.mode")
                }}</label>
                <Select v-model="configForm.mode" :options="modeOptions" />
                <p class="mt-2 text-xs leading-5 text-muted">
                  {{ modeDescription(configForm.mode) }}
                </p>
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.baseUrl")
                }}</label>
                <input
                  v-model.trim="configForm.base_url"
                  type="url"
                  class="input"
                  placeholder="https://api.openai.com"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.model")
                }}</label>
                <input
                  v-model.trim="configForm.model"
                  type="text"
                  class="input"
                  placeholder="omni-moderation-latest"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.timeoutMs")
                }}</label>
                <input
                  v-model.number="configForm.timeout_ms"
                  type="number"
                  min="500"
                  max="30000"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.retryCount")
                }}</label>
                <input
                  v-model.number="configForm.retry_count"
                  type="number"
                  min="0"
                  max="5"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.sampleRate")
                }}</label>
                <div class="relative">
                  <input
                    v-model.number="configForm.sample_rate"
                    type="number"
                    min="0"
                    max="100"
                    step="1"
                    class="input pr-8"
                  />
                  <span
                    class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted-soft"
                    >%</span
                  >
                </div>
              </div>
            </div>

            <div
              class="overflow-hidden rounded-xl border border-hairline-soft bg-canvas shadow-sm"
            >
              <div
                class="flex flex-col gap-4 border-b border-hairline-soft bg-surface-soft px-4 py-4 lg:flex-row lg:items-center lg:justify-between"
              >
                <div class="flex items-start gap-3">
                  <span
                    class="mt-0.5 flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600"
                  >
                    <Icon name="key" size="md" />
                  </span>
                  <div>
                    <label class="text-sm font-semibold text-ink">{{
                      t("admin.riskControl.apiKeys")
                    }}</label>
                    <p class="mt-1 max-w-3xl text-xs leading-5 text-muted">
                      {{
                        t("admin.riskControl.apiKeysHint", {
                          count: configForm.api_key_count,
                        })
                      }}
                    </p>
                  </div>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    class="btn btn-secondary inline-flex items-center gap-2"
                    :disabled="
                      apiKeyTesting ||
                      inputApiKeyCount === 0 ||
                      configForm.clear_api_key
                    "
                    @click="testApiKeys(true)"
                  >
                    <Icon
                      name="beaker"
                      size="sm"
                      :class="apiKeyTesting ? 'animate-pulse' : ''"
                    />
                    {{
                      apiKeyTesting
                        ? t("admin.riskControl.testingApiKeys")
                        : t("admin.riskControl.testInputApiKeys")
                    }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary inline-flex items-center gap-2"
                    :disabled="
                      apiKeyTesting ||
                      effectiveStoredApiKeyCount === 0 ||
                      pendingDeletedApiKeyCount > 0 ||
                      configForm.clear_api_key ||
                      configForm.api_keys_mode === 'replace'
                    "
                    @click="testApiKeys(false)"
                  >
                    <Icon name="shield" size="sm" />
                    {{ storedApiKeyTestButtonText }}
                  </button>
                  <button
                    v-if="configForm.api_key_configured"
                    type="button"
                    class="btn btn-secondary inline-flex items-center gap-2"
                    @click="toggleClearApiKey"
                  >
                    <Icon
                      :name="configForm.clear_api_key ? 'x' : 'trash'"
                      size="sm"
                    />
                    {{
                      configForm.clear_api_key
                        ? t("admin.riskControl.keepApiKey")
                        : t("admin.riskControl.clearApiKey")
                    }}
                  </button>
                </div>
              </div>

              <div
                class="grid grid-cols-1 gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_minmax(360px,440px)]"
              >
                <div class="space-y-3">
                  <div
                    class="flex flex-col gap-2 rounded-lg border border-hairline-soft bg-surface-soft p-2 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div class="text-xs leading-5 text-muted">
                      <span class="font-medium text-body">{{
                        t("admin.riskControl.apiKeysWriteMode")
                      }}</span>
                      <span class="ml-2">{{ apiKeysModeHint }}</span>
                    </div>
                    <div class="inline-flex rounded-lg bg-canvas p-1 shadow-sm">
                      <button
                        type="button"
                        class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors"
                        :class="
                          configForm.api_keys_mode === 'append'
                            ? 'bg-primary-500 text-on-primary shadow-sm'
                            : 'text-body hover:bg-surface-card '
                        "
                        :disabled="configForm.clear_api_key"
                        @click="setAPIKeysMode('append')"
                      >
                        {{ t("admin.riskControl.apiKeysModeAppend") }}
                      </button>
                      <button
                        type="button"
                        class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors"
                        :class="
                          configForm.api_keys_mode === 'replace'
                            ? 'bg-accent-amber text-on-primary shadow-sm'
                            : 'text-body hover:bg-surface-card '
                        "
                        :disabled="configForm.clear_api_key"
                        @click="setAPIKeysMode('replace')"
                      >
                        {{ t("admin.riskControl.apiKeysModeReplace") }}
                      </button>
                    </div>
                  </div>
                  <textarea
                    v-model="configForm.api_keys_text"
                    class="input min-h-44 resize-y font-mono text-sm"
                    :placeholder="apiKeysPlaceholder"
                    autocomplete="new-password"
                    :disabled="configForm.clear_api_key"
                  ></textarea>
                  <div
                    class="flex flex-wrap items-center gap-2 text-xs text-muted"
                  >
                    <span
                      class="inline-flex rounded-md bg-surface-card px-2 py-1"
                    >
                      {{
                        t("admin.riskControl.inputApiKeyCount", {
                          count: inputApiKeyCount,
                        })
                      }}
                    </span>
                    <span
                      v-if="configForm.api_key_configured"
                      class="inline-flex rounded-md bg-surface-card px-2 py-1"
                    >
                      {{
                        t("admin.riskControl.storedApiKeyCount", {
                          count: configForm.api_key_count,
                        })
                      }}
                    </span>
                    <span
                      v-if="configForm.clear_api_key"
                      class="inline-flex rounded-md bg-error/15 px-2 py-1 text-error"
                    >
                      {{ t("admin.riskControl.apiKeyWillClear") }}
                    </span>
                    <span
                      v-else-if="pendingDeletedApiKeyCount > 0"
                      class="inline-flex rounded-md bg-accent-amber/15 px-2 py-1 text-warning"
                    >
                      {{
                        t("admin.riskControl.apiKeyPendingDeleteCount", {
                          count: pendingDeletedApiKeyCount,
                        })
                      }}
                    </span>
                    <span
                      v-if="configForm.api_keys_mode === 'replace'"
                      class="inline-flex rounded-md bg-accent-amber/15 px-2 py-1 text-warning"
                    >
                      {{ t("admin.riskControl.apiKeysReplaceWarning") }}
                    </span>
                  </div>

                  <div
                    class="rounded-lg border border-hairline-soft bg-surface-soft p-3"
                    @paste="handleModerationImagePaste"
                  >
                    <div class="mb-3 flex items-center justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-ink">
                          {{ t("admin.riskControl.auditTestInput") }}
                        </p>
                        <p class="mt-1 text-xs text-muted">
                          {{ t("admin.riskControl.auditTestInputHint") }}
                        </p>
                      </div>
                      <button
                        v-if="
                          moderationTestPrompt ||
                          moderationTestImages.length > 0 ||
                          moderationTestResult
                        "
                        type="button"
                        class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-muted hover:bg-canvas hover:text-ink"
                        @click="clearModerationTestInput"
                      >
                        <Icon name="x" size="xs" />
                        {{ t("admin.riskControl.clearAuditTest") }}
                      </button>
                    </div>
                    <textarea
                      v-model="moderationTestPrompt"
                      class="input min-h-24 resize-y text-sm"
                      :placeholder="
                        t('admin.riskControl.auditTestPromptPlaceholder')
                      "
                    ></textarea>
                    <div
                      class="mt-3 rounded-lg border border-dashed border-hairline bg-canvas p-3"
                      @dragover.prevent
                      @drop.prevent="handleModerationImageDrop"
                    >
                      <div
                        class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
                      >
                        <div class="flex items-start gap-2">
                          <Icon
                            name="upload"
                            size="md"
                            class="mt-0.5 text-muted-soft"
                          />
                          <div>
                            <p class="text-sm font-medium text-body-strong">
                              {{ t("admin.riskControl.auditTestImages") }}
                            </p>
                            <p class="mt-1 text-xs text-muted">
                              {{ t("admin.riskControl.auditTestImagesHint") }}
                            </p>
                          </div>
                        </div>
                        <label
                          class="btn btn-secondary inline-flex cursor-pointer items-center gap-2"
                        >
                          <Icon name="plus" size="sm" />
                          {{ t("admin.riskControl.addAuditTestImage") }}
                          <input
                            type="file"
                            accept="image/*"
                            multiple
                            class="sr-only"
                            @change="handleModerationImageUpload"
                          />
                        </label>
                      </div>
                      <div
                        v-if="moderationTestImages.length > 0"
                        class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4"
                      >
                        <div
                          v-for="(image, index) in moderationTestImages"
                          :key="image.slice(0, 64) + index"
                          class="group relative aspect-square overflow-hidden rounded-lg border border-hairline-soft bg-surface-card"
                        >
                          <img
                            :src="image"
                            alt=""
                            class="h-full w-full object-cover"
                          />
                          <button
                            type="button"
                            class="absolute right-1.5 top-1.5 flex h-7 w-7 items-center justify-center rounded-full bg-ink/60 text-on-dark opacity-0 transition-opacity group-hover:opacity-100"
                            @click="removeModerationTestImage(index)"
                          >
                            <Icon name="x" size="xs" :stroke-width="2" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div
                  class="rounded-lg border border-hairline-soft bg-surface-soft p-3"
                >
                  <div class="mb-3 flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="text-sm font-semibold text-ink">
                        {{ t("admin.riskControl.apiKeyHealth") }}
                      </p>
                      <p class="mt-1 text-xs text-muted">
                        {{ t("admin.riskControl.apiKeyFreezeRule") }}
                      </p>
                    </div>
                    <span
                      class="inline-flex shrink-0 items-center whitespace-nowrap rounded-full bg-canvas px-2 py-0.5 text-[11px] font-medium leading-5 text-body shadow-sm"
                    >
                      {{
                        t("admin.riskControl.apiKeyRows", {
                          count: apiKeyRows.length,
                        })
                      }}
                    </span>
                  </div>

                  <div
                    v-if="apiKeyRows.length === 0"
                    class="flex min-h-32 flex-col items-center justify-center rounded-lg border border-dashed border-hairline bg-canvas px-4 py-6 text-center"
                  >
                    <Icon name="infoCircle" size="lg" class="text-muted-soft" />
                    <p class="mt-2 text-sm font-medium text-body">
                      {{ t("admin.riskControl.apiKeyHealthEmpty") }}
                    </p>
                    <p class="mt-1 text-xs text-muted">
                      {{ t("admin.riskControl.apiKeyHealthEmptyHint") }}
                    </p>
                  </div>
                  <div v-else class="space-y-2">
                    <div
                      class="space-y-2"
                      :class="
                        apiKeyRowsExpanded
                          ? 'max-h-72 overflow-y-auto pr-1'
                          : ''
                      "
                    >
                      <div
                        v-for="(row, index) in visibleApiKeyRows"
                        :key="apiKeyRowKey(row, index)"
                        class="rounded-lg border bg-canvas p-2.5 shadow-sm"
                        :class="
                          isStoredApiKeyPendingDelete(row)
                            ? 'border-accent-amber/30 opacity-70 '
                            : 'border-hairline-soft '
                        "
                      >
                        <div class="flex items-start justify-between gap-2">
                          <div class="min-w-0">
                            <div
                              class="flex min-w-0 flex-wrap items-center gap-2"
                            >
                              <span
                                class="truncate font-mono text-sm font-semibold text-ink"
                                >{{ row.masked || "-" }}</span
                              >
                              <span
                                class="inline-flex rounded-md px-1.5 py-0.5 text-[11px] font-medium"
                                :class="
                                  row.configured
                                    ? 'bg-primary-50 text-primary-700 '
                                    : 'bg-primary-50 text-primary-700 '
                                "
                              >
                                {{
                                  isStoredApiKeyPendingDelete(row)
                                    ? t("admin.riskControl.apiKeyPendingDelete")
                                    : row.configured
                                      ? t("admin.riskControl.apiKeyConfigured")
                                      : t("admin.riskControl.apiKeyTemporary")
                                }}
                              </span>
                            </div>
                            <p class="mt-1 text-xs leading-5 text-muted">
                              {{ apiKeyStatusMeta(row) }}
                            </p>
                          </div>
                          <div class="flex flex-shrink-0 items-center gap-1.5">
                            <span
                              class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium"
                              :class="apiKeyStatusBadgeClass(row.status)"
                            >
                              <span
                                class="h-1.5 w-1.5 rounded-full"
                                :class="apiKeyStatusDotClass(row.status)"
                              ></span>
                              {{ apiKeyStatusLabel(row.status) }}
                            </span>
                            <button
                              v-if="row.configured && !configForm.clear_api_key"
                              type="button"
                              class="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-soft transition-colors hover:bg-surface-card hover:text-body"
                              :title="
                                isStoredApiKeyPendingDelete(row)
                                  ? t('admin.riskControl.undoDeleteApiKey')
                                  : t('admin.riskControl.deleteApiKey')
                              "
                              @click="toggleDeleteStoredApiKey(row)"
                            >
                              <Icon
                                :name="
                                  isStoredApiKeyPendingDelete(row)
                                    ? 'refresh'
                                    : 'trash'
                                "
                                size="xs"
                              />
                            </button>
                          </div>
                        </div>
                        <p
                          v-if="row.last_error"
                          class="mt-1.5 rounded-md bg-accent-amber/15 px-2 py-1.5 text-xs leading-5 text-warning"
                        >
                          {{ row.last_error }}
                        </p>
                      </div>
                    </div>

                    <div
                      v-if="canToggleApiKeyRows"
                      class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-hairline bg-canvas px-3 py-2 text-xs text-muted"
                    >
                      <span class="min-w-0 truncate">
                        {{
                          apiKeyRowsExpanded
                            ? t("admin.riskControl.apiKeyRowsExpanded", {
                                count: apiKeyRows.length,
                              })
                            : t("admin.riskControl.apiKeyRowsCollapsed", {
                                count: hiddenApiKeyRowCount,
                              })
                        }}
                      </span>
                      <button
                        type="button"
                        class="inline-flex shrink-0 items-center gap-1 rounded-md px-2 py-1 font-medium text-primary-600 transition-colors hover:bg-primary-50 hover:text-primary-700"
                        @click="apiKeyRowsExpanded = !apiKeyRowsExpanded"
                      >
                        <Icon
                          :name="
                            apiKeyRowsExpanded ? 'chevronUp' : 'chevronDown'
                          "
                          size="xs"
                        />
                        {{
                          apiKeyRowsExpanded
                            ? t("admin.riskControl.collapseApiKeyRows")
                            : t("admin.riskControl.expandApiKeyRows")
                        }}
                      </button>
                    </div>
                  </div>

                  <div
                    v-if="moderationTestResult"
                    class="mt-4 rounded-lg border border-hairline-soft bg-canvas p-3"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="text-sm font-semibold text-ink">
                          {{ t("admin.riskControl.auditTestResult") }}
                        </p>
                        <p class="mt-1 text-xs text-muted">
                          {{
                            t("admin.riskControl.auditTestHighest", {
                              category:
                                moderationTestResult.highest_category || "-",
                              score: percent(
                                moderationTestResult.highest_score,
                              ),
                            })
                          }}
                        </p>
                      </div>
                      <span
                        class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                        :class="
                          moderationTestResult.flagged
                            ? 'bg-error/15 text-error '
                            : 'bg-success/15 text-success '
                        "
                      >
                        {{
                          moderationTestResult.flagged
                            ? t("admin.riskControl.auditTestFlagged")
                            : t("admin.riskControl.auditTestPassed")
                        }}
                      </span>
                    </div>
                    <div class="mt-3">
                      <div
                        class="mb-2 flex items-center justify-between text-xs text-muted"
                      >
                        <span>{{
                          t("admin.riskControl.auditTestComposite")
                        }}</span>
                        <span class="font-semibold text-ink">{{
                          percent(moderationTestResult.composite_score)
                        }}</span>
                      </div>
                      <div
                        class="h-2 overflow-hidden rounded-full bg-surface-card"
                      >
                        <div
                          class="h-full rounded-full"
                          :class="
                            moderationTestResult.flagged
                              ? 'bg-error'
                              : 'bg-success'
                          "
                          :style="{
                            width: percentWidth(
                              moderationTestResult.composite_score,
                            ),
                          }"
                        ></div>
                      </div>
                    </div>
                    <div class="mt-3 max-h-52 space-y-2 overflow-y-auto pr-1">
                      <div
                        v-for="score in moderationScoreRows"
                        :key="score.category"
                      >
                        <div
                          class="mb-1 flex items-center justify-between gap-3 text-xs"
                        >
                          <span class="truncate text-body">{{
                            score.category
                          }}</span>
                          <span class="font-mono text-muted"
                            >{{ percent(score.score) }} /
                            {{ percent(score.threshold) }}</span
                          >
                        </div>
                        <div
                          class="h-1.5 overflow-hidden rounded-full bg-surface-card"
                        >
                          <div
                            class="h-full rounded-full"
                            :class="score.hit ? 'bg-error' : 'bg-primary-500'"
                            :style="{ width: percentWidth(score.score) }"
                          ></div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'scope'" class="space-y-5">
            <div
              class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
            >
              <div>
                <h3 class="text-base font-semibold text-ink">
                  {{ t("admin.riskControl.groupScope") }}
                </h3>
                <p class="mt-1 text-sm text-muted">
                  {{ t("admin.riskControl.groupScopeHint") }}
                </p>
              </div>
              <div class="inline-flex rounded-lg bg-surface-card p-1">
                <button
                  type="button"
                  class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="
                    configForm.all_groups
                      ? 'bg-canvas text-ink shadow-sm '
                      : 'text-muted '
                  "
                  @click="configForm.all_groups = true"
                >
                  {{ t("admin.riskControl.allGroups") }}
                </button>
                <button
                  type="button"
                  class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="
                    !configForm.all_groups
                      ? 'bg-canvas text-ink shadow-sm '
                      : 'text-muted '
                  "
                  @click="configForm.all_groups = false"
                >
                  {{ t("admin.riskControl.selectedGroups") }}
                </button>
              </div>
            </div>

            <div v-if="!configForm.all_groups" class="space-y-4">
              <div class="relative">
                <Icon
                  name="search"
                  size="sm"
                  class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-soft"
                />
                <input
                  v-model.trim="groupSearch"
                  type="search"
                  class="input pl-9"
                  :placeholder="t('admin.riskControl.searchGroups')"
                />
              </div>
              <div
                class="grid max-h-[420px] grid-cols-1 gap-3 overflow-y-auto pr-1 md:grid-cols-2 xl:grid-cols-3"
              >
                <button
                  v-for="group in filteredGroups"
                  :key="group.id"
                  type="button"
                  class="flex min-h-20 items-center justify-between rounded-lg border p-4 text-left transition-colors"
                  :class="
                    isGroupSelected(group.id)
                      ? 'border-primary-300 bg-primary-50 '
                      : 'border-hairline-soft hover:bg-surface-soft '
                  "
                  @click="toggleGroup(group.id)"
                >
                  <span class="min-w-0">
                    <span
                      class="block truncate text-sm font-semibold text-ink"
                      >{{ group.name }}</span
                    >
                    <span
                      class="mt-1 inline-flex rounded-md bg-surface-card px-2 py-0.5 text-xs text-muted"
                      >{{ group.platform }}</span
                    >
                  </span>
                  <span
                    class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full border"
                    :class="
                      isGroupSelected(group.id)
                        ? 'border-primary-500 bg-primary-500 text-on-primary'
                        : 'border-hairline text-transparent '
                    "
                  >
                    <Icon name="check" size="xs" :stroke-width="2" />
                  </span>
                </button>
                <p
                  v-if="filteredGroups.length === 0"
                  class="text-sm text-muted"
                >
                  {{ t("admin.riskControl.noGroups") }}
                </p>
              </div>
            </div>
          </div>

          <div
            v-else-if="activeSettingsTab === 'runtime'"
            class="grid grid-cols-1 gap-5 lg:grid-cols-2"
          >
            <div>
              <label class="input-label">{{
                t("admin.riskControl.workerCount")
              }}</label>
              <input
                v-model.number="configForm.worker_count"
                type="number"
                min="1"
                max="32"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.riskControl.queueSize")
              }}</label>
              <input
                v-model.number="configForm.queue_size"
                type="number"
                min="100"
                max="100000"
                class="input"
              />
            </div>
            <div
              class="flex items-center justify-between rounded-lg border border-hairline-soft p-4 lg:col-span-2"
            >
              <div>
                <p class="text-sm font-medium text-ink">
                  {{ t("admin.riskControl.recordNonHits") }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ t("admin.riskControl.recordNonHitsHint") }}
                </p>
              </div>
              <Toggle v-model="configForm.record_non_hits" />
            </div>
            <div
              class="space-y-4 rounded-lg border border-hairline-soft p-4 lg:col-span-2"
            >
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-sm font-medium text-ink">
                    {{ t("admin.riskControl.preHashCheck") }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    {{ t("admin.riskControl.preHashCheckHint") }}
                  </p>
                </div>
                <Toggle v-model="configForm.pre_hash_check_enabled" />
              </div>
              <div class="rounded-lg bg-surface-soft p-3">
                <div
                  class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between"
                >
                  <div>
                    <p class="text-sm font-medium text-ink">
                      {{
                        t("admin.riskControl.flaggedHashCount", {
                          count: formatNumber(status?.flagged_hash_count ?? 0),
                        })
                      }}
                    </p>
                    <p class="mt-1 text-xs text-muted">
                      {{ t("admin.riskControl.flaggedHashHint") }}
                    </p>
                  </div>
                  <button
                    type="button"
                    class="btn btn-secondary inline-flex items-center justify-center gap-2 text-error hover:text-error"
                    :disabled="
                      hashActionLoading ||
                      (status?.flagged_hash_count ?? 0) === 0
                    "
                    @click="clearFlaggedHashes"
                  >
                    <Icon
                      name="trash"
                      size="sm"
                      :class="hashActionLoading ? 'animate-pulse' : ''"
                    />
                    {{ t("admin.riskControl.clearFlaggedHashes") }}
                  </button>
                </div>
                <div class="mt-3 flex flex-col gap-2 sm:flex-row">
                  <input
                    v-model.trim="flaggedHashInput"
                    type="text"
                    class="input font-mono text-sm"
                    :placeholder="t('admin.riskControl.flaggedHashPlaceholder')"
                  />
                  <button
                    type="button"
                    class="btn btn-secondary inline-flex items-center justify-center gap-2"
                    :disabled="hashActionLoading || !isFlaggedHashInputValid"
                    @click="deleteFlaggedHash"
                  >
                    <Icon name="trash" size="sm" />
                    {{ t("admin.riskControl.deleteFlaggedHash") }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'response'" class="space-y-5">
            <div class="grid grid-cols-1 gap-5 lg:grid-cols-2">
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.blockStatus")
                }}</label>
                <input
                  v-model.number="configForm.block_status"
                  type="number"
                  min="400"
                  max="599"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.blockMessage")
                }}</label>
                <input
                  v-model.trim="configForm.block_message"
                  type="text"
                  class="input"
                />
              </div>
              <div
                class="flex items-center justify-between rounded-lg border border-hairline-soft p-4"
              >
                <div>
                  <p class="text-sm font-medium text-ink">
                    {{ t("admin.riskControl.emailOnHit") }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    {{ t("admin.riskControl.emailOnHitHint") }}
                  </p>
                </div>
                <Toggle v-model="configForm.email_on_hit" />
              </div>
              <div
                class="flex items-center justify-between rounded-lg border border-hairline-soft p-4"
              >
                <div>
                  <p class="text-sm font-medium text-ink">
                    {{ t("admin.riskControl.autoBan") }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    {{ t("admin.riskControl.autoBanHint") }}
                  </p>
                </div>
                <Toggle v-model="configForm.auto_ban_enabled" />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.banThreshold")
                }}</label>
                <input
                  v-model.number="configForm.ban_threshold"
                  type="number"
                  min="1"
                  max="1000"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{
                  t("admin.riskControl.violationWindowHours")
                }}</label>
                <input
                  v-model.number="configForm.violation_window_hours"
                  type="number"
                  min="1"
                  max="8760"
                  class="input"
                />
              </div>
            </div>
          </div>

          <div v-else-if="activeSettingsTab === 'keywords'" class="space-y-5">
            <div
              class="flex items-start gap-3 rounded-lg border p-4"
              :class="keywordNotice.toneClass"
            >
              <Icon
                :name="keywordNotice.icon"
                size="md"
                :class="keywordNotice.iconClass"
              />
              <div class="text-sm leading-6">
                <p class="font-medium" :class="keywordNotice.titleClass">
                  {{ keywordNotice.title }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ keywordNotice.description }}
                </p>
              </div>
            </div>

            <div class="space-y-2">
              <label class="input-label">{{
                t("admin.riskControl.keywordBlockingMode")
              }}</label>
              <div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
                <button
                  v-for="option in keywordBlockingModeOptions"
                  :key="option.value"
                  type="button"
                  class="rounded-lg border p-3 text-left transition-colors"
                  :class="
                    configForm.keyword_blocking_mode === option.value
                      ? 'border-primary-200 bg-primary-500/10 text-primary-700 shadow-sm '
                      : 'border-hairline bg-canvas text-body hover:bg-surface-card '
                  "
                  @click="configForm.keyword_blocking_mode = option.value"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="text-sm font-semibold">{{ option.label }}</span>
                    <span
                      class="flex h-4 w-4 flex-shrink-0 items-center justify-center rounded-full border"
                      :class="
                        configForm.keyword_blocking_mode === option.value
                          ? 'border-primary-500 bg-primary-500 text-on-primary'
                          : 'border-hairline text-transparent'
                      "
                    >
                      <Icon name="check" size="xs" :stroke-width="2" />
                    </span>
                  </div>
                  <p class="mt-1 text-xs leading-5 text-muted">
                    {{ option.description }}
                  </p>
                </button>
              </div>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between">
                <label class="input-label mb-0">{{
                  t("admin.riskControl.blockedKeywords")
                }}</label>
                <span class="inline-flex rounded-md bg-surface-card px-2 py-1 text-xs text-muted">
                  {{
                    t("admin.riskControl.blockedKeywordCount", {
                      count: blockedKeywordCount,
                    })
                  }}
                </span>
              </div>
              <textarea
                v-model="configForm.blocked_keywords_text"
                class="input min-h-52 resize-y font-mono text-sm"
                :placeholder="t('admin.riskControl.blockedKeywordsPlaceholder')"
                :disabled="configForm.keyword_blocking_mode === 'api_only'"
              ></textarea>
              <p class="mt-2 text-xs text-muted">
                {{
                  t("admin.riskControl.blockedKeywordsLimit", {
                    max: blockedKeywordMax,
                  })
                }}
              </p>
            </div>
          </div>

          <div v-else class="grid grid-cols-1 gap-5 lg:grid-cols-2">
            <div>
              <label class="input-label">{{
                t("admin.riskControl.hitRetentionDays")
              }}</label>
              <input
                v-model.number="configForm.hit_retention_days"
                type="number"
                min="1"
                max="3650"
                class="input"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.riskControl.nonHitRetentionDays")
              }}</label>
              <input
                v-model.number="configForm.non_hit_retention_days"
                type="number"
                min="1"
                max="3"
                class="input"
              />
            </div>
            <div
              class="rounded-lg border border-hairline-soft p-4 text-sm text-muted lg:col-span-2"
            >
              <div class="flex flex-wrap items-center gap-3">
                <Icon name="database" size="md" class="text-muted-soft" />
                <span>{{
                  t("admin.riskControl.cleanupStats", {
                    hit: status?.last_cleanup_deleted_hit ?? 0,
                    nonHit: status?.last_cleanup_deleted_non_hit ?? 0,
                  })
                }}</span>
              </div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              class="btn btn-secondary"
              @click="settingsOpen = false"
            >
              {{ t("common.cancel") }}
            </button>
            <button
              type="button"
              class="btn btn-primary inline-flex items-center gap-2"
              :disabled="saving"
              @click="saveConfig"
            >
              <Icon
                v-if="saving"
                name="refresh"
                size="sm"
                class="animate-spin"
              />
              <Icon v-else name="check" size="sm" />
              {{
                saving ? t("common.saving") : t("admin.riskControl.saveConfig")
              }}
            </button>
          </div>
        </template>
      </BaseDialog>

      <BaseDialog
        :show="inputDetailRow !== null"
        :title="t('admin.riskControl.inputDetailTitle')"
        width="wide"
        @close="closeInputDetail"
      >
        <div v-if="inputDetailRow" class="space-y-5">
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <div
              class="rounded-lg border border-hairline-soft bg-surface-soft p-4"
            >
              <p class="text-xs font-medium text-muted">
                {{ t("admin.riskControl.table.time") }}
              </p>
              <p class="mt-1 truncate text-sm font-semibold text-ink">
                {{ formatDateTime(inputDetailRow.created_at) }}
              </p>
            </div>
            <div
              class="rounded-lg border border-hairline-soft bg-surface-soft p-4"
            >
              <p class="text-xs font-medium text-muted">
                {{ t("admin.riskControl.table.user") }}
              </p>
              <p class="mt-1 truncate text-sm font-semibold text-ink">
                {{ inputDetailRow.user_email || "-" }}
              </p>
            </div>
            <div
              class="rounded-lg border border-hairline-soft bg-surface-soft p-4"
            >
              <p class="text-xs font-medium text-muted">
                {{ t("admin.riskControl.table.result") }}
              </p>
              <span
                class="mt-1 inline-flex rounded-md px-2 py-1 text-xs font-medium"
                :class="resultBadgeClass(inputDetailRow)"
              >
                {{ resultLabel(inputDetailRow) }}
              </span>
            </div>
            <div
              class="rounded-lg border border-hairline-soft bg-surface-soft p-4"
            >
              <p class="text-xs font-medium text-muted">
                {{ t("admin.riskControl.table.highest") }}
              </p>
              <p class="mt-1 truncate text-sm font-semibold text-ink">
                {{ inputDetailRow.highest_category || "-" }} /
                {{ percent(inputDetailRow.highest_score) }}
              </p>
            </div>
          </div>

          <div
            class="rounded-xl border border-hairline-soft bg-canvas p-4 shadow-sm"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-ink">
                  {{ t("admin.riskControl.inputDetailContent") }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ inputDetailRow.endpoint || "-" }} ·
                  {{ inputDetailRow.provider || "-" }} /
                  {{ inputDetailRow.model || "-" }}
                </p>
              </div>
              <span
                v-if="inputDetailRow.group_name"
                class="inline-flex rounded-md bg-accent-teal/15 px-2.5 py-1 text-xs font-medium text-accent-teal"
              >
                {{ inputDetailRow.group_name }}
              </span>
            </div>
            <pre
              class="mt-4 max-h-[420px] overflow-auto whitespace-pre-wrap break-words rounded-lg bg-surface-dark p-4 text-sm leading-6 text-canvas shadow-inner"
              >{{ inputDetailText }}</pre
            >
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end">
            <button
              type="button"
              class="btn btn-secondary"
              @click="closeInputDetail"
            >
              {{ t("common.close") }}
            </button>
          </div>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import Select from "@/components/common/Select.vue";
import Toggle from "@/components/common/Toggle.vue";
import Pagination from "@/components/common/Pagination.vue";
import { adminAPI } from "@/api/admin";
import type {
  ContentModerationAPIKeyStatus,
  ContentModerationConfig,
  ContentModerationLog,
  ContentModerationRuntimeStatus,
  ContentModerationTestAuditResult,
  KeywordBlockingMode,
  ModerationMode,
  UpdateContentModerationConfig,
} from "@/api/admin/riskControl";
import type { AdminGroup, SelectOption } from "@/types";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { formatDateTime as formatDateTimeValue } from "@/utils/format";

type SettingsTab =
  | "basic"
  | "scope"
  | "runtime"
  | "response"
  | "keywords"
  | "retention";
type WorkerSlotState = "active" | "idle" | "disabled";
type APIKeysWriteMode = "append" | "replace";
type OverviewIcon = "shield" | "key" | "users" | "document";
type OverviewItem = {
  key: string;
  label: string;
  value: string;
  meta: string;
  icon: OverviewIcon;
  iconClass: string;
  badge?: string;
  badgeClass?: string;
};
type ModerationScoreRow = {
  category: string;
  score: number;
  threshold: number;
  hit: boolean;
};

const maxModerationTestImages = 1;
const maxModerationTestImageSize = 8 * 1024 * 1024;
const maxVisibleApiKeyRows: number = 3;
const blockedKeywordMax = 10000;

const { t } = useI18n();
const appStore = useAppStore();

const loading = ref(true);
const saving = ref(false);
const logsLoading = ref(false);
const statusLoading = ref(false);
const apiKeyTesting = ref(false);
const hashActionLoading = ref(false);
const unbanningUserID = ref<number | null>(null);
const settingsOpen = ref(false);
const activeSettingsTab = ref<SettingsTab>("basic");
const groupSearch = ref("");
const flaggedHashInput = ref("");
const groups = ref<AdminGroup[]>([]);
const logs = ref<ContentModerationLog[]>([]);
const status = ref<ContentModerationRuntimeStatus | null>(null);
const testedApiKeyStatuses = ref<ContentModerationAPIKeyStatus[]>([]);
const pendingDeleteApiKeyHashes = ref<string[]>([]);
const apiKeyRowsExpanded = ref<boolean>(false);
const moderationTestPrompt = ref("");
const moderationTestImages = ref<string[]>([]);
const moderationTestResult = ref<ContentModerationTestAuditResult | null>(null);
const inputDetailRow = ref<ContentModerationLog | null>(null);
let statusTimer: number | null = null;

const configForm = reactive({
  enabled: false,
  mode: "pre_block" as ModerationMode,
  base_url: "https://api.openai.com",
  model: "omni-moderation-latest",
  api_keys_text: "",
  api_key_configured: false,
  api_key_masked: "",
  api_key_count: 0,
  api_key_masks: [] as string[],
  api_key_statuses: [] as ContentModerationAPIKeyStatus[],
  api_keys_mode: "append" as APIKeysWriteMode,
  clear_api_key: false,
  timeout_ms: 3000,
  retry_count: 2,
  sample_rate: 100,
  all_groups: true,
  group_ids: [] as number[],
  record_non_hits: false,
  worker_count: 4,
  queue_size: 32768,
  block_status: 403,
  block_message: "内容审计命中风险规则，请调整输入后重试",
  email_on_hit: true,
  auto_ban_enabled: true,
  ban_threshold: 10,
  violation_window_hours: 720,
  hit_retention_days: 180,
  non_hit_retention_days: 3,
  pre_hash_check_enabled: false,
  blocked_keywords_text: "",
  keyword_blocking_mode: "keyword_and_api" as KeywordBlockingMode,
});

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 1,
});

const filters = reactive({
  result: "",
  group_id: 0,
  endpoint: "",
  search: "",
  from: "",
  to: "",
});

const settingsTabs = computed<Array<{ id: SettingsTab; label: string }>>(() => [
  { id: "basic", label: t("admin.riskControl.tabs.basic") },
  { id: "scope", label: t("admin.riskControl.tabs.scope") },
  { id: "runtime", label: t("admin.riskControl.tabs.runtime") },
  { id: "response", label: t("admin.riskControl.tabs.response") },
  { id: "keywords", label: t("admin.riskControl.tabs.keywords") },
  { id: "retention", label: t("admin.riskControl.tabs.retention") },
]);

const modeOptions = computed<SelectOption[]>(() => [
  { value: "pre_block", label: t("admin.riskControl.modePreBlock") },
  { value: "observe", label: t("admin.riskControl.modeObserve") },
  { value: "off", label: t("admin.riskControl.modeOff") },
]);

const keywordBlockingModeOptions = computed<Array<{ value: KeywordBlockingMode; label: string; description: string }>>(() => [
  {
    value: 'keyword_and_api',
    label: t('admin.riskControl.keywordModeKeywordAndApi'),
    description: t('admin.riskControl.keywordModeKeywordAndApiDesc'),
  },
  {
    value: 'keyword_only',
    label: t('admin.riskControl.keywordModeKeywordOnly'),
    description: t('admin.riskControl.keywordModeKeywordOnlyDesc'),
  },
  {
    value: 'api_only',
    label: t('admin.riskControl.keywordModeApiOnly'),
    description: t('admin.riskControl.keywordModeApiOnlyDesc'),
  },
])

type KeywordNoticeView = {
  title: string
  description: string
  icon: 'infoCircle' | 'exclamationTriangle'
  toneClass: string
  iconClass: string
  titleClass: string
}

const keywordNoticeTones = {
  info: {
    icon: 'infoCircle' as const,
    toneClass: 'border-primary-200 bg-primary-500/10',
    iconClass: 'mt-0.5 flex-shrink-0 text-primary-600',
    titleClass: 'text-primary-700',
  },
  warning: {
    icon: 'exclamationTriangle' as const,
    toneClass: 'border-accent-amber/30 bg-accent-amber/10',
    iconClass: 'mt-0.5 flex-shrink-0 text-warning',
    titleClass: 'text-warning',
  },
}

const keywordNotice = computed<KeywordNoticeView>(() => {
  const strategy = configForm.keyword_blocking_mode
  if (strategy === 'api_only') {
    return {
      ...keywordNoticeTones.info,
      title: t('admin.riskControl.keywordModeApiOnlyNotice'),
      description: t('admin.riskControl.keywordModeApiOnlyDesc'),
    }
  }
  if (configForm.mode !== 'pre_block') {
    return {
      ...keywordNoticeTones.warning,
      title: t('admin.riskControl.blockedKeywordsModeWarning', { mode: modeLabel(configForm.mode) }),
      description: t('admin.riskControl.blockedKeywordsDescription'),
    }
  }
  if (strategy === 'keyword_only') {
    return {
      ...keywordNoticeTones.info,
      title: t('admin.riskControl.keywordModeKeywordOnlyNotice'),
      description: t('admin.riskControl.keywordModeKeywordOnlyDesc'),
    }
  }
  return {
    ...keywordNoticeTones.info,
    title: t('admin.riskControl.blockedKeywordsPreBlockHint'),
    description: t('admin.riskControl.blockedKeywordsDescription'),
  }
})

const resultOptions = computed<SelectOption[]>(() => [
  { value: "", label: t("admin.riskControl.result.all") },
  { value: "hit", label: t("admin.riskControl.result.hit") },
  { value: "blocked", label: t("admin.riskControl.result.blocked") },
  { value: "pass", label: t("admin.riskControl.result.pass") },
  { value: "error", label: t("admin.riskControl.result.error") },
]);

const endpointOptions = computed<SelectOption[]>(() => [
  { value: "", label: t("admin.riskControl.filters.allEndpoints") },
  { value: "/v1/messages", label: "/v1/messages" },
  { value: "/v1/responses", label: "/v1/responses" },
  { value: "/v1/chat/completions", label: "/v1/chat/completions" },
  { value: "/v1beta/models", label: "/v1beta/models" },
  { value: "/v1/images/generations", label: "/v1/images/generations" },
  { value: "/v1/images/edits", label: "/v1/images/edits" },
]);

const groupFilterOptions = computed<SelectOption[]>(() => [
  { value: 0, label: t("admin.riskControl.filters.allGroups") },
  ...groups.value.map((group) => ({
    value: group.id,
    label: `${group.name} (${group.platform})`,
  })),
]);

const selectedGroupCount = computed(() => String(configForm.group_ids.length));

const filteredGroups = computed(() => {
  const keyword = groupSearch.value.trim().toLowerCase();
  if (!keyword) return groups.value;
  return groups.value.filter((group) => {
    return (
      group.name.toLowerCase().includes(keyword) ||
      String(group.platform).toLowerCase().includes(keyword)
    );
  });
});

const inputApiKeyCount = computed(
  () => parseApiKeys(configForm.api_keys_text).length,
);

const blockedKeywordList = computed(() =>
  parseBlockedKeywords(configForm.blocked_keywords_text),
);

const blockedKeywordCount = computed(() => blockedKeywordList.value.length);

const pendingDeletedApiKeyCount = computed(
  () => pendingDeleteApiKeyHashes.value.length,
);

const effectiveStoredApiKeyCount = computed(() =>
  Math.max(0, configForm.api_key_count - pendingDeletedApiKeyCount.value),
);

const apiKeysPlaceholder = computed(() =>
  configForm.api_keys_mode === "replace"
    ? t("admin.riskControl.apiKeysPlaceholderReplace")
    : t("admin.riskControl.apiKeysPlaceholder"),
);

const apiKeysModeHint = computed(() =>
  configForm.api_keys_mode === "replace"
    ? t("admin.riskControl.apiKeysModeReplaceHint")
    : t("admin.riskControl.apiKeysModeAppendHint"),
);

const hasModerationAuditInput = computed(() => {
  return (
    moderationTestPrompt.value.trim() !== "" ||
    moderationTestImages.value.length > 0
  );
});

const isFlaggedHashInputValid = computed(() =>
  /^[a-fA-F0-9]{64}$/.test(flaggedHashInput.value.trim()),
);

const storedApiKeyTestButtonText = computed(() => {
  if (apiKeyTesting.value) return t("admin.riskControl.testingApiKeys");
  if (hasModerationAuditInput.value)
    return t("admin.riskControl.testContentWithStoredApiKey");
  return t("admin.riskControl.testStoredApiKeys");
});

const savedApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => {
  const rows = status.value?.api_key_statuses?.length
    ? status.value.api_key_statuses
    : configForm.api_key_statuses;
  return Array.isArray(rows) ? rows : [];
});

const apiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => [
  ...savedApiKeyRows.value,
  ...testedApiKeyStatuses.value,
]);

const visibleApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() => {
  if (apiKeyRowsExpanded.value) return apiKeyRows.value;
  return apiKeyRows.value.slice(0, maxVisibleApiKeyRows);
});

const hiddenApiKeyRowCount = computed<number>(() =>
  Math.max(0, apiKeyRows.value.length - visibleApiKeyRows.value.length),
);

const canToggleApiKeyRows = computed<boolean>(
  () => apiKeyRows.value.length > maxVisibleApiKeyRows,
);

const activeSavedApiKeyRows = computed<ContentModerationAPIKeyStatus[]>(() =>
  savedApiKeyRows.value.filter((row) => !isStoredApiKeyPendingDelete(row)),
);

const apiKeyHealthBadges = computed<
  Array<{ status: ContentModerationAPIKeyStatus["status"]; count: number }>
>(() => {
  const counts: Record<ContentModerationAPIKeyStatus["status"], number> = {
    ok: 0,
    error: 0,
    frozen: 0,
    unknown: 0,
  };
  for (const row of activeSavedApiKeyRows.value) {
    counts[row.status] = (counts[row.status] ?? 0) + 1;
  }
  if (
    activeSavedApiKeyRows.value.length === 0 &&
    effectiveStoredApiKeyCount.value > 0
  ) {
    counts.unknown = effectiveStoredApiKeyCount.value;
  }
  return (
    ["ok", "frozen", "error", "unknown"] as Array<
      ContentModerationAPIKeyStatus["status"]
    >
  )
    .map((item) => ({ status: item, count: counts[item] }))
    .filter((item) => item.count > 0);
});

const apiKeyHealthSummary = computed(() => {
  if (!configForm.api_key_configured) return "";
  if (apiKeyHealthBadges.value.length === 0)
    return t("admin.riskControl.apiKeyStatusUnknown");
  return apiKeyHealthBadges.value
    .map((badge) => `${apiKeyStatusLabel(badge.status)} ${badge.count}`)
    .join(" · ");
});

const overviewItems = computed<OverviewItem[]>(() => [
  {
    key: "status",
    label: t("admin.riskControl.overview.status"),
    value: configForm.enabled
      ? t("admin.riskControl.overview.enabled")
      : t("admin.riskControl.overview.disabled"),
    meta: modeLabel(configForm.mode),
    icon: "shield",
    iconClass: configForm.enabled
      ? "bg-success/15 text-success "
      : "bg-surface-card text-muted ",
    badge: runtimeBadgeText.value,
    badgeClass: runtimeBadgeClass.value,
  },
  {
    key: "api-key",
    label: t("admin.riskControl.overview.apiKey"),
    value: configForm.api_key_configured
      ? t("admin.riskControl.apiKeyCount", { count: configForm.api_key_count })
      : t("admin.riskControl.notConfigured"),
    meta: configForm.api_key_configured
      ? apiKeyHealthSummary.value || configForm.model || "-"
      : configForm.model || "-",
    icon: "key",
    iconClass: "bg-accent-teal/15 text-accent-teal ",
  },
  {
    key: "scope",
    label: t("admin.riskControl.overview.groupScope"),
    value: configForm.all_groups
      ? t("admin.riskControl.allGroups")
      : selectedGroupCount.value,
    meta: configForm.all_groups
      ? t("admin.riskControl.allGroupsHint")
      : t("admin.riskControl.selectedGroupsHint"),
    icon: "users",
    iconClass: "bg-primary-100 text-primary-600 ",
  },
  {
    key: "logs",
    label: t("admin.riskControl.overview.logs"),
    value: formatNumber(pagination.total),
    meta: t("admin.riskControl.overview.currentFilter"),
    icon: "document",
    iconClass: "bg-accent-amber/15 text-warning ",
  },
]);

const moderationScoreRows = computed<ModerationScoreRow[]>(() => {
  const result = moderationTestResult.value;
  if (!result) return [];
  return Object.entries(result.category_scores || {})
    .map(([category, score]) => {
      const threshold = result.thresholds?.[category] ?? 1;
      return {
        category,
        score,
        threshold,
        hit: score >= threshold,
      };
    })
    .sort((a, b) => b.score - a.score);
});

const inputDetailText = computed(() => {
  if (!inputDetailRow.value) return "-";
  return (
    inputDetailRow.value.input_excerpt || inputDetailRow.value.error || "-"
  );
});

const queueUsagePercent = computed(
  () =>
    `${Math.min(100, Math.max(0, status.value?.queue_usage_percent ?? 0)).toFixed(1)}%`,
);

const queueUsageStyle = computed(() => ({
  width: queueUsagePercent.value,
}));

const workerSlots = computed(() => {
  const total = Math.max(
    0,
    status.value?.worker_count ?? configForm.worker_count,
  );
  const active = Math.max(0, status.value?.active_workers ?? 0);
  const enabled = Boolean(
    status.value?.risk_control_enabled &&
    status.value?.enabled &&
    status.value?.mode !== "off",
  );
  return Array.from({ length: total }, (_, index) => ({
    id: index + 1,
    state: (!enabled
      ? "disabled"
      : index < active
        ? "active"
        : "idle") as WorkerSlotState,
    label: !enabled
      ? t("admin.riskControl.workerDisabled")
      : index < active
        ? t("admin.riskControl.workerActive")
        : t("admin.riskControl.workerIdle"),
  }));
});

const runtimeBadgeText = computed(() => {
  if (!status.value?.risk_control_enabled)
    return t("admin.riskControl.riskSwitchOff");
  if (!configForm.enabled || configForm.mode === "off")
    return t("admin.riskControl.overview.disabled");
  return t("admin.riskControl.overview.enabled");
});

const runtimeBadgeClass = computed(() => {
  if (
    !status.value?.risk_control_enabled ||
    !configForm.enabled ||
    configForm.mode === "off"
  ) {
    return "bg-surface-card text-body ";
  }
  return "bg-success/15 text-success ";
});

function applyConfig(config: ContentModerationConfig) {
  configForm.enabled = config.enabled;
  configForm.mode = config.mode;
  configForm.base_url = config.base_url || "https://api.openai.com";
  configForm.model = config.model || "omni-moderation-latest";
  configForm.api_keys_text = "";
  configForm.api_key_configured = config.api_key_configured;
  configForm.api_key_masked = config.api_key_masked || "";
  configForm.api_key_count = config.api_key_count || 0;
  configForm.api_key_masks = Array.isArray(config.api_key_masks)
    ? [...config.api_key_masks]
    : [];
  configForm.api_key_statuses = Array.isArray(config.api_key_statuses)
    ? [...config.api_key_statuses]
    : [];
  configForm.api_keys_mode = "append";
  configForm.clear_api_key = false;
  pendingDeleteApiKeyHashes.value = [];
  testedApiKeyStatuses.value = [];
  apiKeyRowsExpanded.value = false;
  configForm.timeout_ms = config.timeout_ms || 3000;
  configForm.retry_count = config.retry_count ?? 2;
  configForm.sample_rate = config.sample_rate ?? 100;
  configForm.all_groups = config.all_groups;
  configForm.group_ids = Array.isArray(config.group_ids)
    ? [...config.group_ids]
    : [];
  configForm.record_non_hits = config.record_non_hits;
  configForm.worker_count = config.worker_count || 4;
  configForm.queue_size = config.queue_size || 32768;
  configForm.block_status = config.block_status || 403;
  configForm.block_message =
    config.block_message || "内容审计命中风险规则，请调整输入后重试";
  configForm.email_on_hit = config.email_on_hit ?? true;
  configForm.auto_ban_enabled = config.auto_ban_enabled ?? true;
  configForm.ban_threshold = config.ban_threshold || 10;
  configForm.violation_window_hours = config.violation_window_hours || 720;
  configForm.hit_retention_days = config.hit_retention_days || 180;
  configForm.non_hit_retention_days = Math.min(
    Math.max(config.non_hit_retention_days || 3, 1),
    3,
  );
  configForm.pre_hash_check_enabled = config.pre_hash_check_enabled ?? false;
  configForm.blocked_keywords_text = Array.isArray(config.blocked_keywords)
    ? config.blocked_keywords.join("\n")
    : "";
  configForm.keyword_blocking_mode = normalizeKeywordBlockingMode(
    config.keyword_blocking_mode,
  );
}

async function loadAll() {
  loading.value = true;
  try {
    const [config, groupItems, runtimeStatus] = await Promise.all([
      adminAPI.riskControl.getConfig(),
      adminAPI.groups.getAll(),
      adminAPI.riskControl.getStatus(),
    ]);
    applyConfig(config);
    groups.value = groupItems;
    status.value = runtimeStatus;
    if (Array.isArray(runtimeStatus.api_key_statuses)) {
      configForm.api_key_statuses = [...runtimeStatus.api_key_statuses];
      prunePendingDeleteAPIKeyHashes();
    }
    await loadLogs();
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(err, t("admin.riskControl.loadFailed")),
    );
  } finally {
    loading.value = false;
  }
}

async function loadStatus(silent = true) {
  statusLoading.value = true;
  try {
    const runtimeStatus = await adminAPI.riskControl.getStatus();
    status.value = runtimeStatus;
    if (Array.isArray(runtimeStatus.api_key_statuses)) {
      configForm.api_key_statuses = [...runtimeStatus.api_key_statuses];
      prunePendingDeleteAPIKeyHashes();
    }
  } catch (err: unknown) {
    if (!silent) {
      appStore.showError(
        extractApiErrorMessage(err, t("admin.riskControl.statusFailed")),
      );
    }
  } finally {
    statusLoading.value = false;
  }
}

async function saveConfig() {
  saving.value = true;
  try {
    const payload: UpdateContentModerationConfig = {
      enabled: configForm.enabled,
      mode: configForm.mode,
      base_url: configForm.base_url,
      model: configForm.model,
      timeout_ms: Number(configForm.timeout_ms) || 3000,
      retry_count: Number(configForm.retry_count) || 0,
      sample_rate: Number(configForm.sample_rate) || 0,
      all_groups: configForm.all_groups,
      group_ids: configForm.all_groups ? [] : [...configForm.group_ids],
      record_non_hits: configForm.record_non_hits,
      clear_api_key: configForm.clear_api_key,
      worker_count: Number(configForm.worker_count) || 4,
      queue_size: Number(configForm.queue_size) || 32768,
      block_status: Number(configForm.block_status) || 403,
      block_message:
        configForm.block_message || "内容审计命中风险规则，请调整输入后重试",
      email_on_hit: configForm.email_on_hit,
      auto_ban_enabled: configForm.auto_ban_enabled,
      ban_threshold: Number(configForm.ban_threshold) || 10,
      violation_window_hours: Number(configForm.violation_window_hours) || 720,
      hit_retention_days: Number(configForm.hit_retention_days) || 180,
      non_hit_retention_days: Math.min(
        Math.max(Number(configForm.non_hit_retention_days) || 3, 1),
        3,
      ),
      pre_hash_check_enabled: configForm.pre_hash_check_enabled,
      blocked_keywords: blockedKeywordList.value,
      keyword_blocking_mode: configForm.keyword_blocking_mode,
    };
    const keys = parseApiKeys(configForm.api_keys_text);
    if (
      !payload.clear_api_key &&
      configForm.api_keys_mode === "replace" &&
      keys.length === 0
    ) {
      appStore.showError(t("admin.riskControl.apiKeysReplaceNoInput"));
      return;
    }
    if (keys.length > 0) {
      payload.api_keys = keys;
      payload.api_keys_mode = configForm.api_keys_mode;
      payload.clear_api_key = false;
    }
    if (
      !payload.clear_api_key &&
      configForm.api_keys_mode !== "replace" &&
      pendingDeleteApiKeyHashes.value.length > 0
    ) {
      payload.delete_api_key_hashes = [...pendingDeleteApiKeyHashes.value];
    }

    const updated = await adminAPI.riskControl.updateConfig(payload);
    applyConfig(updated);
    settingsOpen.value = false;
    appStore.showSuccess(t("admin.riskControl.saved"));
    await Promise.all([loadStatus(true), loadLogs()]);
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(err, t("admin.riskControl.saveFailed")),
    );
  } finally {
    saving.value = false;
  }
}

async function loadLogs() {
  logsLoading.value = true;
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      result: filters.result || undefined,
      group_id: filters.group_id || undefined,
      endpoint: filters.endpoint || undefined,
      search: filters.search || undefined,
      from: normalizeDateTimeLocal(filters.from),
      to: normalizeDateTimeLocal(filters.to),
    };
    const result = await adminAPI.riskControl.listLogs(params);
    logs.value = result.items;
    pagination.total = result.total;
    pagination.page = result.page;
    pagination.page_size = result.page_size;
    pagination.pages = result.pages;
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(err, t("admin.riskControl.logsFailed")),
    );
  } finally {
    logsLoading.value = false;
  }
}

function canUnbanRow(row: ContentModerationLog): boolean {
  return Boolean(
    row.auto_banned && row.user_id && row.user_status === "disabled",
  );
}

function inputSummaryText(row: ContentModerationLog): string {
  return row.input_excerpt || row.error || "-";
}

function openInputDetail(row: ContentModerationLog) {
  inputDetailRow.value = row;
}

function closeInputDetail() {
  inputDetailRow.value = null;
}

async function unbanUser(row: ContentModerationLog) {
  if (!row.user_id || unbanningUserID.value !== null) return;
  unbanningUserID.value = row.user_id;
  try {
    const result = await adminAPI.riskControl.unbanUser(row.user_id);
    logs.value = logs.value.map((item) => {
      if (item.user_id !== row.user_id) return item;
      return { ...item, user_status: result.status };
    });
    appStore.showSuccess(t("admin.riskControl.unbanSuccess"));
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(err, t("admin.riskControl.unbanFailed")),
    );
  } finally {
    unbanningUserID.value = null;
  }
}

async function deleteFlaggedHash() {
  if (!isFlaggedHashInputValid.value || hashActionLoading.value) return;
  hashActionLoading.value = true;
  try {
    const result = await adminAPI.riskControl.deleteFlaggedHash(
      flaggedHashInput.value,
    );
    flaggedHashInput.value = "";
    await loadStatus(true);
    appStore.showSuccess(
      result.deleted
        ? t("admin.riskControl.flaggedHashDeleted")
        : t("admin.riskControl.flaggedHashNotFound"),
    );
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        err,
        t("admin.riskControl.flaggedHashDeleteFailed"),
      ),
    );
  } finally {
    hashActionLoading.value = false;
  }
}

async function clearFlaggedHashes() {
  if (hashActionLoading.value) return;
  const confirmed = window.confirm(
    t("admin.riskControl.clearFlaggedHashesConfirm"),
  );
  if (!confirmed) return;
  hashActionLoading.value = true;
  try {
    const result = await adminAPI.riskControl.clearFlaggedHashes();
    await loadStatus(true);
    appStore.showSuccess(
      t("admin.riskControl.flaggedHashesCleared", { count: result.deleted }),
    );
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(
        err,
        t("admin.riskControl.flaggedHashesClearFailed"),
      ),
    );
  } finally {
    hashActionLoading.value = false;
  }
}

function openSettings() {
  activeSettingsTab.value = "basic";
  settingsOpen.value = true;
}

function reloadLogsFromFirstPage() {
  pagination.page = 1;
  void loadLogs();
}

function onPageChange(page: number) {
  pagination.page = page;
  void loadLogs();
}

function onPageSizeChange(pageSize: number) {
  pagination.page = 1;
  pagination.page_size = pageSize;
  void loadLogs();
}

function toggleClearApiKey() {
  configForm.clear_api_key = !configForm.clear_api_key;
  if (configForm.clear_api_key) {
    configForm.api_keys_text = "";
    configForm.api_keys_mode = "append";
    testedApiKeyStatuses.value = [];
    pendingDeleteApiKeyHashes.value = [];
  }
}

function setAPIKeysMode(mode: APIKeysWriteMode) {
  configForm.api_keys_mode = mode;
  if (mode === "replace") {
    pendingDeleteApiKeyHashes.value = [];
  }
}

async function testApiKeys(useInputKeys: boolean) {
  const keys = useInputKeys ? parseApiKeys(configForm.api_keys_text) : [];
  if (useInputKeys && keys.length === 0) {
    appStore.showError(t("admin.riskControl.apiKeyTestNoInput"));
    return;
  }
  apiKeyTesting.value = true;
  try {
    const result = await adminAPI.riskControl.testAPIKeys({
      api_keys: keys,
      base_url: configForm.base_url,
      model: configForm.model,
      timeout_ms: Number(configForm.timeout_ms) || 3000,
      prompt: moderationTestPrompt.value,
      images: moderationTestImages.value,
    });
    moderationTestResult.value = result.audit_result ?? null;
    if (useInputKeys) {
      testedApiKeyStatuses.value = result.items.map((item) => ({
        ...item,
        configured: false,
      }));
    } else {
      mergeConfiguredAPIKeyStatuses(result.items);
      testedApiKeyStatuses.value = [];
      await loadStatus(true);
    }
    appStore.showSuccess(
      t("admin.riskControl.apiKeyTestDone", { count: result.items.length }),
    );
  } catch (err: unknown) {
    appStore.showError(
      extractApiErrorMessage(err, t("admin.riskControl.apiKeyTestFailed")),
    );
  } finally {
    apiKeyTesting.value = false;
  }
}

function mergeConfiguredAPIKeyStatuses(items: ContentModerationAPIKeyStatus[]) {
  if (
    !hasModerationAuditInput.value ||
    configForm.api_key_statuses.length === 0
  ) {
    configForm.api_key_statuses = items;
    return;
  }
  const updates = new Map(items.map((item) => [item.key_hash, item]));
  configForm.api_key_statuses = configForm.api_key_statuses.map(
    (item) => updates.get(item.key_hash) ?? item,
  );
}

function toggleDeleteStoredApiKey(row: ContentModerationAPIKeyStatus) {
  if (!row.configured || !row.key_hash) return;
  const index = pendingDeleteApiKeyHashes.value.indexOf(row.key_hash);
  if (index >= 0) {
    pendingDeleteApiKeyHashes.value.splice(index, 1);
    return;
  }
  pendingDeleteApiKeyHashes.value.push(row.key_hash);
}

function isStoredApiKeyPendingDelete(
  row: ContentModerationAPIKeyStatus,
): boolean {
  return (
    row.configured &&
    row.key_hash !== "" &&
    pendingDeleteApiKeyHashes.value.includes(row.key_hash)
  );
}

function prunePendingDeleteAPIKeyHashes() {
  const currentHashes = new Set(
    savedApiKeyRows.value.map((row) => row.key_hash).filter(Boolean),
  );
  pendingDeleteApiKeyHashes.value = pendingDeleteApiKeyHashes.value.filter(
    (hash) => currentHashes.has(hash),
  );
}

function clearModerationTestInput() {
  moderationTestPrompt.value = "";
  moderationTestImages.value = [];
  moderationTestResult.value = null;
}

function removeModerationTestImage(index: number) {
  moderationTestImages.value.splice(index, 1);
}

async function handleModerationImageUpload(event: Event) {
  const input = event.target as HTMLInputElement;
  await addModerationTestFiles(input.files);
  input.value = "";
}

async function handleModerationImageDrop(event: DragEvent) {
  await addModerationTestFiles(event.dataTransfer?.files ?? null);
}

async function handleModerationImagePaste(event: ClipboardEvent) {
  const files = Array.from(event.clipboardData?.files ?? []).filter((file) =>
    file.type.startsWith("image/"),
  );
  if (files.length === 0) return;
  event.preventDefault();
  await addModerationTestFiles(files);
}

async function addModerationTestFiles(files: FileList | File[] | null) {
  if (!files) return;
  const items = Array.from(files).filter((file) =>
    file.type.startsWith("image/"),
  );
  for (const file of items) {
    if (moderationTestImages.value.length >= maxModerationTestImages) {
      appStore.showError(
        t("admin.riskControl.auditTestImageLimit", {
          count: maxModerationTestImages,
        }),
      );
      return;
    }
    if (file.size > maxModerationTestImageSize) {
      appStore.showError(t("admin.riskControl.auditTestImageTooLarge"));
      continue;
    }
    try {
      moderationTestImages.value.push(await fileToDataURL(file));
    } catch {
      appStore.showError(t("admin.riskControl.auditTestImageReadFailed"));
    }
  }
}

function fileToDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}

function toggleGroup(groupID: number) {
  const index = configForm.group_ids.indexOf(groupID);
  if (index >= 0) {
    configForm.group_ids.splice(index, 1);
  } else {
    configForm.group_ids.push(groupID);
  }
}

function isGroupSelected(groupID: number): boolean {
  return configForm.group_ids.includes(groupID);
}

function modeLabel(mode: ModerationMode): string {
  const found = modeOptions.value.find((option) => option.value === mode);
  return found?.label ?? mode;
}

function modeDescription(mode: ModerationMode): string {
  const descriptions: Record<ModerationMode, string> = {
    pre_block: t("admin.riskControl.modePreBlockDesc"),
    observe: t("admin.riskControl.modeObserveDesc"),
    off: t("admin.riskControl.modeOffDesc"),
  };
  return descriptions[mode] ?? "";
}

function resultLabel(row: ContentModerationLog): string {
  if (row.action === "keyword_block")
    return t("admin.riskControl.action.keywordBlock");
  if (row.action === "block") return t("admin.riskControl.action.block");
  if (row.action === "error" || row.error)
    return t("admin.riskControl.action.error");
  if (row.flagged) return t("admin.riskControl.result.hit");
  return t("admin.riskControl.result.pass");
}

function resultBadgeClass(row: ContentModerationLog): string {
  if (row.action === "block" || row.action === "keyword_block")
    return "bg-error/15 text-error ";
  if (row.action === "error" || row.error)
    return "bg-accent-amber/15 text-warning ";
  if (row.flagged) return "bg-primary-100 text-primary-700 ";
  return "bg-success/15 text-success ";
}

function workerSlotClass(state: WorkerSlotState): string {
  if (state === "active") {
    return "border-accent-teal/30 bg-accent-teal/15 text-accent-teal ";
  }
  if (state === "idle") {
    return "border-success/25 bg-success/15 text-success ";
  }
  return "border-hairline-soft bg-canvas text-muted-soft ";
}

function workerDotClass(state: WorkerSlotState): string {
  if (state === "active") return "bg-accent-teal";
  if (state === "idle") return "bg-success";
  return "bg-surface-cream-strong ";
}

function percent(value: number): string {
  if (!Number.isFinite(value)) return "-";
  return `${(value * 100).toFixed(1)}%`;
}

function percentWidth(value: number): string {
  if (!Number.isFinite(value)) return "0%";
  return `${Math.min(100, Math.max(0, value * 100)).toFixed(1)}%`;
}

function latencyText(value: number | null): string {
  if (value === null || value === undefined) return "-";
  return `${value} ms`;
}

function apiKeyRowKey(
  row: ContentModerationAPIKeyStatus,
  index: number,
): string {
  return `${row.configured ? "saved" : "test"}-${row.key_hash || index}`;
}

function apiKeyStatusLabel(
  statusValue: ContentModerationAPIKeyStatus["status"],
): string {
  const labels: Record<ContentModerationAPIKeyStatus["status"], string> = {
    ok: t("admin.riskControl.apiKeyStatusOk"),
    error: t("admin.riskControl.apiKeyStatusError"),
    frozen: t("admin.riskControl.apiKeyStatusFrozen"),
    unknown: t("admin.riskControl.apiKeyStatusUnknown"),
  };
  return labels[statusValue] ?? labels.unknown;
}

function apiKeyStatusBadgeClass(
  statusValue: ContentModerationAPIKeyStatus["status"],
): string {
  const classes: Record<ContentModerationAPIKeyStatus["status"], string> = {
    ok: "bg-success/15 text-success ",
    error: "bg-accent-amber/15 text-warning ",
    frozen: "bg-error/15 text-error ",
    unknown: "bg-surface-card text-body ",
  };
  return classes[statusValue] ?? classes.unknown;
}

function apiKeyStatusDotClass(
  statusValue: ContentModerationAPIKeyStatus["status"],
): string {
  const classes: Record<ContentModerationAPIKeyStatus["status"], string> = {
    ok: "bg-success",
    error: "bg-accent-amber",
    frozen: "bg-error",
    unknown: "bg-muted-soft",
  };
  return classes[statusValue] ?? classes.unknown;
}

function apiKeyStatusMeta(row: ContentModerationAPIKeyStatus): string {
  const parts: string[] = [];
  parts.push(
    t("admin.riskControl.apiKeyFailureCount", {
      count: row.failure_count || 0,
    }),
  );
  if (row.last_latency_ms > 0) {
    parts.push(
      t("admin.riskControl.apiKeyLatency", { ms: row.last_latency_ms }),
    );
  }
  if (row.last_http_status > 0) {
    parts.push(
      t("admin.riskControl.apiKeyHTTPStatus", { status: row.last_http_status }),
    );
  }
  if (row.frozen_until) {
    parts.push(
      t("admin.riskControl.apiKeyFrozenUntil", {
        time: formatDateTime(row.frozen_until),
      }),
    );
  } else if (row.last_checked_at) {
    parts.push(
      t("admin.riskControl.apiKeyLastChecked", {
        time: formatDateTime(row.last_checked_at),
      }),
    );
  } else {
    parts.push(t("admin.riskControl.apiKeyNotTested"));
  }
  return parts.join(" / ");
}

function parseApiKeys(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter((item, index, arr) => item && arr.indexOf(item) === index);
}

function normalizeKeywordBlockingMode(value: unknown): KeywordBlockingMode {
  if (value === 'keyword_only' || value === 'api_only' || value === 'keyword_and_api') {
    return value
  }
  return 'keyword_and_api'
}

function parseBlockedKeywords(value: string): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  for (const line of value.split(/\r?\n/)) {
    const kw = line.trim()
    if (!kw) continue
    const key = kw.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(kw)
  }
  return out
}

function violationCountText(row: ContentModerationLog): string {
  if (!row.flagged) return "-";
  return t("admin.riskControl.violationCount", {
    count: row.violation_count || 1,
  });
}

function normalizeDateTimeLocal(value: string): string | undefined {
  if (!value) return undefined;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return undefined;
  return date.toISOString();
}

function formatDateTime(value: string): string {
  return formatDateTimeValue(value) || "-";
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value);
}

onMounted(() => {
  void loadAll();
  statusTimer = window.setInterval(() => {
    void loadStatus(true);
  }, 15000);
});

onUnmounted(() => {
  if (statusTimer !== null) {
    window.clearInterval(statusTimer);
    statusTimer = null;
  }
});
</script>
