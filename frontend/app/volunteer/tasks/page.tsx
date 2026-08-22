"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, SkillGroup, Task } from "@/lib/api";
import { fmtDate, skillLabel, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, Modal } from "@/components/ui";

export default function TasksPage() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [busy, setBusy] = useState<string>("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [okOpen, setOkOpen] = useState(false);

  useEffect(() => {
    api.tasks().then((r) => setItems(r.items || [])).catch((e) => setMsg(e.message));
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
  }, []);

  const titleById = useMemo(() => {
    const m = new Map<string, string>();
    for (const g of catalog) {
      for (const s of g.skills || []) m.set(s.id, `${g.title} / ${s.title}`);
    }
    return m;
  }, [catalog]);

  async function accept(id: string) {
    setBusy(id);
    setErr("");
    setMsg("");
    try {
      await api.acceptTask(id);
      setOkOpen(true);
      const r = await api.tasks();
      setItems(r.items || []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در ارسال درخواست");
    } finally {
      setBusy("");
    }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">فعالیت‌های قابل درخواست</h1>
      <p className="text-sm text-stone-500">
        پس از ارسال درخواست و تایید ادمین، از صفحه{" "}
        <Link className="text-mahak-700" href="/volunteer/work">کارهای من</Link>
        {" "}فعالیت را شروع کنید و نتیجه را بفرستید.
      </p>
      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Modal open={okOpen} title="درخواست ارسال شد" onClose={() => setOkOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست شما ارسال شد و در حال بررسی است. پس از تایید یا رد ادمین، نتیجه در اعلان‌ها و صفحه «کارهای من» نمایش داده می‌شود.
        </p>
        <div className="mt-4 flex flex-wrap justify-end gap-2">
          <Link href="/volunteer/work" className="rounded-2xl border border-mahak-200 px-4 py-2.5 text-sm text-mahak-700">رفتن به کارهای من</Link>
          <Button onClick={() => setOkOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>
      {items.length === 0 && (
        <Card className="p-6 text-stone-500">
          فعالیت واجد شرایطی برای مهارت‌های شما نیست، هنوز تایید نشده‌اید، یا همه درخواست‌هایتان ثبت شده‌اند.
        </Card>
      )}
      <div className="grid gap-4">
        {items.map((t) => (
          <Card key={t.id} className="p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 className="text-lg font-bold">{t.title}</h2>
                <p className="mt-1 text-sm text-stone-600">{t.description}</p>
                <p className="mt-2 text-xs text-stone-500">
                  {workModeLabel(t.work_mode)} · {t.location || (t.work_mode === "remote" ? "دورکار" : "—")} · {fmtDate(t.starts_at)} تا {fmtDate(t.ends_at)} · ظرفیت تاییدشده {t.reserved_count}/{t.capacity} · معادل {t.hour_weight} ساعت
                </p>
                {t.work_mode === "remote" && t.delivery_hint && (
                  <p className="mt-1 text-xs text-mahak-700">تحویل: {t.delivery_hint}</p>
                )}
                <div className="mt-2 flex flex-wrap gap-1">
                  {(t.required_skill_ids || []).length > 0
                    ? (t.required_skill_ids || []).map((id) => (
                        <span key={id} className="rounded-full bg-stone-100 px-2 py-0.5 text-xs">{titleById.get(id) || id}</span>
                      ))
                    : (t.required_skills || []).map((s) => (
                        <span key={s} className="rounded-full bg-stone-100 px-2 py-0.5 text-xs">{skillLabel(s)}</span>
                      ))}
                  {t.min_score > 0 && <span className="text-xs text-stone-500">حداقل امتیاز {t.min_score}</span>}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge status={t.status} />
                <Button disabled={busy === t.id} onClick={() => accept(t.id)}>ارسال درخواست</Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
