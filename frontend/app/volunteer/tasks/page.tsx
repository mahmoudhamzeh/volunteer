"use client";

import { useEffect, useMemo, useState } from "react";
import { api, SkillGroup, Task } from "@/lib/api";
import { fmtDate, skillLabel } from "@/lib/labels";
import { Badge, Button, Card } from "@/components/ui";

export default function TasksPage() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [busy, setBusy] = useState<string>("");
  const [msg, setMsg] = useState("");

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
    try {
      await api.acceptTask(id);
      setMsg("درخواست ارسال شد و در انتظار تایید ادمین است");
      const r = await api.tasks();
      setItems(r.items || []);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    } finally {
      setBusy("");
    }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">فعالیت‌های قابل درخواست</h1>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
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
                  {t.location} · {fmtDate(t.starts_at)} تا {fmtDate(t.ends_at)} · ظرفیت تاییدشده {t.reserved_count}/{t.capacity} · معادل {t.hour_weight} ساعت
                </p>
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
