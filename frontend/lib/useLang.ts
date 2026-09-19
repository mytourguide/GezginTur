"use client";

import { useEffect, useState } from "react";

// Istemci bilesenlerinde mevcut dili (lang cookie) dondurur
export function useLang(): "tr" | "en" {
  const [lang, setLang] = useState<"tr" | "en">("tr");
  useEffect(() => {
    if (/(?:^|;\s*)lang=en/.test(document.cookie)) setLang("en");
  }, []);
  return lang;
}
