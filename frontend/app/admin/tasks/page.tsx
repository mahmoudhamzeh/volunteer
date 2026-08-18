"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment, SkillGroup, Task } from "@/lib/api";
import { fmtDate, skillLabel } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";
import { ShamsiDateTimeField } from "@/components/shamsi";
import { gregorianToJalali, jalaliToIsoDateTime } from "@/lib/jalali";

function defaultTaskTimes() {
  const n = new Date();
  const j = gregorianToJalali(n.getFullYear(), n.getMonth() + 1, n.getDate());
  return {
    starts_at: jalaliToIsoDateTime(j.jy, j.jm, j.jd, 9, 0),
    ends_at: jalaliToIsoDateTime(j.jy, j.jm, j.jd, 13, 0),
  };
}

export default function AdminTasks() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [applicants, setApplicants] = useState<Record<string, Assignment[]>>({});
  const [openTask, setOpenTask] = useState("");
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [msg, setMsg] = useState("");
  const [groupId, setGroupId] = useState("");
  const [form, setForm] = useState({
    title: "", description: "", location: "", ...defaultTaskTimes(),
    capacity: 5, hour_weight: 4, min_score: 0, required_education: "",
    required_skills: [] as string[],
    required_skill_ids: [] as string[],
  });

  async function load() {
    const r = await api.adminTasks();
    setItems(r.items || []);
  }

  async function loadApplicants(taskId: string) {
    const r = await api.adminAssignments(`?task_id=${taskId}&limit=100`);
    setApplicants((prev) => ({ ...prev, [taskId]: r.items || [] }));
  }
  useEffect(() => {
    load();
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
  }, []);

  const group = catalog.find((g) => g.id === groupId);

  function toggleSub(id: string) {
    const ids = form.required_skill_ids.includes(id)
      ? form.required_skill_ids.filter((x) => x !== id)
      : [...form.required_skill_ids, id];
    const slugs = catalog
      .filter((g) => (g.skills || []).some((s) => ids.includes(s.id)))
      .map((g) => g.slug);
    setForm({ ...form, required_skill_ids: ids, required_skills: slugs });
  }

  const selectedLabels = useMemo(() => {
    const out: { id: string; label: string }[] = [];
    for (const g of catalog) {
      for (const s of g.skills || []) {
        if (form.required_skill_ids.includes(s.id)) out.push({ id: s.id, label: `${g.title} / ${s.title}` });
      }
    }
    return out;
  }, [catalog, form.required_skill_ids]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    await api.createTask({
      ...form,
      starts_at: form.starts_at,
      ends_at: form.ends_at,
    });
    setForm({ ...form, title: "", description: "", required_skill_ids: [], required_skills: [] });
    await load();
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">تعریف فعالیت عملیاتی</h1>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      <Card className="p-5">
        <form onSubmit={onSubmit} className="grid gap-3 md:grid-cols-2">
          <Field label="عنوان"><input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></Field>
          <Field label="مکان"><input className={inputClass} value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} /></Field>
          <div className="md:col-span-2">
            <Field label="شرح کار"><textarea className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} required /></Field>
          </div>
          <ShamsiDateTimeField label="شروع (شمسی)" value={form.starts_at} onChange={(starts_at) => setForm({ ...form, starts_at })} />
          <ShamsiDateTimeField label="پایان (شمسی)" value={form.ends_at} onChange={(ends_at) => setForm({ ...form, ends_at })} />
          <Field label="ظرفیت"><input type="number" className={inputClass} value={form.capacity} onChange={(e) => setForm({ ...form, capacity: Number(e.target.value) })} /></Field>
          <Field label="وزن ساعتی"><input type="number" step="0.5" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} /></Field>
          <Field label="حداقل امتیاز"><input type="number" step="0.1" className={inputClass} value={form.min_score} onChange={(e) => setForm({ ...form, min_score: Number(e.target.value) })} /></Field>
          <Field label="رشته تحصیلی الزامی"><input className={inputClass} value={form.required_education} onChange={(e) => setForm({ ...form, required_education: e.target.value })} /></Field>
          <div className="md:col-span-2 space-y-3">
            <Field label="گروه مهارت">
              <select className={inputClass} value={groupId} onChange={(e) => setGroupId(e.target.value)}>
                <option value="">انتخاب گروه</option>
                {catalog.map((g) => (
                  <option key={g.id} value={g.id}>{g.title}</option>
                ))}
              </select>
            </Field>
            {group && (
              <div>
                <div className="mb-2 text-sm text-stone-600">زیرمجموعه {group.title}</div>
                <div className="flex flex-wrap gap-2">
                  {(group.skills || []).filter((s) => s.status !== "inactive").map((s) => (
                    <button
                      type="button"
                      key={s.id}
                      onClick={() => toggleSub(s.id)}
                      className={`rounded-full border px-3 py-1 text-xs ${form.required_skill_ids.includes(s.id) ? "border-mahak-400 bg-mahak-50 text-mahak-800" : "border-stone-200"}`}
                    >
                      {s.title}
                    </button>
                  ))}
                </div>
              </div>
            )}
            {selectedLabels.length > 0 && (
              <div className="flex flex-wrap gap-2 text-xs">
                {selectedLabels.map((s) => (
                  <span key={s.id} className="rounded-full bg-mahak-50 px-2 py-1 text-mahak-800">{s.label}</span>
                ))}
              </div>
            )}
          </div>
          <Button type="submit">ایجاد فعالیت</Button>
        </form>
      </Card>
      {items.map((t) => (
        <Card key={t.id} className="p-4">
          <div className="flex items-start justify-between">
            <div>
              <div className="font-bold">{t.title}</div>
              <div className="text-xs text-stone-500">{t.location} · {fmtDate(t.starts_at)} تا {fmtDate(t.ends_at)} · تاییدشده {t.reserved_count}/{t.capacity} · {t.hour_weight} ساعت</div>
              <div className="mt-1 flex flex-wrap gap-1">
                {(t.required_skill_ids || []).length > 0
                  ? (t.required_skill_ids || []).map((id) => {
                      const hit = catalog.flatMap((g) => (g.skills || []).map((s) => ({ id: s.id, label: `${g.title} / ${s.title}` }))).find((x) => x.id === id);
                      return <span key={id} className="rounded-full bg-stone-100 px-2 py-0.5 text-[11px]">{hit?.label || id}</span>;
                    })
                  : (t.required_skills || []).map((s) => (
                      <span key={s} className="rounded-full bg-stone-100 px-2 py-0.5 text-[11px]">{skillLabel(s)}</span>
                    ))}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge status={t.status} />
              <Button variant="outline" onClick={async () => {
                const next = openTask === t.id ? "" : t.id;
                setOpenTask(next);
                if (next) await loadApplicants(t.id);
              }}>{openTask === t.id ? "بستن درخواست‌ها" : "درخواست‌ها"}</Button>
            </div>
          </div>
          {openTask === t.id && (
            <div className="mt-4 space-y-3 border-t border-stone-100 pt-4">
              {(applicants[t.id] || []).length === 0 && <p className="text-sm text-stone-400">هنوز درخواستی ثبت نشده</p>}
              {(applicants[t.id] || []).map((a) => (
                <div key={a.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-3">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <Link className="font-medium text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                      {a.volunteer?.full_name || "داوطلب"}
                    </Link>
                    <Badge status={a.status} />
                  </div>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    {a.status === "requested" && (
                      <>
                        <Button onClick={async () => {
                          try {
                            await api.approveAssignment(a.id);
                            setMsg("تایید شد");
                            await load();
                            await loadApplicants(t.id);
                          } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                        }}>تایید و رزرو ظرفیت</Button>
                        <Button variant="danger" onClick={async () => {
                          try {
                            await api.rejectAssignment(a.id);
                            setMsg("رد شد");
                            await loadApplicants(t.id);
                          } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                        }}>رد</Button>
                      </>
                    )}
                    {a.status === "reserved" && (
                      <Button variant="danger" onClick={async () => {
                        try {
                          await api.rejectAssignment(a.id);
                          setMsg("رزرو لغو شد");
                          await load();
                          await loadApplicants(t.id);
                        } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                      }}>لغو رزرو</Button>
                    )}
                    <input
                      className={inputClass + " max-w-xs"}
                      placeholder="پیام به داوطلب"
                      value={notes[a.id] || ""}
                      onChange={(e) => setNotes({ ...notes, [a.id]: e.target.value })}
                    />
                    <Button variant="ghost" onClick={async () => {
                      try {
                        await api.messageAssignment(a.id, notes[a.id] || "");
                        setNotes({ ...notes, [a.id]: "" });
                        setMsg("پیام ارسال شد");
                      } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                    }}>ارسال پیام</Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      ))}
    </div>
  );
}
