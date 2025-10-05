import type { AIProvider, GenerationRequest, GenerationResult, BackendResponse, ProvidersResponse } from './types';
import type { SSEEventPayload } from './utils/sse';
import { appendSanitizedImages, sanitizeImages } from './utils/images';
import { DEFAULT_HTTP_TIMEOUT_MS, requestWithTimeout } from './utils/http';
import { parseSSEEvent } from './utils/sse';
import { safeParseJSON } from './utils/json';

const LLM_GENERATE_ENDPOINT = '/api/llm';
const LLM_PROVIDERS_ENDPOINT = `/api/llm/providers`;

export const generateImage = async (request: GenerationRequest): Promise<GenerationResult> => {
  const prompt = request.prompt.trim();
  if (!prompt) {
    throw new Error('请输入图片描述');
  }

  if (!request.provider?.id) {
    throw new Error('请选择服务商');
  }

  const payload: Record<string, unknown> = {
    prompt,
    images: sanitizeImages(request.images),
    provider: request.provider.id,
    model: request.model,
  };

  if (typeof request.size === 'string' && request.size.trim()) {
    payload.size = request.size.trim();
  }

  const timeoutMs = DEFAULT_HTTP_TIMEOUT_MS;
  const controller = new AbortController();
  let timeoutId: number | undefined;
  let timedOut = false;

  const resetTimer = () => {
    if (timeoutMs <= 0) {
      return;
    }
    if (timeoutId !== undefined) {
      window.clearTimeout(timeoutId);
    }
    timeoutId = window.setTimeout(() => {
      timedOut = true;
      controller.abort();
    }, timeoutMs);
  };

  const clearTimer = () => {
    if (timeoutId !== undefined) {
      window.clearTimeout(timeoutId);
      timeoutId = undefined;
    }
  };

  resetTimer();

  let response: Response;
  try {
    response = await fetch(LLM_GENERATE_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
      },
      body: JSON.stringify(payload),
      signal: controller.signal,
    });
  } catch (error) {
    clearTimer();
    const isAbortError = error instanceof DOMException && error.name === 'AbortError';
    if (timedOut || isAbortError) {
      const timeoutError = new Error('请求超时，请稍后重试');
      timeoutError.name = 'TimeoutError';
      throw timeoutError;
    }
    throw new Error(`请求后端服务失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }

  resetTimer();

  const contentType = response.headers.get('Content-Type')?.toLowerCase() ?? '';

  if (!contentType.includes('text/event-stream')) {
    clearTimer();

    let body: BackendResponse | undefined;
    try {
      body = await response.json();
    } catch {
      body = undefined;
    }

    if (!response.ok) {
      const message = body?.error ?? `服务请求失败: ${response.status} ${response.statusText}`;
      throw new Error(message);
    }

    if (body?.error) {
      throw new Error(body.error);
    }

    const imageSet = new Set<string>();

    appendSanitizedImages(imageSet, body?.images);



    if (imageSet.size === 0 && body?.text) {
      return { images: [], text: body.text };
    }

    if (imageSet.size === 0) {
      throw new Error('后端未返回有效的图片数据');
    }

    return { images: Array.from(imageSet), text: body?.text };
  }

  if (!response.ok) {
    clearTimer();
    throw new Error(`服务请求失败: ${response.status} ${response.statusText}`);
  }

  const reader = response.body?.getReader();
  if (!reader) {
    clearTimer();
    throw new Error('当前环境不支持流式响应');
  }

  const decoder = new TextDecoder();
  let buffer = '';
  let backendResponse: BackendResponse | undefined;

  const processEvent = (event: SSEEventPayload): BackendResponse | undefined => {
    const eventName = event.event || 'message';
    const data = event.data;

    if (eventName === 'ping' || eventName === 'status' || eventName === 'message') {
      return undefined;
    }

    if (eventName === 'error') {
      const payload = safeParseJSON<BackendResponse>(data);
      const message = payload?.error ?? data ?? '流式请求失败';
      throw new Error(message);
    }

    if (eventName === 'result') {
      const payload = safeParseJSON<BackendResponse>(data);
      if (!payload) {
        throw new Error('流式响应解析失败');
      }
      return payload;
    }

    return undefined;
  };

  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) {
        break;
      }

      resetTimer();
      buffer += decoder.decode(value, { stream: true });

      let delimiter: number;
      while ((delimiter = buffer.indexOf('\n\n')) !== -1) {
        const rawEvent = buffer.slice(0, delimiter);
        buffer = buffer.slice(delimiter + 2);

        if (!rawEvent.trim()) {
          continue;
        }

        const event = parseSSEEvent(rawEvent);
        if (!event) {
          continue;
        }

        const maybeResponse = processEvent(event);
        if (maybeResponse) {
          backendResponse = maybeResponse;
          break;
        }
      }

      if (backendResponse) {
        break;
      }
    }

    if (!backendResponse && buffer.trim()) {
      const event = parseSSEEvent(buffer);
      if (event) {
        backendResponse = processEvent(event);
      }
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      if (timedOut) {
        const timeoutError = new Error('请求超时，请稍后重试');
        timeoutError.name = 'TimeoutError';
        throw timeoutError;
      }
      throw new Error('请求被中断，请重试');
    }

    if (error instanceof Error && error.name === 'TimeoutError') {
      throw error;
    }

    throw error;
  } finally {
    clearTimer();
    void reader.cancel().catch(() => undefined);
  }

  if (!backendResponse) {
    throw new Error('流式响应未返回结果');
  }

  if (backendResponse.error) {
    throw new Error(backendResponse.error);
  }

  const imageSet = new Set<string>();

  appendSanitizedImages(imageSet, backendResponse.images);

  if (imageSet.size === 0 && backendResponse.text) {
    return { images: [], text: backendResponse.text };
  }

  if (imageSet.size === 0) {
    throw new Error('后端未返回有效的图片数据');
  }

  return { images: Array.from(imageSet), text: backendResponse.text };
};

export const fetchProviders = async (): Promise<AIProvider[]> => {
  let response: Response;
  try {
    response = await requestWithTimeout(LLM_PROVIDERS_ENDPOINT);
  } catch (error) {
    throw new Error(`请求模型列表失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }

  let body: ProvidersResponse | undefined;
  try {
    body = await response.json();
  } catch {
    body = undefined;
  }

  if (!response.ok) {
    const message = body?.error ?? `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!body?.providers || !Array.isArray(body.providers)) {
    throw new Error('后端未返回有效的模型列表');
  }

  return body.providers;
};
