<template>
  <div>
    <label class="mb-2 block text-sm font-medium text-body">
      {{ t("payment.paymentMethod") }}
    </label>
    <div class="grid grid-cols-2 gap-3 sm:flex">
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :disabled="!method.available"
        :class="[
          'relative flex h-[64px] min-w-0 flex-col items-center justify-center rounded-lg border px-3 text-center transition-all sm:flex-1',
          !method.available
            ? 'cursor-not-allowed border-hairline bg-surface-soft opacity-50 '
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'border-hairline bg-canvas text-body hover:border-muted-soft ',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="flex items-center gap-2">
          <img
            :src="methodIcon(method.type)"
            :alt="t(`payment.methods.${method.type}`)"
            class="h-7 w-7 object-contain"
          />
          <span class="flex min-w-0 flex-col items-start leading-none">
            <span class="text-sm font-semibold leading-tight">{{
              t(`payment.methods.${method.type}`)
            }}</span>
            <span
              v-if="method.fee_rate > 0"
              class="mt-0.5 text-[10px] tracking-wide text-muted"
            >
              {{ t("payment.fee") }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { METHOD_ORDER } from "./providerConfig";
import alipayIcon from "@/assets/icons/alipay.svg";
import wxpayIcon from "@/assets/icons/wxpay.svg";
import stripeIcon from "@/assets/icons/stripe.svg";
import airwallexIcon from "@/assets/icons/airwallex.svg";

export interface PaymentMethodOption {
  type: string;
  fee_rate: number;
  available: boolean;
}

const props = defineProps<{
  methods: PaymentMethodOption[];
  selected: string;
}>();

const emit = defineEmits<{
  select: [type: string];
}>();

const { t } = useI18n();

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
};

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER;
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type);
    const bi = order.indexOf(b.type);
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi);
  });
});

function methodIcon(type: string): string {
  if (type.includes("alipay")) return METHOD_ICONS.alipay;
  if (type.includes("wxpay")) return METHOD_ICONS.wxpay;
  if (type === "airwallex") return METHOD_ICONS.airwallex;
  return METHOD_ICONS[type] || alipayIcon;
}

function methodSelectedClass(type: string): string {
  if (type.includes("alipay"))
    return "border-primary-200 bg-surface-card text-ink shadow-card ";
  if (type.includes("wxpay"))
    return "border-primary-200 bg-surface-card text-ink shadow-card ";
  if (type === "stripe")
    return "border-primary-200 bg-surface-card text-ink shadow-card ";
  if (type === "airwallex")
    return "border-primary-200 bg-surface-card text-ink shadow-card ";
  return "border-primary-200 bg-surface-card text-ink shadow-card ";
}
</script>
