"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";
import type { Tour } from "@/lib/types";

export default function AdminToursPage() {
  const [tours, setTours] = useState<Tour[]>([]);
  const [search, setSearch] = useState("");

  const load = () => api.get(`/admin/tours?search=${encodeURIComponent(search)}`, true).then(setTours);
  useEffect(() => { load(); }, []); // eslint-disable-line

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Turlar</h1>
        <a href="/admin/turlar/yeni" className="bg-brand-600 text-white px-4 py-2 rounded-lg">+ Yeni Tur</a>
      </div>
      <input value={search} onChange={(e) => setSearch(e.target.value)} onKeyDown={(e) => e.key === "Enter" && load()}
        placeholder="Tur ara..." className="border rounded-lg px-3 py-2 mb-4 w-full sm:w-80" />
      <div className="bg-white border rounded-xl overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left">
            <tr><th className="p-3">Tur</th><th className="p-3">Kategori</th><th className="p-3">Lokasyon</th><th className="p-3">Fiyat</th><th className="p-3">Durum</th><th className="p-3"></th></tr>
          </thead>
          <tbody className="divide-y">
            {tours.map((t) => (
              <tr key={t.id}>
                <td className="p-3 font-medium">{t.title}</td>
                <td className="p-3">{t.category_name}</td>
                <td className="p-3">{t.location}</td>
                <td className="p-3">{formatPrice(t.base_price)}</td>
                <td className="p-3">{t.active ? "Aktif" : "Pasif"}</td>
                <td className="p-3"><a className="text-brand-600 hover:underline" href={`/admin/turlar/${t.id}`}>Duzenle</a></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
