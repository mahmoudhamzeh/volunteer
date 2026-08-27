"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, Assignment, Certificate, openAuth } from "@/lib/api";
import { Badge, Button, Card, Field, Modal, StarRating, AttachmentButton, inputClass } from "@/components/ui";
import { STATUS_LABEL, fmtDate, sortAssignmentsOpenFirst, weekdayLabel, workModeLabel } from "@/lib/labels";
import { AttendancePanel } from "@/components/attendance-panel";
import { TrainingBadge } from "@/components/training-notice";
import { DeliveryHistory } from "@/components/delivery-history";

const FILTERS: { id: string; label: string; match: (s: string) => boolean }[] = [
  { id: "action", label: "نیاز به اقدام", match: (s: string) => ["requested", "training_pending", "reserved", "in_progress", "attended", "submitted", "revision_requested"].includes(s) },
  { id: "submitted", label: "نتیجه ارسال‌شده", match: (s) => s === "submitted" || s === "revision_requested" },
  { id: "completed", label: "تکمیل‌شده", match: (s) => s === "completed" },
  { id: "all", label: "همه", match: () => true },
];

export default function AssignmentsAdmin() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [scores, setScores] = useState<Record<string, { d: number; e: number; t: number; c: string }>>({});
  const [filter, setFilter] = useState("action");
  const [q, setQ] = useState("");
  const [msg, setMsg] = useState("");
  const [openId, setOpenId] = useState("");
  const [volQ, setVolQ] = useState("");
  const [issued, setIssued] = useState<Record<string, string>>({});
  const [revNotes, setRevNotes] = useState<Record<string, string>>({});

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
      const key = a.task?.kind === "occurrence" && a.task.series_id ? a.task.series_id : a.task_id;
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
    return Array.from(map.entries()).map(([id, g]) => ({ id, ...g }));
  }, [filtered]);

  const activeItems = useMemo(() => {
    if (!openId) return [] as Assignment[];
    return (items || []).filter((a) => a.task_id === openId || a.task?.series_id === openId);
  }, [items, openId]);

  const activeMeta = activeItems[0];
  const activeTitle = activeMeta?.task?.title || groups.find((g) => g.id === openId)?.title || "داوطلبان فعالیت";

  const visibleVolunteers = useMemo(() => {
    const needle = volQ.trim();
    const list = sortAssignmentsOpenFirst(activeItems);
    if (!needle) return list;
    return list.filter((a) => `${a.volunteer?.full_name || ""} ${a.volunteer?.phone || ""}`.includes(needle));
  }, [activeItems, volQ]);

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
        <p className="mt-1 text-sm text-stone-500">هر ردیف یک فعالیت است. با کلیک، فهرست داوطلبان همان فعالیت در پاپ‌آپ باز می‌شود.</p>
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
        <span className="text-xs text-stone-400">{groups.length} فعالیت</span>
      </div>
      {groups.length === 0 && <Card className="p-6 text-stone-500">موردی با این فیلتر نیست.</Card>}
      <div className="grid gap-3">
        {groups.map((g) => {
          const action = g.items.filter((a) => ["requested", "training_pending", "reserved", "in_progress", "attended", "submitted", "revision_requested"].includes(a.status)).length;
          return (
            <button
              key={g.id}
              type="button"
              onClick={() => { setOpenId(g.id); setVolQ(""); }}
              className="w-full rounded-3xl border border-white/70 bg-white/90 p-4 text-right shadow-card hover:border-mahak-200"
            >
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div>
                  <div className="text-xs text-mahak-700">فعالیت</div>
                  <h2 className="text-lg font-black">{g.title}</h2>
                  <p className="mt-1 text-sm text-stone-600">
                    {workModeLabel(g.mode)} · {g.location || (g.mode === "remote" ? "دورکار" : "—")} · {fmtDate(g.starts)}
                    {g.ends ? ` تا ${fmtDate(g.ends)}` : ""}
                  </p>
                </div>
                <div className="flex flex-col items-end gap-1 text-xs">
                  <span className="rounded-full bg-stone-100 px-3 py-1">{g.items.length} داوطلب</span>
                  {action > 0 && <span className="rounded-full bg-amber-50 px-3 py-1 text-amber-800">{action} نیازمند اقدام</span>}
                </div>
              </div>
            </button>
          );
        })}
      </div>

      <Modal open={!!openId} size="lg" title={activeTitle} onClose={() => setOpenId("")}>
        {openId && (
          <div className="space-y-3">
            <p className="text-sm text-stone-500">
              {workModeLabel(activeMeta?.task?.work_mode)} · {activeMeta?.task?.location || "—"} · {fmtDate(activeMeta?.task?.starts_at)} · معادل {activeMeta?.task?.hour_weight || 0} ساعت
            </p>
            <TrainingBadge task={activeMeta?.task} />
            <input className={inputClass} placeholder="جستجو نام یا موبایل داوطلب" value={volQ} onChange={(e) => setVolQ(e.target.value)} />
            {visibleVolunteers.length === 0 && <p className="text-sm text-stone-400">داوطلبی با این جستجو نیست.</p>}
            {visibleVolunteers.map((a) => (
              <div key={a.id} className="space-y-3 rounded-2xl border border-stone-100 p-3">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <Link className="font-bold text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                      {a.volunteer?.full_name || "داوطلب"}
                    </Link>
                    <div className="text-xs text-stone-500">
                      {a.volunteer?.phone ? `${a.volunteer.phone} · ` : ""}
                      {a.task?.kind === "occurrence" ? `${weekdayLabel(a.task.weekday)} · ${fmtDate(a.task.starts_at)} · ` : ""}
                      ثبت {fmtDate(a.created_at)}
                    </div>
                  </div>
                  <Badge status={a.status} reason={a.admin_comment} />
                </div>
                <div className="rounded-2xl bg-stone-50 px-3 py-2 text-sm">
                  {a.status === "requested" && <p>درخواست داده؛ هنوز توسط واحد پشتیبانی تایید نشده است.</p>}
                  {a.status === "training_pending" && <p>درخواست تایید شده؛ تا تایید آموزش در بخش آموزش، امکان ادامه فرایند فعالیت نیست.</p>}
                  {a.status === "reserved" && a.task?.work_mode === "remote" && <p>تایید شده؛ داوطلب باید از پنل کارها فعالیت را شروع و نتیجه را بارگذاری کند.</p>}
                  {a.status === "reserved" && a.task?.work_mode !== "remote" && <p>تایید شده؛ واحد پشتیبانی حضور یا عدم حضور را ثبت می‌کند. داوطلب نیازی به شروع ندارد.</p>}
                  {a.status === "in_progress" && <p>داوطلب کار دورکار را شروع کرده است.</p>}
                  {a.status === "submitted" && <p>نتیجه دورکار ارسال شده و آماده بررسی است. می‌توانید تکمیل کنید یا درخواست اصلاح بفرستید.</p>}
                  {a.status === "revision_requested" && <p>درخواست اصلاح برای داوطلب ارسال شد{a.admin_comment ? ` — ${a.admin_comment}` : "."}</p>}
                  {a.status === "attended" && (
                    <p>
                      حضور تایید شد
                      {a.check_in_at ? ` · ورود ${fmtDate(a.check_in_at)}` : a.attended_at ? ` در ${fmtDate(a.attended_at)}` : ""}
                      {a.check_out_at ? ` · خروج ${fmtDate(a.check_out_at)}` : ""}
                    </p>
                  )}
                  {a.status === "absent" && <p>عدم حضور ثبت شد.</p>}
                  {a.volunteer_comment && <p>نظر داوطلب: {a.volunteer_comment}</p>}
                  <DeliveryHistory
                    items={a.history}
                    assignmentId={a.id}
                    fileHref={(aid, fid) => `/api/v1/admin/assignments/${aid}/files/${fid}`}
                  />
                  {!a.history?.length && (a.delivery_note || a.delivery_file_name) && (
                    <div className="space-y-1">
                      {a.delivery_note && <p>شرح نتیجه: {a.delivery_note}</p>}
                      {a.delivery_file_name && (
                        <AttachmentButton
                          name={a.delivery_file_name}
                          label="دانلود پیوست نتیجه"
                          onOpen={() => void openAuth(`/api/v1/admin/assignments/${a.id}/delivery`)}
                        />
                      )}
                    </div>
                  )}
                  {a.status === "completed" && (
                    <div className="mt-2 grid gap-2 sm:grid-cols-3">
                      <StarRating label="انضباط" value={a.admin_discipline || 0} readOnly size="sm" />
                      <StarRating label="تخصص" value={a.admin_expertise || 0} readOnly size="sm" />
                      <StarRating label="اخلاق" value={a.admin_ethics || 0} readOnly size="sm" />
                    </div>
                  )}
                  {(a.status === "cancelled" || a.status === "rejected" || a.status === "absent") && (
                    <p>{STATUS_LABEL[a.status]}{a.status === "rejected" && a.admin_comment ? ` — ${a.admin_comment}` : ""}</p>
                  )}
                </div>
                <div className="flex flex-wrap gap-2">
                  {a.status === "requested" && (
                    <Button onClick={() => run(() => api.approveAssignment(a.id), "تایید شد")}>تایید درخواست</Button>
                  )}
                  {a.status === "training_pending" && (
                    <Link className="rounded-2xl bg-mahak-500 px-4 py-2 text-sm font-bold text-white" href="/admin/trainings">
                      تایید در بخش آموزش
                    </Link>
                  )}
                  {(a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "attended") && a.task?.work_mode !== "remote" && (
                    <Button variant="danger" onClick={() => run(() => api.markAbsent(a.id), "عدم حضور ثبت شد")}>عدم حضور</Button>
                  )}
                  {(a.status === "requested" || a.status === "training_pending" || a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "revision_requested") && (
                    <Button variant="danger" onClick={() => run(() => api.rejectAssignment(a.id), "فعالیت رد شد")}>رد کل فعالیت</Button>
                  )}
                </div>
                {(a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "attended") && a.task?.work_mode !== "remote" && (
                  <AttendancePanel assignment={a} onDone={async (ok) => { setMsg(ok); await load(); }} />
                )}
                {a.status === "submitted" && a.task?.work_mode === "remote" && (
                  <div className="space-y-2 rounded-2xl border border-amber-100 bg-amber-50/70 p-3">
                    <Field label="درخواست اصلاح یا تکمیل (برای داوطلب ارسال می‌شود)">
                      <textarea className={inputClass} rows={2} value={revNotes[a.id] || ""} onChange={(e) => setRevNotes({ ...revNotes, [a.id]: e.target.value })} placeholder="مثلاً فایل نهایی را هم بارگذاری کنید" />
                    </Field>
                    <Button variant="outline" onClick={() => {
                      const comment = (revNotes[a.id] || "").trim();
                      if (!comment) {
                        setMsg("برای درخواست اصلاح، توضیح را بنویسید");
                        return;
                      }
                      return run(() => api.requestRevision(a.id, comment), "درخواست اصلاح برای داوطلب ارسال شد");
                    }}>ارسال درخواست اصلاح / تکمیل</Button>
                  </div>
                )}
                {(a.status === "submitted" && a.task?.work_mode === "remote") || (a.status === "attended" && a.task?.work_mode !== "remote") ? (
                  <div className="space-y-3 rounded-2xl border border-stone-100 p-3">
                    <div className="grid gap-3 sm:grid-cols-3">
                      <StarRating label="انضباط" value={sc(a.id).d} onChange={(d) => setScores({ ...scores, [a.id]: { ...sc(a.id), d } })} />
                      <StarRating label="تخصص" value={sc(a.id).e} onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), e } })} />
                      <StarRating label="اخلاق" value={sc(a.id).t} onChange={(t) => setScores({ ...scores, [a.id]: { ...sc(a.id), t } })} />
                    </div>
                    <Field label="نظر پشتیبانی (اختیاری)">
                      <input className={inputClass} value={sc(a.id).c} onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), c: e.target.value } })} />
                    </Field>
                    <Button onClick={() => {
                      const s = sc(a.id);
                      return run(() => api.complete(a.id, { discipline: s.d, expertise: s.e, ethics: s.t, comment: s.c }), "امتیاز ثبت و تکمیل شد");
                    }}>ثبت امتیاز و تکمیل</Button>
                  </div>
                ) : null}
                {a.status === "completed" && (
                  <div className="flex flex-wrap items-center gap-2">
                    <Button variant="outline" onClick={() => run(async () => {
                      const c = await api.issueCert(a.id) as Certificate;
                      if (c?.verification_code) {
                        setIssued((prev) => ({ ...prev, [a.id]: c.verification_code }));
                        window.open(`/api/v1/certificates/${c.verification_code}/pdf`, "_blank");
                      }
                    }, "تقدیرنامه صادر شد")}>صدور تقدیرنامه این فعالیت</Button>
                    {(issued[a.id]) && (
                      <a className="text-sm text-mahak-700" href={`/api/v1/certificates/${issued[a.id]}/pdf`} target="_blank">دانلود PDF</a>
                    )}
                  </div>
                )}
              </div>
            ))}
            <div className="flex justify-end">
              <Button variant="ghost" onClick={() => setOpenId("")}>بستن</Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
