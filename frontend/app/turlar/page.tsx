// Cloudflare Pages (next-on-pages) icin edge runtime zorunlu
export const runtime = "edge";
import TourCard from "@/components/TourCard";
import type { TourListResponse } from "@/lib/types";
import { apiUrl, getLang } from "@/lib/lang";

const TXT = {
  tr: { title: "Turlar", searchPh: "Lokasyon / tur adi", min: "Min TL", max: "Max TL", sortPh: "Siralama",
        sortAsc: "Fiyat (Artan)", sortDesc: "Fiyat (Azalan)", sortPop: "Populerlik", filter: "Filtrele", found: "tur bulundu" },
  en: { title: "Tours", searchPh: "Location / tour name", min: "Min TRY", max: "Max TRY", sortPh: "Sort",
        sortAsc: "Price (Low to High)", sortDesc: "Price (High to Low)", sortPop: "Popularity", filter: "Filter", found: "tours found" },
};

// searchParams ile filtre/siralama (fiyat, tarih, sure, kategori)
export default async function ToursPage({ searchParams }: { searchParams: Record<string, string> }) {
  const qs = new URLSearchParams(searchParams).toString();
  // API hatasi durumunda 500 yerine bos liste goster
  let data: TourListResponse = { tours: [], total: 0, page: 1, limit: 0 };
  try {
    data = await fetch(apiUrl(`/tours?${qs}`), { cache: "no-store" }).then((r) => r.json());
  } catch { /* API ulasilamazsa bos liste ile devam */ }
  const t = TXT[getLang()];

  return (
    <div className="max-w-6xl mx-auto px-4 py-10">
      <h1 className="text-3xl font-bold mb-6">{t.title}</h1>
      {/* Filtre cubugu: GET formu oldugu icin SSR ile calisir */}
      <form className="grid sm:grid-cols-4 gap-3 bg-white p-4 rounded-2xl shadow mb-8">
        <input name="search" defaultValue={searchParams.search} placeholder={t.searchPh} className="border rounded-lg px-3 py-2" />
        <input name="min_price" type="number" defaultValue={searchParams.min_price} placeholder={t.min} className="border rounded-lg px-3 py-2" />
        <input name="max_price" type="number" defaultValue={searchParams.max_price} placeholder={t.max} className="border rounded-lg px-3 py-2" />
        <select name="sort" defaultValue={searchParams.sort} className="border rounded-lg px-3 py-2">
          <option value="">{t.sortPh}</option>
          <option value="price_asc">{t.sortAsc}</option>
          <option value="price_desc">{t.sortDesc}</option>
          <option value="popular">{t.sortPop}</option>
        </select>
        <input type="hidden" name="category" value={searchParams.category || ""} />
        <button className="sm:col-span-4 bg-brand-600 text-white py-2 rounded-lg hover:bg-brand-700">{t.filter}</button>
      </form>

      <p className="text-sm text-gray-500 mb-4">{data.total} {t.found}</p>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {(data.tours ?? []).map((t) => <TourCard key={t.id} tour={t} lang={getLang()} />)}
      </div>
    </div>
  );
}
