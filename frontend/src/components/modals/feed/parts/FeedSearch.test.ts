import { flushPromises, mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';
import FeedSearch from './FeedSearch.vue';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

describe('FeedSearch', () => {
  it('searches the feed directory and emits the selected feed', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => ({
        ok: true,
        json: async () => [
          {
            title: '网易新闻',
            url: 'https://example.com/rss.xml',
            description: '网易新闻 RSS',
            inactive: false,
          },
        ],
      }))
    );
    const wrapper = mount(FeedSearch);
    await wrapper.get('input').setValue('网易');
    await wrapper.get('button').trigger('click');
    await flushPromises();

    expect(fetch).toHaveBeenCalledWith('/api/feeds/search?q=%E7%BD%91%E6%98%93', {
      signal: expect.objectContaining({ aborted: false }),
    });
    await wrapper.get('ul button').trigger('click');
    expect(wrapper.emitted('select')?.[0]?.[0]).toMatchObject({
      title: '网易新闻',
      url: 'https://example.com/rss.xml',
    });
    wrapper.unmount();
    vi.unstubAllGlobals();
  });
});
