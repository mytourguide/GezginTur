"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { User } from "@/lib/types";

export default function AdminCustomersPage() {
  const [list, setList] = useState<User[]>([]);
  const [search, setSearch] = useState("");

  const load = () => api.get(`/admin/users?search=${encodeURIComponent(search)}`, true).then(setList);
  useEffect(() => { load(); }, []); // eslint-disable-line

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Musteriler</h1>
      <input value={search} onChange={(e) => setSearch(e.target.value)} onKeyDown={(e) => e.key === "Enter" && load()}
        placeholder="Ad veya e-posta ara..." className="border rounded-lg px-3 py-2 mb-4 w-full sm:w-80" />
      <div className="bg-white border rounded-xl overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left">
            <tr><th className="p-3">Ad Soyad</th><th className="p-3">E-posta</th><th className="p-3">Telefon</th><th className="p-3">Rol</th><th className="p-3">Kayit</th></tr>
          </thead>
          <tbody className="divide-y">
            {list.map((u) => (
              <tr key={u.id}>
                <td className="p-3 font-medium">{u.full_name}</td>
                <td className="p-3">{u.email}</td>
                <td className="p-3">{u.phone || "-"}</td>
                <td className="p-3">{u.role}</td>
                <td className="p-3">{u.created_at ? new Date(u.created_at).toLocaleDateString("tr-TR") : "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
