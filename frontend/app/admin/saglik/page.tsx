"use client";

import { useState } from "react";
import { api } from "@/lib/api";

interface Check { group: string; name: string; status: string; detail: string; ms: number }

const GROUPS: Record<string, string> = {
  altyapi: "Altyapi",
  guvenlik: "Guvenlik",
  icerik: "icerik",
  odeme: "Odeme",
  saklama: "Saklama",
};
const GROUP_ORDER = ["altyapi", "guvenlik", "icerik", "odeme", "saklama"];
const BADGE: Record<string, string> = {
  ok: "bg-green-100 text-green-700",
  warn: "bg-amber-100 text-amber-700",
  fail: "bg-red-100 text-red-700",
};
const LABEL: Record<string, string> = { ok: "OK", warn: "UYARI", fail: "HATA" };

export default function HealthPage() {
  const [result, setResult] = useState<{ generated_at: string; checks: Check[] } | null>(null);
  const [loading, setLoading] = useState<string | null>(null); //  calisan grup
  const [err, setErr] = useState("");
  const [copied, setCopied] = useState(false);

  async function run(group?: string) {
    setErr(""); setCopied(false); setLoading(group ?? "all");
    try {
      const data = await api.get("/admin/health" + (group ? `?group=${group}` : ""), true);
      setResult((prev) =>
        group && prev
          ? { generated_at: data.generated_at, checks: [...prev.checks.filter((c) => c.group !== group), ...data.checks] }
          : data
      );
    } catch (e: any) {
      setErr(e.message);
    } finally { setLoading(null); }
  }

  // AI'a yapistirilmak uzere okunakli rapor metni
  function reportText(): string {
    if (!result) return "";
    const lines: string[] = [
      `GEZGIN TUR - SISTEM SAGLIK RAPORU (${result.generated_at})`,
      `Ozet: ${result.checks.filter((c) => c.status === "ok").length} OK, ` +
        `${result.checks.filter((c) => c.status === "warn").length} UYARI, ` +
        `${result.checks.filter((c) => c.status === "fail").length} HATA`,
      "",
    ];
    for (const g of GROUP_ORDER) {
      const items = result.checks.filter((c) => c.group === g);
      if (!items.length) continue;
      lines.push(`[${(GROUPS[g] || g).toUpperCase()}]`);
      for (const c of items) lines.push(`  - ${c.name}: ${LABEL[c.status]} — ${c.detail} (${c.ms}ms)`);
      lines.push("");
    }
    return lines.join("\n");
  }

  function copyReport() {
    navigator.clipboard.writeText(reportText());
    setCopied(true); setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="bg-white border rounded-2xl p-6">
      <div className="flex flex-wrap gap-3 items-center justify-between">
        <div>
          <h1 className="text-xl font-bold">Sistem Sagligi</h1>
          <p className="text-sm text-gray-500">Tum modulleri tek tikla kontrol edin; raporu kopyalayip AI'ya yapistirabilirsiniz.</p>
        </div>
        <button onClick={() => run()} disabled={loading !== null}
          className="bg-brand-600 text-white px-5 py-2 rounded-lg hover:bg-brand-700 disabled:opacity-50 font-semibold">
          {loading === "all" ? "Calisiyor..." : "Tumunu Calistir"}
        </button>
      </div>

      {err && <p className="mt-3 text-red-600 text-sm">{err}</p>}

      {result && (
        <>
          <div className="mt-4 space-y-6">
            {GROUP_ORDER.map((g) => {
              const items = result.checks.filter((c) => c.group === g);
              if (!items.length) return null;
              return (
                <div key={g}>
                  <div className="flex items-center justify-between mb-2">
                    <h2 className="font-bold">{GROUPS[g]}</h2>
                    <button onClick={() => run(g)} disabled={loading !== null}
                      className="text-xs text-brand-600 hover:underline disabled:opacity-50">
                      {loading === g ? "..." : "Yeniden calistir"}
                    </button>
                  </div>
                  <table className="w-full text-sm">
                    <tbody>
                      {items.map((c) => (
                        <tr key={c.group + c.name} className="border-t">
                          <td className="py-2 pr-3">{c.name}</td>
                          <td className="w-20">
                            <span className={`text-xs px-2 py-1 rounded-full font-semibold ${BADGE[c.status]}`}>{LABEL[c.status]}</span>
                          </td>
                          <td className="py-2 px-3 text-gray-600">{c.detail}</td>
                          <td className="text-xs text-gray-400 w-16 text-right">{c.ms} ms</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              );
            })}
          </div>

          <div className="mt-6 border-t pt-4">
            <div className="flex items-center justify-between mb-2">
              <h2 className="font-bold">Kopyalanabilir Rapor</h2>
              <button onClick={copyReport}
                className="text-sm bg-gray-800 text-white px-4 py-1.5 rounded-lg hover:bg-gray-700">
                {copied ? "Kopyalandi ✓" : "Kopyala"}
              </button>
            </div>
            <textarea readOnly rows={12} value={reportText()}
              className="w-full font-mono text-xs border rounded-lg p-3 bg-gray-50" />
          </div>
        </>
      )}
    </div>
  );
}
