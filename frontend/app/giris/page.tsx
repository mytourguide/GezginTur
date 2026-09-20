"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { login } from "@/lib/auth";
import SocialLoginButtons from "@/components/SocialLoginButtons";

import { Suspense } from "react";

function LoginForm() {
  const router = useRouter();
  const next = useSearchParams().get("next") || "/";
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [need2fa, setNeed2fa] = useState(false); // 2FA aciksa kod alani gorunur
  const [error, setError] = useState("");

  async function submit(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      const user = await login(email, password, totpCode);
      router.push(user.role === "admin" ? "/admin" : next);
    } catch (err: any) {
      const msg = err.message || "Giris basarisiz";
      // Backend 2FA gerekince ozel mesaj dondurur: kod alanini ac
      if (msg.toLowerCase().includes("dogrulama kodu") || msg.includes("2FA")) setNeed2fa(true);
      setError(msg);
    }
  }

  return (
    <div className="max-w-sm mx-auto px-4 py-16">
      <h1 className="text-2xl font-bold mb-6">Giris Yap</h1>
      <form onSubmit={submit} className="space-y-4 bg-white border rounded-2xl p-6">
        <input type="email" required placeholder="E-posta" value={email} onChange={(e) => setEmail(e.target.value)}
          className="w-full border rounded-lg px-3 py-2" />
        <input type="password" required placeholder="Sifre" value={password} onChange={(e) => setPassword(e.target.value)}
          className="w-full border rounded-lg px-3 py-2" />
        {need2fa && (
          <input inputMode="numeric" maxLength={6} placeholder="Authenticator kodu (6 hane)"
            value={totpCode} onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ""))}
            className="w-full border rounded-lg px-3 py-2 tracking-widest" />
        )}
        {error && <p className="text-red-600 text-sm">{error}</p>}
        <button className="w-full bg-brand-600 text-white py-2 rounded-lg hover:bg-brand-700">Giris Yap</button>
        <p className="text-sm text-center">Hesabiniz yok mu? <a href="/kayit" className="text-brand-600">Kayit olun</a></p>
        <SocialLoginButtons />
      </form>
    </div>
  );
}

// useSearchParams Suspense sinirina muhtactir (prerender hatasi onlemi)
export default function LoginPage() {
  return (
    <Suspense fallback={<div className="max-w-sm mx-auto px-4 py-16 text-gray-500">Yukleniyor...</div>}>
      <LoginForm />
    </Suspense>
  );
}
