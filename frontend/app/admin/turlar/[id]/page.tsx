// Cloudflare Pages (next-on-pages)
export const runtime = "edge";
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import TourForm from "@/components/admin/TourForm";
import type { Category, Tour } from "@/lib/types";

export default function EditTourPage() {
  const { id } = useParams();
  const [tour, setTour] = useState<Tour | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);

  useEffect(() => {
    api.get(`/admin/tours/${id}`, true).then(setTour);
    api.get("/categories").then(setCategories);
  }, [id]);

  if (!tour) return <p>Yukleniyor...</p>;
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Tur Duzenle: {tour.title}</h1>
      <TourForm tour={tour} categories={categories} />
    </div>
  );
}
