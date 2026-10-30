import { ref, watch, onBeforeUnmount } from 'vue';
import type { Article } from '@/types/models';
import { isMediaCacheEnabled, proxyImagesInHtml } from '@/utils/mediaProxy';
import { invalidateArticleContent } from '@/utils/articleContentCache';

interface Options {
  article: () => Article;
  enabled: () => boolean;
  automatic: () => boolean;
  empty?: () => boolean;
  loading: () => boolean;
  onContent: (content: string) => void | Promise<void>;
  onError: () => void;
  onSuccess: () => void;
}

export function useFullArticle(options: Options) {
  const content = ref('');
  const loading = ref(false);
  let generation = 0;
  let controller: AbortController | null = null;
  let attempted: number | null = null;
  const recovered = new Map<number, string>();
  const finished = ref(false);

  function reset() {
    generation += 1;
    controller?.abort();
    controller = null;
    content.value = recovered.get(options.article()?.id) || '';
    finished.value = !!content.value;
    loading.value = false;
    attempted = null;
  }

  async function fetchFullArticle(showErrors = true) {
    const article = options.article();
    if (!article?.id || !options.enabled() || loading.value) return;
    const requestId = ++generation;
    attempted = article.id;
    loading.value = true;
    finished.value = false;
    const current = () => requestId === generation && options.article()?.id === article.id;
    const attempts = showErrors ? 1 : 3;
    let fetched = false;
    for (let attempt = 0; attempt < attempts && current(); attempt++) {
      controller?.abort();
      controller = new AbortController();
      try {
        const response = await fetch(`/api/articles/fetch-full?id=${article.id}`, {
          method: 'POST',
          signal: controller.signal,
        });
        if (!response.ok) throw new Error(`Full article: ${response.status}`);
        const data = await response.json();
        invalidateArticleContent(article.id);
        const cacheEnabled = await isMediaCacheEnabled();
        if (!current()) return;
        if (typeof data.content !== 'string' || !data.content.trim())
          throw new Error('Empty article');
        content.value = cacheEnabled
          ? proxyImagesInHtml(data.content, data.feed_url || article.url)
          : data.content;
        recovered.set(article.id, content.value);
        if (recovered.size > 100) recovered.delete(recovered.keys().next().value!);
        fetched = true;
        if (showErrors) options.onSuccess();
        await options.onContent(content.value);
        break;
      } catch {
        if (!current() || controller.signal.aborted) return;
        if (attempt < attempts - 1) {
          await new Promise((resolve) => window.setTimeout(resolve, 500 * (attempt + 1)));
        }
      }
    }
    if (current() && !fetched && showErrors) options.onError();
    if (current()) {
      loading.value = false;
      finished.value = true;
    }
  }

  watch(() => options.article()?.id, reset, { flush: 'sync' });
  watch(
    () => [
      options.article()?.id,
      options.enabled(),
      options.automatic(),
      options.loading(),
      options.empty?.(),
    ],
    () => {
      if (
        options.enabled() &&
        (options.automatic() || options.empty?.()) &&
        !options.loading() &&
        attempted !== options.article()?.id &&
        !content.value
      )
        void fetchFullArticle(false);
    },
    { immediate: true, flush: 'post' }
  );
  function onReload(event: Event) {
    if ((event as CustomEvent<number>).detail === options.article()?.id) {
      recovered.delete(options.article().id);
      reset();
    }
  }
  window.addEventListener('article-content-reloaded', onReload);
  onBeforeUnmount(() => {
    reset();
    window.removeEventListener('article-content-reloaded', onReload);
  });
  return {
    fullArticleFetchFinished: finished,
    fullArticleContent: content,
    isFetchingFullArticle: loading,
    fetchFullArticle,
  };
}
