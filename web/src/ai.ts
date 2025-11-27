import type {
  AIProvider,
  GenerationRequest,
  GenerationJob,
  GenerationEventPayload,
  ProvidersResponse,
  UsageRecordListResponse,
  UsageRecord,
  PaginationMeta,
  UsageRecordDetailResponse,
  Tag,
  TagListResponse,
  TagDetailResponse,
} from "./types";
import { sanitizeImages } from "./utils/images";
import { requestWithTimeout } from "./utils/http";
import { emitUnauthorized, getStoredToken } from "./utils/authStorage";
import { parseSSEEvent } from "./utils/sse";
import { safeParseJSON } from "./utils/json";

const LLM_GENERATE_ENDPOINT = "/api/llm";
const LLM_EVENTS_ENDPOINT = "/api/llm/events";
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
  hasImages?: boolean;
}

export const generateContent = async (
  request: GenerationRequest,
  clientId?: string,
): Promise<GenerationJob> => {
  const prompt = request.prompt.trim();
  if (!prompt) {
    throw new Error("请输入图片描述");
  }

  if (!request.provider?.id) {
    throw new Error("请选择服务商");
  }

  const images = sanitizeImages(request.inputs?.images ?? []);
  const videos = sanitizeImages(request.inputs?.videos ?? []);

  const payload: Record<string, unknown> = {
    prompt,
    provider_id: request.provider.id,
    model_id: request.model,
    inputs: {
      images,
      ...(videos.length > 0 ? { videos } : {}),
    },
  };

  const size = request.options?.size?.trim();
  const duration =
    typeof request.options?.duration === "number"
      ? request.options.duration
      : undefined;

  if (size || duration) {
    payload.options = {
      ...(size ? { size } : {}),
      ...(duration && duration > 0 ? { duration } : {}),
    };
  }

  if (request.tag_ids?.length) {
    payload.tag_ids = request.tag_ids;
  }

  if (clientId?.trim()) {
    payload.client_id = clientId.trim();
  }

  let response: Response;
  try {
    response = await requestWithTimeout(LLM_GENERATE_ENDPOINT, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "请求后端服务失败";
    throw new Error(message);
  }

  let body: Partial<GenerationJob> & { error?: string } = {};
  try {
    body = (await response.json()) as Partial<GenerationJob> & {
      error?: string;
    };
  } catch {
    body = {};
  }

  if (!response.ok) {
    const message =
      body?.error ??
      `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  const recordId = Number(body?.record_id);
  if (!Number.isFinite(recordId) || recordId <= 0) {
    throw new Error("后端未返回有效的记录ID");
  }

  return {
    record_id: recordId,
    status: body?.status,
  };
};

export const subscribeGenerationEvents = (
  clientId: string,
  onEvent: (event: GenerationEventPayload) => void,
  onError?: (error: Error) => void,
): { close: () => void } => {
  const trimmedClientId = clientId.trim();
  const controller = new AbortController();

  const emitError = (message: string | Error) => {
    if (!onError) {
      return;
    }
    const error =
      message instanceof Error ? message : new Error(message || "未知错误");
    onError(error);
  };

  if (!trimmedClientId) {
    emitError("缺少客户端标识");
    return { close: () => undefined };
  }

  const connect = async () => {
    const token = getStoredToken();
    const headers: Record<string, string> = {
      Accept: "text/event-stream",
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    let response: Response;
    try {
      response = await fetch(
        `${LLM_EVENTS_ENDPOINT}?client_id=${encodeURIComponent(trimmedClientId)}`,
        {
          method: "GET",
          headers,
          signal: controller.signal,
        },
      );
    } catch (error) {
      if (controller.signal.aborted) {
        return;
      }
      emitError(
        `连接事件流失败: ${error instanceof Error ? error.message : "未知错误"}`,
      );
      return;
    }

    if (response.status === 401) {
      emitUnauthorized();
    }

    const contentType = response.headers.get("Content-Type")?.toLowerCase() ?? "";
    if (!contentType.includes("text/event-stream")) {
      emitError("后端不支持事件流连接");
      return;
    }

    if (!response.ok) {
      emitError(
        `事件流连接失败: ${response.status} ${response.statusText}`,
      );
      return;
    }

    const reader = response.body?.getReader();
    if (!reader) {
      emitError("当前环境不支持事件流");
      return;
    }

    const decoder = new TextDecoder();
    let buffer = "";

    const handleEvent = (rawEvent: string) => {
      if (!rawEvent.trim()) {
        return;
      }
      const parsed = parseSSEEvent(rawEvent);
      if (!parsed) {
        return;
      }
      const eventName = parsed.event || "message";
      if (eventName === "ping" || eventName === "message") {
        return;
      }
      if (eventName === "generation_completed") {
        const payload = safeParseJSON<GenerationEventPayload>(parsed.data);
        const recordId = Number(payload?.record_id);
        if (!Number.isFinite(recordId) || recordId <= 0) {
          return;
        }
        const status =
          payload?.status === "failure" ? "failure" : "success";
        onEvent({
          record_id: recordId,
          status,
          error: payload?.error,
        });
      }
    };

    try {
      while (true) {
        const { value, done } = await reader.read();
        if (done) {
          break;
        }
        buffer += decoder.decode(value, { stream: true });
        let delimiter: number;
        while ((delimiter = buffer.indexOf("\n\n")) !== -1) {
          const rawEvent = buffer.slice(0, delimiter);
          buffer = buffer.slice(delimiter + 2);
          handleEvent(rawEvent);
        }
      }
      if (!controller.signal.aborted) {
        emitError("事件流已关闭");
      }
    } catch (error) {
      if (!controller.signal.aborted) {
        emitError(
          `事件流中断: ${error instanceof Error ? error.message : "未知错误"}`,
        );
      }
    } finally {
      void reader.cancel().catch(() => undefined);
    }
  };

  void connect();

  return {
    close: () => controller.abort(),
  };
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
  if (filters?.hasImages) {
    params.set("has_output_images", "true");
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
