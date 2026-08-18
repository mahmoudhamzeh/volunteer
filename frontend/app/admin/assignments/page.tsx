"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, Assignment, openAuth } from "@/lib/api";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";
import { STATUS_LABEL, fmtDate, workModeLabel } from "@/lib/labels";

const FILTERS: { id: string; label: string; match: (s: string) => boolean }[] = [
  { id: "action", label: "نیاز به اقدام", match: (s: string) => ["requested", "reserved", "in_progress", "attended", "submitted"].includes(s) },
  { id: "submitted", label: "نتیجه ارسال‌شده", match: (s) => s === "submitted" },
  { id: "completed", label: "تکمیل‌شده", match: (s) => s === "completed" },
  { id: "all", label: "همه", match: () => true },
];

export default function AssignmentsAdmin() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [scores, setScores] = useState<Record<string, { d: number; e: number; t: number; c: string }>>({});
  const [filter, setFilter] = useState("action");
  const [q, setQ] = useState("");
  const [msg, setMsg] = useState("");

  async function load() {
    const r = await api.adminAssignments("?limit=200");
    setItems(r.items || []);
  }
  useEffect(() => { void load(); }, []);

  function sc(id: string) {
    return scores[id] || { d: 5, e: 5, t: 5, c: "" };
  }

  const filtered = useMemo(() => {
    const f = FILTERS.find((x) => x.id === filter) || FILTERS[0];
    const needle = q.trim();
    return (items || []).filter((a) => {
      if (!f.match(a.status)) return false;
      if (!needle) return true;
      const hay = `${a.task?.title || ""} ${a.volunteer?.full_name || ""} ${a.task?.location || ""} ${a.volunteer?.phone || ""}`;
      return hay.includes(needle);
    });
  }, [items, filter, q]);

  const groups = useMemo(() => {
    const map = new Map<string, { title: string; location: string; starts: string; ends?: string; mode?: string; hint?: string; hours: number; items: Assignment[] }>();
    for (const a of filtered) {
      const key = a.task_id;
      const cur = map.get(key);
      if (cur) {
        cur.items.push(a);
        continue;
      }
      map.set(key, {
        title: a.task?.title || "فعالیت بدون عنوان",
        location: a.task?.location || "",
        starts: a.task?.starts_at || "",
        ends: a.task?.ends_at,
        mode: a.task?.work_mode,
        hint: a.task?.delivery_hint,
        hours: a.task?.hour_weight || a.hours_awarded || 0,
        items: [a],
      });
    }
    return Array.from(map.values());
  }, [filtered]);

  async function run(fn: () => Promise<unknown>, ok: string) {
    try {
      await fn();
      setMsg(ok);
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">حضور، نتیجه و امتیاز</h1>
        <p className="mt-1 text-sm text-stone-500">ردیف‌ها بر اساس فعالیت گروه‌بندی شده‌اند تا در تعداد بالا مشخص باشد کار مربوط به کدام فعالیت و چه نتیجه‌ای داشته است.</p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      <div className="flex flex-wrap items-center gap-2">
        {FILTERS.map((f) => (
          <button
            key={f.id}
            onClick={() => setFilter(f.id)}
            className={`rounded-full px-3 py-1 text-sm ${filter === f.id ? "bg-mahak-500 text-white" : "bg-white"}`}
          >
            {f.label}
          </button>
        ))}
        <input className={inputClass + " max-w-xs"} placeholder="جستجو فعالیت یا داوطلب" value={q} onChange={(e) => setQ(e.target.value)} />
        <span className="text-xs text-stone-400">{filtered.length} مورد در {groups.length} فعالیت</span>
      </div>
      {groups.length === 0 && <Card className="p-6 text-stone-500">موردی با این فیلتر نیست.</Card>}
      {groups.map((g) => (
        <Card key={g.items[0]?.task_id || g.title} className="overflow-hidden">
          <div className="border-b border-mahak-50 bg-mahak-50/40 px-5 py-4">
            <div className="flex flex-wrap items-start justify-between gap-2">
              <div>
                <div className="text-xs font-medium text-mahak-700">فعالیت</div>
                <h2 className="text-lg font-black text-ink-900">{g.title}</h2>
                <p className="mt-1 text-sm text-stone-600">
                  {workModeLabel(g.mode)} · {g.location || (g.mode === "remote" ? "دورکار" : "—")} · {fmtDate(g.starts)}
                  {g.ends ? ` تا ${fmtDate(g.ends)}` : ""} · معادل {g.hours} ساعت
                </p>
                {g.mode === "remote" && g.hint && <p className="mt-1 text-xs text-mahak-700">تحویل مورد انتظار: {g.hint}</p>}
              </div>
              <span className="rounded-full bg-white px-3 py-1 text-xs text-stone-600">{g.items.length} داوطلب</span>
            </div>
          </div>
          <div className="divide-y divide-stone-100">
            {g.items.map((a) => (
              <div key={a.id} className="space-y-3 px-5 py-4">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <Link className="font-bold text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                      {a.volunteer?.full_name || "داوطلب"}
                    </Link>
                    <div className="text-xs text-stone-500">
                      {a.volunteer?.phone ? `${a.volunteer.phone} · ` : ""}
                      ثبت {fmtDate(a.created_at)}
                    </div>
                  </div>
                  <Badge status={a.status} />
                </div>

                <div className="rounded-2xl bg-stone-50 px-3 py-2 text-sm">
                  <div className="text-xs text-stone-500">نتیجه / وضعیت کار</div>
                  {a.status === "requested" && <p>درخواست داده؛ هنوز رزرو نشده است.</p>}
                  {a.status === "reserved" && (
                    <p>تایید شده؛ داوطلب باید از پنل «کارهای من» فعالیت را شروع کند.</p>
                  )}
                  {a.status === "in_progress" && (
                    <p>داوطلب کار را شروع کرده و هنوز نتیجه نهایی ارسال نشده یا در حال انجام است.</p>
                  )}
                  {a.status === "attended" && <p>حضور تایید شد{a.attended_at ? ` در ${fmtDate(a.attended_at)}` : ""}.</p>}
                  {(a.status === "submitted" || a.status === "completed" || a.status === "in_progress") && (a.delivery_note || a.delivery_file_name) && (
                    <div className="space-y-1">
                      {a.delivery_note && <p><span className="text-stone-500">شرح نتیجه: </span>{a.delivery_note}</p>}
                      {a.delivery_file_name && (
                        <button className="text-mahak-700" onClick={() => openAuth(`/api/v1/admin/assignments/${a.id}/delivery`)}>
                          فایل نتیجه: {a.delivery_file_name}
                        </button>
                      )}
                      {a.delivered_at && <p className="text-xs text-stone-400">ارسال {fmtDate(a.delivered_at)}</p>}
                    </div>
                  )}
                  {a.status === "submitted" && !a.delivery_note && !a.delivery_file_name && <p>نتیجه ارسال شده؛ فایل یا شرح ثبت نشده است.</p>}
                  {a.status === "completed" && (
                    <div className="mt-2 grid gap-1 sm:grid-cols-3">
                      <p>امتیاز نهایی: <b>{a.composite_score != null ? a.composite_score.toFixed(1) : "—"}</b></p>
                      <p>ساعات: <b>{a.hours_awarded || "—"}</b></p>
                      <p>انضباط / تخصص / اخلاق: {[a.admin_discipline, a.admin_expertise, a.admin_ethics].map((n) => n ?? "—").join(" / ")}</p>
                      {a.admin_comment && <p className="sm:col-span-3">نظر ادمین: {a.admin_comment}</p>}
                      {a.completed_at && <p className="text-xs text-stone-400 sm:col-span-3">تکمیل {fmtDate(a.completed_at)}</p>}
                    </div>
                  )}
                  {(a.status === "cancelled" || a.status === "rejected") && <p>{STATUS_LABEL[a.status]}</p>}
                </div>

                <div className="flex flex-wrap gap-2">
                  {a.status === "requested" && (
                    <Button onClick={() => run(() => api.approveAssignment(a.id), "تایید و رزرو شد")}>تایید درخواست</Button>
                  )}
                  {(a.status === "reserved" || a.status === "in_progress") && a.task?.work_mode !== "remote" && (
                    <Button onClick={() => run(() => api.attendance(a.id), "حضور تایید شد")}>تایید حضور</Button>
                  )}
                  {(a.status === "requested" || a.status === "reserved" || a.status === "in_progress" || a.status === "submitted") && (
                    <Button variant="danger" onClick={() => run(() => api.rejectAssignment(a.id), "رد شد")}>رد / لغو</Button>
                  )}
                </div>

                {(a.status === "submitted" || a.status === "attended" || (a.task?.work_mode !== "remote" && (a.status === "in_progress" || a.status === "reserved"))) && (
                  <div className="grid gap-2 rounded-2xl border border-stone-100 p-3 md:grid-cols-4">
                    <Field label="انضباط (۱ تا ۵)">
                      <input className={inputClass} type="number" min={1} max={5} value={sc(a.id).d}
                        onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), d: Number(e.target.value) } })} />
                    </Field>
                    <Field label="تخصص (۱ تا ۵)">
                      <input className={inputClass} type="number" min={1} max={5} value={sc(a.id).e}
                        onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), e: Number(e.target.value) } })} />
                    </Field>
                    <Field label="اخلاق (۱ تا ۵)">
                      <input className={inputClass} type="number" min={1} max={5} value={sc(a.id).t}
                        onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), t: Number(e.target.value) } })} />
                    </Field>
                    <div className="md:col-span-4">
                      <Field label="نظر ادمین (اختیاری)">
                        <input className={inputClass} value={sc(a.id).c}
                          onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), c: e.target.value } })} />
                      </Field>
                    </div>
                    <Button onClick={() => {
                      const s = sc(a.id);
                      return run(() => api.complete(a.id, { discipline: s.d, expertise: s.e, ethics: s.t, comment: s.c }), "امتیاز ثبت و تکمیل شد");
                    }}>ثبت امتیاز و تکمیل این فعالیت</Button>
                  </div>
                )}

                {a.status === "completed" && (
                  <Button variant="outline" onClick={() => run(() => api.issueCert(a.id), "گواهی موردی صادر شد")}>صدور گواهی این فعالیت</Button>
                )}
              </div>
            ))}
          </div>
        </Card>
      ))}
    </div>
  );
}
