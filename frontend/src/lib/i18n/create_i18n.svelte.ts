// 2026-08-30 国际化功能已冻结，暂时不要继续开发或接入业务。

type MessageTable = Record<string, string>;
type MessageCatalog = Record<string, MessageTable>;
type TranslationParam = string | number;
type TranslationParams = Record<string, TranslationParam>;

export type I18nLocale<TMessages extends MessageCatalog> = Extract<keyof TMessages, string>;
export type I18nMessageKey<TMessages extends MessageCatalog> = Extract<
  keyof TMessages[I18nLocale<TMessages>],
  string
>;

type CreateI18nOptions<TMessages extends MessageCatalog> = {
  messages: TMessages;
  fallbackLocale: I18nLocale<TMessages>;
  storageKey: string;
};

// createI18n 根据项目文字表创建独立的响应式国际化实例。
export function createI18n<const TMessages extends MessageCatalog>(
  options: CreateI18nOptions<TMessages>,
) {
  type Locale = I18nLocale<TMessages>;
  type MessageKey = I18nMessageKey<TMessages>;

  const locales = Object.keys(options.messages) as Locale[];
  const state = $state({
    locale: getInitialLocale(),
  });

  // setLocale 切换并持久化当前语言。
  function setLocale(locale: Locale): void {
    state.locale = locale;
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(options.storageKey, locale);
    }
  }

  // toggleLocale 按项目语言表顺序切换到下一种语言。
  function toggleLocale(): void {
    const currentIndex = locales.indexOf(state.locale);
    setLocale(locales[(currentIndex + 1) % locales.length]);
  }

  // t 返回当前语言的文案并替换动态参数。
  function t(key: MessageKey, params: TranslationParams = {}): string {
    const table = options.messages[state.locale] as Record<MessageKey, string>;
    const template = table[key];

    return template.replace(/\{(\w+)\}/g, (placeholder, name: string) => {
      const value = params[name];
      return value === undefined ? placeholder : String(value);
    });
  }

  // getInitialLocale 优先读取用户设置，未设置时使用系统语言。
  function getInitialLocale(): Locale {
    if (typeof localStorage !== 'undefined') {
      const savedLocale = localStorage.getItem(options.storageKey);
      const matchedLocale = locales.find((locale) => locale === savedLocale);
      if (matchedLocale) {
        return matchedLocale;
      }
    }
    return detectSystemLocale();
  }

  // detectSystemLocale 将系统首选语言匹配到项目支持的语言。
  function detectSystemLocale(): Locale {
    for (const language of getSystemLanguages()) {
      const normalizedLanguage = language.toLowerCase();
      const exactLocale = locales.find(
        (locale) => locale.toLowerCase() === normalizedLanguage,
      );
      if (exactLocale) {
        return exactLocale;
      }

      const baseLanguage = normalizedLanguage.split('-')[0];
      const baseLocale = locales.find(
        (locale) => locale.toLowerCase().split('-')[0] === baseLanguage,
      );
      if (baseLocale) {
        return baseLocale;
      }
    }
    return options.fallbackLocale;
  }

  return {
    state,
    locales,
    setLocale,
    toggleLocale,
    t,
  };
}

// getSystemLanguages 读取浏览器报告的系统首选语言。
function getSystemLanguages(): readonly string[] {
  if (typeof navigator === 'undefined') {
    return [];
  }
  if (navigator.languages.length > 0) {
    return navigator.languages;
  }
  return navigator.language ? [navigator.language] : [];
}
