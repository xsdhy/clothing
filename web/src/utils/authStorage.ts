const AUTH_TOKEN_KEY = "clothing:auth_token";
export const AUTH_EVENTS = {
  unauthorized: "auth:unauthorized",
};

const safeWindow = typeof window !== "undefined" ? window : undefined;

export const getStoredToken = (): string | null => {
  if (!safeWindow) {
    return null;
  }
  try {
    return safeWindow.localStorage.getItem(AUTH_TOKEN_KEY);
  } catch {
    return null;
  }
};

export const storeToken = (token: string): void => {
  if (!safeWindow) {
    return;
  }
  try {
    safeWindow.localStorage.setItem(AUTH_TOKEN_KEY, token);
  } catch {
    // ignore persistence errors
  }
};

export const clearStoredToken = (): void => {
  if (!safeWindow) {
    return;
  }
  try {
    safeWindow.localStorage.removeItem(AUTH_TOKEN_KEY);
  } catch {
    // ignore persistence errors
  }
};

export const emitUnauthorized = (): void => {
  if (!safeWindow) {
    return;
  }
  safeWindow.dispatchEvent(new CustomEvent(AUTH_EVENTS.unauthorized));
};

export const hasStoredToken = (): boolean => Boolean(getStoredToken());
