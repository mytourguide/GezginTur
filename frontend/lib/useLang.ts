"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

export function useLang(): "tr" | "en" {
  const searchParams = useSearchParams();
  const [lang, setLang] = useState<"tr" | "en">("tr");

  useEffect(() => {
    const fromQuery = searchParams.get("lang") as "en" | "tr" | null;
    const fromCookie = /(?:^|;\s*)lang=(en|tr)/.exec(document.cookie)?.[1] as "en"|"tr"|undefined;
    const l = (fromQuery || fromCookie || "tr") as "tr"|"en";
    setLang(l);
    if (fromQuery) {
      document.cookie = `lang=${fromQuery};path=/;max-age=31536000`;
    }
  }, [searchParams]);

  return lang;
}
