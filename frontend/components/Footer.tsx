"use client";

import { useLang } from "@/lib/useLang";

const F = {
  tr: {
    about: "TURSAB uyesi seyahat acentasi. 2008'den beri guvenli seyahat.",
    contact: "Iletisim",
    legal: "Yasal",
    legal1: "KVKK Aydinlatma Metni",
    legal2: "Mesafeli Satis Sozlesmesi",
  },
  en: {
    about: "TURSAB member travel agency. Safe travel since 2008.",
    contact: "Contact",
    legal: "Legal",
    legal1: "Privacy Notice (KVKK)",
    legal2: "Distance Sales Agreement",
  },
};

export default function Footer() {
  const t = F[useLang()];
  return (
    <footer className="border-t bg-white mt-16">
      <div className="max-w-6xl mx-auto px-4 py-10 grid gap-6 md:grid-cols-3 text-sm text-gray-600">
        <div>
          <p className="font-bold text-gray-900 mb-2">Gezgin Tur</p>
          <p>{t.about}</p>
        </div>
        <div>
          <p className="font-bold text-gray-900 mb-2">{t.contact}</p>
          <p>info@gezgintur.com</p>
          <p>+90 212 000 00 00</p>
        </div>
        <div>
          <p className="font-bold text-gray-900 mb-2">{t.legal}</p>
          <p>{t.legal1}</p>
          <p>{t.legal2}</p>
        </div>
      </div>
    </footer>
  );
}
