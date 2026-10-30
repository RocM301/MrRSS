import { describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import ArticleSummary from './ArticleSummary.vue';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key, locale: ref('en') }),
}));

describe('ArticleSummary manual recovery', () => {
  it('shows a manual generate action when an automatic AI summary is missing', async () => {
    const wrapper = mount(ArticleSummary, {
      props: {
        summaryResult: null,
        isLoadingSummary: false,
        translationEnabled: false,
        summaryProvider: 'ai',
        isLoadingContent: false,
      },
    });

    expect(wrapper.text()).toContain('setting.content.generateSummary');
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('setting.content.generateSummary'))!
      .trigger('click');
    expect(wrapper.emitted('generate-summary')).toHaveLength(1);
  });

  it('offers retry after summary generation fails', async () => {
    const wrapper = mount(ArticleSummary, {
      props: {
        summaryResult: {
          summary: '',
          sentence_count: 0,
          is_too_short: false,
          error: 'AI unavailable',
        },
        isLoadingSummary: false,
        translationEnabled: false,
        summaryProvider: 'ai',
      },
    });

    expect(wrapper.text()).toContain('AI unavailable');
    await wrapper.get('button[title="setting.content.regenerateSummary"]').trigger('click');
    expect(wrapper.emitted('generate-summary')).toHaveLength(1);
  });
});
