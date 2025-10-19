import { requestWithTimeout } from "../utils/http";
import type {
  ProviderAdmin,
  ProviderModelAdmin,
  ProviderCreatePayload,
  ProviderUpdatePayload,
  ProviderModelCreatePayload,
  ProviderModelUpdatePayload,
} from "../types";

type ErrorPayload = { error?: string };

const jsonHeaders = {
  "Content-Type": "application/json",
  Accept: "application/json",
};

const parseError = async (response: Response): Promise<string> => {
  try {
    const payload = (await response.json()) as ErrorPayload;
    if (payload?.error) {
      return payload.error;
    }
  } catch {
    // ignore JSON parse error
  }
  return `服务请求失败: ${response.status} ${response.statusText}`;
};

export const listProvidersAdmin = async (): Promise<ProviderAdmin[]> => {
  const response = await requestWithTimeout("/api/providers", {
    headers: { Accept: "application/json" },
  });
  const body = (await response.json()) as
    | { providers?: ProviderAdmin[] }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `获取厂商列表失败: ${response.status}`,
    );
  }
  return (body as { providers?: ProviderAdmin[] }).providers ?? [];
};

export const createProvider = async (
  payload: ProviderCreatePayload,
): Promise<ProviderAdmin> => {
  const response = await requestWithTimeout("/api/providers", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
  const body = (await response.json()) as
    | { provider?: ProviderAdmin }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `创建厂商失败: ${response.status} ${response.statusText}`,
    );
  }
  const provider = (body as { provider?: ProviderAdmin }).provider;
  if (!provider) {
    throw new Error("后端未返回有效的厂商信息");
  }
  return provider;
};

export const getProviderDetail = async (
  id: string,
): Promise<ProviderAdmin> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(id)}`,
    {
      headers: { Accept: "application/json" },
    },
  );
  const body = (await response.json()) as
    | { provider?: ProviderAdmin }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `获取厂商详情失败: ${response.status}`,
    );
  }
  const provider = (body as { provider?: ProviderAdmin }).provider;
  if (!provider) {
    throw new Error("后端未返回有效的厂商信息");
  }
  return provider;
};

export const updateProvider = async (
  id: string,
  payload: ProviderUpdatePayload,
): Promise<ProviderAdmin> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(id)}`,
    {
      method: "PATCH",
      headers: jsonHeaders,
      body: JSON.stringify(payload),
    },
  );
  const body = (await response.json()) as
    | { provider?: ProviderAdmin }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `更新厂商失败: ${response.status} ${response.statusText}`,
    );
  }
  const provider = (body as { provider?: ProviderAdmin }).provider;
  if (!provider) {
    throw new Error("后端未返回有效的厂商信息");
  }
  return provider;
};

export const deleteProvider = async (id: string): Promise<void> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(id)}`,
    {
      method: "DELETE",
      headers: { Accept: "application/json" },
    },
  );
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
};

export const listProviderModels = async (
  id: string,
): Promise<ProviderModelAdmin[]> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(id)}/models`,
    {
      headers: { Accept: "application/json" },
    },
  );
  const body = (await response.json()) as
    | { models?: ProviderModelAdmin[] }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `获取模型列表失败: ${response.status}`,
    );
  }
  return (body as { models?: ProviderModelAdmin[] }).models ?? [];
};

export const createProviderModel = async (
  providerId: string,
  payload: ProviderModelCreatePayload,
): Promise<ProviderModelAdmin> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(providerId)}/models`,
    {
      method: "POST",
      headers: jsonHeaders,
      body: JSON.stringify(payload),
    },
  );
  const body = (await response.json()) as
    | { model?: ProviderModelAdmin }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `创建模型失败: ${response.status} ${response.statusText}`,
    );
  }
  const model = (body as { model?: ProviderModelAdmin }).model;
  if (!model) {
    throw new Error("后端未返回有效的模型信息");
  }
  return model;
};

export const updateProviderModel = async (
  providerId: string,
  modelId: string,
  payload: ProviderModelUpdatePayload,
): Promise<ProviderModelAdmin> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(providerId)}/models/${encodeURIComponent(
      modelId,
    )}`,
    {
      method: "PATCH",
      headers: jsonHeaders,
      body: JSON.stringify(payload),
    },
  );
  const body = (await response.json()) as
    | { model?: ProviderModelAdmin }
    | ErrorPayload;
  if (!response.ok) {
    throw new Error(
      (body as ErrorPayload)?.error ??
        `更新模型失败: ${response.status} ${response.statusText}`,
    );
  }
  const model = (body as { model?: ProviderModelAdmin }).model;
  if (!model) {
    throw new Error("后端未返回有效的模型信息");
  }
  return model;
};

export const deleteProviderModel = async (
  providerId: string,
  modelId: string,
): Promise<void> => {
  const response = await requestWithTimeout(
    `/api/providers/${encodeURIComponent(providerId)}/models/${encodeURIComponent(
      modelId,
    )}`,
    {
      method: "DELETE",
      headers: { Accept: "application/json" },
    },
  );
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
};
