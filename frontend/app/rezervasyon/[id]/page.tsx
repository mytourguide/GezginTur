"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";

// Backend'den alinan checkoutFormContent (iyzico'nun urettigi <script> HTML'i)
// dogrudan sayfaya gomulur. Kart bilgileri yalnizca iyzico altyapisinda islenir.
export default function CheckoutPage({ params }: { params: { id: string } }) {
  const [content, setContent] = useState("");
  const [summary, setSummary] = useState<any>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api.get(`/bookings/${params.id}`, true).then(setSummary).catch((e) => setError(e.message));
    api.post("/payments/checkout/init", { booking_id: params.id }, true)
      .then((res) => setContent(res.checkout_form_content))
      .catch((e) => setError(e.message + " - lutfen sayfayi yenileyip tekrar deneyin"));
  }, [params.id]);

  return (
    <div className="max-w-3xl mx-auto px-4 py-10">
      <h1 className="text-2xl font-bold mb-2">Guvenli Odeme</h1>
      {summary && (
        <div className="bg-white border rounded-xl p-4 mb-6 text-sm">
          <p><b>{summary.tour_title}</b></p>
          <p>{summary.adult_count} yetiskin, {summary.child_count} cocuk</p>
          <p className="text-lg font-bold mt-1">Toplam: {formatPrice(summary.total_price)}</p>
          <p className="text-xs text-gray-500">Taksit secenekleri: 2 / 3 / 6 / 9 - 3D Secure zorunludur</p>
        </div>
      )}
      {error && <p className="text-red-600 mb-4">{error}</p>}
      {/* iyzico formu buraya render edilir */}
      <div dangerouslySetInnerHTML={{ __html: content }} />
    </div>
  );
}
