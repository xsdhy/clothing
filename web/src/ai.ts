import type {
  AIProvider,
  GenerationRequest,
  GenerationResult,
  BackendResponse,
  ProvidersResponse,
  UsageRecordListResponse,
  UsageRecord,
  PaginationMeta,
  UsageRecordDetailResponse,
  Tag,
  TagListResponse,
  TagDetailResponse,
} from "./types";
import type { SSEEventPayload } from "./utils/sse";
import { appendSanitizedImages, sanitizeImages } from "./utils/images";
import { DEFAULT_HTTP_TIMEOUT_MS, requestWithTimeout } from "./utils/http";
import { emitUnauthorized, getStoredToken } from "./utils/authStorage";
import { parseSSEEvent } from "./utils/sse";
import { safeParseJSON } from "./utils/json";

const LLM_GENERATE_ENDPOINT = "/api/llm";
const LLM_PROVIDERS_ENDPOINT = `/api/llm/providers`;
const USAGE_RECORDS_ENDPOINT = "/api/usage-records";
const TAGS_ENDPOINT = "/api/tags";

const normaliseUsageRecord = (record: UsageRecord): UsageRecord => ({
  ...record,
  tags: Array.isArray((record as unknown as { tags?: Tag[] }).tags)
    ? ((record as unknown as { tags?: Tag[] }).tags as Tag[])
    : [],
});

export interface UsageRecordsResult {
  records: UsageRecord[];
  meta: PaginationMeta;
}

export type UsageRecordResultFilter = "success" | "failure" | "all";

export interface UsageRecordFilters {
  provider?: string;
  model?: string;
  result?: UsageRecordResultFilter;
  tags?: number[];
}

export const generateImage = async (
  request: GenerationRequest,
): Promise<GenerationResult> => {
  const prompt = request.prompt.trim();
  if (!prompt) {
    throw new Error("请输入图片描述");
  }

  if (!request.provider?.id) {
    throw new Error("请选择服务商");
  }

  const images = sanitizeImages(request.images);
  const videos = sanitizeImages(request.videos);

  const payload: Record<string, unknown> = {
    prompt,
    images,
    provider: request.provider.id,
    model: request.model,
  };

  if (typeof request.duration === "number" && request.duration > 0) {
    payload.duration = request.duration;
  }

  if (videos.length > 0) {
    payload.videos = videos;
  }

  if (typeof request.size === "string" && request.size.trim()) {
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
    const token = getStoredToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    response = await fetch(LLM_GENERATE_ENDPOINT, {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
      signal: controller.signal,
    });
  } catch (error) {
    clearTimer();
    const isAbortError =
      error instanceof DOMException && error.name === "AbortError";
    if (timedOut || isAbortError) {
      const timeoutError = new Error("请求超时，请稍后重试");
      timeoutError.name = "TimeoutError";
      throw timeoutError;
    }
    throw new Error(
      `请求后端服务失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }
  resetTimer();

  const contentType = response.headers.get("Content-Type")?.toLowerCase() ?? "";

  if (response.status === 401) {
    emitUnauthorized();
  }

  if (!contentType.includes("text/event-stream")) {
    clearTimer();

    let body: BackendResponse | undefined;
    try {
      body = await response.json();
    } catch {
      body = undefined;
    }

    if (!response.ok) {
      const message =
        body?.error ??
        `服务请求失败: ${response.status} ${response.statusText}`;
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
      throw new Error("后端未返回有效的图片数据");
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
    throw new Error("当前环境不支持流式响应");
  }

  const decoder = new TextDecoder();
  let buffer = "";
  let backendResponse: BackendResponse | undefined;

  const processEvent = (
    event: SSEEventPayload,
  ): BackendResponse | undefined => {
    const eventName = event.event || "message";
    const data = event.data;

    if (
      eventName === "ping" ||
      eventName === "status" ||
      eventName === "message"
    ) {
      return undefined;
    }

    if (eventName === "error") {
      const payload = safeParseJSON<BackendResponse>(data);
      const message = payload?.error ?? data ?? "流式请求失败";
      throw new Error(message);
    }

    if (eventName === "result") {
      const payload = safeParseJSON<BackendResponse>(data);
      if (!payload) {
        throw new Error("流式响应解析失败");
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
      while ((delimiter = buffer.indexOf("\n\n")) !== -1) {
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
    if (error instanceof DOMException && error.name === "AbortError") {
      if (timedOut) {
        const timeoutError = new Error("请求超时，请稍后重试");
        timeoutError.name = "TimeoutError";
        throw timeoutError;
      }
      throw new Error("请求被中断，请重试");
    }

    if (error instanceof Error && error.name === "TimeoutError") {
      throw error;
    }

    throw error;
  } finally {
    clearTimer();
    void reader.cancel().catch(() => undefined);
  }

  if (!backendResponse) {
    throw new Error("流式响应未返回结果");
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
    throw new Error("后端未返回有效的图片数据");
  }

  return { images: Array.from(imageSet), text: backendResponse.text };
};

export const fetchProviders = async (): Promise<AIProvider[]> => {
  let response: Response;
  try {
    response = await requestWithTimeout(LLM_PROVIDERS_ENDPOINT);
  } catch (error) {
    throw new Error(
      `请求模型列表失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let body: ProvidersResponse | undefined;
  try {
    body = await response.json();
  } catch {
    body = undefined;
  }

  if (!response.ok) {
    const message =
      body?.error ?? `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!body?.providers || !Array.isArray(body.providers)) {
    throw new Error("后端未返回有效的模型列表");
  }

  return body.providers;
};

export const fetchUsageRecords = async (
  page = 1,
  pageSize = 20,
  filters?: UsageRecordFilters,
): Promise<UsageRecordsResult> => {
  const safePage = Number.isFinite(page) && page > 0 ? Math.floor(page) : 1;
  const safePageSize =
    Number.isFinite(pageSize) && pageSize > 0
      ? Math.min(Math.floor(pageSize), 100)
      : 20;

  const params = new URLSearchParams({
    page: String(safePage),
    page_size: String(safePageSize),
  });

  const normaliseFilterValue = (value?: string) => {
    const trimmed = value?.trim();
    return trimmed ? trimmed : undefined;
  };

  const resultFilter = (
    filters?.result ?? "success"
  ).toLowerCase() as UsageRecordResultFilter;

  const provider = normaliseFilterValue(filters?.provider);
  const model = normaliseFilterValue(filters?.model);
  const rawTagIDs = Array.isArray(filters?.tags) ? filters?.tags ?? [] : [];
  const tagIDs = Array.from(
    new Set(
      rawTagIDs
        .map((id) => Number(id))
        .filter((id) => Number.isFinite(id) && id > 0),
    ),
  );

  if (provider) {
    params.set("provider", provider);
  }
  if (model) {
    params.set("model", model);
  }
  if (tagIDs.length > 0) {
    params.set("tags", tagIDs.join(","));
  }

  if (resultFilter === "failure") {
    params.set("result", "failure");
  } else if (resultFilter === "all") {
    params.set("result", "all");
  } else {
    params.set("result", "success");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(
      `${USAGE_RECORDS_ENDPOINT}?${params.toString()}`,
      {
        headers: {
          Accept: "application/json",
        },
      },
    );
  } catch (error) {
    throw new Error(
      `获取生成记录失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<UsageRecordListResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<UsageRecordListResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  const records = Array.isArray(payload.records)
    ? (payload.records as UsageRecord[])
    : [];
  const meta = payload.meta ?? {
    page: safePage,
    page_size: safePageSize,
    total: records.length,
  };

  const normalisedRecords = records.map((record) => normaliseUsageRecord(record));

  return { records: normalisedRecords, meta };
};

export const deleteUsageRecord = async (id: number): Promise<void> => {
  if (!Number.isFinite(id) || id <= 0) {
    throw new Error("无效的记录 ID");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(`${USAGE_RECORDS_ENDPOINT}/${id}`, {
      method: "DELETE",
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new Error(
      `删除生成记录失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  if (!response.ok) {
    let message = `服务请求失败: ${response.status} ${response.statusText}`;
    try {
      const payload = (await response.json()) as { error?: string };
      if (payload?.error) {
        message = payload.error;
      }
    } catch {
      // ignore invalid json
    }
    throw new Error(message);
  }
};

export const fetchUsageRecordDetail = async (
  id: number,
): Promise<UsageRecord> => {
  if (!Number.isFinite(id) || id <= 0) {
    throw new Error("无效的记录 ID");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(`${USAGE_RECORDS_ENDPOINT}/${id}`, {
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new Error(
      `获取记录详情失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<UsageRecordDetailResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<UsageRecordDetailResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!payload?.record) {
    throw new Error("后端未返回有效的记录详情");
  }

  return normaliseUsageRecord(payload.record as UsageRecord);
};

export const fetchTags = async (): Promise<Tag[]> => {
  let response: Response;
  try {
    response = await requestWithTimeout(TAGS_ENDPOINT, {
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new Error(
      `获取标签失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<TagListResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<TagListResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  return Array.isArray(payload.tags) ? payload.tags : [];
};

export const createTag = async (name: string): Promise<Tag> => {
  const trimmed = name.trim();
  if (!trimmed) {
    throw new Error("请输入标签名称");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(TAGS_ENDPOINT, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ name: trimmed }),
    });
  } catch (error) {
    throw new Error(
      `创建标签失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<TagDetailResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<TagDetailResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!payload.tag) {
    throw new Error("后端未返回有效的标签");
  }

  return payload.tag;
};

export const updateTag = async (id: number, name: string): Promise<Tag> => {
  if (!Number.isFinite(id) || id <= 0) {
    throw new Error("无效的标签 ID");
  }
  const trimmed = name.trim();
  if (!trimmed) {
    throw new Error("请输入标签名称");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(`${TAGS_ENDPOINT}/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ name: trimmed }),
    });
  } catch (error) {
    throw new Error(
      `更新标签失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<TagDetailResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<TagDetailResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!payload.tag) {
    throw new Error("后端未返回有效的标签");
  }

  return payload.tag;
};

export const deleteTag = async (id: number): Promise<void> => {
  if (!Number.isFinite(id) || id <= 0) {
    throw new Error("无效的标签 ID");
  }

  let response: Response;
  try {
    response = await requestWithTimeout(`${TAGS_ENDPOINT}/${id}`, {
      method: "DELETE",
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new Error(
      `删除标签失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  if (!response.ok) {
    let message = `服务请求失败: ${response.status} ${response.statusText}`;
    try {
      const payload = (await response.json()) as { error?: string };
      if (payload?.error) {
        message = payload.error;
      }
    } catch {
      // ignore
    }
    throw new Error(message);
  }
};

export const updateUsageRecordTags = async (
  recordId: number,
  tagIds: number[],
): Promise<UsageRecord> => {
  if (!Number.isFinite(recordId) || recordId <= 0) {
    throw new Error("无效的记录 ID");
  }

  const uniqueTagIds = Array.from(
    new Set(
      tagIds
        .map((id) => Number(id))
        .filter((id) => Number.isFinite(id) && id > 0),
    ),
  );

  let response: Response;
  try {
    response = await requestWithTimeout(
      `${USAGE_RECORDS_ENDPOINT}/${recordId}/tags`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ tag_ids: uniqueTagIds }),
      },
    );
  } catch (error) {
    throw new Error(
      `更新记录标签失败: ${error instanceof Error ? error.message : "未知错误"}`,
    );
  }

  let payload: Partial<UsageRecordDetailResponse> & { error?: string };
  try {
    payload = (await response.json()) as Partial<UsageRecordDetailResponse> & {
      error?: string;
    };
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const message =
      payload?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!payload.record) {
    throw new Error("后端未返回更新后的记录");
  }

  return normaliseUsageRecord(payload.record as UsageRecord);
};
