import { describe, expect, it } from 'vitest'
import {
  MUYIN_ALIPAY_CHANNEL_OPTIONS,
  MUYIN_WXPAY_CHANNEL_OPTIONS,
  PAYMENT_CURRENCY_OPTIONS,
  PROVIDER_CALLBACK_PATHS,
  PROVIDER_CONFIG_FIELDS,
  PROVIDER_SUPPORTED_TYPES,
} from '@/components/payment/providerConfig'

function findField(providerKey: string, key: string) {
  const fields = PROVIDER_CONFIG_FIELDS[providerKey] || []
  return fields.find(field => field.key === key)
}

describe('PROVIDER_CONFIG_FIELDS.wxpay', () => {
  it('keeps admin form validation aligned with backend-required credentials', () => {
    expect(findField('wxpay', 'publicKeyId')?.optional).toBeFalsy()
    expect(findField('wxpay', 'certSerial')?.optional).toBeFalsy()
  })

  it('only keeps the simplified visible credential set in the admin form', () => {
    expect(findField('wxpay', 'mpAppId')).toBeUndefined()
    expect(findField('wxpay', 'h5AppName')).toBeUndefined()
    expect(findField('wxpay', 'h5AppUrl')).toBeUndefined()
  })
})

describe('PROVIDER_CONFIG_FIELDS.airwallex', () => {
  it('adds currency config with CNY as the default', () => {
    const currency = findField('airwallex', 'currency')

    expect(currency?.defaultValue).toBe('CNY')
    expect(currency?.hintKey).toBe('admin.settings.payment.field_paymentCurrencyHint')
    expect(currency?.options).toBe(PAYMENT_CURRENCY_OPTIONS)
  })

  it('marks accountId as optional and explains when it can be left blank', () => {
    const accountId = findField('airwallex', 'accountId')

    expect(accountId?.optional).toBe(true)
    expect(accountId?.clearable).toBe(true)
    expect(accountId?.hintKey).toBe('admin.settings.payment.field_accountIdHint')
  })

  it('explains that apiBase must match the Airwallex key environment', () => {
    expect(findField('airwallex', 'apiBase')?.hintKey).toBe('admin.settings.payment.field_airwallexApiBaseHint')
  })
})

describe('PROVIDER_CONFIG_FIELDS.muyin', () => {
  it('supports the visible Alipay and WeChat Pay methods through one provider', () => {
    expect(PROVIDER_SUPPORTED_TYPES.muyin).toEqual(['alipay', 'wxpay'])
  })

  it('defaults to the documented API base and channel values', () => {
    expect(findField('muyin', 'apiBase')?.defaultValue).toBe('https://auth.muyin.site')
    expect(findField('muyin', 'alipayChannel')?.defaultValue).toBe('FACE_TO_FACE_PAYMENT')
    expect(findField('muyin', 'wxpayChannel')?.defaultValue).toBe('WECHATPAY_H5')
    expect(findField('muyin', 'alipayChannel')?.options).toBe(MUYIN_ALIPAY_CHANNEL_OPTIONS)
    expect(findField('muyin', 'wxpayChannel')?.options).toBe(MUYIN_WXPAY_CHANNEL_OPTIONS)
  })

  it('marks the bearer token as sensitive and configures callback paths', () => {
    expect(findField('muyin', 'token')?.sensitive).toBe(true)
    expect(PROVIDER_CALLBACK_PATHS.muyin).toEqual({
      notifyUrl: '/api/v1/payment/webhook/muyin',
      returnUrl: '/payment/result',
    })
  })

  it('requires admins to enter their MuYin username as the platform identifier', () => {
    const platform = findField('muyin', 'platform')

    expect(platform?.defaultValue).toBeUndefined()
    expect(platform?.hintKey).toBe('admin.settings.payment.field_muyinPlatformHint')
    expect(platform?.optional).toBeFalsy()
  })
})

describe('PROVIDER_CONFIG_FIELDS.stripe', () => {
  it('adds currency config with CNY as the default', () => {
    const currency = findField('stripe', 'currency')

    expect(currency?.defaultValue).toBe('CNY')
    expect(currency?.hintKey).toBe('admin.settings.payment.field_paymentCurrencyHint')
    expect(currency?.options).toBe(PAYMENT_CURRENCY_OPTIONS)
  })
})
