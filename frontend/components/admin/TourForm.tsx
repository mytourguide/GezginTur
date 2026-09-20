"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import type { Category, Tour } from "@/lib/types";

// AdminTourInput modelinin tam UI karsiligi:
// temel bilgiler + dinamik gorsel/dahil/haric/getirilecekler/gun-programi + kalkis yonetimi
export default function TourForm({ tour, categories }: { tour?: Tour; categories: Category[] }) {
  const router = useRouter();
  const [form, setForm] = useState({
    title: tour?.title ?? "",
    slug: tour?.slug ?? "",
    description: tour?.description ?? "",
    category_id: tour?.category_id ?? "",
    cover_image: tour?.cover_image ?? "",
    duration_days: tour?.duration_days ?? 1,
    duration_nights: tour?.duration_nights ?? 0,
    location: tour?.location ?? "",
    base_price: tour?.base_price ?? 0,
    currency: tour?.currency ?? "TRY",
    active: tour?.active ?? true,
    featured: tour?.featured ?? false,
    publish_order: tour?.publish_order ?? 0,
    images: tour?.images?.map((i) => i.image_url) ?? [],
    included: tour?.included?.map((i) => i.description) ?? [],
    excluded: tour?.excluded?.map((i) => i.description) ?? [],
    bring_items: tour?.bring_items?.map((i) => i.description) ?? [],
    itinerary: tour?.itinerary?.map((d) => ({ day_no: d.day_no, title: d.title, description: d.description })) ?? [],
  });
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);

  function set<K extends keyof typeof form>(k: K, v: (typeof form)[K]) {
    setForm({ ...form, [k]: v });
  }

  // Dinamik liste yardimcilari
  function listOps(key: "images" | "included" | "excluded" | "bring_items") {
    return {
      add: () => set(key, [...(form[key] as string[]), ""]),
      update: (i: number, v: string) => {
        const arr = [...(form[key] as string[])]; arr[i] = v; set(key, arr);
      },
      remove: (i: number) => set(key, (form[key] as string[]).filter((_, j) => j !== i)),
    };
  }

  // S3 imzali yukleme: once PUT ile dogrudan S3'e, sonra public URL'i listeye ekle
  async function uploadImage(file: File) {
    setUploading(true); setError("");
    try {
      const sign = await api.post("/uploads/sign", { filename: file.name, content_type: file.type }, true);
      const put = await fetch(sign.upload_url, { method: "PUT", body: file, headers: { "Content-Type": file.type } });
      if (!put.ok) throw new Error("S3 yukleme basarisiz");
      set("images", [...form.images, sign.public_url]);
    } catch (e: any) {
      setError(e.message + " — S3 ayarlari (.env) kontrol edin");
    } finally { setUploading(false); }
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault(); setError(""); setMsg("");
    try {
      // Bos satirlari temizle
      const payload = {
        ...form,
        images: form.images.filter((s) => s.trim()),
        included: form.included.filter((s) => s.trim()),
        excluded: form.excluded.filter((s) => s.trim()),
        bring_items: form.bring_items.filter((s) => s.trim()),
        itinerary: form.itinerary.filter((d) => d.title.trim()),
      };
      if (tour) await api.put(`/admin/tours/${tour.id}`, payload, true);
      else await api.post("/admin/tours", payload, true);
      setMsg("Kaydedildi. Turlar listesine yonlendiriliyorsunuz...");
      setTimeout(() => router.push("/admin/turlar"), 800);
    } catch (err: any) { setError(err.message); }
  }

  const img = listOps("images"), inc = listOps("included"), exc = listOps("excluded"), bring = listOps("bring_items");
  const input = "border rounded-lg px-3 py-2 w-full";

  return (
    <form onSubmit={submit} className="space-y-6">
      {/* Temel bilgiler */}
      <section className="bg-white border rounded-xl p-4 grid sm:grid-cols-2 gap-3">
        <h2 className="font-bold sm:col-span-2">Temel Bilgiler</h2>
        <input required placeholder="Tur basligi" value={form.title} className={input}
          onChange={(e) => set("title", e.target.value)} />
        <input required placeholder="slug (ornek: kapadokya-turu)" value={form.slug} className={input}
          onChange={(e) => set("slug", e.target.value)} />
        <textarea required placeholder="Aciklama" value={form.description} rows={3} className={`${input} sm:col-span-2`}
          onChange={(e) => set("description", e.target.value)} />
        <select required value={form.category_id} className={input} onChange={(e) => set("category_id", e.target.value)}>
          <option value="">Kategori secin</option>
          {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
        <input placeholder="Lokasyon" value={form.location} className={input}
          onChange={(e) => set("location", e.target.value)} />
        <label className="flex items-center gap-2 text-sm">
          <input type="number" min={1} value={form.duration_days} className={input}
            onChange={(e) => set("duration_days", +e.target.value)} /> gun
          <input type="number" min={0} value={form.duration_nights} className={input}
            onChange={(e) => set("duration_nights", +e.target.value)} /> gece
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="number" step="0.01" min={0} value={form.base_price} className={input}
            onChange={(e) => set("base_price", +e.target.value)} /> {form.currency} (taban fiyat)
        </label>
        <label className="flex items-center gap-4 text-sm flex-wrap">
          <span><input type="checkbox" checked={form.active} onChange={(e) => set("active", e.target.checked)} /> Aktif</span>
          <span><input type="checkbox" checked={form.featured} onChange={(e) => set("featured", e.target.checked)} /> Anasayfada Yayinla</span>
          {form.featured && (
            <span className="flex items-center gap-1">
              Siralama:
              <input type="number" value={form.publish_order} className={`${input} w-20`}
                onChange={(e) => set("publish_order", +e.target.value)} />
              <span className="text-gray-400">(buyuk sayi once gosterilir)</span>
            </span>
          )}
        </label>
      </section>

      {/* Gorseller: dosya yukleme (S3) + URL listesi */}
      <section className="bg-white border rounded-xl p-4">
        <h2 className="font-bold mb-3">Gorseller</h2>
        <input type="file" accept="image/*" disabled={uploading} className="text-sm"
          onChange={(e) => e.target.files?.[0] && uploadImage(e.target.files[0])} />
        {uploading && <p className="text-sm text-gray-500 mt-1">S3&apos;e yukleniyor...</p>}
        {form.images.map((url, i) => (
          <div key={i} className="flex gap-2 mt-2">
            <input value={url} className={input} onChange={(e) => img.update(i, e.target.value)} />
            <button type="button" onClick={() => img.remove(i)} className="text-red-600 px-3">Sil</button>
          </div>
        ))}
        <button type="button" onClick={img.add} className="mt-2 text-sm text-brand-600 hover:underline">+ URL ile ekle</button>
      </section>

      {/* Dahil / Haric / Getirilecekler */}
      {([["included", "Tura Dahil", inc], ["excluded", "Tura Dahil Degil", exc], ["bring_items", "Yaninizda Getirin", bring]] as const).map(([key, title, ops]) => (
        <section key={key} className="bg-white border rounded-xl p-4">
          <h2 className="font-bold mb-3">{title}</h2>
          {(form[key] as string[]).map((v, i) => (
            <div key={i} className="flex gap-2 mt-2">
              <input value={v} className={input} onChange={(e) => ops.update(i, e.target.value)} />
              <button type="button" onClick={() => ops.remove(i)} className="text-red-600 px-3">Sil</button>
            </div>
          ))}
          <button type="button" onClick={() => ops.add()} className="mt-2 text-sm text-brand-600 hover:underline">+ Madde ekle</button>
        </section>
      ))}

      {/* Gun gun program */}
      <section className="bg-white border rounded-xl p-4">
        <h2 className="font-bold mb-3">Gun Gun Program</h2>
        {form.itinerary.map((d, i) => (
          <div key={i} className="grid sm:grid-cols-[80px_1fr_auto] gap-2 mt-2">
            <input type="number" min={1} value={d.day_no} className={input}
              onChange={(e) => {
                const it = [...form.itinerary]; it[i] = { ...d, day_no: +e.target.value }; set("itinerary", it);
              }} />
            <div className="space-y-2">
              <input placeholder="Gun basligi" value={d.title} className={input}
                onChange={(e) => { const it = [...form.itinerary]; it[i] = { ...d, title: e.target.value }; set("itinerary", it); }} />
              <textarea placeholder="Aciklama" value={d.description} rows={2} className={input}
                onChange={(e) => { const it = [...form.itinerary]; it[i] = { ...d, description: e.target.value }; set("itinerary", it); }} />
            </div>
            <button type="button" onClick={() => set("itinerary", form.itinerary.filter((_, j) => j !== i))}
              className="text-red-600 px-3">Sil</button>
          </div>
        ))}
        <button type="button"
          onClick={() => set("itinerary", [...form.itinerary, { day_no: form.itinerary.length + 1, title: "", description: "" }])}
          className="mt-2 text-sm text-brand-600 hover:underline">+ Gun ekle</button>
      </section>

      {/* Kalkis tarihleri (duzenleme modunda mevcut listeyi goster) */}
      {tour && <DeparturesEditor tourId={tour.id} departures={tour.departures ?? []} />}

      {error && <p className="text-red-600 text-sm">{error}</p>}
      {msg && <p className="text-green-600 text-sm">{msg}</p>}
      <button className="bg-brand-600 text-white px-6 py-3 rounded-lg font-semibold hover:bg-brand-700">
        {tour ? "Guncelle" : "Olustur"}
      </button>
    </form>
  );
}

// Kalkis (tarih/kontenjan/fiyat) yonetimi — yalnizca duzenleme modunda
function DeparturesEditor({ tourId, departures }: { tourId: string; departures: any[] }) {
  const [rows, setRows] = useState(departures);
  const [form, setForm] = useState({ start_date: "", end_date: "", capacity: 20, price: 0 });

  async function add(e: React.FormEvent) {
    e.preventDefault();
    const d = await api.post("/admin/departures", { tour_id: tourId, ...form }, true);
    setRows([d, ...rows]);
  }
  async function remove(id: string) {
    if (!confirm("Kalkis silinsin mi? (Satisi varsa silinmez)")) return;
    await api.del(`/admin/departures/${id}`);
    setRows(rows.filter((r) => r.id !== id));
  }

  return (
    <section className="bg-white border rounded-xl p-4">
      <h2 className="font-bold mb-3">Kalkis Tarihleri</h2>
      <form onSubmit={add} className="grid sm:grid-cols-5 gap-2 text-sm">
        <input type="date" required value={form.start_date} onChange={(e) => setForm({ ...form, start_date: e.target.value })} className="border rounded-lg px-2 py-2" />
        <input type="date" required value={form.end_date} onChange={(e) => setForm({ ...form, end_date: e.target.value })} className="border rounded-lg px-2 py-2" />
        <input type="number" min={1} value={form.capacity} onChange={(e) => setForm({ ...form, capacity: +e.target.value })} className="border rounded-lg px-2 py-2" placeholder="Kontenjan" />
        <input type="number" step="0.01" min={0} value={form.price} onChange={(e) => setForm({ ...form, price: +e.target.value })} className="border rounded-lg px-2 py-2" placeholder="Fiyat" />
        <button className="bg-brand-600 text-white rounded-lg">Ekle</button>
      </form>
      <div className="mt-3 divide-y text-sm">
        {rows.map((d) => (
          <div key={d.id} className="flex justify-between py-2">
            <span>{new Date(d.start_date).toLocaleDateString("tr-TR")} → {new Date(d.end_date).toLocaleDateString("tr-TR")}</span>
            <span>{d.filled}/{d.capacity} — {new Intl.NumberFormat("tr-TR", { style: "currency", currency: "TRY", maximumFractionDigits: 0 }).format(d.price)}</span>
            <button type="button" onClick={() => remove(d.id)} className="text-red-600 hover:underline">Sil</button>
          </div>
        ))}
      </div>
    </section>
  );
}
