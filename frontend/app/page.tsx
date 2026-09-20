import Link from "next/link";
import TourCard from "@/components/TourCard";
import type { Category, TourListResponse } from "@/lib/types";
import { apiUrl, getLang } from "@/lib/lang";

// Ana sayfa statik metinleri (dil seciciyle SSR eslesir)
const TXT = {
  tr: {
    heroTitle: "Hayalinizdeki Turu Kesfedin",
    heroSub: "Yurt ici ve yurt disinda yuzlerce rota",
    searchPlaceholder: "Nereye gitmek istersiniz?",
    searchBtn: "Ara",
    featuredTitle: "One Cikan Turlar",
  },
  en: {
    heroTitle: "Discover the Tour of Your Dreams",
    heroSub: "Hundreds of routes domestically and abroad",
    searchPlaceholder: "Where would you like to go?",
    searchBtn: "Search",
    featuredTitle: "Featured Tours",
  },
};

async function getData() {
  // SSR: her istekte taze veri (SEO icin metadata layout'ta tanimli)
  // Anasayfa adedi admin ayarlarindan okunur (varsayilan 6)
  let limit = 6;
  try {
    const s = await fetch(apiUrl("/settings/public"), { cache: "no-store" }).then((r) => r.json());
    const n = parseInt(s.homepage_featured_count, 10);
    if (Number.isFinite(n) && n > 0) limit = n;
  } catch { /* varsayilan kullanilir */ }
  const [featured, categories] = await Promise.all([
    fetch(apiUrl(`/tours?featured=true&limit=${limit}`), { cache: "no-store" }).then((r) => r.json() as Promise<TourListResponse>),
    fetch(apiUrl("/categories"), { cache: "no-store" }).then((r) => r.json() as Promise<Category[]>),
  ]);
  return { featured: featured.tours ?? [], categories: categories ?? [], limit };
}

export default async function HomePage() {
  const { featured, categories } = await getData();
  const lang = getLang();
  const tx = TXT[lang];

  return (
    <div>
      {/* Hero + arama */}
      <section className="relative h-[420px] flex items-center justify-center text-white"
        style={{ backgroundImage: "url(https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=1600&q=80)", backgroundSize: "cover", backgroundPosition: "center" }}>
        <div className="absolute inset-0 bg-black/40" />
        <div className="relative text-center px-4">
          <h1 className="text-4xl md:text-5xl font-extrabold">{tx.heroTitle}</h1>
          <p className="mt-3 text-lg">{tx.heroSub}</p>
          {/* Lokasyon/tarih aramasi liste sayfasina yonlendirir */}
          <form action="/turlar" className="mt-6 flex flex-col sm:flex-row gap-2 max-w-xl mx-auto">
            <input name="search" placeholder={tx.searchPlaceholder} className="flex-1 px-4 py-3 rounded-lg text-gray-900" />
            <input name="start_date" type="date" className="px-4 py-3 rounded-lg text-gray-900" />
            <button className="bg-brand-600 px-6 py-3 rounded-lg font-semibold hover:bg-brand-700">{tx.searchBtn}</button>
          </form>
        </div>
      </section>

      {/* Kategori filtreleri */}
      <section className="max-w-6xl mx-auto px-4 mt-10">
        <div className="flex flex-wrap gap-2">
          {categories.map((c) => (
            <Link key={c.id} href={`/turlar?category=${c.slug}`}
              className="px-4 py-2 bg-white rounded-full border text-sm hover:border-brand-500 hover:text-brand-600">
              {c.name}
            </Link>
          ))}
        </div>
      </section>

      {/* One cikan turlar */}
      <section className="max-w-6xl mx-auto px-4 mt-10">
        <h2 className="text-2xl font-bold mb-6">{tx.featuredTitle}</h2>
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {featured.map((t) => <TourCard key={t.id} tour={t} lang={lang} />)}
        </div>
      </section>
    </div>
  );
}
