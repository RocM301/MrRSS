import { ref, onBeforeUnmount } from 'vue';

// localStorage keys for persisting the user's manually-dragged panel widths so
// they survive reloads, layout/settings changes and app restarts.
const LS_SIDEBAR_WIDTH = 'mrrss.sidebarWidth';
const LS_ARTICLE_WIDTH = 'mrrss.articleListWidth';

function readSavedWidth(key: string): number | null {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return null;
    const n = Number(raw);
    return Number.isFinite(n) && n > 0 ? n : null;
  } catch {
    return null;
  }
}

export function useResizablePanels() {
  const savedSidebar = readSavedWidth(LS_SIDEBAR_WIDTH);
  const savedArticle = readSavedWidth(LS_ARTICLE_WIDTH);

  const sidebarWidth = ref<number>(Math.min(500, Math.max(180, savedSidebar ?? 256)));
  const articleListWidth = ref<number>(savedArticle ?? 350);
  const isResizingSidebar = ref<boolean>(false);
  const isResizingArticleList = ref<boolean>(false);
  const compactMode = ref<boolean>(false);

  let previousCursor = '';
  let previousSelect = '';

  const storageKey = () => `articleListWidth:${compactMode.value ? 'compact' : 'normal'}`;

  function restoreArticleListWidth(): void {
    const fallback = compactMode.value ? 500 : 350;
    try {
      const saved = localStorage.getItem(storageKey()) ?? localStorage.getItem(LS_ARTICLE_WIDTH);
      const width = saved === null ? fallback : Number(saved);
      articleListWidth.value =
        Number.isFinite(width) && width > 0
          ? Math.min(compactMode.value ? 800 : 600, Math.max(compactMode.value ? 300 : 280, width))
          : fallback;
    } catch {
      articleListWidth.value = fallback;
    }
  }

  // Track initial mouse position when starting resize
  const initialMouseX = ref<number>(0);
  const initialArticleListWidth = ref<number>(400);

  function setCompactMode(enabled: boolean): void {
    stopResizeArticleList();
    compactMode.value = enabled;
    restoreArticleListWidth();
  }

  restoreArticleListWidth();

  // Sidebar resize handlers
  function startResizeSidebar(): void {
    isResizingSidebar.value = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeSidebar);
    window.addEventListener('mouseup', stopResizeSidebar);
  }

  function handleResizeSidebar(event: MouseEvent): void {
    if (!isResizingSidebar.value) return;
    const newWidth = event.clientX;
    if (newWidth >= 180 && newWidth <= 450) {
      sidebarWidth.value = newWidth;
    }
  }

  function stopResizeSidebar(): void {
    isResizingSidebar.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener('mousemove', handleResizeSidebar);
    window.removeEventListener('mouseup', stopResizeSidebar);
    // Persist the chosen sidebar width
    try {
      localStorage.setItem(LS_SIDEBAR_WIDTH, String(Math.round(sidebarWidth.value)));
    } catch {
      // ignore storage errors
    }
  }

  // Article list resize handlers
  function startResizeArticleList(event: MouseEvent): void {
    if (event.button !== 0 || isResizingArticleList.value) return;
    event.preventDefault();
    isResizingArticleList.value = true;
    // Store initial mouse position and article list width
    initialMouseX.value = event.clientX;
    const renderedWidth =
      event.currentTarget instanceof HTMLElement
        ? event.currentTarget.previousElementSibling?.getBoundingClientRect().width
        : undefined;
    initialArticleListWidth.value =
      renderedWidth && renderedWidth > 0 ? renderedWidth : articleListWidth.value;
    previousCursor = document.body.style.cursor;
    previousSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', handleResizeArticleList);
    window.addEventListener('mouseup', stopResizeArticleList);
    window.addEventListener('blur', stopResizeArticleList);
  }

  function handleResizeArticleList(event: MouseEvent): void {
    if (!isResizingArticleList.value) return;
    const currentMouseX = event.clientX;
    // Calculate the delta from the initial position and apply to initial width
    const deltaX = currentMouseX - initialMouseX.value;
    const newWidth = initialArticleListWidth.value + deltaX;
    // In compact mode, allow wider range (300-800), in normal mode (250-600)
    const minWidth = compactMode.value ? 300 : 280;
    const maxWidth = compactMode.value ? 800 : 600;
    if (newWidth >= minWidth && newWidth <= maxWidth) {
      articleListWidth.value = newWidth;
    }
  }

  function stopResizeArticleList(): void {
    if (!isResizingArticleList.value) return;
    isResizingArticleList.value = false;
    document.body.style.cursor = previousCursor;
    document.body.style.userSelect = previousSelect;
    window.removeEventListener('mousemove', handleResizeArticleList);
    window.removeEventListener('mouseup', stopResizeArticleList);
    window.removeEventListener('blur', stopResizeArticleList);
    try {
      localStorage.setItem(storageKey(), String(articleListWidth.value));
    } catch {
      // Keep resizing usable when browser storage is unavailable.
    }
  }

  // Cleanup
  onBeforeUnmount(() => {
    stopResizeArticleList();
    window.removeEventListener('mousemove', handleResizeSidebar);
    window.removeEventListener('mouseup', stopResizeSidebar);
    window.removeEventListener('mousemove', handleResizeArticleList);
    window.removeEventListener('mouseup', stopResizeArticleList);
  });

  return {
    sidebarWidth,
    articleListWidth,
    startResizeSidebar,
    startResizeArticleList,
    setCompactMode,
  };
}
