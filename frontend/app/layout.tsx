import type { Metadata } from "next";
import "./globals.css";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

// Site geneli SEO metadata (App Router SSR/SSG ile otomatik optimize edilir)
export const metadata: Metadata = {
  title: { default: "Gezgin Tur — Hayalinizdeki Turu Kesfedin", template: "%s | Gezgin Tur" },
  description: "Yurt ici ve yurt disi turlar, kultur ve doga rotalari. Guvenli online odeme ile rezervasyon yapin.",
  openGraph: { siteName: "Gezgin Tur", locale: "tr_TR", type: "website" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="tr">
      <body className="min-h-screen flex flex-col">
        <Navbar />
        <main className="flex-1">{children}</main>
        <Footer />
      </body>
    </html>
  );
}
