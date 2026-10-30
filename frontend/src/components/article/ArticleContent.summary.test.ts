import { afterEach, describe, expect, it, vi } from 'vitest';
import { shallowMount, flushPromises } from '@vue/test-utils';
import { ref } from 'vue';
import ArticleContent from './ArticleContent.vue';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key, locale: ref('en') }),
}));
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    feeds: [{ id: 1, auto_expand_content: 'enabled' }],
    updateArticleSummary: vi.fn(),
  }),
}));
vi.mock('@/composables/core/useSettings', () => ({
  useSettings: () => ({
    settings: ref({ full_text_fetch_enabled: true }),
    fetchSettings: async () => ({}),
  }),
}));
vi.mock('@/composables/article/useArticleRendering', () => ({
  useArticleRendering: () => ({
    enhanceRendering: vi.fn(),
    renderMathFormulas: vi.fn(),
    highlightCodeBlocks: vi.fn(),
  }),
}));
vi.mock('@/composables/article/useArticleTranslation', () => ({
  useArticleTranslation: () => ({
    translationSettings: ref({ enabled: false }),
    loadTranslationSettings: async () => {},
  }),
}));
vi.mock('@/utils/mediaProxy', () => ({ isMediaCacheEnabled: async () => false }));

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe('automatic summary with full text', () => {
  it('starts again after returning to cached full text with an interrupted summary', async () => {
    vi.useFakeTimers();
    const calls: number[] = [];
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string, options?: RequestInit) => {
        let data: unknown = {};
        if (url === '/api/settings')
          data = {
            summary_enabled: 'true',
            summary_provider: 'ai',
            summary_trigger_mode: 'auto',
          };
        if (url.includes('fetch-full')) data = { content: '<p>Full article body</p>' };
        if (url.includes('summarize')) {
          calls.push(JSON.parse(options!.body as string).article_id);
          return new Promise((_resolve, reject) => {
            options?.signal?.addEventListener('abort', () =>
              reject(new DOMException('Aborted', 'AbortError'))
            );
          });
        }
        return { ok: true, json: async () => data };
      })
    );
    const article = (id: number) =>
      ({ id, feed_id: 1, title: 'Title', url: 'https://example.com' }) as Article;
    const wrapper = shallowMount(ArticleContent, {
      props: {
        article: article(1),
        articleContent: '<p>RSS content</p>',
        isLoadingContent: false,
      },
    });
    await flushPromises();
    await vi.advanceTimersByTimeAsync(300);
    await flushPromises();
    expect(calls).toContain(1);
    await wrapper.setProps({ article: article(2) });
    await flushPromises();
    await wrapper.setProps({ article: article(1) });
    await flushPromises();
    expect(calls.filter((id) => id === 1)).toHaveLength(2);
    wrapper.unmount();
    await flushPromises();
  });
});
