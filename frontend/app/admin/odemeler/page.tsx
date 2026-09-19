"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";
import type { Payment } from "@/lib/types";

export default function AdminPaymentsPage() {
  const [list, setList] = useState<Payment[]>([]);
  useEffect(() => { api.get("/admin/payments", true).then(setList); }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Odeme Kayitlari (iyzico)</h1>
      <div className="bg-white border rounded-xl overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left">
            <tr><th className="p-3">Rezervasyon</th><th className="p-3">iyzico Payment ID</th><th className="p-3">Tutar</th><th className="p-3">Taksit</th><th className="p-3">Durum</th><th className="p-3">Tarih</th></tr>
          </thead>
          <tbody className="divide-y">
            {list.map((p) => (
              <tr key={p.id}>
                <td className="p-3">{p.booking_id.slice(0, 8)}</td>
                <td className="p-3 font-mono text-xs">{p.iyzico_payment_id || "-"}</td>
                <td className="p-3">{formatPrice(p.amount)}</td>
                <td className="p-3">{p.installment}</td>
                <td className="p-3">{p.status}</td>
                <td className="p-3">{new Date(p.created_at).toLocaleString("tr-TR")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
