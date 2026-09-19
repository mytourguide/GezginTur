// Liste sayfalarinda kullanilan tur karti

import Image from "next/image";
import Link from "next/link";
import type { Tour } from "@/lib/types";

export function formatPrice(price: number) {
  return new Intl.NumberFormat("tr-TR", { style: "currency", currency: "TRY", maximumFractionDigits: 0 }).format(price);
}

export default function TourCard({ tour, lang = "tr" }: { tour: Tour; lang?: "tr" | "en" }) {
  const en = lang === "en";
  return (
    <Link
      href={`/turlar/${tour.slug}`}
      className="group block bg-white rounded-2xl overflow-hidden shadow hover:shadow-lg transition"
    >
      <div className="relative h-48 w-full">
        <Image src={tour.cover_image} alt={tour.title} fill className="object-cover group-hover:scale-105 transition" />
      </div>
      <div className="p-4">
        <p className="text-xs text-brand-600 font-semibold uppercase">{tour.category_name}</p>
        <h3 className="font-bold mt-1 group-hover:text-brand-700">{tour.title}</h3>
        <p className="text-sm text-gray-500 mt-1">{tour.location} - {tour.duration_days} {en ? "day(s)" : "gun"} {tour.duration_nights} {en ? "night(s)" : "gece"}</p>
        <p className="mt-2 font-bold text-lg">{formatPrice(tour.base_price)}<span className="text-xs font-normal text-gray-500"> {en ? "from" : "'den itibaren"}</span></p>
      </div>
    </Link>
  );
}
