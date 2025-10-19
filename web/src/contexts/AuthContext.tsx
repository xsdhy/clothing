import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import type { ReactNode } from "react";
import type { AuthResponse, UserSummary } from "../types";
import {
  AUTH_EVENTS,
  clearStoredToken,
  getStoredToken,
  storeToken,
} from "../utils/authStorage";
import {
  fetchCurrentUser,
  fetchAuthStatus,
  login as loginRequest,
  registerInitialUser,
} from "../api/auth";

interface AuthContextValue {
  user: UserSummary | null;
  token: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isSuperAdmin: boolean;
  login: (email: string, password: string) => Promise<AuthResponse>;
  registerInitial: (payload: {
    email: string;
    password: string;
    displayName?: string;
  }) => Promise<AuthResponse>;
  logout: () => void;
  refresh: () => Promise<UserSummary | null>;
  ensureStatus: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const applyDisplayName = (user: UserSummary): UserSummary => ({
  ...user,
  display_name: user.display_name?.trim() || user.email,
});

export const AuthProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<UserSummary | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const applySession = useCallback((response: AuthResponse) => {
    storeToken(response.token);
    setToken(response.token);
    setUser(applyDisplayName(response.user));
  }, []);

  const logout = useCallback(() => {
    clearStoredToken();
    setToken(null);
    setUser(null);
  }, []);

  const refresh = useCallback(async (): Promise<UserSummary | null> => {
    try {
      const profile = await fetchCurrentUser();
      const enriched = applyDisplayName(profile);
      setUser(enriched);
      return enriched;
    } catch (error) {
      logout();
      return null;
    }
  }, [logout]);

  useEffect(() => {
    const stored = getStoredToken();
    if (!stored) {
      setLoading(false);
      return;
    }
    setToken(stored);

    void (async () => {
      try {
        const profile = await fetchCurrentUser();
        setUser(applyDisplayName(profile));
      } catch {
        logout();
      } finally {
        setLoading(false);
      }
    })();
  }, [logout]);

  useEffect(() => {
    if (token) {
      return;
    }
    setLoading(false);
  }, [token]);

  useEffect(() => {
    const handler = () => {
      logout();
    };
    window.addEventListener(AUTH_EVENTS.unauthorized, handler);
    return () => {
      window.removeEventListener(AUTH_EVENTS.unauthorized, handler);
    };
  }, [logout]);

  const login = useCallback(
    async (email: string, password: string) => {
      const response = await loginRequest({ email, password });
      applySession(response);
      return response;
    },
    [applySession],
  );

  const registerInitial = useCallback(
    async ({
      email,
      password,
      displayName,
    }: {
      email: string;
      password: string;
      displayName?: string;
    }) => {
      const response = await registerInitialUser({
        email,
        password,
        display_name: displayName,
      });
      applySession(response);
      return response;
    },
    [applySession],
  );

  const ensureStatus = useCallback(async () => {
    const status = await fetchAuthStatus();
    return status.has_user;
  }, []);

  const value = useMemo<AuthContextValue>(() => {
    const currentRole = user?.role ?? "";
    const isSuperAdmin = currentRole === "super_admin";
    const isAdmin = isSuperAdmin || currentRole === "admin";
    return {
      user,
      token,
      loading,
      isAuthenticated: Boolean(user),
      isAdmin,
      isSuperAdmin,
      login,
      registerInitial,
      logout,
      refresh,
      ensureStatus,
    };
  }, [loading, login, logout, refresh, registerInitial, token, user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextValue => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
};
