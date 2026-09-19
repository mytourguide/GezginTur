// Dashboard: istatistik + satis grafigi (basit SVG) + doluluk
"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";
import type { DashboardStats, OccupancyRow, SalesPoint } from "@/lib/types";

export default function AdminDashboardPage() {
  const [data, setData] = useState<{ stats: DashboardStats; sales: SalesPoint[]; occupancy: OccupancyRow[] } | null>(null);

  useEffect(() => { api.get("/admin/dashboard", true).then(setData); }, []);
  if (!data) return <p>Yukleniyor...</p>;
  const { stats, sales, occupancy } = data;
  const max = Math.max(...sales.map((s) => s.total), 1);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      {/* Hizli islemler */}
      <div className="flex flex-wrap gap-2 mb-6 text-sm">
        <a href="/admin/turlar" className="px-3 py-1.5 bg-white border rounded-lg hover:border-brand-400 hover:text-brand-700">+ Yeni Tur</a>
        <a href="/admin/kuponlar" className="px-3 py-1.5 bg-white border rounded-lg hover:border-brand-400 hover:text-brand-700">+ Yeni Kupon</a>
        <a href="/admin/rezervasyonlar" className="px-3 py-1.5 bg-white border rounded-lg hover:border-brand-400 hover:text-brand-700">Bekleyen Rezervasyonlar</a>
        <a href="/admin/saglik" className="px-3 py-1.5 bg-white border rounded-lg hover:border-brand-400 hover:text-brand-700">Sistem Sagligi</a>
      </div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {([["Toplam Rezervasyon", stats.total_bookings], ["Bekleyen Odeme", stats.pending_payments],
          ["Toplam Musteri", stats.total_customers]] as const).map(([label, value]) => (
          <div key={label} className="bg-white border rounded-xl p-4">
            <p className="text-sm text-gray-500">{label}</p>
            <p className="text-2xl font-bold mt-1">{value}</p>
          </div>
        ))}
        <div className="bg-brand-600 text-white rounded-xl p-4">
          <p className="text-sm opacity-80">Ciro (Odenen)</p>
          <p className="text-2xl font-bold mt-1">{formatPrice(stats.total_revenue)}</p>
        </div>
      </div>

      {/* Gunluk satis grafigi (son 30 gun, SVG bar) */}
      <h2 className="font-bold mt-8 mb-3">Satis Grafigi (Son 30 Gun)</h2>
      <div className="bg-white border rounded-xl p-4 flex items-end gap-1 h-40 overflow-hidden">
        {sales.map((s, i) => (
          <div key={i} title={`${new Date(s.date).toLocaleDateString("tr-TR")}: ${formatPrice(s.total)}`}
            className="flex-1 bg-brand-500 rounded-t hover:bg-brand-600"
            style={{ height: `${(s.total / max) * 100}%` }} />
        ))}
      </div>

      {/* Kalkis doluluk oranlari */}
      <h2 className="font-bold mt-8 mb-3">Yaklasan Kalkis Doluluklari</h2>
      <div className="bg-white border rounded-xl divide-y">
        {occupancy.map((o, i) => (
          <div key={i} className="p-4">
            <div className="flex justify-between text-sm mb-1">
              <span>{o.tour} - {new Date(o.start_date).toLocaleDateString("tr-TR")}</span>
              <span>{o.filled}/{o.capacity} (%{Math.round((o.filled / o.capacity) * 100)})</span>
            </div>
            <div className="h-2 bg-gray-100 rounded-full">
              <div className="h-2 bg-brand-500 rounded-full" style={{ width: `${(o.filled / o.capacity) * 100}%` }} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
