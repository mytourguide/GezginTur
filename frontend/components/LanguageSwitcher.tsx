"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

// Dil secici: "lang" cookie'sini ayarlar ve sayfayi yeniler (SSR ile trambasiz calisir)
export default function LanguageSwitcher() {
  const router = useRouter();
  const [lang, setLang] = useState<"tr" | "en">("tr");

  useEffect(() => {
    const m = document.cookie.match(/(?:^|;\s*)lang=(\w+)/);
    if (m && m[1] === "en") setLang("en");
  }, []);

  function set(l: "tr" | "en") {
    document.cookie = `lang=${l};path=/;max-age=31536000`;
    setLang(l);
    router.refresh();
  }

  return (
    <div className="flex border rounded-lg overflow-hidden text-xs font-semibold">
      <button
        onClick={() => set("tr")}
        className={`px-2 py-1 ${lang === "tr" ? "bg-brand-600 text-white" : "text-gray-600 hover:bg-gray-100"}`}
      >
        TR
      </button>
      <button
        onClick={() => set("en")}
        className={`px-2 py-1 ${lang === "en" ? "bg-brand-600 text-white" : "text-gray-600 hover:bg-gray-100"}`}
      >
        EN
      </button>
    </div>
  );
}
