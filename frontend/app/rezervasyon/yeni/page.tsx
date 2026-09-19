// Akis: yolcu bilgileri → ozet → iyzico odeme formu

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

interface Draft { departure_id: string; adult_count: number; child_count: number;
  travelers: { full_name: string; birth_date: string; id_number: string; is_child: boolean }[] }

export default function NewBookingPage() {
  const router = useRouter();
  const [draft, setDraft] = useState<Draft | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const raw = localStorage.getItem("booking-draft");
    if (!raw) { router.replace("/"); return; }
    setDraft(JSON.parse(raw));
  }, [router]);

  if (!draft) return <p className="text-center py-20">Yukleniyor...</p>;

  function updateTraveler(i: number, field: string, value: string) {
    const t = [...(draft!.travelers)];
    t[i] = { ...t[i], [field]: value };
    setDraft({ ...(draft as Draft), travelers: t });
  }

  async function createBooking() {
    setError(""); setLoading(true);
    try {
      const b = await api.post("/bookings", draft, true); // rezervasyon durumu: pending
      localStorage.removeItem("booking-draft");
      router.push(`/rezervasyon/${b.id}`); // odeme adimina gec
    } catch (e: any) {
      setError(e.message);
    } finally { setLoading(false); }
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <h1 className="text-2xl font-bold mb-6">Yolcu Bilgileri</h1>
      <div className="space-y-4">
        {draft.travelers.map((t, i) => (
          <div key={i} className="bg-white border rounded-xl p-4">
            <p className="font-semibold mb-3">{t.is_child ? `Cocuk ${i - draft.adult_count + 1}` : `Yetiskin ${i + 1}`}</p>
            <div className="grid sm:grid-cols-3 gap-3">
              <input placeholder="Ad Soyad" value={t.full_name} onChange={(e) => updateTraveler(i, "full_name", e.target.value)}
                className="border rounded-lg px-3 py-2" />
              <input type="date" value={t.birth_date} onChange={(e) => updateTraveler(i, "birth_date", e.target.value)}
                className="border rounded-lg px-3 py-2" />
              <input placeholder="T.C. / Pasaport No" value={t.id_number} onChange={(e) => updateTraveler(i, "id_number", e.target.value)}
                className="border rounded-lg px-3 py-2" />
            </div>
          </div>
        ))}
      </div>
      {error && <p className="text-red-600 text-sm mt-4">{error}</p>}
      <button onClick={createBooking} disabled={loading}
        className="w-full mt-6 bg-brand-600 text-white py-3 rounded-lg font-semibold hover:bg-brand-700 disabled:opacity-50">
        {loading ? "Rezervasyon olusturuluyor..." : "Ozet: Odemeye Gec"}
      </button>
      <p className="text-xs text-gray-400 mt-2">Kimlik/pasaport bilgileriniz KVKK uyarinca sifrelenerek saklanir.</p>
    </div>
  );
}
