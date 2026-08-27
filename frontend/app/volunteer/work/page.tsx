"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment } from "@/lib/api";
import { WORK_STATUS_FILTERS, fmtDate, matchesWorkStatusFilter, sortAssignmentsOpenFirst, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, Modal, StarRating, inputClass } from "@/components/ui";
import { TrainingBadge } from "@/components/training-notice";
import { DeliveryHistory } from "@/components/delivery-history";
import { ShamsiDateField } from "@/components/shamsi";
import { currentJalaliYear } from "@/lib/jalali";

function assignmentDay(a: Assignment) {
  const iso = a.task?.starts_at || a.created_at || "";
  return iso.slice(0, 10);
}

export default function WorkPage() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [openId, setOpenId] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [rating, setRating] = useState<Record<string, number>>({});
  const [comments, setComments] = useState<Record<string, string>>({});
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [files, setFiles] = useState<Record<string, File[]>>({});
  const [fileKey, setFileKey] = useState<Record<string, number>>({});
  const [busy, setBusy] = useState("");
  const [msg, setMsg] = useState("");

  async function load() {
    const asg = await api.myAssignments();
    setItems(asg || []);
  }
  useEffect(() => { void load(); }, []);

  async function run(id: string, fn: () => Promise<unknown>, ok: string) {
    setBusy(id);
    try {
      await fn();
      setMsg(ok);
      setNotes((cur) => ({ ...cur, [id]: "" }));
      setFiles((cur) => ({ ...cur, [id]: [] }));
      setFileKey((cur) => ({ ...cur, [id]: (cur[id] || 0) + 1 }));
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    } finally {
      setBusy("");
    }
  }

  const visible = useMemo(() => {
    const list = sortAssignmentsOpenFirst(items.filter((a) => a.status !== "requested"));
    return list.filter((a) => {
      if (!matchesWorkStatusFilter(a.status, statusFilter)) return false;
      const day = assignmentDay(a);
      if (from && day && day < from) return false;
      if (to && day && day > to) return false;
      return true;
    });
  }, [items, statusFilter, from, to]);

  const open = items.find((a) => a.id === openId) || null;

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">کارهای من</h1>
        <p className="mt-1 text-sm text-stone-500">
          عنوان و وضعیت هر فعالیت را ببینید. برای شروع، ارسال نتیجه یا جزئیات، روی ردیف بزنید.
        </p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-4">
        <div className="grid gap-3 md:grid-cols-4">
          <label className="block space-y-1.5">
            <span className="text-sm text-stone-600">وضعیت</span>
            <select className={inputClass} value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              {WORK_STATUS_FILTERS.map((f) => (
                <option key={f.id || "all"} value={f.id}>{f.label}</option>
              ))}
            </select>
          </label>
          <ShamsiDateField label="از تاریخ" value={from} onChange={setFrom} minYear={currentJalaliYear() - 2} maxYear={currentJalaliYear() + 2} />
          <ShamsiDateField label="تا تاریخ" value={to} onChange={setTo} minYear={currentJalaliYear() - 2} maxYear={currentJalaliYear() + 2} />
          <div className="flex items-end">
            <Button variant="ghost" type="button" onClick={() => { setStatusFilter(""); setFrom(""); setTo(""); }}>
              پاک کردن فیلتر
            </Button>
          </div>
        </div>
      </Card>

      {items.filter((a) => a.status !== "requested").length === 0 && (
        <Card className="p-6 text-stone-500">
          پس از تایید واحد پشتیبانی، فعالیت اینجا نمایش داده می‌شود. درخواست‌های در انتظار را در{" "}
          <Link className="text-mahak-700" href="/volunteer/tasks">فعالیت‌ها</Link>
          {" "}ببینید.
        </Card>
      )}

      {items.filter((a) => a.status !== "requested").length > 0 && visible.length === 0 && (
        <Card className="p-6 text-stone-500">با این فیلتر موردی نیست.</Card>
      )}

      <ul className="space-y-2">
        {visible.map((a) => (
          <li key={a.id}>
            <button
              type="button"
              onClick={() => setOpenId(a.id)}
              className="flex w-full items-center justify-between gap-3 rounded-2xl border border-stone-100 bg-white px-4 py-3 text-right hover:border-mahak-200 hover:bg-mahak-50/40"
            >
              <div className="min-w-0">
                <div className="truncate font-bold">{a.task?.title || "فعالیت"}</div>
                <div className="mt-0.5 text-xs text-stone-500">
                  {workModeLabel(a.task?.work_mode)}
                  {a.task?.starts_at ? ` · ${fmtDate(a.task.starts_at)}` : ""}
                </div>
              </div>
              <Badge status={a.status} reason={a.admin_comment} />
            </button>
          </li>
        ))}
      </ul>

      <Modal open={!!open} size="lg" title={open?.task?.title || "جزئیات فعالیت"} onClose={() => setOpenId("")}>
        {open && (
          <WorkDetail
            a={open}
            busy={busy === open.id}
            rating={rating[open.id] || 0}
            comment={comments[open.id] || ""}
            note={notes[open.id] ?? ""}
            files={files[open.id] || []}
            fileKey={fileKey[open.id] || 0}
            onRating={(n) => setRating({ ...rating, [open.id]: n })}
            onComment={(v) => setComments({ ...comments, [open.id]: v })}
            onNote={(v) => setNotes({ ...notes, [open.id]: v })}
            onFiles={(list) => setFiles({ ...files, [open.id]: list })}
            onRun={(fn, ok) => run(open.id, fn, ok)}
            onClose={() => setOpenId("")}
          />
        )}
      </Modal>
    </div>
  );
}

function WorkDetail({
  a,
  busy,
  rating,
  comment,
  note,
  files,
  fileKey,
  onRating,
  onComment,
  onNote,
  onFiles,
  onRun,
  onClose,
}: {
  a: Assignment;
  busy: boolean;
  rating: number;
  comment: string;
  note: string;
  files: File[];
  fileKey: number;
  onRating: (n: number) => void;
  onComment: (v: string) => void;
  onNote: (v: string) => void;
  onFiles: (files: File[]) => void;
  onRun: (fn: () => Promise<unknown>, ok: string) => void;
  onClose: () => void;
}) {
  const remote = a.task?.work_mode === "remote";
  const canDeliver = remote && (a.status === "in_progress" || a.status === "submitted" || a.status === "revision_requested");
  const canStart = remote && a.status === "reserved";
  const canCancel = a.status === "requested" || a.status === "training_pending" || a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "revision_requested";
  const canRate = (a.status === "completed" || a.status === "attended") && !a.volunteer_rating;

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <p className="text-sm text-stone-500">
          {workModeLabel(a.task?.work_mode)} · {a.task?.location || (remote ? "دورکار" : "—")} · {fmtDate(a.task?.starts_at)}
        </p>
        <Badge status={a.status} reason={a.admin_comment} />
      </div>
      {a.task?.delivery_hint && <p className="text-xs text-mahak-700">تحویل مورد انتظار: {a.task.delivery_hint}</p>}
      {a.composite_score && (
        <StarRating label="امتیاز پشتیبانی" value={a.composite_score} readOnly size="sm" />
      )}
      <TrainingBadge task={a.task} completed={["reserved", "in_progress", "attended", "submitted", "revision_requested", "completed", "absent"].includes(a.status)} />

      {a.status === "absent" && (
        <p className="rounded-2xl bg-rose-50 px-3 py-2 text-sm text-rose-800">عدم حضور برای این فعالیت ثبت شده است.</p>
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
          درخواست شما تایید شده است. ابتدا در جلسه آموزش این فعالیت شرکت کنید. تا تایید آموزش توسط واحد پشتیبانی، امکان ادامه فرایند فعالیت نیست.
        </p>
      )}
      {a.status === "reserved" && !remote && (
        <p className="rounded-2xl border border-mahak-100 bg-mahak-50/60 px-4 py-3 text-sm text-ink-800">
          درخواست شما تایید شده است. برای انجام فعالیت متناسب با زمان‌بندی فعالیت در محل حضور داشته باشید.
        </p>
      )}

      {canStart && (
        <div className="rounded-2xl border border-mahak-100 bg-mahak-50/60 p-4 space-y-3">
          <p className="text-sm text-ink-800">برای کار دورکار، شروع فعالیت را بزنید و سپس نتیجه را بارگذاری کنید.</p>
          <Button disabled={busy} onClick={() => onRun(() => api.startAssignment(a.id), "فعالیت شروع شد")}>شروع فعالیت</Button>
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
          <DeliveryHistory
            items={a.history}
            assignmentId={a.id}
            fileHref={(aid, fid) => `/api/v1/assignments/${aid}/files/${fid}`}
          />
          <textarea
            className={inputClass}
            rows={3}
            placeholder="مثلاً: تست انجام دادم / شرح کار انجام‌شده"
            value={note}
            onChange={(e) => onNote(e.target.value)}
          />
          <input
            key={fileKey}
            type="file"
            multiple
            onChange={(e) => onFiles(Array.from(e.target.files || []))}
          />
          {files.length > 0 && (
            <p className="text-xs text-stone-500">{files.length} فایل انتخاب شده: {files.map((f) => f.name).join("، ")}</p>
          )}
          <Button
            disabled={busy}
            onClick={() => onRun(() => api.deliverAssignment(a.id, note, files), "نتیجه ارسال شد و در انتظار بررسی واحد پشتیبانی است")}
          >
            {a.status === "revision_requested" ? "ارسال نتیجه اصلاح‌شده" : a.status === "submitted" ? "ارسال مجدد نتیجه" : "ارسال نتیجه — انجام دادم"}
          </Button>
        </div>
      )}

      {a.status === "completed" && a.task?.work_mode === "remote" && (
        <DeliveryHistory
          items={a.history}
          assignmentId={a.id}
          fileHref={(aid, fid) => `/api/v1/assignments/${aid}/files/${fid}`}
        />
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
          <StarRating label="امتیاز به سازماندهی" value={rating} onChange={onRating} />
          <textarea
            className={inputClass}
            rows={3}
            placeholder="نظر خود را درباره این فعالیت بنویسید"
            value={comment}
            onChange={(e) => onComment(e.target.value)}
          />
          <Button variant="outline" disabled={busy || !(rating > 0)} onClick={() => onRun(() => api.rateAssignment(a.id, rating, comment), "امتیاز و نظر ثبت شد")}>
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

      {a.status === "completed" && (
        <Link className="inline-block text-sm text-mahak-700" href="/volunteer/certificates">
          درخواست تقدیرنامه این فعالیت
        </Link>
      )}

      <div className="flex flex-wrap gap-2 pt-1">
        {canCancel && (
          <Button variant="danger" disabled={busy} onClick={() => { onRun(() => api.cancelMyAssignment(a.id), "انصراف ثبت شد"); onClose(); }}>
            انصراف
          </Button>
        )}
        <Button variant="ghost" onClick={onClose}>بستن</Button>
      </div>
    </div>
  );
}
