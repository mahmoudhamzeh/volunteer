"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Assignment } from "@/lib/api";
import { fmtDate, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, inputClass } from "@/components/ui";

export default function WorkPage() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [rating, setRating] = useState<Record<string, number>>({});
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [files, setFiles] = useState<Record<string, File | undefined>>({});
  const [busy, setBusy] = useState("");
  const [msg, setMsg] = useState("");

  async function load() {
    setItems((await api.myAssignments()) || []);
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
          بعد از تایید ادمین، فعالیت را شروع کنید و نتیجه را اینجا ارسال کنید.
        </p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {items.length === 0 && (
        <Card className="p-6 text-stone-500">
          هنوز درخواستی ثبت نکرده‌اید.{" "}
          <Link className="text-mahak-700" href="/volunteer/tasks">مشاهده فعالیت‌ها</Link>
        </Card>
      )}
      {items.map((a) => {
        const remote = a.task?.work_mode === "remote";
        const canDeliver = a.status === "in_progress" || a.status === "submitted";
        const canCancel = a.status === "requested" || a.status === "reserved" || a.status === "in_progress" || a.status === "submitted";
        return (
          <Card key={a.id} className="p-5 space-y-3">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h2 className="font-bold">{a.task?.title}</h2>
                <p className="text-sm text-stone-500">
                  {workModeLabel(a.task?.work_mode)} · {a.task?.location || (remote ? "دورکار" : "—")} · {fmtDate(a.task?.starts_at)}
                </p>
                {a.task?.delivery_hint && <p className="mt-1 text-xs text-mahak-700">تحویل مورد انتظار: {a.task.delivery_hint}</p>}
                {a.composite_score && <p className="text-sm">امتیاز مدیر: {a.composite_score.toFixed(1)}</p>}
              </div>
              <Badge status={a.status} />
            </div>

            {a.status === "requested" && (
              <p className="rounded-2xl bg-amber-50 px-3 py-2 text-sm text-amber-900">
                درخواست ثبت شد. پس از تایید ادمین می‌توانید فعالیت را شروع کنید.
              </p>
            )}

            {a.status === "reserved" && (
              <div className="rounded-2xl border border-mahak-100 bg-mahak-50/60 p-4 space-y-3">
                <p className="text-sm text-ink-800">درخواست شما تایید شد. برای انجام کار، شروع فعالیت را بزنید.</p>
                <Button disabled={busy === a.id} onClick={() => run(a.id, () => api.startAssignment(a.id), "فعالیت شروع شد")}>
                  شروع فعالیت
                </Button>
              </div>
            )}

            {canDeliver && (
              <div className="rounded-2xl border border-stone-100 bg-stone-50/70 p-4 space-y-2">
                <p className="text-sm font-medium text-ink-800">
                  {a.status === "submitted" ? "نتیجه ارسال شده؛ در صورت نیاز می‌توانید دوباره بفرستید." : "نتیجه کار را بنویسید و در صورت نیاز فایل بارگذاری کنید."}
                </p>
                {a.delivery_note && <p className="text-sm text-stone-600">آخرین نتیجه: {a.delivery_note}</p>}
                {a.delivery_file_name && <p className="text-xs text-stone-500">فایل: {a.delivery_file_name}</p>}
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
                  onClick={() => run(a.id, () => api.deliverAssignment(a.id, notes[a.id] || "", files[a.id]), "نتیجه ارسال شد و در انتظار بررسی ادمین است")}
                >
                  {a.status === "submitted" ? "ارسال مجدد نتیجه" : "ارسال نتیجه — انجام دادم"}
                </Button>
              </div>
            )}

            {canCancel && (
              <Button variant="danger" disabled={busy === a.id} onClick={() => run(a.id, () => api.cancelMyAssignment(a.id), "انصراف ثبت شد")}>
                انصراف
              </Button>
            )}

            {(a.status === "completed" || a.status === "attended") && !a.volunteer_rating && (
              <div className="flex items-center gap-2">
                <input className={inputClass + " w-20"} type="number" min={1} max={5} placeholder="1-5"
                  onChange={(e) => setRating({ ...rating, [a.id]: Number(e.target.value) })} />
                <Button variant="outline" onClick={async () => {
                  await api.rateAssignment(a.id, rating[a.id] || 5, "");
                  await load();
                }}>امتیاز به سازماندهی</Button>
              </div>
            )}
          </Card>
        );
      })}
    </div>
  );
}
