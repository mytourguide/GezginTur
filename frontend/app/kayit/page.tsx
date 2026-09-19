"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { register } from "@/lib/auth";
import SocialLoginButtons from "@/components/SocialLoginButtons";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({ full_name: "", email: "", phone: "", password: "" });
  const [error, setError] = useState("");

  async function submit(e: React.FormEvent) {
    e.preventDefault(); setError("");
    try {
      await register(form.full_name, form.email, form.phone, form.password);
      router.push("/hesabim");
    } catch (err: any) { setError(err.message); }
  }

  return (
    <div className="max-w-sm mx-auto px-4 py-16">
      <h1 className="text-2xl font-bold mb-6">Kayit Ol</h1>
      <form onSubmit={submit} className="space-y-4 bg-white border rounded-2xl p-6">
        <input required placeholder="Ad Soyad" value={form.full_name}
          onChange={(e) => setForm({ ...form, full_name: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
        <input type="email" required placeholder="E-posta" value={form.email}
          onChange={(e) => setForm({ ...form, email: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
        <input placeholder="Telefon" value={form.phone}
          onChange={(e) => setForm({ ...form, phone: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
        <input type="password" required minLength={8} placeholder="Sifre (en az 8 karakter)" value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })} className="w-full border rounded-lg px-3 py-2" />
        {error && <p className="text-red-600 text-sm">{error}</p>}
        <button className="w-full bg-brand-600 text-white py-2 rounded-lg hover:bg-brand-700">Kayit Ol</button>
        <SocialLoginButtons />
      </form>
    </div>
  );
}
