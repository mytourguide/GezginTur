"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Coupon } from "@/lib/types";

export default function AdminCouponsPage() {
  const [list, setList] = useState<Coupon[]>([]);
  const [form, setForm] = useState({ code: "", discount_type: "percent", discount_value: 10, valid_from: "", valid_until: "" });

  const load = () => api.get("/admin/coupons", true).then(setList);
  useEffect(() => { load(); }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    await api.post("/admin/coupons", form, true);
    setForm({ code: "", discount_type: "percent", discount_value: 10, valid_from: "", valid_until: "" });
    load();
  }
  async function toggle(c: Coupon) {
    await api.patch(`/admin/coupons/${c.id}/active`, { active: !c.active }, true);
    load();
  }
  async function remove(id: string) {
    if (!confirm("Kupon silinsin mi?")) return;
    await api.del(`/admin/coupons/${id}`);
    load();
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Kuponlar</h1>
      <form onSubmit={submit} className="bg-white border rounded-xl p-4 mb-6 grid sm:grid-cols-5 gap-3 text-sm">
        <input required placeholder="KOD (ornek: YAZ2026)" value={form.code}
          onChange={(e) => setForm({ ...form, code: e.target.value.toUpperCase() })} className="border rounded-lg px-3 py-2" />
        <select value={form.discount_type} onChange={(e) => setForm({ ...form, discount_type: e.target.value })}
          className="border rounded-lg px-3 py-2">
          <option value="percent">Yuzde (%)</option>
          <option value="fixed">Sabit (TL)</option>
        </select>
        <input type="number" min={1} max={form.discount_type === "percent" ? 100 : undefined}
          value={form.discount_value} onChange={(e) => setForm({ ...form, discount_value: +e.target.value })}
          className="border rounded-lg px-3 py-2" placeholder="Indirim" />
        <input type="date" value={form.valid_from} onChange={(e) => setForm({ ...form, valid_from: e.target.value })}
          className="border rounded-lg px-3 py-2" title="Gecerlilik baslangici" />
        <div className="flex gap-2">
          <input type="date" value={form.valid_until} onChange={(e) => setForm({ ...form, valid_until: e.target.value })}
            className="border rounded-lg px-3 py-2 w-full" title="Gecerlilik bitisi" />
          <button className="bg-brand-600 text-white px-4 rounded-lg whitespace-nowrap">Ekle</button>
        </div>
      </form>
      <div className="bg-white border rounded-xl divide-y">
        {list.map((c) => (
          <div key={c.id} className="p-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="font-mono font-bold">{c.code}</p>
              <p className="text-sm text-gray-500">
                {c.discount_type === "percent" ? `%${c.discount_value}` : `${c.discount_value} TL`} indirim
                {c.valid_from && ` - ${c.valid_from} → ${c.valid_until || "sinirsiz"}`}
              </p>
            </div>
            <div className="flex gap-3 text-sm">
              <button onClick={() => toggle(c)} className={c.active ? "text-amber-600 hover:underline" : "text-green-600 hover:underline"}>
                {c.active ? "Pasifleştir" : "Aktifleştir"}
              </button>
              <button onClick={() => remove(c.id)} className="text-red-600 hover:underline">Sil</button>
            </div>
          </div>
        ))}
        {list.length === 0 && <p className="p-4 text-gray-500">Henüz kupon yok.</p>}
      </div>
    </div>
  );
}
