"use client";

import { useEffect, useState } from "react";

export function useLang(): "tr" | "en" {
  const [lang, setLang] = useState<"tr" | "en">("tr");
  useEffect(() => {
    const fromQuery = /(?:^|[?&])lang=(en|tr)/.exec(window.location.search)?.[1] as "en"|"tr"|undefined;
    const fromCookie = /(?:^|;\s*)lang=(en|tr)/.exec(document.cookie)?.[1] as "en"|"tr"|undefined;
    const l = (fromQuery || fromCookie || "tr") as "tr"|"en";
    setLang(l);
    if (fromQuery) {
      document.cookie = `lang=${fromQuery};path=/;max-age=31536000`;
    }
  }, []);
  return lang;
}
