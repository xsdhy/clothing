export const DEFAULT_HTTP_TIMEOUT_MS = 120_000;

import { emitUnauthorized, getStoredToken } from "./authStorage";

export const requestWithTimeout = async (
  input: RequestInfo | URL,
  init: RequestInit = {},
  timeout = DEFAULT_HTTP_TIMEOUT_MS,
): Promise<Response> => {
  const controller = new AbortController();
  const { signal, headers, ...rest } = init;

  const requestHeaders = new Headers(headers ?? {});
  const token = getStoredToken();
  if (token && !requestHeaders.has("Authorization")) {
    requestHeaders.set("Authorization", `Bearer ${token}`);
  }

  let timedOut = false;
  const timeoutId = window.setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeout);

  const abortHandler = () => controller.abort();
  signal?.addEventListener?.("abort", abortHandler);

  try {
    const response = await fetch(input, {
      ...rest,
      headers: requestHeaders,
      signal: controller.signal,
    });
    if (response.status === 401) {
      emitUnauthorized();
    }
    return response;
  } catch (error) {
    if (
      timedOut ||
      (error instanceof DOMException && error.name === "AbortError")
    ) {
      const timeoutError = new Error("请求超时，请稍后重试");
      timeoutError.name = "TimeoutError";
      throw timeoutError;
    }
    throw error;
  } finally {
    clearTimeout(timeoutId);
    signal?.removeEventListener?.("abort", abortHandler);
  }
};
