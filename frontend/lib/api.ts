// Backend REST API'sine erisim icin merkezi istemci

// Istemci: ayni-origin proxy uzerinden (Next rewrite) — CORS/engelleyici sorunlari icin.
// Sunucu tarafi cagiricilar zaten dogrudan backend'e gider.
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

export class ApiError extends Error {
  status: number;
  constructor(status: number, body: string) {
    super(body);
    this.status = status;
  }
}

// apiFetch, JSON istekleri gonderir; auth=true ise localStorage'daki token'i ekler.
export async function apiFetch<T = any>(
  path: string,
  options: RequestInit & { auth?: boolean } = {}
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (options.auth && typeof window !== "undefined") {
    const token = localStorage.getItem("access_token");
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }
  const res = await fetch(`${API_URL}${path}`, { ...options, headers });
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    try {
      const data = await res.json();
      msg = data.error || msg;
    } catch { /* govde JSON degilse durum kodu yeterli */ }
    throw new ApiError(res.status, msg);
  }
  return res.json();
}

export const api = {
  get: <T = any>(path: string, auth = false) => apiFetch<T>(path, { method: "GET", auth }),
  post: <T = any>(path: string, body?: unknown, auth = false) =>
    apiFetch<T>(path, { method: "POST", body: body ? JSON.stringify(body) : undefined, auth }),
  put: <T = any>(path: string, body?: unknown, auth = false) =>
    apiFetch<T>(path, { method: "PUT", body: body ? JSON.stringify(body) : undefined, auth }),
  patch: <T = any>(path: string, body?: unknown, auth = false) =>
    apiFetch<T>(path, { method: "PATCH", body: body ? JSON.stringify(body) : undefined, auth }),
  del: <T = any>(path: string, auth = true) => apiFetch<T>(path, { method: "DELETE", auth }),
};
