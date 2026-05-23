<template>
  <AppLayout>
    <div
      data-testid="profile-shell"
      class="mx-auto max-w-[950px] space-y-6"
    >
      <div class="flex flex-wrap gap-2 border-b border-hairline pb-2">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="rounded-md px-3 py-2 text-sm font-medium transition-colors"
          :class="
            activeTab === tab.key
              ? 'bg-surface-card text-ink ring-1 ring-inset ring-hairline'
              : 'text-muted hover:bg-surface-soft hover:text-body'
          "
          type="button"
          @click="setActiveTab(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>

      <template v-if="activeTab === 'account'">
        <ProfileInfoCard
          :user="user"
          :linuxdo-enabled="linuxdoOAuthEnabled"
          :dingtalk-enabled="dingtalkOAuthEnabled"
          :oidc-enabled="oidcOAuthEnabled"
          :oidc-provider-name="oidcOAuthProviderName"
          :wechat-enabled="wechatOAuthEnabled"
          :wechat-open-enabled="wechatOAuthOpenEnabled"
          :wechat-mp-enabled="wechatOAuthMPEnabled"
        />

        <div
          v-if="contactInfo"
          class="card border-primary-200 bg-surface-card p-6"
        >
          <div class="flex items-center gap-4">
            <div class="rounded-xl bg-primary-100 p-3 text-primary-600">
              <Icon name="chat" size="lg" />
            </div>
            <div>
              <h3 class="text-[22px] leading-tight text-ink">
                {{ t('common.contactSupport') }}
              </h3>
              <p class="text-sm font-medium text-body">{{ contactInfo }}</p>
            </div>
          </div>
        </div>

        <ProfilePasswordForm />

        <ProfileBalanceNotifyCard
          v-if="user && balanceLowNotifyEnabled"
          :enabled="user.balance_notify_enabled ?? true"
          :threshold="user.balance_notify_threshold"
          :extra-emails="user.balance_notify_extra_emails ?? []"
          :system-default-threshold="systemDefaultThreshold"
          :user-email="user.email"
        />

        <ProfileTotpCard />
      </template>

      <UserSubscriptionsPanel v-else />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import UserSubscriptionsPanel from '@/components/user/subscriptions/UserSubscriptionsPanel.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
type ProfileTab = 'account' | 'subscriptions'
const activeTab = ref<ProfileTab>(
  route.query.section === 'subscriptions' ? 'subscriptions' : 'account'
)
const tabs = computed(() => [
  { key: 'account' as const, label: t('profile.tabs.account') },
  { key: 'subscriptions' as const, label: t('profile.tabs.subscriptions') }
])

const contactInfo = ref('')
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const dingtalkOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')

function setActiveTab(tab: ProfileTab) {
  activeTab.value = tab
  router.replace({
    query: {
      ...route.query,
      section: tab === 'subscriptions' ? 'subscriptions' : undefined
    }
  })
}

watch(
  () => route.query.section,
  (section) => {
    activeTab.value = section === 'subscriptions' ? 'subscriptions' : 'account'
  }
)

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || ''
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>
