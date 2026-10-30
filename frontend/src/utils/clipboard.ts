/**
 * Clipboard utilities for MrRSS
 * Uses Wails v3 native Clipboard API
 */

import { Clipboard } from '@wailsio/runtime';

/**
 * Copy text to clipboard using Wails v3 native API
 * @param text Text to copy
 * @returns Promise that resolves to true if successful, false otherwise
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (!text) {
    console.warn('copyToClipboard: text is empty');
    return false;
  }

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // WebView clipboard permission may be unavailable; try the native API.
  }

  try {
    await Clipboard.SetText(text);
    return true;
  } catch {
    // HTTP browser sessions may expose neither the secure nor native clipboard.
  }

  const previousFocus = document.activeElement as HTMLElement | null;
  const input = document.createElement('textarea');
  input.value = text;
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  try {
    input.select();
    return document.execCommand('copy');
  } catch {
    return false;
  } finally {
    input.remove();
    previousFocus?.focus();
  }
}

export const copyText = copyToClipboard;

/**
 * Copy article URL to clipboard
 * @param url Article URL
 * @returns Promise that resolves to true if successful
 */
export async function copyArticleLink(url: string): Promise<boolean> {
  return copyToClipboard(url);
}

/**
 * Copy article title to clipboard
 * @param title Article title
 * @returns Promise that resolves to true if successful
 */
export async function copyArticleTitle(title: string): Promise<boolean> {
  return copyToClipboard(title);
}

/**
 * Copy feed URL to clipboard
 * @param url Feed URL
 * @returns Promise that resolves to true if successful
 */
export async function copyFeedURL(url: string): Promise<boolean> {
  return copyToClipboard(url);
}
