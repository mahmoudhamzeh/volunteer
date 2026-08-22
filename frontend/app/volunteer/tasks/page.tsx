"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, SkillGroup, Task } from "@/lib/api";
import { weekdayLabel, fmtDate, skillLabel, workModeLabel, WEEKDAYS } from "@/lib/labels";
import { Badge, Button, Card, Modal } from "@/components/ui";

export default function TasksPage() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [busy, setBusy] = useState<string>("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [okOpen, setOkOpen] = useState(false);
  const [pick, setPick] = useState<Record<string, string[]>>({});

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

  const groups = useMemo(() => {
    const map = new Map<string, Task[]>();
    for (const t of items) {
      const key = t.series_id || t.id;
      const cur = map.get(key) || [];
      cur.push(t);
      map.set(key, cur);
    }
    return Array.from(map.values()).map((list) => {
      const sorted = [...list].sort((a, b) => a.starts_at.localeCompare(b.starts_at));
      return { key: sorted[0].series_id || sorted[0].id, items: sorted, head: sorted[0] };
    });
  }, [items]);

  function selectedIds(key: string) {
    return pick[key] || [];
  }

  function toggleDate(key: string, id: string) {
    const cur = selectedIds(key);
    const next = cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id];
    setPick({ ...pick, [key]: next });
  }

  function toggleWeekday(key: string, items: Task[], wd: number) {
    const ids = items.filter((o) => o.weekday === wd).map((o) => o.id);
    const cur = selectedIds(key);
    const allOn = ids.every((id) => cur.includes(id));
    const next = allOn ? cur.filter((id) => !ids.includes(id)) : Array.from(new Set([...cur, ...ids]));
    setPick({ ...pick, [key]: next });
  }

  async function accept(ids: string[], key: string) {
    if (!ids.length) {
      setErr("حداقل یک روز را برای درخواست انتخاب کنید");
      return;
    }
    setBusy(key);
    setErr("");
    setMsg("");
    const failed: string[] = [];
    let ok = 0;
    for (const id of ids) {
      try {
        await api.acceptTask(id);
        ok += 1;
      } catch (e) {
        failed.push(e instanceof Error ? e.message : "خطا");
      }
    }
    try {
      const r = await api.tasks();
      setItems(r.items || []);
    } catch {
      /* ignore */
    }
    setBusy("");
    if (ok > 0) {
      setOkOpen(true);
      setPick({ ...pick, [key]: [] });
    }
    if (failed.length) {
      setErr(ok > 0 ? `${ok} نوبت ثبت شد. بقیه: ${failed[0]}` : failed[0]);
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
          درخواست شما برای روزهای انتخاب‌شده ارسال شد و در حال بررسی است. پس از تایید یا رد ادمین، نتیجه در اعلان‌ها و صفحه «کارهای من» نمایش داده می‌شود.
        </p>
        <div className="mt-4 flex flex-wrap justify-end gap-2">
          <Link href="/volunteer/work" className="rounded-2xl border border-mahak-200 px-4 py-2.5 text-sm text-mahak-700">رفتن به کارهای من</Link>
          <Button onClick={() => setOkOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>
      {groups.length === 0 && (
        <Card className="p-6 text-stone-500">
          فعالیت واجد شرایطی برای مهارت‌های شما نیست، هنوز تایید نشده‌اید، یا همه درخواست‌هایتان ثبت شده‌اند.
        </Card>
      )}
      <div className="grid gap-4">
        {groups.map((g) => {
          const recurring = g.items.length > 1 || g.head.kind === "occurrence";
          const selected = recurring ? selectedIds(g.key) : [g.head.id];
          const t = g.head;
          const weekdays = Array.from(new Set(g.items.map((o) => o.weekday))).sort((a, b) => a - b);
          return (
            <Card key={g.key} className="p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <h2 className="text-lg font-bold">{g.head.title}</h2>
                  {recurring && <p className="text-xs text-mahak-700">فعالیت جاری — روزهایی که می‌خواهید درخواست بدهید را مشخص کنید</p>}
                  <p className="mt-1 text-sm text-stone-600">{g.head.description}</p>
                  {recurring && (
                    <div className="mt-3 space-y-2">
                      <div className="flex flex-wrap gap-2">
                        {weekdays.map((wd) => {
                          const ids = g.items.filter((o) => o.weekday === wd).map((o) => o.id);
                          const on = ids.length > 0 && ids.every((id) => selected.includes(id));
                          return (
                            <button
                              type="button"
                              key={wd}
                              onClick={() => toggleWeekday(g.key, g.items, wd)}
                              className={`rounded-full border px-3 py-1 text-xs ${on ? "border-mahak-400 bg-mahak-50 text-mahak-800" : "border-stone-200"}`}
                            >
                              همه {WEEKDAYS[wd]}‌ها
                            </button>
                          );
                        })}
                      </div>
                      <div className="max-h-48 space-y-1 overflow-y-auto rounded-2xl border border-stone-100 p-2">
                        {g.items.map((o) => {
                          const on = selected.includes(o.id);
                          return (
                            <label key={o.id} className="flex cursor-pointer items-center gap-2 rounded-xl px-2 py-1.5 text-sm hover:bg-stone-50">
                              <input type="checkbox" checked={on} onChange={() => toggleDate(g.key, o.id)} />
                              <span>{weekdayLabel(o.weekday)} · {fmtDate(o.starts_at)} · ظرفیت {o.reserved_count}/{o.capacity}</span>
                            </label>
                          );
                        })}
                      </div>
                      <p className="text-xs text-stone-500">
                        {selected.length ? `${selected.length} نوبت انتخاب شده` : "هنوز روزی انتخاب نشده است"}
                      </p>
                    </div>
                  )}
                  <p className="mt-2 text-xs text-stone-500">
                    {workModeLabel(t.work_mode)} · {t.location || (t.work_mode === "remote" ? "دورکار" : "—")}
                    {recurring ? "" : ` · ${fmtDate(t.starts_at)} تا ${fmtDate(t.ends_at)}`}
                    {" "}· معادل {t.hour_weight} ساعت
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
                  <Button
                    disabled={busy === g.key || (recurring && selected.length === 0)}
                    onClick={() => accept(recurring ? selected : [g.head.id], g.key)}
                  >
                    {recurring ? `ارسال درخواست (${selected.length} روز)` : "ارسال درخواست"}
                  </Button>
                </div>
              </div>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
