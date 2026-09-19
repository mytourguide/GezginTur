"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/components/TourCard";
import type { Booking } from "@/lib/types";

const STATUS_TR: Record<string, string> = {
  pending: "Odeme Bekliyor", paid: "Odendi", confirmed: "Onaylandi", cancelled: "Iptal", failed: "Basarisiz",
};

export default function AccountPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [profile, setProfile] = useState({ full_name: "", phone: "" });
  const [msg, setMsg] = useState("");

  useEffect(() => {
    api.get("/bookings/mine", true).then(setBookings);
    api.get("/auth/me", true).then((u) => setProfile({ full_name: u.full_name, phone: u.phone || "" }));
  }, []);

  async function cancel(id: string) {
    if (!confirm("Rezervasyonu iptal etmek istediginize emin misiniz?")) return;
    await api.post(`/bookings/${id}/cancel`, {}, true);
    setBookings((bs) => bs.map((b) => (b.id === id ? { ...b, status: "cancelled" } : b)));
  }

  async function saveProfile(e: React.FormEvent) {
    e.preventDefault();
    await api.put("/auth/me", profile, true);
    setMsg("Profil guncellendi");
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-10">
      <h1 className="text-2xl font-bold mb-6">Hesabim</h1>

      <form onSubmit={saveProfile} className="bg-white border rounded-xl p-4 mb-8 grid sm:grid-cols-3 gap-3">
        <input value={profile.full_name} onChange={(e) => setProfile({ ...profile, full_name: e.target.value })}
          className="border rounded-lg px-3 py-2" placeholder="Ad Soyad" />
        <input value={profile.phone} onChange={(e) => setProfile({ ...profile, phone: e.target.value })}
          className="border rounded-lg px-3 py-2" placeholder="Telefon" />
        <button className="bg-brand-600 text-white rounded-lg">Kaydet</button>
        {msg && <p className="text-green-600 text-sm sm:col-span-3">{msg}</p>}
      </form>

      <TwoFactorSection />

      <h2 className="font-bold mb-4">Rezervasyonlarim</h2>
      <div className="space-y-3">
        {bookings.map((b) => (
          <div key={b.id} className="bg-white border rounded-xl p-4 flex flex-wrap justify-between items-center gap-3">
            <div>
              <p className="font-semibold">{b.tour_title}</p>
              <p className="text-sm text-gray-500">
                {b.adult_count} yetiskin + {b.child_count} cocuk - {formatPrice(b.total_price)} - {new Date(b.created_at).toLocaleDateString("tr-TR")}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span className={`text-sm font-medium px-3 py-1 rounded-full ${
                b.status === "paid" || b.status === "confirmed" ? "bg-green-100 text-green-700"
                : b.status === "cancelled" || b.status === "failed" ? "bg-red-100 text-red-700" : "bg-amber-100 text-amber-700"}`}>
                {STATUS_TR[b.status] || b.status}
              </span>
              {b.status === "pending" && <a href={`/rezervasyon/${b.id}`} className="text-brand-600 text-sm hover:underline">Ode</a>}
              {b.status === "pending" && <button onClick={() => cancel(b.id)} className="text-red-600 text-sm hover:underline">Iptal</button>}
            </div>
          </div>
        ))}
        {bookings.length === 0 && <p className="text-gray-500">Henuz rezervasyonunuz yok.</p>}
      </div>
    </div>
  );
}

// Iki asamali dogrulama bolumu: kurulum (secret + otpauth URL) → dogrulama → aktif/pasif
function TwoFactorSection() {
  const [enabled, setEnabled] = useState(false);
  const [setup, setSetup] = useState<{ secret: string; otp_url: string } | null>(null);
  const [code, setCode] = useState("");
  const [msg, setMsg] = useState("");

  useEffect(() => {
    api.get("/auth/2fa/status", true).then((r) => setEnabled(r.enabled));
  }, []);

  async function startSetup() {
    setMsg("");
    setSetup(await api.post("/auth/2fa/setup", {}, true));
  }
  async function enable(e: React.FormEvent) {
    e.preventDefault(); setMsg("");
    try {
      await api.post("/auth/2fa/enable", { code }, true);
      setEnabled(true); setSetup(null); setCode("");
      setMsg("2FA aktif edildi. Bundan sonra girislerde kod istenecek.");
    } catch (err: any) { setMsg(err.message); }
  }
  async function disable() {
    if (!confirm("2FA kapatilsin mi? Hesabiniz daha az korunur hale gelir.")) return;
    await api.post("/auth/2fa/disable", {}, true);
    setEnabled(false);
    setMsg("2FA kapatildi.");
  }

  return (
    <div className="bg-white border rounded-xl p-4 mb-8">
      <div className="flex justify-between items-center">
        <h2 className="font-bold">Iki Asamali Dogrulama (2FA)</h2>
        <span className={`text-sm px-3 py-1 rounded-full ${enabled ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-600"}`}>
          {enabled ? "Aktif" : "Pasif"}
        </span>
      </div>
      <p className="text-sm text-gray-500 mt-1">
        Google Authenticator gibi bir uygulama ile hesabiniza ek guvenlik katmani ekleyin.
      </p>
      {enabled ? (
        <button onClick={disable} className="mt-3 text-red-600 text-sm hover:underline">2FA'yı Kapat</button>
      ) : setup ? (
        <form onSubmit={enable} className="mt-4 space-y-3">
          <div className="bg-gray-50 rounded-lg p-3 text-sm">
            <p className="font-mono break-all"><b>Gizli anahtar:</b> {setup.secret}</p>
            <p className="mt-1 break-all text-xs text-gray-500">{setup.otp_url}</p>
            <p className="text-xs text-gray-500 mt-1">Bu URL'i Google Authenticator'a taratın veya anahtarı elle girin.</p>
          </div>
          <div className="flex gap-2">
            <input value={code} onChange={(e) => setCode(e.target.value)} placeholder="6 haneli kod"
              maxLength={6} className="border rounded-lg px-3 py-2 w-40 text-center tracking-widest" />
            <button className="bg-brand-600 text-white px-4 rounded-lg">Dogrula ve Aktif Et</button>
          </div>
        </form>
      ) : (
        <button onClick={startSetup} className="mt-3 bg-brand-600 text-white px-4 py-2 rounded-lg text-sm">
          2FA Kurulumunu Baslat
        </button>
      )}
      {msg && <p className="text-sm mt-2 text-gray-600">{msg}</p>}
    </div>
  );
}
