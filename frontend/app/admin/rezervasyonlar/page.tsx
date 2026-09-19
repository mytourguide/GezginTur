"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";
import type { Booking } from "@/lib/types";

export default function AdminBookingsPage() {
  const [list, setList] = useState<Booking[]>([]);
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");

  const load = () => {
    const qs = new URLSearchParams({ status, search }).toString();
    api.get(`/admin/bookings?${qs}`, true).then(setList);
  };
  useEffect(load, [status]); // eslint-disable-line

  async function refund(id: string) {
    if (!confirm("Tam iade yapilsin mi? Kontenjan otomatik iade edilir.")) return;
    await api.post(`/admin/bookings/${id}/refund`, {}, true);
    load();
  }
  async function setBookingStatus(id: string, s: string) {
    await api.patch(`/admin/bookings/${id}/status`, { status: s }, true);
    load();
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Rezervasyonlar</h1>
      <div className="flex gap-3 mb-4">
        <select value={status} onChange={(e) => setStatus(e.target.value)} className="border rounded-lg px-3 py-2">
          <option value="">Tumu</option>
          {["pending", "paid", "confirmed", "cancelled", "failed"].map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
        <input value={search} onChange={(e) => setSearch(e.target.value)} onKeyDown={(e) => e.key === "Enter" && load()}
          placeholder="Musteri ara..." className="border rounded-lg px-3 py-2" />
      </div>
      <div className="bg-white border rounded-xl overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left">
            <tr><th className="p-3">Tur</th><th className="p-3">Kisi</th><th className="p-3">Tutar</th><th className="p-3">Durum</th><th className="p-3">Islemler</th></tr>
          </thead>
          <tbody className="divide-y">
            {list.map((b) => (
              <tr key={b.id}>
                <td className="p-3">{b.tour_title}<br /><span className="text-xs text-gray-400">{b.id.slice(0, 8)}</span></td>
                <td className="p-3">{b.adult_count}+{b.child_count}</td>
                <td className="p-3">{formatPrice(b.total_price)}</td>
                <td className="p-3">
                  <select value={b.status} onChange={(e) => setBookingStatus(b.id, e.target.value)}
                    className="border rounded px-2 py-1 text-xs">
                    {["pending", "paid", "confirmed", "cancelled", "failed"].map((s) => <option key={s} value={s}>{s}</option>)}
                  </select>
                </td>
                <td className="p-3">
                  {(b.status === "paid" || b.status === "confirmed") && (
                    <button onClick={() => refund(b.id)} className="text-red-600 text-xs hover:underline">Iade Et</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
