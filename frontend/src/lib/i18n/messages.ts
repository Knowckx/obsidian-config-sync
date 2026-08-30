// 2026-08-30 国际化功能已冻结，暂时不要继续开发或接入业务。

const zhMessages = {
  'common.finish': '完成',
  'common.next': '下一步',
  'common.previous': '上一步',
  'language.en': '英文',
  'language.zh': '中文',
} as const;

const enMessages = {
  'common.finish': 'Finish',
  'common.next': 'Next',
  'common.previous': 'Previous',
  'language.en': 'English',
  'language.zh': 'Chinese',
} as const satisfies Record<keyof typeof zhMessages, string>;

export const messages = {
  zh: zhMessages,
  en: enMessages,
} as const;
