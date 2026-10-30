import { describe, expect, it } from 'vitest';
import { isInvalidArticleTranslation } from './articleTranslation';

describe('isInvalidArticleTranslation', () => {
  it('recognizes explanatory AI responses', () => {
    expect(isInvalidArticleTranslation('这段文字已经是简体中文了，无需翻译。原文：标题')).toBe(
      true
    );
    expect(isInvalidArticleTranslation('The text is already in Simplified Chinese')).toBe(true);
  });

  it('keeps actual translated titles', () => {
    expect(isInvalidArticleTranslation('群晖 NAS 遭勒索病毒攻击')).toBe(false);
    expect(isInvalidArticleTranslation(undefined)).toBe(false);
  });
});
