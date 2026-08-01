import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { useArticleRendering } from './useArticleRendering';

// Mock KaTeX
vi.mock('katex', () => ({
  default: {
    render: vi.fn((formula, element) => {
      element.innerHTML = `<span class="katex">${formula}</span>`;
    }),
  },
}));

describe('useArticleRendering', () => {
  let container: HTMLElement;

  beforeEach(() => {
    // Create a fresh container for each test
    container = document.createElement('div');
    container.className = 'prose-content';
    document.body.appendChild(container);
  });

  afterEach(() => {
    // Clean up
    document.body.removeChild(container);
  });

  describe('renderMathFormulas', () => {
    it('should render inline math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>The formula is $E = mc^2$ here.</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('E = mc^2');
    });

    it('should render display math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>$$\\int_{0}^{\\infty} e^{-x} dx$$</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
    });

    it('should not render math inside code blocks', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<pre><code>$E = mc^2$</code></pre>';
      const originalHTML = container.innerHTML;
      renderMathFormulas(container);

      // HTML should remain unchanged
      expect(container.innerHTML).toBe(originalHTML);
    });

    it('should handle text without math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>This is just plain text.</p>';
      const originalHTML = container.innerHTML;
      renderMathFormulas(container);

      // HTML should remain unchanged
      expect(container.innerHTML).toBe(originalHTML);
    });

    it('should not render dollar prices as math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>Apple TV: $199, up from $129 (+$70)</p>';
      const originalHTML = container.innerHTML;
      renderMathFormulas(container);

      expect(container.innerHTML).toBe(originalHTML);
      expect(container.innerHTML).not.toContain('katex-inline');
    });

    it('should handle mixed content with both inline and display math', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>Inline $a = b$ and display $$c = d$$</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('katex-display');
    });

    it('should not crash on empty container', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '';
      expect(() => renderMathFormulas(container)).not.toThrow();
    });

    it('should skip math in script tags', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<script>var x = "$E = mc^2$";</script><p>Normal text</p>';
      const scriptContent = container.querySelector('script')?.textContent;

      renderMathFormulas(container);

      // Script content should remain unchanged
      expect(container.querySelector('script')?.textContent).toBe(scriptContent);
    });
  });

  describe('enhanceRendering', () => {
    it('should enhance rendering for the specified container', async () => {
      const { enhanceRendering } = useArticleRendering();

      container.innerHTML = '<p>The equation is $E = mc^2$.</p>';
      await enhanceRendering('.prose-content');

      expect(container.innerHTML).toContain('katex');
    });

    it('should handle non-existent container gracefully', async () => {
      const { enhanceRendering } = useArticleRendering();

      // Should not throw error
      await expect(enhanceRendering('.non-existent')).resolves.not.toThrow();
    });
  });
});
