import { createContext, useContext, useEffect, useMemo, useState } from "react";
import * as api from "../api/client";
import type { AuthUser, Role } from "../types";

type LoginInput = {
  email: string;
  password: string;
};

type AuthContextValue = {
  loading: boolean;
  role: Role | null;
  user: AuthUser | null;
  accessToken: string | null;
  login: (input: LoginInput) => Promise<Role>;
  updateUser: (updates: { name: string; email: string }) => Promise<void>;
  logout: () => Promise<void>;
};

const STORAGE_KEY = "hrms-auth-user";
const TOKEN_KEY = "hrms-access-token";
const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const bootstrap = async () => {
      const storedToken = window.localStorage.getItem(TOKEN_KEY);

      if (!storedToken) {
        window.localStorage.removeItem(STORAGE_KEY);
        setLoading(false);
        return;
      }

      try {
        const meResponse = await api.me(storedToken);
        setAccessToken(storedToken);
        const nextUser = api.toAuthUser(meResponse);
        setUser(nextUser);
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextUser));
      } catch {
        try {
          const refreshed = await api.refresh();
          const nextUser = api.toAuthUser(refreshed.user);
          setAccessToken(refreshed.access_token);
          setUser(nextUser);
          window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextUser));
          window.localStorage.setItem(TOKEN_KEY, refreshed.access_token);
        } catch {
          window.localStorage.removeItem(STORAGE_KEY);
          window.localStorage.removeItem(TOKEN_KEY);
          setUser(null);
          setAccessToken(null);
        }
      } finally {
        setLoading(false);
      }
    };

    void bootstrap();
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      loading,
      role: user?.role ?? null,
      user,
      accessToken,
      login: async ({ email, password }) => {
        const response = await api.login(email, password);
        const nextUser = api.toAuthUser(response.user);
        setAccessToken(response.access_token);
        setUser(nextUser);
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextUser));
        window.localStorage.setItem(TOKEN_KEY, response.access_token);
        return nextUser.role;
      },
      updateUser: async ({ name, email }) => {
        if (!accessToken) throw new Error("Missing access token");
        const [firstName = "", ...rest] = name.trim().split(" ");
        const profile = await api.updateProfile(accessToken, {
          email,
          first_name: firstName,
          last_name: rest.join(" "),
        });
        const nextUser = api.toAuthUser(profile);
        setUser(nextUser);
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextUser));
      },
      logout: async () => {
        const token = accessToken;
        window.localStorage.removeItem(STORAGE_KEY);
        window.localStorage.removeItem(TOKEN_KEY);
        setUser(null);
        setAccessToken(null);

        try {
          await api.logout(token);
        } catch {
          // ignore logout errors during local cleanup
        }
      },
    }),
    [accessToken, loading, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
