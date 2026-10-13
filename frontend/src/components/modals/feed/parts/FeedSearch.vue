<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhMagnifyingGlass, PhSpinner } from '@phosphor-icons/vue';
interface Result {
  title: string;
  url: string;
  description: string;
  inactive: boolean;
}
const emit = defineEmits<{ select: [result: Result] }>();
const { t } = useI18n();
const query = ref('');
const results = ref<Result[]>([]);
const loading = ref(false);
const searched = ref(false);
const error = ref(false);
let controller: ReturnType<typeof globalThis.AbortController> | undefined;
async function search() {
  controller?.abort();
  if (!query.value.trim()) return;
  const current = new globalThis.AbortController();
  controller = current;
  loading.value = true;
  error.value = false;
  searched.value = false;
  results.value = [];
  try {
    const response = await fetch(`/api/feeds/search?q=${encodeURIComponent(query.value.trim())}`, {
      signal: current.signal,
    });
    if (!response.ok) throw new Error('Search failed');
    const data: Result[] = await response.json();
    if (controller !== current) return;
    results.value = data;
    searched.value = true;
  } catch {
    if (!current.signal.aborted) error.value = true;
  } finally {
    if (controller === current) loading.value = false;
  }
}
function select(result: Result) {
  emit('select', result);
  results.value = [];
  searched.value = false;
}
onBeforeUnmount(() => controller?.abort());
</script>
<template>
  <div class="mb-4">
    <label
      for="rss-directory-query"
      class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary"
      >{{ t('modal.feed.searchRSS') }}</label
    >
    <div class="flex gap-2">
      <input
        id="rss-directory-query"
        v-model="query"
        maxlength="200"
        type="search"
        :placeholder="t('modal.feed.searchRSSPlaceholder')"
        class="min-w-0 flex-1 p-2 border border-border rounded-md bg-bg-tertiary text-text-primary text-sm focus:border-accent focus:outline-none"
        @keydown.enter.prevent="search"
      />
      <button
        type="button"
        :disabled="!query.trim() || loading"
        :title="t('modal.feed.searchRSS')"
        :aria-label="t('modal.feed.searchRSS')"
        class="shrink-0 w-10 h-10 flex items-center justify-center rounded-md bg-accent text-white disabled:opacity-50"
        @click="search"
      >
        <PhSpinner v-if="loading" class="w-5 h-5 animate-spin" />
        <PhMagnifyingGlass v-else class="w-5 h-5" />
      </button>
    </div>
    <div aria-live="polite" class="text-xs mt-2">
      <p v-if="error" class="text-red-500">{{ t('modal.feed.searchRSSError') }}</p>
      <p v-else-if="searched && !results.length" class="text-text-secondary">
        {{ t('modal.feed.searchRSSEmpty') }}
      </p>
    </div>
    <ul
      v-if="results.length"
      class="mt-2 max-h-64 overflow-y-auto divide-y divide-border border border-border rounded-md"
    >
      <li v-for="result in results" :key="result.url">
        <button
          type="button"
          class="w-full text-left p-3 hover:bg-bg-tertiary transition-colors"
          @click="select(result)"
        >
          <span class="block text-sm font-semibold text-text-primary break-words"
            >{{ result.title }}
            <span v-if="result.inactive" class="text-xs font-normal text-text-tertiary">{{
              t('modal.feed.searchRSSInactive')
            }}</span></span
          >
          <span
            v-if="result.description"
            class="block text-xs text-text-secondary line-clamp-2 mt-1"
            >{{ result.description }}</span
          >
          <span class="block text-xs text-accent break-all mt-1">{{ result.url }}</span>
        </button>
      </li>
    </ul>
  </div>
</template>
