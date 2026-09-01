"use client";

import { useEffect, useMemo, useState } from "react";
import { api, Assignment, SkillGroup, Task, VolunteerTraining } from "@/lib/api";
import { weekdayLabel, catalogLabelMap, fmtDate, fmtDay, fmtTime, skillLabel, trainingSatisfied, workModeLabel, WEEKDAYS } from "@/lib/labels";
import { Badge, Button, Card, Modal } from "@/components/ui";
import { TrainingBadge, TrainingNotice } from "@/components/training-notice";

const DAYS_PAGE_SIZE = 10;

function faNum(n: number) {
  return n.toLocaleString("fa-IR");
}

function RecurringDaysBox({
  items,
  selected,
  page,
  onPage,
  onToggle,
}: {
  items: Task[];
  selected: string[];
  page: number;
  onPage: (page: number) => void;
  onToggle: (id: string) => void;
}) {
  const total = items.length;
  const pages = Math.max(1, Math.ceil(total / DAYS_PAGE_SIZE));
  const safePage = Math.min(Math.max(0, page), pages - 1);
  const from = safePage * DAYS_PAGE_SIZE;
  const visible = items.slice(from, from + DAYS_PAGE_SIZE);
  return (
    <div className="space-y-3">
      <ul className="grid gap-2 sm:grid-cols-2">
        {visible.map((o) => {
          const on = selected.includes(o.id);
          const time = fmtTime(o.starts_at);
          return (
            <li key={o.id}>
              <label
                className={`flex cursor-pointer items-start gap-3 rounded-2xl border px-3 py-3 ${on ? "border-mahak-300 bg-mahak-50" : "border-stone-200 bg-white"}`}
              >
                <input
                  type="checkbox"
                  className="mt-1 h-4 w-4 shrink-0"
                  checked={on}
                  onChange={() => onToggle(o.id)}
                />
                <span className="min-w-0 flex-1">
                  <span className="block text-sm font-medium leading-6">
                    {weekdayLabel(o.weekday)} {fmtDay(o.starts_at)}
                  </span>
                  <span className="mt-0.5 flex flex-wrap gap-x-3 text-xs text-stone-500">
                    {time ? <span>ساعت {time}</span> : null}
                    <span>ظرفیت {o.reserved_count}/{o.capacity}</span>
                  </span>
                </span>
              </label>
            </li>
          );
        })}
      </ul>
      {pages > 1 && (
        <div className="flex flex-wrap items-center justify-between gap-2 border-t border-stone-100 pt-3">
          <Button
            variant="outline"
            className="min-h-11 px-3 py-2"
            disabled={safePage <= 0}
            onClick={() => onPage(safePage - 1)}
          >
            صفحه قبلی
          </Button>
          <p className="min-w-0 flex-1 text-center text-xs text-stone-500">
            {faNum(from + 1)} تا {faNum(from + visible.length)} از {faNum(total)}
            {" — "}صفحه {faNum(safePage + 1)} از {faNum(pages)}
          </p>
          <Button
            variant="outline"
            className="min-h-11 px-3 py-2"
            disabled={safePage >= pages - 1}
            onClick={() => onPage(safePage + 1)}
          >
            صفحه بعدی
          </Button>
        </div>
      )}
    </div>
  );
}

export default function TasksPage() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [busy, setBusy] = useState<string>("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [okOpen, setOkOpen] = useState(false);
  const [pick, setPick] = useState<Record<string, string[]>>({});
  const [pending, setPending] = useState<Assignment[]>([]);
  const [courses, setCourses] = useState<VolunteerTraining[]>([]);
  const [confirm, setConfirm] = useState<{ ids: string[]; key: string; task: Task } | null>(null);
  const [ackTrain, setAckTrain] = useState(false);
  const [dayPage, setDayPage] = useState<Record<string, number>>({});
  const [pickOpen, setPickOpen] = useState<string | null>(null);

  async function loadTasks() {
    const [r, mine, train] = await Promise.all([
      api.tasks(),
      api.myAssignments().catch(() => [] as Assignment[]),
      api.myTrainings().catch(() => [] as VolunteerTraining[]),
    ]);
    setItems(r.items || []);
    setPending((mine || []).filter((a) => a.status === "requested"));
    setCourses(train || []);
  }

  useEffect(() => {
    loadTasks().catch((e) => setMsg(e instanceof Error ? e.message : "خطا"));
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
  }, []);

  const titleById = useMemo(() => {
    const m = new Map<string, string>();
    for (const g of catalog) {
      for (const s of g.skills || []) m.set(s.id, `${g.title} / ${s.title}`);
    }
    return m;
  }, [catalog]);
  const skillNames = useMemo(() => catalogLabelMap(catalog), [catalog]);

  function seriesKey(t: Task) {
    const sid = t.series_id || "";
    const missing = !sid || sid.startsWith("00000000-0000-0000-0000");
    if (t.kind === "occurrence" && !missing) return sid;
    return t.id;
  }

  const groups = useMemo(() => {
    const map = new Map<string, Task[]>();
    for (const t of items) {
      const key = seriesKey(t);
      const cur = map.get(key) || [];
      cur.push(t);
      map.set(key, cur);
    }
    return Array.from(map.values()).map((list) => {
      const sorted = [...list].sort((a, b) => a.starts_at.localeCompare(b.starts_at));
      return { key: seriesKey(sorted[0]), items: sorted, head: sorted[0] };
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

  function request(ids: string[], key: string, task: Task) {
    if (!ids.length) {
      setErr("حداقل یک روز را برای درخواست انتخاب کنید");
      return;
    }
    if (task.requires_training && !trainingSatisfied(task, courses)) {
      setAckTrain(false);
      setConfirm({ ids, key, task });
      return;
    }
    void accept(ids, key);
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
      await loadTasks();
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
        پس از ارسال درخواست، تا تایید واحد پشتیبانی در همین صفحه با وضعیت «در انتظار تایید» می‌ماند. پس از تایید به «کارهای من» می‌رود.
      </p>
      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Modal open={okOpen} title="درخواست در انتظار تایید است" onClose={() => setOkOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست شما ثبت شد و تا تایید واحد پشتیبانی در همین صفحه می‌ماند. پس از تایید، فعالیت به «کارهای من» منتقل می‌شود.
        </p>
        <div className="mt-4 flex flex-wrap justify-end gap-2">
          <Button onClick={() => setOkOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>
      {(() => {
        const g = pickOpen ? groups.find((x) => x.key === pickOpen) : undefined;
        if (!g) return null;
        const selected = selectedIds(g.key);
        const weekdays = Array.from(
          new Set(g.items.map((o) => o.weekday).filter((wd): wd is number => typeof wd === "number")),
        ).sort((a, b) => a - b);
        return (
          <Modal
            open
            size="lg"
            title={`انتخاب روز — ${g.head.title}`}
            onClose={() => setPickOpen(null)}
          >
            <p className="text-sm text-stone-600">
              روزهایی که می‌خواهید درخواست بدهید را مشخص کنید.
              {selected.length ? ` ${faNum(selected.length)} روز انتخاب شده است.` : " هنوز روزی انتخاب نشده است."}
            </p>
            {weekdays.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {weekdays.map((wd) => {
                  const ids = g.items.filter((o) => o.weekday === wd).map((o) => o.id);
                  const on = ids.length > 0 && ids.every((id) => selected.includes(id));
                  return (
                    <button
                      type="button"
                      key={wd}
                      onClick={() => toggleWeekday(g.key, g.items, wd)}
                      className={`min-h-11 rounded-2xl border px-3 py-2 text-sm ${on ? "border-mahak-400 bg-mahak-50 font-medium text-mahak-800" : "border-stone-200 bg-white"}`}
                    >
                      همه {WEEKDAYS[wd]}‌ها
                    </button>
                  );
                })}
              </div>
            )}
            <div className="mt-3">
              <RecurringDaysBox
                items={g.items}
                selected={selected}
                page={dayPage[g.key] || 0}
                onPage={(p) => setDayPage({ ...dayPage, [g.key]: p })}
                onToggle={(id) => toggleDate(g.key, id)}
              />
            </div>
            <div className="mt-4 flex flex-wrap justify-end gap-2">
              <Button variant="ghost" onClick={() => setPickOpen(null)}>بستن</Button>
              <Button onClick={() => setPickOpen(null)}>تایید انتخاب</Button>
            </div>
          </Modal>
        );
      })()}
      <Modal
        open={!!confirm}
        title="این فعالیت نیاز به آموزش دارد"
        onClose={() => { setConfirm(null); setAckTrain(false); }}
      >
        <p className="text-sm leading-7 text-stone-700">
          برای شرکت در این فعالیت باید در جلسه آموزش حاضر شوید. جزئیات را ببینید و در صورت موافقت، درخواست را ارسال کنید.
        </p>
        {confirm && <TrainingNotice task={confirm.task} className="mt-3 rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-950" />}
        <label className="mt-3 flex cursor-pointer items-start gap-2 text-sm leading-6 text-stone-700">
          <input type="checkbox" className="mt-1" checked={ackTrain} onChange={(e) => setAckTrain(e.target.checked)} />
          زمان، محل و نوع آموزش را دیدم و برای حضور در آموزش تایید می‌کنم.
        </label>
        <div className="mt-4 flex flex-wrap justify-end gap-2">
          <Button variant="ghost" onClick={() => { setConfirm(null); setAckTrain(false); }}>انصراف</Button>
          <Button
            disabled={!ackTrain}
            onClick={() => {
              if (!confirm || !ackTrain) return;
              const { ids, key } = confirm;
              setConfirm(null);
              setAckTrain(false);
              void accept(ids, key);
            }}
          >
            تایید و ارسال درخواست
          </Button>
        </div>
      </Modal>
      {pending.length > 0 && (
        <Card className="p-5">
          <h2 className="font-bold">در انتظار تایید واحد پشتیبانی</h2>
          <ul className="mt-3 space-y-2">
            {pending.map((a) => (
              <li key={a.id} className="flex flex-col gap-3 rounded-2xl bg-amber-50 px-3 py-3 text-sm sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0 flex-1">
                  <div className="font-medium">{a.task?.title || "فعالیت"}</div>
                  <div className="text-xs text-stone-500">{workModeLabel(a.task?.work_mode)} · {fmtDate(a.task?.starts_at)}</div>
                  <TrainingBadge task={a.task} className="mt-2" />
                </div>
                <div className="flex items-center gap-2">
                  <Badge status="requested" />
                  <Button variant="danger" onClick={async () => {
                    try {
                      await api.cancelMyAssignment(a.id);
                      await loadTasks();
                    } catch (e) {
                      setErr(e instanceof Error ? e.message : "خطا");
                    }
                  }}>انصراف</Button>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}
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
          return (
            <Card key={g.key} className="p-4 sm:p-5">
              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="text-lg font-bold">{g.head.title}</h2>
                      <Badge status={t.status} />
                      <TrainingBadge task={t} completed={trainingSatisfied(t, courses)} />
                    </div>
                    {recurring && (
                      <p className="mt-1 text-xs text-mahak-700">فعالیت جاری — روزهایی که می‌خواهید درخواست بدهید را مشخص کنید</p>
                    )}
                    <p className="mt-1 text-sm text-stone-600">{g.head.description}</p>
                  </div>
                  <div className="flex w-full shrink-0 flex-col gap-2 sm:w-auto">
                    {recurring && (
                      <Button variant="outline" className="w-full sm:w-auto" onClick={() => setPickOpen(g.key)}>
                        انتخاب روز{selected.length ? ` (${faNum(selected.length)})` : ""}
                      </Button>
                    )}
                    <Button
                      className="w-full sm:w-auto"
                      disabled={busy === g.key || (recurring && selected.length === 0)}
                      onClick={() => request(recurring ? selected : [g.head.id], g.key, t)}
                    >
                      {recurring ? `ارسال درخواست (${faNum(selected.length)} روز)` : "ارسال درخواست"}
                    </Button>
                  </div>
                </div>
                {recurring && (
                  <p className="text-xs text-stone-500">
                    {selected.length ? `${faNum(selected.length)} روز انتخاب شده — برای تغییر، «انتخاب روز» را باز کنید` : "هنوز روزی انتخاب نشده است. با «انتخاب روز» نوبت‌ها را ببینید."}
                  </p>
                )}
                <p className="text-xs text-stone-500">
                  {workModeLabel(t.work_mode)} · {t.location || (t.work_mode === "remote" ? "دورکار" : "—")}
                  {recurring ? "" : ` · ${fmtDate(t.starts_at)} تا ${fmtDate(t.ends_at)}`}
                  {" "}· معادل {t.hour_weight} ساعت
                </p>
                {t.work_mode === "remote" && t.delivery_hint && (
                  <p className="text-xs text-mahak-700">تحویل: {t.delivery_hint}</p>
                )}
                <div className="flex flex-wrap gap-1">
                  {(t.required_skill_ids || []).length > 0
                    ? (t.required_skill_ids || []).map((id) => (
                        <span key={id} className="rounded-full bg-stone-100 px-2 py-0.5 text-xs">{titleById.get(id) || id}</span>
                      ))
                    : (t.required_skills || []).map((s) => (
                        <span key={s} className="rounded-full bg-stone-100 px-2 py-0.5 text-xs">{skillLabel(s, skillNames)}</span>
                      ))}
                  {t.min_score > 0 && <span className="text-xs text-stone-500">حداقل امتیاز {t.min_score}</span>}
                </div>
              </div>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
