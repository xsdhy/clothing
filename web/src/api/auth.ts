import { requestWithTimeout } from "../utils/http";
import type {
  AuthResponse,
  AuthStatusResponse,
  UserListResponse,
  UserSummary,
} from "../types";

type ErrorPayload = { error?: string };

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  email: string;
  password: string;
  display_name?: string;
}

export interface CreateUserPayload {
  email: string;
  password: string;
  display_name?: string;
  role: string;
  is_active?: boolean;
}

export interface UpdateUserPayload {
  display_name?: string | null;
  password?: string | null;
  role?: string | null;
  is_active?: boolean | null;
}

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
    // ignore
  }
  return `服务请求失败: ${response.status} ${response.statusText}`;
};

export const fetchAuthStatus = async (): Promise<AuthStatusResponse> => {
  const response = await requestWithTimeout("/api/auth/status", {
    headers: { Accept: "application/json" },
  });
  const payload = (await response.json()) as AuthStatusResponse & ErrorPayload;
  if (!response.ok) {
    throw new Error(payload?.error ?? `服务请求失败: ${response.status}`);
  }
  return { has_user: Boolean(payload.has_user) };
};

export const login = async (payload: LoginPayload): Promise<AuthResponse> => {
  const response = await requestWithTimeout("/api/auth/login", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
  const body = (await response.json()) as AuthResponse & ErrorPayload;
  if (!response.ok) {
    throw new Error(body?.error ?? `登录失败: ${response.status}`);
  }
  return body;
};

export const registerInitialUser = async (
  payload: RegisterPayload,
): Promise<AuthResponse> => {
  const response = await requestWithTimeout("/api/auth/register", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
  const body = (await response.json()) as AuthResponse & ErrorPayload;
  if (!response.ok) {
    throw new Error(body?.error ?? `注册失败: ${response.status}`);
  }
  return body;
};

export const fetchCurrentUser = async (): Promise<UserSummary> => {
  const response = await requestWithTimeout("/api/auth/me", {
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
  return (await response.json()) as UserSummary;
};

export const listUsers = async (
  page = 1,
  pageSize = 20,
): Promise<UserListResponse> => {
  const params = new URLSearchParams({
    page: String(Math.max(1, page)),
    page_size: String(Math.min(Math.max(1, pageSize), 100)),
  });
  const response = await requestWithTimeout(`/api/users?${params.toString()}`);
  const body = (await response.json()) as UserListResponse & ErrorPayload;
  if (!response.ok) {
    throw new Error(body?.error ?? `获取用户失败: ${response.status}`);
  }
  return body;
};

export const createUser = async (
  payload: CreateUserPayload,
): Promise<UserSummary> => {
  const response = await requestWithTimeout("/api/users", {
    method: "POST",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
  const body = (await response.json()) as UserSummary & ErrorPayload;
  if (!response.ok) {
    throw new Error(body?.error ?? `创建用户失败: ${response.status}`);
  }
  return body;
};

export const updateUser = async (
  id: number,
  payload: UpdateUserPayload,
): Promise<UserSummary> => {
  const response = await requestWithTimeout(`/api/users/${id}`, {
    method: "PATCH",
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  });
  const body = (await response.json()) as UserSummary & ErrorPayload;
  if (!response.ok) {
    throw new Error(body?.error ?? `更新用户失败: ${response.status}`);
  }
  return body;
};

export const deleteUser = async (id: number): Promise<void> => {
  const response = await requestWithTimeout(`/api/users/${id}`, {
    method: "DELETE",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
};
