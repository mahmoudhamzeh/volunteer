"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Assignment, VolunteerTraining } from "@/lib/api";
import { fmtDate, trainingKindLabel, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, StarRating, inputClass } from "@/components/ui";
import { TrainingBadge } from "@/components/training-notice";

export default function WorkPage() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [courses, setCourses] = useState<VolunteerTraining[]>([]);
  const [rating, setRating] = useState<Record<string, number>>({});
  const [comments, setComments] = useState<Record<string, string>>({});
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [files, setFiles] = useState<Record<string, File | undefined>>({});
  const [busy, setBusy] = useState("");
  const [msg, setMsg] = useState("");

  async function load() {
    const [asg, train] = await Promise.all([
      api.myAssignments(),
      api.myTrainings().catch(() => [] as VolunteerTraining[]),
    ]);
    setItems(asg || []);
    setCourses(train || []);
  }
  useEffect(() => { void load(); }, []);

  async function run(id: string, fn: () => Promise<unknown>, ok: string) {
    setBusy(id);
    try {
      await fn();
      setMsg(ok);
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    } finally {
      setBusy("");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">کارهای من</h1>
        <p className="mt-1 text-sm text-stone-500">
          پس از تایید واحد پشتیبانی، فعالیت اینجا می‌آید. اگر فعالیت نیاز به آموزش داشته باشد، ابتدا آموزش تایید می‌شود و سپس مرحله حضور برای انجام فعالیت شروع می‌شود.
        </p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {courses.length > 0 && (
        <Card className="p-5">
          <h2 className="font-bold">دوره‌های آموزشی گذرانده‌شده</h2>
          <p className="mt-1 text-xs text-stone-500">برای فعالیت‌هایی با همین آموزش، نیاز به حضور مجدد در کلاس نیست.</p>
          <ul className="mt-3 space-y-2">
            {courses.map((c) => (
              <li key={c.id} className="rounded-2xl bg-emerald-50 px-3 py-2 text-sm text-emerald-950">
                <div className="font-medium">{c.source_task_title || "دوره آموزشی"}</div>
                <div className="text-xs leading-6">
                  نوع: {trainingKindLabel(c.training_kind)} · محل: {c.training_location || "—"}
                  {c.training_at ? ` · ${fmtDate(c.training_at)}` : ""}
                  {c.confirmed_at ? ` · تایید ${fmtDate(c.confirmed_at)}` : ""}
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}
      {items.filter((a) => a.status !== "requested").length === 0 && (
        <Card className="p-6 text-stone-500">
          پس از تایید واحد پشتیبانی، فعالیت اینجا نمایش داده می‌شود. درخواست‌های در انتظار را در{" "}
          <Link className="text-mahak-700" href="/volunteer/tasks">فعالیت‌ها</Link>
          {" "}ببینید.
        </Card>
      )}
      {items.filter((a) => a.status !== "requested").map((a) => {
        const remote = a.task?.work_mode === "remote";
        const canDeliver = remote && (a.status === "in_progress" || a.status === "submitted" || a.status === "revision_requested");
        const canStart = remote && a.status === "reserved";
        const canCancel = a.status === "requested" || a.status === "training_pending" || a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "revision_requested";
        const canRate = (a.status === "completed" || a.status === "attended") && !a.volunteer_rating;
        return (
          <Card key={a.id} className="p-5 space-y-3">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h2 className="font-bold">{a.task?.title}</h2>
                <p className="text-sm text-stone-500">
                  {workModeLabel(a.task?.work_mode)} · {a.task?.location || (remote ? "دورکار" : "—")} · {fmtDate(a.task?.starts_at)}
                </p>
                {a.task?.delivery_hint && <p className="mt-1 text-xs text-mahak-700">تحویل مورد انتظار: {a.task.delivery_hint}</p>}
                {a.composite_score && (
                  <StarRating label="امتیاز پشتیبانی" value={a.composite_score} readOnly size="sm" />
                )}
              </div>
              <Badge status={a.status} reason={a.admin_comment} />
            </div>
            <TrainingBadge task={a.task} completed={["reserved", "in_progress", "attended", "submitted", "revision_requested", "completed", "absent"].includes(a.status)} />

            {a.status === "absent" && (
              <p className="rounded-2xl bg-rose-50 px-3 py-2 text-sm text-rose-800">
                عدم حضور برای این فعالیت ثبت شده است.
              </p>
            )}
            {a.status === "revision_requested" && (
              <p className="rounded-2xl bg-amber-50 px-3 py-2 text-sm text-amber-900">
                واحد پشتیبانی درخواست اصلاح یا تکمیل نتیجه کرده است
                {a.admin_comment ? ` — ${a.admin_comment}` : "."} لطفاً نتیجه را اصلاح و دوباره ارسال کنید.
              </p>
            )}
            {a.status === "rejected" && (
              <p className="rounded-2xl bg-rose-50 px-3 py-2 text-sm text-rose-800">
                درخواست رد شد{a.admin_comment ? ` — ${a.admin_comment}` : "."}
              </p>
            )}

            {a.status === "training_pending" && (
              <p className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950">
                درخواست شما تایید شده است. ابتدا در آموزش این فعالیت شرکت کنید. پس از برگزاری، واحد پشتیبانی حضور شما در آموزش را تایید می‌کند و سپس مرحله انجام فعالیت شروع می‌شود.
              </p>
            )}

            {a.status === "reserved" && !remote && (
              <p className="rounded-2xl border border-mahak-100 bg-mahak-50/60 px-4 py-3 text-sm text-ink-800">
                درخواست شما تایید شده است. برای انجام فعالیت متناسب با زمان‌بندی فعالیت در محل حضور داشته باشید.
              </p>
            )}

            {canStart && (
              <div className="rounded-2xl border border-mahak-100 bg-mahak-50/60 p-4 space-y-3">
                <p className="text-sm text-ink-800">درخواست شما تایید شده است. برای کار دورکار، شروع فعالیت را بزنید و سپس نتیجه را بارگذاری کنید.</p>
                <Button disabled={busy === a.id} onClick={() => run(a.id, () => api.startAssignment(a.id), "فعالیت شروع شد")}>
                  شروع فعالیت
                </Button>
              </div>
            )}

            {canDeliver && (
              <div className="rounded-2xl border border-stone-100 bg-stone-50/70 p-4 space-y-2">
                <p className="text-sm font-medium text-ink-800">
                  {a.status === "revision_requested"
                    ? "نتیجه را اصلاح یا تکمیل کنید و دوباره بفرستید."
                    : a.status === "submitted"
                      ? "نتیجه ارسال شده؛ در صورت نیاز می‌توانید دوباره بفرستید."
                      : "نتیجه کار را بنویسید و در صورت نیاز فایل بارگذاری کنید."}
                </p>
                {a.delivery_note && <p className="text-sm text-stone-600">آخرین نتیجه: {a.delivery_note}</p>}
                {a.delivery_file_name && (
                  <p className="text-xs text-stone-500">پیوست: {a.delivery_file_name.length > 40 ? `${a.delivery_file_name.slice(0, 20)}…${a.delivery_file_name.slice(-8)}` : a.delivery_file_name}</p>
                )}
                <textarea
                  className={inputClass}
                  rows={3}
                  placeholder="مثلاً: تست انجام دادم / شرح کار انجام‌شده"
                  value={notes[a.id] ?? a.delivery_note ?? ""}
                  onChange={(e) => setNotes({ ...notes, [a.id]: e.target.value })}
                />
                <input type="file" onChange={(e) => setFiles({ ...files, [a.id]: e.target.files?.[0] })} />
                <Button
                  disabled={busy === a.id}
                  onClick={() => run(a.id, () => api.deliverAssignment(a.id, notes[a.id] || "", files[a.id]), "نتیجه ارسال شد و در انتظار بررسی واحد پشتیبانی است")}
                >
                  {a.status === "revision_requested" ? "ارسال نتیجه اصلاح‌شده" : a.status === "submitted" ? "ارسال مجدد نتیجه" : "ارسال نتیجه — انجام دادم"}
                </Button>
              </div>
            )}

            {canCancel && (
              <Button variant="danger" disabled={busy === a.id} onClick={() => run(a.id, () => api.cancelMyAssignment(a.id), "انصراف ثبت شد")}>
                انصراف
              </Button>
            )}

            {a.status === "completed" && (
              <Link className="inline-block text-sm text-mahak-700" href="/volunteer/certificates">
                درخواست تقدیرنامه این فعالیت
              </Link>
            )}

            {a.status === "attended" && !remote && (
              <p className="rounded-2xl bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
                حضور شما توسط واحد پشتیبانی ثبت شد
                {a.check_in_at ? ` · ورود ${fmtDate(a.check_in_at)}` : ""}
                {a.check_out_at ? ` · خروج ${fmtDate(a.check_out_at)}` : "."}
              </p>
            )}

            {canRate && (
              <div className="space-y-2 rounded-2xl border border-stone-100 p-4">
                <StarRating label="امتیاز به سازماندهی" value={rating[a.id] || 0} onChange={(n) => setRating({ ...rating, [a.id]: n })} />
                <textarea
                  className={inputClass}
                  rows={3}
                  placeholder="نظر خود را درباره این فعالیت بنویسید"
                  value={comments[a.id] || ""}
                  onChange={(e) => setComments({ ...comments, [a.id]: e.target.value })}
                />
                <Button variant="outline" disabled={busy === a.id || !(rating[a.id] > 0)} onClick={() => run(a.id, () => api.rateAssignment(a.id, rating[a.id], comments[a.id] || ""), "امتیاز و نظر ثبت شد")}>
                  ثبت امتیاز و نظر
                </Button>
              </div>
            )}
            {a.volunteer_rating ? (
              <div className="rounded-2xl bg-stone-50 px-3 py-2 text-sm">
                <StarRating label="امتیاز شما" value={a.volunteer_rating} readOnly size="sm" />
                {a.volunteer_comment && <p className="mt-1 text-stone-600">{a.volunteer_comment}</p>}
              </div>
            ) : null}
          </Card>
        );
      })}
    </div>
  );
}
