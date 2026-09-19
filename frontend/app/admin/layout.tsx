"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession, isAdmin } from "@/lib/auth";

const MENU = [
  ["Dashboard", "/admin"],
  ["Turlar", "/admin/turlar"],
  ["Rezervasyonlar", "/admin/rezervasyonlar"],
  ["Odemeler", "/admin/odemeler"],
  ["Musteriler", "/admin/musteriler"],
  ["Kategoriler", "/admin/kategoriler"],
  ["Kuponlar", "/admin/kuponlar"],
  ["Sistem Sagligi", "/admin/saglik"],
];

// Sidebar navigasyon + admin kontrolu (API tarafi RequireAdmin middleware gercek yetkilendirmeyi yapar)
export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [ok, setOk] = useState(false);
  const [email, setEmail] = useState("");

  useEffect(() => {
    const u = getSession().user;
    if (!u || !isAdmin()) router.replace("/giris");
    else { setEmail(u.email); setOk(true); }
  }, [router]);

  if (!ok) return null;

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 grid md:grid-cols-[220px_1fr] gap-8">
      <aside className="bg-white border rounded-2xl p-4 h-fit">
        <p className="font-bold text-brand-700 mb-1 px-2">Yonetim Paneli</p>
        <p className="text-xs text-gray-500 mb-3 px-2 break-all">{email}</p>
        <nav className="flex md:flex-col gap-1 overflow-x-auto text-sm">
          {MENU.map(([label, href]) => (
            <Link key={href} href={href} className="px-3 py-2 rounded-lg hover:bg-brand-50 hover:text-brand-700 whitespace-nowrap">{label}</Link>
          ))}
        </nav>
        <Link href="/" className="block mt-4 px-3 py-2 text-sm text-gray-500 border rounded-lg hover:border-brand-300 hover:text-brand-700">
          ← Siteyi Goruntule
        </Link>
      </aside>
      <section>{children}</section>
    </div>
  );
}
