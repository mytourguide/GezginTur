// Cloudflare Pages (next-on-pages)
export const runtime = "edge";
// iyzico callback'i kullaniciyi buraya yonlendirir (?status=success|failed)
import Link from "next/link";

export default function PaymentResultPage({ searchParams }: { searchParams: { status?: string; reason?: string } }) {
  const ok = searchParams.status === "success";
  return (
    <div className="max-w-xl mx-auto px-4 py-20 text-center">
      <div className={`text-6xl mb-4 ${ok ? "text-green-500" : "text-red-500"}`}>{ok ? "✓" : "✕"}</div>
      <h1 className="text-2xl font-bold">{ok ? "Rezervasyonunuz Onaylandi!" : "Odeme Tamamlanamadi"}</h1>
      <p className="text-gray-600 mt-2">
        {ok
          ? "Rezervasyon detaylari ve onay e-postasi gonderildi. Iyi yolculuklar!"
          : (searchParams.reason ? `Sebep: ${searchParams.reason}. ` : "") + "Lutfen tekrar deneyin; kartinizdan tahsilat yapilmamistir."}
      </p>
      <div className="flex gap-3 justify-center mt-6">
        <Link href="/hesabim" className="bg-brand-600 text-white px-6 py-2 rounded-lg">Rezervasyonlarim</Link>
        {!ok && <Link href="/turlar" className="border px-6 py-2 rounded-lg">Turlara Don</Link>}
      </div>
    </div>
  );
}
