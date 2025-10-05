export const DEFAULT_HTTP_TIMEOUT_MS = 120_000;

export const requestWithTimeout = async (
  input: RequestInfo | URL,
  init: RequestInit = {},
  timeout = DEFAULT_HTTP_TIMEOUT_MS,
): Promise<Response> => {
  const controller = new AbortController();
  const { signal, ...rest } = init;

  let timedOut = false;
  const timeoutId = window.setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeout);

  const abortHandler = () => controller.abort();
  signal?.addEventListener?.('abort', abortHandler);

  try {
    return await fetch(input, { ...rest, signal: controller.signal });
  } catch (error) {
    if (timedOut || (error instanceof DOMException && error.name === 'AbortError')) {
      const timeoutError = new Error('请求超时，请稍后重试');
      timeoutError.name = 'TimeoutError';
      throw timeoutError;
    }
    throw error;
  } finally {
    clearTimeout(timeoutId);
    signal?.removeEventListener?.('abort', abortHandler);
  }
};
