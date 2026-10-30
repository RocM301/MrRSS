const invalidTranslationMarkers = [
  '无需翻译',
  '不需要翻译',
  '已经是简体中文',
  '已经是繁体中文',
  '原文：',
  '原文:',
  'already in simplified chinese',
  'already in traditional chinese',
  'no translation is needed',
  'no translation needed',
  'the original text is',
  'here is the translation',
];

export function isInvalidArticleTranslation(value?: string): boolean {
  const normalized = value?.trim().toLocaleLowerCase();
  return !!normalized && invalidTranslationMarkers.some((marker) => normalized.includes(marker));
}
