"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import TourForm from "@/components/admin/TourForm";
import type { Category } from "@/lib/types";

export default function NewTourPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  useEffect(() => { api.get("/categories").then(setCategories); }, []);
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Yeni Tur</h1>
      <TourForm categories={categories} />
    </div>
  );
}
