import { nextTick } from 'vue';
import katex from 'katex';
import hljs from 'highlight.js/lib/core';

// Import commonly used languages
import javascript from 'highlight.js/lib/languages/javascript';
import typescript from 'highlight.js/lib/languages/typescript';
import python from 'highlight.js/lib/languages/python';
import java from 'highlight.js/lib/languages/java';
import cpp from 'highlight.js/lib/languages/cpp';
import c from 'highlight.js/lib/languages/c';
import csharp from 'highlight.js/lib/languages/csharp';
import go from 'highlight.js/lib/languages/go';
import rust from 'highlight.js/lib/languages/rust';
import ruby from 'highlight.js/lib/languages/ruby';
import php from 'highlight.js/lib/languages/php';
import swift from 'highlight.js/lib/languages/swift';
import kotlin from 'highlight.js/lib/languages/kotlin';
import scala from 'highlight.js/lib/languages/scala';
import sql from 'highlight.js/lib/languages/sql';
import bash from 'highlight.js/lib/languages/bash';
import shell from 'highlight.js/lib/languages/shell';
import powershell from 'highlight.js/lib/languages/powershell';
import json from 'highlight.js/lib/languages/json';
import xml from 'highlight.js/lib/languages/xml';
import yaml from 'highlight.js/lib/languages/yaml';
import markdown from 'highlight.js/lib/languages/markdown';
import css from 'highlight.js/lib/languages/css';
import scss from 'highlight.js/lib/languages/scss';
import less from 'highlight.js/lib/languages/less';
import dockerfile from 'highlight.js/lib/languages/dockerfile';
import nginx from 'highlight.js/lib/languages/nginx';
import ini from 'highlight.js/lib/languages/ini';
import diff from 'highlight.js/lib/languages/diff';
import plaintext from 'highlight.js/lib/languages/plaintext';

// Register languages
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('js', javascript);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('ts', typescript);
hljs.registerLanguage('python', python);
hljs.registerLanguage('py', python);
hljs.registerLanguage('java', java);
hljs.registerLanguage('cpp', cpp);
hljs.registerLanguage('c', c);
hljs.registerLanguage('csharp', csharp);
hljs.registerLanguage('cs', csharp);
hljs.registerLanguage('go', go);
hljs.registerLanguage('rust', rust);
hljs.registerLanguage('rs', rust);
hljs.registerLanguage('ruby', ruby);
hljs.registerLanguage('rb', ruby);
hljs.registerLanguage('php', php);
hljs.registerLanguage('swift', swift);
hljs.registerLanguage('kotlin', kotlin);
hljs.registerLanguage('kt', kotlin);
hljs.registerLanguage('scala', scala);
hljs.registerLanguage('sql', sql);
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('sh', bash);
hljs.registerLanguage('shell', shell);
hljs.registerLanguage('powershell', powershell);
hljs.registerLanguage('ps1', powershell);
hljs.registerLanguage('json', json);
hljs.registerLanguage('xml', xml);
hljs.registerLanguage('html', xml);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('yml', yaml);
hljs.registerLanguage('markdown', markdown);
hljs.registerLanguage('md', markdown);
hljs.registerLanguage('css', css);
hljs.registerLanguage('scss', scss);
hljs.registerLanguage('less', less);
hljs.registerLanguage('dockerfile', dockerfile);
hljs.registerLanguage('docker', dockerfile);
hljs.registerLanguage('nginx', nginx);
hljs.registerLanguage('ini', ini);
hljs.registerLanguage('diff', diff);
hljs.registerLanguage('plaintext', plaintext);
hljs.registerLanguage('text', plaintext);

/**
 * Composable for enhanced article content rendering
 * Handles math formulas, code syntax highlighting, and other advanced rendering
 */
export function useArticleRendering() {
  /**
   * Render math formulas in the content
   * Supports multiple formats:
   * - Display math: $$...$$ or \[...\]
   * - Inline math: $...$ or \(...\)
   */
  function renderMathFormulas(container: HTMLElement) {
    if (!container) return;

    try {
      // First, handle pre-existing math elements with class 'math' or 'MathJax'
      const existingMathElements = container.querySelectorAll(
        '.math, .MathJax, [data-math], script[type*="math"]'
      );
      existingMathElements.forEach((el) => {
        if (el.classList.contains('katex') || el.querySelector('.katex')) return;

        const mathContent = el.getAttribute('data-math') || el.textContent || '';
        if (!mathContent.trim()) return;

        try {
          const isDisplay =
            el.classList.contains('math-display') ||
            el.classList.contains('display') ||
            el.tagName === 'DIV';
          const mathElement = document.createElement(isDisplay ? 'div' : 'span');
          mathElement.className = isDisplay ? 'katex-display' : 'katex-inline';
          katex.render(mathContent, mathElement, {
            displayMode: isDisplay,
            throwOnError: false,
            strict: false,
          });
          el.replaceWith(mathElement);
        } catch (e) {
          console.error('Error rendering existing math element:', e);
        }
      });

      // Find all text nodes that might contain math
      const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
        acceptNode: (node) => {
          // Skip if already inside a math element
          const parent = node.parentElement;
          if (
            parent?.classList.contains('katex') ||
            parent?.classList.contains('katex-display') ||
            parent?.classList.contains('katex-inline') ||
            parent?.closest('.katex') ||
            parent?.closest('.katex-display') ||
            parent?.tagName === 'CODE' ||
            parent?.tagName === 'PRE' ||
            parent?.tagName === 'SCRIPT' ||
            parent?.tagName === 'STYLE'
          ) {
            return NodeFilter.FILTER_REJECT;
          }
          // Accept nodes that contain math delimiters
          const text = node.textContent || '';
          if (text.includes('$') || text.includes('\\(') || text.includes('\\[')) {
            return NodeFilter.FILTER_ACCEPT;
          }
          return NodeFilter.FILTER_REJECT;
        },
      });

      const nodesToProcess: Text[] = [];
      let currentNode: Node | null;
      while ((currentNode = walker.nextNode())) {
        nodesToProcess.push(currentNode as Text);
      }

      // Process each text node
      for (const node of nodesToProcess) {
        const text = node.textContent || '';
        if (!text) continue;

        const fragments: (string | HTMLElement)[] = [];
        let lastIndex = 0;

        // Match all math patterns
        // Order matters: longer/more specific patterns first
        const mathPatterns = [
          // Display math: $$...$$ (must not be empty)
          { regex: /\$\$([^$]+)\$\$/g, isDisplay: true },
          // Display math: \[...\]
          { regex: /\\\[([^\]]+)\\\]/g, isDisplay: true },
          // Inline math: \(...\)
          { regex: /\\\(([^)]+)\\\)/g, isDisplay: false },
          // Inline math: $...$ (single line, not empty, not currency like $199)
          { regex: /\$(?![+-]?\d)([^\s$][^$\n]*[^\s$]|\S)\$/g, isDisplay: false },
        ];

        // Collect all matches with their positions
        const allMatches: Array<{
          start: number;
          end: number;
          math: string;
          isDisplay: boolean;
        }> = [];

        for (const { regex, isDisplay } of mathPatterns) {
          let match;
          regex.lastIndex = 0; // Reset regex state
          while ((match = regex.exec(text)) !== null) {
            // Check if this position is already covered by another match
            const isOverlapping = allMatches.some(
              (m) =>
                (match!.index >= m.start && match!.index < m.end) ||
                (match!.index + match![0].length > m.start &&
                  match!.index + match![0].length <= m.end)
            );
            if (!isOverlapping) {
              allMatches.push({
                start: match.index,
                end: match.index + match[0].length,
                math: match[1],
                isDisplay,
              });
            }
          }
        }

        // Sort matches by position
        allMatches.sort((a, b) => a.start - b.start);

        // Build fragments
        for (const { start, end, math, isDisplay } of allMatches) {
          // Add text before match
          if (start > lastIndex) {
            fragments.push(text.substring(lastIndex, start));
          }

          // Render math
          try {
            const mathElement = document.createElement(isDisplay ? 'div' : 'span');
            mathElement.className = isDisplay ? 'katex-display' : 'katex-inline';
            katex.render(math.trim(), mathElement, {
              displayMode: isDisplay,
              throwOnError: false,
              strict: false,
            });
            fragments.push(mathElement);
          } catch (e) {
            console.error('Error rendering math:', e, 'Content:', math);
            // On error, keep the original text
            fragments.push(text.substring(start, end));
          }

          lastIndex = end;
        }

        // Add remaining text
        if (lastIndex < text.length) {
          fragments.push(text.substring(lastIndex));
        }

        // Only replace if we found any math
        if (fragments.length > 0 && allMatches.length > 0) {
          const parent = node.parentNode;
          if (parent) {
            // Insert fragments
            for (const fragment of fragments) {
              if (typeof fragment === 'string') {
                parent.insertBefore(document.createTextNode(fragment), node);
              } else {
                parent.insertBefore(fragment, node);
              }
            }
            // Remove original text node
            parent.removeChild(node);
          }
        }
      }
    } catch (e) {
      console.error('Error processing math formulas:', e);
    }
  }

  /**
   * Apply syntax highlighting to code blocks
   */
  function highlightCodeBlocks(container: HTMLElement) {
    if (!container) return;

    try {
      // Find all code blocks within pre elements
      const codeBlocks = container.querySelectorAll('pre code');

      codeBlocks.forEach((block) => {
        // Skip if already highlighted
        if (block.classList.contains('hljs')) return;

        const codeElement = block as HTMLElement;

        // Try to detect language from class
        let language = '';
        const classList = Array.from(codeElement.classList);
        for (const cls of classList) {
          if (cls.startsWith('language-') || cls.startsWith('lang-')) {
            language = cls.replace(/^(language-|lang-)/, '');
            break;
          }
        }

        // Highlight the code
        if (language && hljs.getLanguage(language)) {
          // Use specific language
          const result = hljs.highlight(codeElement.textContent || '', { language });
          codeElement.innerHTML = result.value;
          codeElement.classList.add('hljs');
        } else {
          // Auto-detect language
          const result = hljs.highlightAuto(codeElement.textContent || '');
          codeElement.innerHTML = result.value;
          codeElement.classList.add('hljs');
          if (result.language) {
            codeElement.dataset.detectedLanguage = result.language;
          }
        }
      });

      // Also handle standalone pre elements without code tags
      const preElements = container.querySelectorAll('pre:not(:has(code))');
      preElements.forEach((pre) => {
        if (pre.classList.contains('hljs')) return;

        const preElement = pre as HTMLElement;
        const result = hljs.highlightAuto(preElement.textContent || '');
        preElement.innerHTML = result.value;
        preElement.classList.add('hljs');
      });
    } catch (e) {
      console.error('Error highlighting code blocks:', e);
    }
  }

  /**
   * Apply all rendering enhancements to the container
   */
  async function enhanceRendering(containerSelector: string = '.prose-content') {
    await nextTick();
    const container = document.querySelector(containerSelector) as HTMLElement;
    if (!container) return;

    // Render math formulas
    renderMathFormulas(container);

    // Apply syntax highlighting to code blocks
    highlightCodeBlocks(container);
  }

  return {
    renderMathFormulas,
    highlightCodeBlocks,
    enhanceRendering,
  };
}
