"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Category } from "@/lib/types";

export default function AdminCategoriesPage() {
  const [list, setList] = useState<Category[]>([]);
  const [form, setForm] = useState({ name: "", slug: "", description: "" });

  const load = () => api.get("/categories").then(setList);
  useEffect(() => { load(); }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    await api.post("/admin/categories", form, true);
    setForm({ name: "", slug: "", description: "" });
    load();
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Kategoriler</h1>
      <form onSubmit={submit} className="bg-white border rounded-xl p-4 mb-6 grid sm:grid-cols-4 gap-3">
        <input required placeholder="Kategori adi" value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })} className="border rounded-lg px-3 py-2" />
        <input required placeholder="slug (ornek: kultur-turlari)" value={form.slug}
          onChange={(e) => setForm({ ...form, slug: e.target.value })} className="border rounded-lg px-3 py-2" />
        <input placeholder="Aciklama" value={form.description}
          onChange={(e) => setForm({ ...form, description: e.target.value })} className="border rounded-lg px-3 py-2" />
        <button className="bg-brand-600 text-white rounded-lg">Ekle</button>
      </form>
      <div className="bg-white border rounded-xl divide-y">
        {list.map((c) => (
          <div key={c.id} className="p-4 flex justify-between items-center">
            <div><p className="font-medium">{c.name}</p><p className="text-xs text-gray-400">/{c.slug}</p></div>
            <p className="text-sm text-gray-500">{c.description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
