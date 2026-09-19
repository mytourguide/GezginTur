"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getSession, logout } from "@/lib/auth";
import type { User } from "@/lib/types";
import LanguageSwitcher from "@/components/LanguageSwitcher";
import { useLang } from "@/lib/useLang";

const NAV = {
  tr: { tours: "Turlar", account: "Hesabim", admin: "Yonetim", logout: "Cikis", login: "Giris Yap", register: "Kayit Ol" },
  en: { tours: "Tours", account: "My Account", admin: "Admin", logout: "Log Out", login: "Sign In", register: "Sign Up" },
};

export default function Navbar() {
  const [user, setUser] = useState<User | null>(null);
  const lang = useLang();
  const t = NAV[lang];
  useEffect(() => { setUser(getSession().user); }, [lang]);

  return (
    <header className="sticky top-0 z-40 bg-white/90 backdrop-blur border-b">
      <nav className="max-w-6xl mx-auto flex items-center justify-between px-4 h-16">
        <Link href="/" className="text-xl font-bold text-brand-700">Gezgin Tur</Link>
        <div className="flex items-center gap-5 text-sm font-medium">
          <Link href="/turlar" className="hover:text-brand-600">{t.tours}</Link>
          <LanguageSwitcher />
          {user ? (
            <>
              <Link href="/hesabim" className="hover:text-brand-600">{t.account}</Link>
              {user.role === "admin" && (
                <Link href="/admin" className="hover:text-brand-600">{t.admin}</Link>
              )}
              <button onClick={logout} className="text-red-600 hover:underline">{t.logout}</button>
            </>
          ) : (
            <>
              <Link href="/giris" className="hover:text-brand-600">{t.login}</Link>
              <Link href="/kayit" className="bg-brand-600 text-white px-4 py-2 rounded-lg hover:bg-brand-700">
                {t.register}
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
}
