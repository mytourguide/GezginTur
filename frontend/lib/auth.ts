// Istemci tarafi oturum yonetimi (JWT access + refresh token)

import { api } from "./api";
import type { User } from "./types";

const KEY_ACCESS = "access_token";
const KEY_REFRESH = "refresh_token";
const KEY_USER = "user";

export function getSession(): { accessToken: string | null; refreshToken: string | null; user: User | null } {
  if (typeof window === "undefined") return { accessToken: null, refreshToken: null, user: null };
  const userRaw = localStorage.getItem(KEY_USER);
  return {
    accessToken: localStorage.getItem(KEY_ACCESS),
    refreshToken: localStorage.getItem(KEY_REFRESH),
    user: userRaw ? JSON.parse(userRaw) : null,
  };
}

export function isLoggedIn(): boolean {
  return typeof window !== "undefined" && !!localStorage.getItem(KEY_ACCESS);
}

export function isAdmin(): boolean {
  const { user } = getSession();
  return user?.role === "admin";
}

// saveSession, giris/kayit sonrasi token'lari saklar ve middleware icin cookie yazar.
export function saveSession(tokens: { access_token: string; refresh_token: string }, user: User) {
  localStorage.setItem(KEY_ACCESS, tokens.access_token);
  localStorage.setItem(KEY_REFRESH, tokens.refresh_token);
  localStorage.setItem(KEY_USER, JSON.stringify(user));
  document.cookie = "auth=1; path=/; max-age=604800";
}

export function clearSession() {
  localStorage.removeItem(KEY_ACCESS);
  localStorage.removeItem(KEY_REFRESH);
  localStorage.removeItem(KEY_USER);
  document.cookie = "auth=; path=/; max-age=0";
}

export async function login(email: string, password: string, totpCode?: string) {
  const res = await api.post("/auth/login", { email, password, totp_code: totpCode || undefined });
  saveSession(res.tokens, res.user);
  return res.user as User;
}

export async function register(fullName: string, email: string, phone: string, password: string) {
  const res = await api.post("/auth/register", { full_name: fullName, email, phone, password });
  saveSession(res.tokens, res.user);
  return res.user as User;
}

export async function logout() {
  const { refreshToken, user } = getSession();
  if (refreshToken && user) {
    try {
      await api.post("/auth/logout", { refresh_token: refreshToken, user_id: user.id });
    } catch { /* ag hatasi olsa bile oturum yerel olarak kapatilir */ }
  }
  clearSession();
  window.location.href = "/";
}
