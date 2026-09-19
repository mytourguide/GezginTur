import { cookies } from "next/headers";

// Istege bagli dil: ?lang=en oncelikli, yoksa "lang" cookie'si (varsayilan tr)
export function getLang(): "tr" | "en" {
  return cookies().get("lang")?.value === "en" ? "en" : "tr";
}

// API istegine dil parametresi ekler
export function apiUrl(path: string): string {
  const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
  const lang = getLang();
  const sep = path.includes("?") ? "&" : "?";
  return `${API}${path}${sep}lang=${lang}`;
}
