"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { isLoggedIn } from "@/lib/auth";
import { formatPrice } from "./TourCard";
import type { Departure, Tour } from "@/lib/types";

const CHILD_RATIO = 0.7;

// Dil destegi: lang=tr|en cookie'sine gore statik metinler
const STR = {
  tr: {
    perPerson: "/ kisi basi", departureDate: "Kalkis Tarihi", noDates: "Bu tura ait gelecek tarih yok",
    lastSeats: "Son", seatsLeft: "yer!", adult: "Yetiskin", child: "Cocuk (2-12 yas, %70 fiyat)",
    coupon: "Indirim kuponu (opsiyonel)", total: "Toplam",
    couponNote: "Kupon indirimi sonraki adimda uygulanir.", book: "Rezervasyon Yap / Odemeye Gec",
    secure: "iyzico ile 3D Secure guvenli odeme - Taksit secenekleri",
    noSeat: "Secilen tarihte yeterli kontenjan yok",
  },
  en: {
    perPerson: "/ per person", departureDate: "Departure Date", noDates: "No upcoming dates for this tour",
    lastSeats: "Only", seatsLeft: "left!", adult: "Adult", child: "Child (2-12 yrs, 70% price)",
    coupon: "Discount coupon (optional)", total: "Total",
    couponNote: "Coupon discount is applied at the next step.", book: "Book Now / Proceed to Payment",
    secure: "Secure payment with iyzico 3D Secure - Installment options",
    noSeat: "Not enough availability on the selected date",
  },
};

// Tarih + kisi sayisi secimi, anlik fiyat hesaplama ve rezervasyon olusturma.
// Kullanici giris yapmamissa once giris sayfasina yonlendirir.
export default function BookingWidget({ tour, lang: langProp }: { tour: Tour; lang?: "tr" | "en" }) {
  const router = useRouter();
  const departures = tour.departures ?? [];
  // Server bilesen dili prop ile gecirebilir; yoksa cookie'den okunur
  const [lang, setLang] = useState<"tr" | "en">(langProp ?? "tr");

  useEffect(() => {
    if (!langProp && /(?:^|;\s*)lang=en/.test(document.cookie)) setLang("en");
    else if (langProp === "en") setLang("en");
  }, [langProp]);
  const t = STR[lang];
  const [departureId, setDepartureId] = useState(departures[0]?.id ?? "");
  const [coupon, setCoupon] = useState("");
  const [adults, setAdults] = useState(2);
  const [children, setChildren] = useState(0);
  const [error, setError] = useState("");

  const dep: Departure | undefined = departures.find((d) => d.id === departureId);
  const total = useMemo(
    () => (dep ? adults * dep.price + children * dep.price * CHILD_RATIO : 0),
    [dep, adults, children]
  );

  function submit() {
    setError("");
    if (!dep) return;
    if (!isLoggedIn()) { router.push("/giris?next=" + encodeURIComponent(`/turlar/${tour.slug}`)); return; }
    if (adults + children > (dep.available ?? 0)) { setError(t.noSeat); return; }
    // Yolcu bilgileri adimina gec; taslak localStorage'da tutulur
    const travelers = Array.from({ length: adults + children }, (_, i) => ({
      full_name: "", birth_date: "", id_number: "", is_child: i >= adults,
    }));
    localStorage.setItem("booking-draft", JSON.stringify({
      departure_id: dep.id, adult_count: adults, child_count: children, travelers,
      coupon_code: coupon.trim().toUpperCase() || undefined,
    }));
    router.push(`/rezervasyon/yeni?tour=${tour.id}&slug=${tour.slug}`);
  }

  return (
    <div className="bg-white rounded-2xl shadow-lg border p-5">
      <p className="text-2xl font-bold">{formatPrice(dep?.price ?? tour.base_price)}
        <span className="text-sm font-normal text-gray-500"> {t.perPerson}</span></p>

      <label className="block text-sm font-medium mt-4 mb-1">{t.departureDate}</label>
      <select value={departureId} onChange={(e) => setDepartureId(e.target.value)}
        className="w-full border rounded-lg px-3 py-2">
        {departures.length === 0 && <option>{t.noDates}</option>}
        {departures.map((d) => (
          <option key={d.id} value={d.id}>
            {new Date(d.start_date).toLocaleDateString(lang === "en" ? "en-GB" : "tr-TR")} - {formatPrice(d.price)}
            {(d.available ?? 0) <= 3 ? ` (${t.lastSeats} ${d.available} ${t.seatsLeft})` : ""}
          </option>
        ))}
      </select>

      {/* Yetiskin / cocuk sayaci */}
      <Counter label={t.adult} value={adults} min={1} onChange={setAdults} />
      <Counter label={t.child} value={children} min={0} onChange={setChildren} />

      {/* Indirim kuponu (backend dogrular ve rezervasyona uygular) */}
      <input value={coupon} onChange={(e) => setCoupon(e.target.value.toUpperCase())}
        placeholder={t.coupon}
        className="w-full border rounded-lg px-3 py-2 mt-4 text-sm uppercase" />

      <div className="border-t mt-4 pt-3 flex justify-between font-bold">
        <span>{t.total}</span><span>{formatPrice(total)}</span>
      </div>
      {coupon && <p className="text-xs text-brand-600 mt-1">{t.couponNote}</p>}
      {error && <p className="text-red-600 text-sm mt-2">{error}</p>}

      <button onClick={submit} disabled={!dep}
        className="w-full mt-4 bg-brand-600 text-white py-3 rounded-lg font-semibold hover:bg-brand-700 disabled:opacity-50">
        {t.book}
      </button>
      <p className="text-xs text-gray-400 mt-2 text-center">{t.secure}</p>
    </div>
  );
}

function Counter({ label, value, min, onChange }: { label: string; value: number; min: number; onChange: (n: number) => void }) {
  return (
    <div className="flex items-center justify-between mt-3">
      <span className="text-sm">{label}</span>
      <div className="flex items-center gap-3">
        <button onClick={() => onChange(Math.max(min, value - 1))} className="w-8 h-8 rounded-full border hover:bg-gray-100">−</button>
        <span className="w-6 text-center font-semibold">{value}</span>
        <button onClick={() => onChange(Math.min(20, value + 1))} className="w-8 h-8 rounded-full border hover:bg-gray-100">+</button>
      </div>
    </div>
  );
}
