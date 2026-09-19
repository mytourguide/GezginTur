"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { saveSession } from "@/lib/auth";
import { useLang } from "@/lib/useLang";

// Google Identity Services ve Sign in with Apple butonlari.
// Ilgili env degiskeni (NEXT_PUBLIC_GOOGLE_CLIENT_ID / NEXT_PUBLIC_APPLE_SERVICE_ID)
// yoksa buton gosterilmez.

declare global {
  interface Window { google?: any; AppleID?: any; }
}

export default function SocialLoginButtons() {
  const router = useRouter();
  const lang = useLang();
  const [err, setErr] = useState("");
  const googleRef = useRef<HTMLDivElement>(null);
  const gClientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "";
  const appleId = process.env.NEXT_PUBLIC_APPLE_SERVICE_ID || "";

  const t = lang === "en"
    ? { or: "or", googleFail: "Google sign-in failed", appleFail: "Apple sign-in failed", apple: "Sign in with Apple" }
    : { or: "veya", googleFail: "Google girisi basarisiz", appleFail: "Apple girisi basarisiz", apple: "Apple ile Giris Yap" };

  async function backendLogin(path: string, body: Record<string, string>, failMsg: string) {
    setErr("");
    try {
      const res = await api.post(path, body);
      saveSession(res.tokens, res.user);
      router.push(res.user.role === "admin" ? "/admin" : "/hesabim");
    } catch (e: any) {
      setErr(e.message || failMsg);
    }
  }

  // Google: GSI script'i yukle ve butonu yerlestir
  useEffect(() => {
    if (!gClientId || !googleRef.current) return;
    const s = document.createElement("script");
    s.src = "https://accounts.google.com/gsi/client";
    s.async = true;
    s.onload = () => {
      window.google.accounts.id.initialize({
        client_id: gClientId,
        callback: (resp: any) => backendLogin("/auth/oauth/google", { id_token: resp.credential }, t.googleFail),
      });
      window.google.accounts.id.renderButton(googleRef.current, {
        theme: "outline", size: "large", width: 320, locale: lang === "en" ? "en" : "tr",
      });
    };
    document.head.appendChild(s);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gClientId, lang]);

  // Apple: popup akisi; id_token'i backend'e gonder
  async function appleLogin() {
    if (!window.AppleID) return;
    try {
      const res = await window.AppleID.auth.signIn();
      const idToken = res.authorization.id_token;
      const name = res.user ? `${res.user.name?.firstName ?? ""} ${res.user.name?.lastName ?? ""}`.trim() : "";
      await backendLogin("/auth/oauth/apple", { id_token: idToken, name }, t.appleFail);
    } catch {
      setErr(t.appleFail);
    }
  }

  useEffect(() => {
    if (!appleId) return;
    const s = document.createElement("script");
    s.src = "https://appleid.cdn-apple.com/appleauth/static/jsapi/appleid/1/en_US/appleid.auth.js";
    s.async = true;
    s.onload = () => {
      window.AppleID.auth.init({
        clientId: appleId,
        scope: "name email",
        redirectURI: window.location.origin,
        usePopup: true,
        locale: "tr",
      });
    };
    document.head.appendChild(s);
  }, [appleId]);

  if (!gClientId && !appleId) return null;

  return (
    <div className="mt-4">
      <div className="flex items-center gap-3 my-3 text-xs text-gray-400">
        <span className="flex-1 border-t" />{t.or}<span className="flex-1 border-t" />
      </div>
      {gClientId && (
        <div className="flex justify-center min-h-[44px]">
          <div ref={googleRef} />
        </div>
      )}
      {appleId && (
        <button type="button" onClick={appleLogin}
          className="w-full flex items-center justify-center gap-2 bg-black text-white py-2.5 rounded-lg text-sm font-medium hover:bg-zinc-800">
          <svg viewBox="0 0 24 24" className="w-4 h-4 fill-current"><path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.53 4.08ZM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25Z"/></svg>
          {t.apple}
        </button>
      )}
      {err && <p className="text-red-600 text-sm mt-2 text-center">{err}</p>}
    </div>
  );
}
