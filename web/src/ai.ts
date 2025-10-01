import type { AIProvider, GenerationRequest, GenerationResult } from './types';


const LLM_GENERATE_ENDPOINT = '/api/llm';
const LLM_PROVIDERS_ENDPOINT = `/api/llm/providers`;
const DEFAULT_HTTP_TIMEOUT_MS = 120_000;

interface BackendResponse {
  image?: string;
  images?: string[];
  text?: string;
  error?: string;
}

interface ProvidersResponse {
  providers?: AIProvider[];
  error?: string;
}

const sanitizeImages = (images: string[] = []): string[] =>
  images
    .map((image) => (typeof image === 'string' ? image.trim() : ''))
    .filter((image): image is string => image.length > 0);

interface SSEEventPayload {
  event: string;
  data: string;
}

const parseSSEEvent = (chunk: string): SSEEventPayload | null => {
  let eventName = 'message';
  const dataLines: string[] = [];

  const lines = chunk.split(/\r?\n/);
  for (const rawLine of lines) {
    const line = rawLine.trimEnd();
    if (!line) {
      continue;
    }
    if (line.startsWith(':')) {
      continue;
    }
    if (line.startsWith('event:')) {
      eventName = line.slice(6).trim();
      continue;
    }
    if (line.startsWith('data:')) {
      const value = line.slice(5);
      dataLines.push(value.startsWith(' ') ? value.slice(1) : value);
    }
  }

  return {
    event: eventName || 'message',
    data: dataLines.join('\n'),
  };
};

const safeParseJSON = <T>(input: string): T | undefined => {
  try {
    return input ? (JSON.parse(input) as T) : undefined;
  } catch {
    return undefined;
  }
};

const requestWithTimeout = async (
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

  const timeoutMs = request.timeoutMs ?? DEFAULT_HTTP_TIMEOUT_MS;
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

    if (Array.isArray(body?.images)) {
      sanitizeImages(body.images).forEach((item) => {
        if (!imageSet.has(item)) {
          imageSet.add(item);
        }
      });
    }

    if (typeof body?.image === 'string') {
      const trimmed = body.image.trim();
      if (trimmed) {
        imageSet.add(trimmed);
      }
    }

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

  if (Array.isArray(backendResponse.images)) {
    sanitizeImages(backendResponse.images).forEach((item) => {
      if (!imageSet.has(item)) {
        imageSet.add(item);
      }
    });
  }

  if (typeof backendResponse.image === 'string') {
    const trimmed = backendResponse.image.trim();
    if (trimmed) {
      imageSet.add(trimmed);
    }
  }

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
