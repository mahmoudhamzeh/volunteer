"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, CertificateRequest } from "@/lib/api";
import { CERT_REQ_LABEL, certRequestTitle, deliveryMethodLabel, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminCertificatesPage() {
  const [items, setItems] = useState<CertificateRequest[]>([]);
  const [filter, setFilter] = useState("pending");
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");

  async function load(status = filter) {
    if (status === "action") {
      const [pending, preparing, ready] = await Promise.all([
        api.adminCertRequests("pending"),
        api.adminCertRequests("preparing"),
        api.adminCertRequests("ready"),
      ]);
      setItems([...(pending || []), ...(preparing || []), ...(ready || [])]);
      return;
    }
    const list = await api.adminCertRequests(status);
    setItems(list || []);
  }

  useEffect(() => {
    load("action").catch((e) => setErr(e instanceof Error ? e.message : "خطا"));
    setFilter("action");
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
        <h1 className="text-2xl font-black">تقدیرنامه و گواهی‌نامه</h1>
        <p className="mt-1 text-sm text-stone-500">
          تقدیرنامه با تایید صادر می‌شود. گواهی‌نامه فعالیت داوطلبانه پس از آماده‌سازی صادر و سپس ارسال یا حضوری تحویل می‌گردد.
        </p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}
      <div className="flex flex-wrap items-center gap-2">
        {[
          ["action", "نیاز به اقدام"],
          ["pending", "تقدیرنامه در انتظار"],
          ["preparing", "در حال آماده‌سازی"],
          ["ready", "آماده تحویل"],
          ["delivered", "تحویل‌شده"],
          ["approved", "تقدیرنامه صادرشده"],
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
                <th className="px-4 py-3 font-medium">نوع</th>
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
                    <div>{certRequestTitle(r)}</div>
                    <div className="text-xs text-stone-400">{r.kind === "official" ? "گواهی‌نامه" : "تقدیرنامه"}</div>
                  </td>
                  <td className="px-4 py-3 text-stone-500">{fmtDate(r.created_at)}</td>
                  <td className="px-4 py-3">
                    <Badge status={r.status} reason={r.admin_note} label={CERT_REQ_LABEL[r.status]} />
                    {r.delivery_method && <div className="mt-1 text-xs text-stone-500">{deliveryMethodLabel(r.delivery_method)}</div>}
                  </td>
                  <td className="px-4 py-3">
                    {r.status === "pending" || r.status === "preparing" ? (
                      <div className="space-y-2">
                        <Field label="یادداشت / دلیل رد">
                          <input
                            className={inputClass}
                            value={notes[r.id] || ""}
                            onChange={(e) => setNotes({ ...notes, [r.id]: e.target.value })}
                          />
                        </Field>
                        <div className="flex flex-wrap gap-2">
                          <Button onClick={() => run(() => api.reviewCertRequest(r.id, "approve", notes[r.id] || ""), r.kind === "official" ? "آماده تحویل شد" : "تقدیرنامه صادر شد")}>
                            {r.kind === "official" ? "بررسی و صدور" : "تایید و صدور"}
                          </Button>
                          <Button variant="danger" onClick={() => run(() => api.reviewCertRequest(r.id, "reject", notes[r.id] || ""), "رد شد")}>رد</Button>
                        </div>
                      </div>
                    ) : r.status === "ready" ? (
                      <div className="flex flex-wrap gap-2">
                        <Button onClick={() => run(() => api.reviewCertRequest(r.id, "deliver", notes[r.id] || "", "send"), "ارسال ثبت شد")}>ارسال برای داوطلب</Button>
                        <Button variant="outline" onClick={() => run(() => api.reviewCertRequest(r.id, "deliver", notes[r.id] || "", "in_person"), "تحویل حضوری ثبت شد")}>تحویل حضوری</Button>
                      </div>
                    ) : (
                      <p className="text-xs text-stone-500">{r.admin_note || deliveryMethodLabel(r.delivery_method) || "—"}</p>
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
