import Image from "next/image";
import type { Metadata } from "next";
import BookingWidget from "@/components/BookingWidget";
import type { Tour, TourImage } from "@/lib/types";
import { apiUrl, getLang } from "@/lib/lang";

async function getTour(slug: string): Promise<Tour> {
  const res = await fetch(apiUrl(`/tours/${slug}`), { cache: "no-store" });
  if (!res.ok) throw new Error("Tur bulunamadi");
  return res.json();
}

// SEO: tur basligi/aciklamasi meta etiketlere islenir
export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const tour = await getTour(params.slug);
  return { title: tour.title, description: tour.description.slice(0, 160), openGraph: { images: [tour.cover_image] } };
}

// SSS (ornek icerik — admin panelden yonetilebilir hale getirilebilir)
const FAQS = {
  tr: [
    { q: "Iptal politikasi nedir?", a: "Kalkistan 7 gun oncesine kadar ucretsiz iptal edebilirsiniz." },
    { q: "Cocuk indirimi var mi?", a: "2-12 yas cocuklar yetiskin fiyatinin %70'i uzerinden ucretlendirilir." },
    { q: "Odeme guvenli mi?", a: "Tum odemeler iyzico altyapisinda, 3D Secure ile islenir; kart bilgileriniz tarafimizca saklanmaz." },
  ],
  en: [
    { q: "What is the cancellation policy?", a: "You can cancel free of charge up to 7 days before departure." },
    { q: "Is there a child discount?", a: "Children aged 2-12 are charged at 70% of the adult price." },
    { q: "Is payment secure?", a: "All payments are processed via iyzico with 3D Secure; your card details are never stored by us." },
  ],
};

const TXT = {
  tr: { itinerary: "Gun Gun Program", included: "Tura Dahil", excluded: "Tura Dahil Degil", bring: "Yaninizda Getirin",
        faq: "Sik Sorulan Sorular", similar: "Benzer Turlar", viewAll: "kategorisindeki tum turlari goruntule →", day: "gun", night: "gece" },
  en: { itinerary: "Day by Day Itinerary", included: "Included", excluded: "Not Included", bring: "What to Bring",
        faq: "Frequently Asked Questions", similar: "Similar Tours", viewAll: "View all tours in this category →", day: "day(s)", night: "night(s)" },
};

export default async function TourDetailPage({ params }: { params: { slug: string } }) {
  const tour = await getTour(params.slug);
  const lang = getLang();
  const txt = TXT[lang];
  const faqs = FAQS[lang];
  const images: TourImage[] = tour.images?.length ? tour.images : [{ id: "x", tour_id: tour.id, image_url: tour.cover_image, sort_order: 0 }];

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Galeri */}
      <div className="grid grid-cols-2 md:grid-cols-3 gap-2 rounded-2xl overflow-hidden">
        {images.slice(0, 3).map((img, i) => (
          <div key={i} className={`relative h-64 ${i === 0 ? "col-span-2 md:col-span-2 md:row-span-2 h-80 md:h-full" : ""}`}>
            <Image src={img.image_url} alt={tour.title} fill className="object-cover" priority={i === 0} />
          </div>
        ))}
      </div>

      <div className="grid lg:grid-cols-3 gap-8 mt-8">
        <div className="lg:col-span-2">
          <p className="text-sm text-brand-600 font-semibold uppercase">{tour.category_name}</p>
          <h1 className="text-3xl font-bold mt-1">{tour.title}</h1>
          <p className="text-gray-500 mt-2">{tour.location} - {tour.duration_days} {txt.day} / {tour.duration_nights} {txt.night}</p>
          <p className="mt-4 leading-relaxed">{tour.description}</p>

          {/* Gun gun program */}
          {(tour.itinerary?.length ?? 0) > 0 && (
          <>
          <h2 className="text-xl font-bold mt-8 mb-4">{txt.itinerary}</h2>
          <ol className="relative border-l-2 border-brand-200 ml-2 space-y-6">
            {tour.itinerary?.map((d) => (
              <li key={d.id} className="ml-6">
                <span className="absolute -left-3 flex h-6 w-6 items-center justify-center rounded-full bg-brand-600 text-white text-xs font-bold">
                  {d.day_no}
                </span>
                <h3 className="font-semibold">{d.title}</h3>
                <p className="text-sm text-gray-600 mt-1">{d.description}</p>
              </li>
            ))}
          </ol>
          </>
          )}

          {/* Dahil / Haric / Yaninizda getirin */}
          <div className="grid md:grid-cols-3 gap-6 mt-8">
            <ListCard title={txt.included} items={tour.included} color="text-green-600" icon="✓" />
            <ListCard title={txt.excluded} items={tour.excluded} color="text-red-500" icon="✕" />
            <ListCard title={txt.bring} items={tour.bring_items} color="text-blue-600" icon="🎒" />
          </div>

          {/* SSS */}
          <h2 className="text-xl font-bold mt-8 mb-4">{txt.faq}</h2>
          <div className="space-y-2">
            {faqs.map((f) => (
              <details key={f.q} className="bg-white rounded-lg border p-4">
                <summary className="font-medium cursor-pointer">{f.q}</summary>
                <p className="text-sm text-gray-600 mt-2">{f.a}</p>
              </details>
            ))}
          </div>
        </div>

        {/* Sag kolon: rezervasyon widget'i */}
        <aside className="lg:sticky lg:top-24 self-start">
          <BookingWidget tour={tour} lang={lang} />
        </aside>
      </div>

      {/* Benzer turlar */}
      <div className="mt-12">
        <h2 className="text-xl font-bold mb-4">{txt.similar}</h2>
        <a href={`/turlar?category=${tour.category_id}`} className="text-brand-600 hover:underline text-sm">
          {tour.category_name} {txt.viewAll}
        </a>
      </div>
    </div>
  );
}

function ListCard({ title, items, color, icon }: { title: string; items?: { id: string; description: string }[]; color: string; icon: string }) {
  return (
    <div className="bg-white rounded-xl border p-4">
      <h3 className="font-semibold mb-3">{title}</h3>
      <ul className="space-y-2 text-sm">
        {items?.map((it) => (
          <li key={it.id} className="flex gap-2">
            <span className={color}>{icon}</span> {it.description}
          </li>
        ))}
      </ul>
    </div>
  );
}
