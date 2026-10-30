import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { setImageDragData } from './imageDrag';

describe('article image file dragging', () => {
  beforeEach(() => {
    vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue('Chrome/130.0.0.0');
  });
  afterEach(() => vi.restoreAllMocks());
  it('preserves native image file dragging in macOS WKWebView', () => {
    vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue('AppleWebKit/605.1.15');
    const image = document.createElement('img');
    image.src = 'https://example.org/photo.jpg';
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData({ dataTransfer } as unknown as DragEvent, image);
    expect(dataTransfer.setData).not.toHaveBeenCalled();
    expect(dataTransfer.effectAllowed).toBe('');
  });
  it('offers a proxy download with a safe original filename instead of a link shortcut', () => {
    const image = document.createElement('img');
    image.src = 'https://example.org/pictures/photo.jpg?token=secret';
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData(
      { dataTransfer } as unknown as DragEvent,
      image,
      'https://example.org/article'
    );
    expect(dataTransfer.effectAllowed).toBe('copy');
    const payload = dataTransfer.setData.mock.calls.find(([type]) => type === 'DownloadURL')?.[1];
    expect(payload).toContain('application/octet-stream:photo.jpg:');
    expect(payload).toContain('/api/media/proxy?url_b64=');
  });
  it('does not nest proxies and strips reserved filenames', () => {
    const image = document.createElement('img');
    image.src = `/api/media/proxy?url_b64=${btoa('https://example.org/CON.jpg')}`;
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData({ dataTransfer } as unknown as DragEvent, image);
    expect(dataTransfer.setData).toHaveBeenCalledWith(
      'DownloadURL',
      `application/octet-stream:image:${image.src}`
    );
  });
  it('does not offer non-HTTP resources as downloads', () => {
    const image = document.createElement('img');
    image.src = 'file:///private.png';
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData({ dataTransfer } as unknown as DragEvent, image);
    expect(dataTransfer.setData).not.toHaveBeenCalled();
  });
});
