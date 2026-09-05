// 2026-08-30 国际化功能已冻结，暂时不要继续开发或接入业务。

import { createI18n } from './create_i18n.svelte';
import { messages } from './messages';

// i18n 是当前项目唯一的国际化实例。
export const i18n = createI18n({
  messages,
  fallbackLocale: 'en',
  storageKey: 'obsidian-config-sync.locale',
});

export type AppLocale = keyof typeof messages;
export type AppMessageKey = keyof (typeof messages)['zh'];
