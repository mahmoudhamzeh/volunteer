"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, CertificateRequest } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminCertificatesPage() {
  const [items, setItems] = useState<CertificateRequest[]>([]);
  const [filter, setFilter] = useState("pending");
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");

  async function load(status = filter) {
    const list = await api.adminCertRequests(status);
    setItems(list || []);
  }

  useEffect(() => {
    load("pending").catch((e) => setErr(e instanceof Error ? e.message : "خطا"));
  }, []);

  async function run(fn: () => Promise<unknown>, ok: string) {
    setErr("");
    try {
      await fn();
      setMsg(ok);
      await load(filter);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">درخواست‌های گواهی‌نامه</h1>
        <p className="mt-1 text-sm text-stone-500">تایید درخواست، گواهی را صادر می‌کند و به داوطلب اطلاع می‌دهد. رد هم با دلیل به داوطلب اعلام می‌شود.</p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}
      <div className="flex flex-wrap items-center gap-2">
        {[
          ["pending", "در انتظار"],
          ["approved", "تایید شده"],
          ["rejected", "رد شده"],
          ["", "همه"],
        ].map(([id, label]) => (
          <button
            key={id || "all"}
            onClick={() => {
              setFilter(id);
              load(id).catch((e) => setErr(e instanceof Error ? e.message : "خطا"));
            }}
            className={`rounded-full px-3 py-1 text-sm ${filter === id ? "bg-mahak-500 text-white" : "bg-white"}`}
          >
            {label}
          </button>
        ))}
      </div>
      {(items || []).length === 0 && <Card className="p-6 text-stone-500">موردی نیست.</Card>}
      <div className="overflow-x-auto rounded-3xl border border-white/70 bg-white/90 shadow-card">
        {(items || []).length > 0 && (
          <table className="w-full min-w-[720px] text-sm">
            <thead className="bg-mahak-50/60 text-right text-xs text-stone-500">
              <tr>
                <th className="px-4 py-3 font-medium">داوطلب</th>
                <th className="px-4 py-3 font-medium">نوع / فعالیت</th>
                <th className="px-4 py-3 font-medium">تاریخ</th>
                <th className="px-4 py-3 font-medium">وضعیت</th>
                <th className="px-4 py-3 font-medium">اقدام</th>
              </tr>
            </thead>
            <tbody>
              {(items || []).map((r) => (
                <tr key={r.id} className="border-t border-stone-100 align-top">
                  <td className="px-4 py-3">
                    <Link className="font-medium text-mahak-700" href={`/admin/volunteers/${r.volunteer_id}`}>
                      {r.volunteer_name || "داوطلب"}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <div>{r.assignment_title || (r.kind === "aggregated" ? "گواهی تجمیعی" : "گواهی فعالیت")}</div>
                    <div className="text-xs text-stone-400">{r.kind === "aggregated" ? "تجمیعی" : "موردی"}</div>
                  </td>
                  <td className="px-4 py-3 text-stone-500">{fmtDate(r.created_at)}</td>
                  <td className="px-4 py-3"><Badge status={r.status} reason={r.admin_note} /></td>
                  <td className="px-4 py-3">
                    {r.status === "pending" ? (
                      <div className="space-y-2">
                        <Field label="یادداشت / دلیل رد">
                          <input
                            className={inputClass}
                            value={notes[r.id] || ""}
                            onChange={(e) => setNotes({ ...notes, [r.id]: e.target.value })}
                          />
                        </Field>
                        <div className="flex flex-wrap gap-2">
                          <Button onClick={() => run(() => api.reviewCertRequest(r.id, "approve", notes[r.id] || ""), "گواهی صادر شد")}>تایید و صدور</Button>
                          <Button variant="danger" onClick={() => run(() => api.reviewCertRequest(r.id, "reject", notes[r.id] || ""), "رد شد")}>رد</Button>
                        </div>
                      </div>
                    ) : (
                      <p className="text-xs text-stone-500">{r.admin_note || "—"}</p>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
