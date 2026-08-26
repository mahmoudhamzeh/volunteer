"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { api, Assignment, SkillGroup, Task, TaskSlot, Volunteer, openAuth } from "@/lib/api";
import { WEEKDAYS, TRAINING_KINDS, fmtDate, skillLabel, weekdayLabel, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, AttachmentButton, inputClass } from "@/components/ui";
import { TrainingBadge } from "@/components/training-notice";
import { ShamsiDateField, ShamsiDateTimeField } from "@/components/shamsi";
import { AttendancePanel } from "@/components/attendance-panel";
import { gregorianToJalali, jalaliToIsoDateTime, currentJalaliYear } from "@/lib/jalali";

function defaultTaskTimes() {
  const n = new Date();
  const j = gregorianToJalali(n.getFullYear(), n.getMonth() + 1, n.getDate());
  return {
    starts_at: jalaliToIsoDateTime(j.jy, j.jm, j.jd, 9, 0),
    ends_at: jalaliToIsoDateTime(j.jy, j.jm, j.jd, 13, 0),
  };
}

const emptyForm = () => ({
  title: "", description: "", location: "", ...defaultTaskTimes(),
  capacity: 5, hour_weight: 4, min_score: 0, required_education: "",
  work_mode: "onsite",
  delivery_hint: "",
  kind: "one_off",
  slots: [] as TaskSlot[],
  required_skills: [] as string[],
  required_skill_ids: [] as string[],
  requires_training: false,
  training_kind: "in_person",
  training_location: "",
  training_at: defaultTaskTimes().starts_at,
});

export default function AdminTasks() {
  const [items, setItems] = useState<Task[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [applicants, setApplicants] = useState<Record<string, Assignment[]>>({});
  const [manageId, setManageId] = useState("");
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [msg, setMsg] = useState("");
  const [groupId, setGroupId] = useState("");
  const [editingId, setEditingId] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [volunteers, setVolunteers] = useState<Volunteer[]>([]);
  const [volQuery, setVolQuery] = useState("");
  const [pick, setPick] = useState<Record<string, string>>({});
  const [listFilter, setListFilter] = useState("open");
  const [seriesId, setSeriesId] = useState("");
  const [occurrences, setOccurrences] = useState<Task[]>([]);
  const [manageOccs, setManageOccs] = useState<Task[]>([]);
  const [assignOccs, setAssignOccs] = useState<string[]>([]);
  const openedManage = useRef("");

  async function load() {
    const r = await api.adminTasks();
    setItems(r.items || []);
    try {
      const a = await api.adminAssignments("?limit=200");
      const byTask: Record<string, Assignment[]> = {};
      for (const x of a.items || []) {
        (byTask[x.task_id] ||= []).push(x);
        const sid = x.task?.series_id;
        if (x.task?.kind === "occurrence" && sid && sid !== x.task_id) {
          (byTask[sid] ||= []).push(x);
        }
      }
      setApplicants(byTask);
    } catch {
      /* list still usable */
    }
  }

  async function loadApplicants(taskId: string) {
    const t = items.find((x) => x.id === taskId);
    const q = t?.kind === "recurring" || t?.kind === "occurrence"
      ? `?series_id=${t.kind === "occurrence" && t.series_id ? t.series_id : taskId}&limit=200`
      : `?task_id=${taskId}&limit=100`;
    const r = await api.adminAssignments(q);
    setApplicants((prev) => ({ ...prev, [taskId]: r.items || [] }));
  }
  useEffect(() => {
    load();
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!manageId) return;
    const q = volQuery.trim();
    const t = window.setTimeout(() => {
      const qs = new URLSearchParams({ status: "approved", limit: "100" });
      if (q) qs.set("q", q);
      api.adminVolunteers(`?${qs.toString()}`).then((r) => setVolunteers(r.items || [])).catch(() => undefined);
    }, q ? 280 : 0);
    return () => window.clearTimeout(t);
  }, [manageId, volQuery]);

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

  function resetForm() {
    setEditingId("");
    setShowForm(false);
    setForm(emptyForm());
    setGroupId("");
  }

  function startCreate() {
    setEditingId("");
    setForm(emptyForm());
    setGroupId("");
    setShowForm(true);
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function startEdit(t: Task) {
    setEditingId(t.id);
    setShowForm(true);
    const gid = catalog.find((g) => (g.skills || []).some((s) => (t.required_skill_ids || []).includes(s.id)))?.id || "";
    setGroupId(gid);
    setForm({
      title: t.title,
      description: t.description,
      location: t.location || "",
      starts_at: t.starts_at,
      ends_at: t.ends_at,
      capacity: t.capacity,
      hour_weight: t.hour_weight,
      min_score: t.min_score,
      required_education: t.required_education || "",
      work_mode: t.work_mode || "onsite",
      delivery_hint: t.delivery_hint || "",
      kind: t.kind === "recurring" ? "recurring" : "one_off",
      slots: t.slots || [],
      required_skills: t.required_skills || [],
      required_skill_ids: t.required_skill_ids || [],
      requires_training: Boolean(t.requires_training),
      training_kind: t.training_kind || "in_person",
      training_location: t.training_location || "",
      training_at: t.training_at || defaultTaskTimes().starts_at,
    });
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const title = form.title.trim();
    const description = form.description.trim();
    if (!title) {
      setMsg("عنوان فعالیت را وارد کنید");
      return;
    }
    if (!description) {
      setMsg("شرح فعالیت را وارد کنید");
      return;
    }
    const start = new Date(form.starts_at);
    const end = new Date(form.ends_at);
    if (!form.starts_at || Number.isNaN(start.getTime())) {
      setMsg("تاریخ شروع نامعتبر است؛ تاریخ و ساعت شروع را از تقویم انتخاب کنید");
      return;
    }
    if (!form.ends_at || Number.isNaN(end.getTime())) {
      setMsg("تاریخ پایان نامعتبر است؛ تاریخ و ساعت پایان را از تقویم انتخاب کنید");
      return;
    }
    if (form.kind === "recurring") {
      if (!form.slots.length) {
        setMsg("برای فعالیت جاری حداقل یک روز هفته را با ظرفیت انتخاب کنید");
        return;
      }
    } else {
      if (!(end.getTime() > start.getTime())) {
        setMsg("تاریخ پایان باید بعد از تاریخ شروع باشد");
        return;
      }
      if (!Number.isFinite(form.capacity) || form.capacity < 1) {
        setMsg("ظرفیت باید حداقل ۱ نفر باشد");
        return;
      }
    }
    if (!Number.isFinite(form.hour_weight) || form.hour_weight <= 0) {
      setMsg("وزن ساعتی باید بزرگ‌تر از صفر باشد");
      return;
    }
    if (form.requires_training) {
      if (!form.training_kind) {
        setMsg("نوع آموزش را مشخص کنید");
        return;
      }
      if (!form.training_location.trim()) {
        setMsg("محل آموزش را وارد کنید");
        return;
      }
      const trainAt = new Date(form.training_at);
      if (!form.training_at || Number.isNaN(trainAt.getTime())) {
        setMsg("زمان آموزش را مشخص کنید");
        return;
      }
    }
    const body: Record<string, unknown> = {
      ...form,
      status: editingId ? items.find((x) => x.id === editingId)?.status : "open",
      training_kind: form.requires_training ? form.training_kind : "",
      training_location: form.requires_training ? form.training_location.trim() : "",
      training_at: form.requires_training ? form.training_at : null,
    };
    if (form.kind === "recurring") {
      body.starts_at = form.starts_at.length <= 10 ? `${form.starts_at}T06:00:00+03:30` : form.starts_at;
      body.ends_at = form.ends_at.length <= 10 ? `${form.ends_at}T18:00:00+03:30` : form.ends_at;
    }
    try {
      if (editingId) await api.updateTask(editingId, body);
      else await api.createTask(body);
      setMsg(editingId ? "فعالیت ویرایش شد" : "فعالیت ایجاد شد");
      resetForm();
      await load();
      if (manageId) await openManage(manageId);
      if (seriesId) await openSeries(seriesId);
    } catch (err) {
      setMsg(err instanceof Error ? err.message : "خطا");
    }
  }

  async function setStatus(id: string, status: string, ok: string) {
    try {
      await api.setTaskStatus(id, status);
      setMsg(ok);
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  const volunteerChoices = useMemo(() => {
    const q = volQuery.trim();
    return (volunteers || []).filter((v) => {
      if (!q) return true;
      const hay = `${v.full_name} ${v.city || ""} ${v.province || ""} ${v.phone || ""} ${v.national_id || ""} ${v.email || ""}`;
      return hay.includes(q);
    });
  }, [volunteers, volQuery]);

  const visibleItems = useMemo(() => {
    if (listFilter === "all") return items;
    return items.filter((t) => t.status === listFilter);
  }, [items, listFilter]);

  const manageTask = items.find((t) => t.id === manageId);

  async function assign(taskId: string) {
    const vid = pick[taskId];
    if (!vid) {
      setMsg("داوطلب را انتخاب کنید");
      return;
    }
    const t = items.find((x) => x.id === taskId);
    const targets = t?.kind === "recurring" ? assignOccs : [taskId];
    if (t?.kind === "recurring" && !targets.length) {
      setMsg("روزهای تخصیص را مثل درخواست داوطلب انتخاب کنید");
      return;
    }
    try {
      let ok = 0;
      let lastErr = "";
      for (const id of targets) {
        try {
          await api.assignVolunteer(id, vid);
          ok += 1;
        } catch (e) {
          lastErr = e instanceof Error ? e.message : "خطا";
        }
      }
      if (ok === 0) {
        setMsg(lastErr || "تخصیص انجام نشد");
        return;
      }
      setMsg(ok === targets.length ? "داوطلب به روزهای انتخاب‌شده تخصیص داده شد" : `${ok} نوبت تخصیص شد. ${lastErr}`);
      setPick({ ...pick, [taskId]: "" });
      setAssignOccs([]);
      await load();
      await openManage(taskId);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  function toggleWeekday(wd: number) {
    const exists = form.slots.some((s) => s.weekday === wd);
    const slots = exists
      ? form.slots.filter((s) => s.weekday !== wd)
      : [...form.slots, { weekday: wd, capacity: 5, start_time: "09:00", end_time: "13:00" }];
    setForm({ ...form, slots });
  }

  function patchSlot(wd: number, patch: Partial<TaskSlot>) {
    setForm({ ...form, slots: form.slots.map((s) => s.weekday === wd ? { ...s, ...patch } : s) });
  }

  async function openSeries(id: string) {
    setSeriesId(id);
    const r = await api.adminTasks(`?series_id=${id}&limit=500`);
    setOccurrences(r.items || []);
  }

  async function openManage(id: string) {
    let t = items.find((x) => x.id === id) || null;
    if (!t) {
      try {
        t = await api.getTask(id);
      } catch {
        t = null;
      }
    }
    const manage = t?.kind === "occurrence" && t.series_id ? t.series_id : id;
    setManageId(manage);
    setAssignOccs([]);
    const series = t?.kind === "recurring" ? t.id : (t?.kind === "occurrence" ? (t.series_id || t.id) : "");
    if (series) {
      const [apps, occ] = await Promise.all([
        api.adminAssignments(`?series_id=${series}&limit=200`),
        api.adminTasks(`?series_id=${series}&limit=500`),
      ]);
      setApplicants((prev) => ({ ...prev, [manage]: apps.items || [] }));
      setManageOccs((occ.items || []).filter((o) => o.status !== "closed" && o.status !== "cancelled"));
      return;
    }
    setManageOccs([]);
    await loadApplicants(manage);
  }

  useEffect(() => {
    const id = typeof window === "undefined" ? "" : new URLSearchParams(window.location.search).get("manage") || "";
    if (!id || !items.length || openedManage.current === id) return;
    openedManage.current = id;
    void openManage(id);
  }, [items.length]);

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black">فعالیت‌های عملیاتی</h1>
          <p className="mt-1 text-sm text-stone-500">تعریف فعالیت، بررسی درخواست‌ها و تخصیص داوطلب در همین صفحه انجام می‌شود.</p>
        </div>
        <Button onClick={showForm && !editingId ? resetForm : startCreate}>
          {showForm && !editingId ? "بستن فرم" : "تعریف فعالیت جدید"}
        </Button>
      </div>
      {msg && (
        <p className={`text-sm ${/خطا|نامعتبر|باید|وارد کنید|بزرگ‌تر|READONLY|replica/.test(msg) ? "text-rose-600" : "text-mahak-700"}`}>
          {msg}
        </p>
      )}

      {showForm && (
        <Card className="p-5">
          <div className="mb-4 flex items-center justify-between gap-2">
            <h2 className="font-bold">{editingId ? "ویرایش فعالیت" : "فعالیت جدید"}</h2>
            <Button variant="ghost" onClick={resetForm}>انصراف</Button>
          </div>
          <form onSubmit={onSubmit} className="space-y-4">
            <section className="grid gap-3 rounded-2xl border border-stone-100 bg-stone-50/50 p-4 md:grid-cols-2">
              <h3 className="text-sm font-bold text-stone-700 md:col-span-2">مشخصات</h3>
              <Field label="عنوان"><input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></Field>
              <Field label="مکان">
                <input
                  className={inputClass}
                  placeholder={form.work_mode === "remote" ? "دورکار" : "محل حضور"}
                  value={form.location}
                  onChange={(e) => setForm({ ...form, location: e.target.value })}
                />
              </Field>
              <div className="md:col-span-2">
                <Field label="شرح کار"><textarea className={inputClass} rows={3} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} required /></Field>
              </div>
            </section>

            <section className="grid gap-3 rounded-2xl border border-stone-100 bg-stone-50/50 p-4 md:grid-cols-2">
              <h3 className="text-sm font-bold text-stone-700 md:col-span-2">زمان، ظرفیت و نحوه اجرا</h3>
              <Field label="نوع اجرا">
                <select className={inputClass} value={form.work_mode} onChange={(e) => setForm({ ...form, work_mode: e.target.value })}>
                  <option value="onsite">حضوری — واحد پشتیبانی حضور و غیاب را ثبت می‌کند</option>
                  <option value="remote">دورکار — داوطلب شروع می‌زند و نتیجه را بارگذاری می‌کند</option>
                </select>
              </Field>
              <Field label="مدل فعالیت">
                <select className={inputClass} value={form.kind} onChange={(e) => setForm({ ...form, kind: e.target.value, slots: e.target.value === "recurring" ? form.slots : [] })}>
                  <option value="one_off">موردی — یک بازه زمانی مشخص</option>
                  <option value="recurring">جاری — بازه + روزهای هفته با ظرفیت جدا</option>
                </select>
              </Field>
              <Field label="راهنمای تحویل نتیجه (اختیاری)">
                <input className={inputClass} placeholder="مثلا فایل پوستر، گزارش تست، یا لینک" value={form.delivery_hint} onChange={(e) => setForm({ ...form, delivery_hint: e.target.value })} />
              </Field>
              {form.kind === "recurring" ? (
                <>
                  <ShamsiDateField label="شروع بازه (شمسی)" value={form.starts_at} onChange={(starts_at) => setForm({ ...form, starts_at })} minYear={currentJalaliYear() - 1} maxYear={currentJalaliYear() + 2} />
                  <ShamsiDateField label="پایان بازه (شمسی)" value={form.ends_at} onChange={(ends_at) => setForm({ ...form, ends_at })} minYear={currentJalaliYear() - 1} maxYear={currentJalaliYear() + 2} />
                </>
              ) : (
                <>
                  <ShamsiDateTimeField label="شروع (شمسی)" value={form.starts_at} onChange={(starts_at) => setForm({ ...form, starts_at })} />
                  <ShamsiDateTimeField label="پایان (شمسی)" value={form.ends_at} onChange={(ends_at) => setForm({ ...form, ends_at })} />
                </>
              )}
              {form.kind !== "recurring" && (
                <Field label="ظرفیت"><input type="number" className={inputClass} value={form.capacity} onChange={(e) => setForm({ ...form, capacity: Number(e.target.value) })} /></Field>
              )}
              <Field label="وزن ساعتی"><input type="number" step="0.5" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} /></Field>
              <Field label="حداقل امتیاز"><input type="number" step="0.1" className={inputClass} value={form.min_score} onChange={(e) => setForm({ ...form, min_score: Number(e.target.value) })} /></Field>
              <Field label="رشته تحصیلی (اختیاری)">
                <input
                  className={inputClass}
                  placeholder="اگر محدودیتی نیست خالی بگذارید"
                  value={form.required_education}
                  onChange={(e) => setForm({ ...form, required_education: e.target.value })}
                />
              </Field>
            </section>

            <section className="grid gap-3 rounded-2xl border border-amber-100 bg-amber-50/40 p-4 md:grid-cols-2">
              <h3 className="text-sm font-bold text-stone-700 md:col-span-2">آموزش</h3>
              <label className="flex items-center gap-2 text-sm md:col-span-2">
                <input
                  type="checkbox"
                  checked={form.requires_training}
                  onChange={(e) => setForm({ ...form, requires_training: e.target.checked })}
                />
                این فعالیت نیاز به آموزش دارد
              </label>
              {form.requires_training && (
                <>
                  <Field label="نوع آموزش">
                    <select className={inputClass} value={form.training_kind} onChange={(e) => setForm({ ...form, training_kind: e.target.value })}>
                      {TRAINING_KINDS.map((k) => (
                        <option key={k.id} value={k.id}>{k.label}</option>
                      ))}
                    </select>
                  </Field>
                  <Field label="محل آموزش">
                    <input
                      className={inputClass}
                      placeholder="سالن آموزش، لینک کلاس آنلاین، یا آدرس"
                      value={form.training_location}
                      onChange={(e) => setForm({ ...form, training_location: e.target.value })}
                    />
                  </Field>
                  <div className="md:col-span-2">
                    <ShamsiDateTimeField label="زمان آموزش (شمسی)" value={form.training_at} onChange={(training_at) => setForm({ ...form, training_at })} />
                  </div>
                </>
              )}
            </section>

            {form.kind === "recurring" && (
              <section className="space-y-3 rounded-2xl border border-stone-100 bg-stone-50/50 p-4">
                <h3 className="text-sm font-bold text-stone-700">روزهای هفته و ظرفیت هر روز</h3>
                <p className="text-xs text-stone-500">مثلا دوشنبه ظرفیت ۳ و سه‌شنبه ظرفیت ۸. سامانه برای هر روز داخل بازه یک نوبت می‌سازد.</p>
                <div className="flex flex-wrap gap-2">
                  {WEEKDAYS.map((label, wd) => {
                    const on = form.slots.some((s) => s.weekday === wd);
                    return (
                      <button
                        type="button"
                        key={wd}
                        onClick={() => toggleWeekday(wd)}
                        className={`rounded-full border px-3 py-1 text-xs ${on ? "border-mahak-400 bg-mahak-50 text-mahak-800" : "border-stone-200"}`}
                      >
                        {label}
                      </button>
                    );
                  })}
                </div>
                { [...form.slots].sort((a, b) => a.weekday - b.weekday).map((s) => (
                  <div key={s.weekday} className="grid gap-2 md:grid-cols-4">
                    <div className="flex items-end text-sm font-medium">{WEEKDAYS[s.weekday]}</div>
                    <Field label="ظرفیت"><input type="number" className={inputClass} value={s.capacity} onChange={(e) => patchSlot(s.weekday, { capacity: Number(e.target.value) })} /></Field>
                    <Field label="ساعت شروع"><input className={inputClass} placeholder="09:00" value={s.start_time} onChange={(e) => patchSlot(s.weekday, { start_time: e.target.value })} /></Field>
                    <Field label="ساعت پایان"><input className={inputClass} placeholder="13:00" value={s.end_time} onChange={(e) => patchSlot(s.weekday, { end_time: e.target.value })} /></Field>
                  </div>
                ))}
              </section>
            )}

            <section className="space-y-3 rounded-2xl border border-stone-100 bg-stone-50/50 p-4">
              <h3 className="text-sm font-bold text-stone-700">مهارت مورد نیاز</h3>
              <p className="text-xs text-stone-500">اگر مهارت «عمومی» انتخاب شود، همه داوطلبان فعال این فعالیت را می‌بینند.</p>
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
              {form.required_skills.includes("general") && (
                <p className="rounded-2xl bg-mahak-50 px-3 py-2 text-sm text-mahak-800">
                  مهارت عمومی انتخاب شده است؛ این فعالیت برای همه داوطلبان فعال نمایش داده می‌شود.
                </p>
              )}
              {selectedLabels.length > 0 && (
                <div className="flex flex-wrap gap-2 text-xs">
                  {selectedLabels.map((s) => (
                    <span key={s.id} className="rounded-full bg-mahak-50 px-2 py-1 text-mahak-800">{s.label}</span>
                  ))}
                </div>
              )}
            </section>

            <div className="flex flex-wrap gap-2">
              <Button type="submit">{editingId ? "ذخیره تغییرات" : "ایجاد فعالیت"}</Button>
              <Button variant="ghost" onClick={resetForm}>انصراف</Button>
            </div>
          </form>
        </Card>
      )}

      <div className="flex flex-wrap items-center gap-2">
        {[
          { id: "open", label: "باز" },
          { id: "closed", label: "اتمام‌یافته" },
          { id: "inactive", label: "غیرفعال" },
          { id: "all", label: "همه" },
        ].map((f) => (
          <button
            key={f.id}
            onClick={() => setListFilter(f.id)}
            className={`rounded-full px-3 py-1 text-sm ${listFilter === f.id ? "bg-mahak-500 text-white" : "bg-white"}`}
          >
            {f.label}
          </button>
        ))}
        <span className="text-xs text-stone-400">{visibleItems.length} فعالیت</span>
      </div>

      {visibleItems.length === 0 && <Card className="p-6 text-stone-500">فعالیتی در این وضعیت نیست.</Card>}

      <div className="space-y-3">
        {visibleItems.map((t) => {
          const apps = applicants[t.id] || [];
          const pending = apps.filter((a) => a.status === "requested" || a.status === "training_pending").length;
          return (
            <Card key={t.id} className="p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="font-bold">{t.title}</div>
                  <div className="text-xs text-stone-500">
                    {t.kind === "recurring" ? "فعالیت جاری · " : ""}                    {workModeLabel(t.work_mode)} · {t.location || (t.work_mode === "remote" ? "دورکار" : "—")} · {fmtDate(t.starts_at)} تا {fmtDate(t.ends_at)} · تاییدشده {t.reserved_count}/{t.capacity} · {t.hour_weight} ساعت
                  </div>
                  <TrainingBadge task={t} className="mt-2" />
                  {t.kind === "recurring" && (t.slots || []).length > 0 && (
                    <div className="mt-1 text-xs text-mahak-700">
                    {(t.slots || []).map((s) => `${WEEKDAYS[s.weekday]} ظرفیت ${s.capacity}`).join("، ")}
                    </div>
                  )}
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
                  {pending > 0 && <span className="text-xs text-amber-700">{pending} درخواست جدید</span>}
                  <Badge status={t.status} />
                </div>
              </div>
              <div className="mt-3 flex flex-wrap gap-2">
                <Button variant="outline" onClick={() => startEdit(t)}>ویرایش</Button>
                <Button variant="outline" onClick={() => openManage(t.id)}>درخواست‌ها و تخصیص</Button>
                {t.kind === "recurring" && (
                  <Button variant="outline" onClick={() => void openSeries(t.id)}>نوبت‌های روزانه</Button>
                )}
                {t.status === "open" && (
                  <>
                    <Button variant="outline" onClick={() => setStatus(t.id, "closed", "فعالیت به اتمام رسید")}>اتمام</Button>
                    <Button variant="danger" onClick={() => setStatus(t.id, "inactive", "فعالیت غیرفعال شد")}>غیرفعالسازی</Button>
                  </>
                )}
                {t.status !== "open" && (
                  <Button variant="ghost" onClick={() => setStatus(t.id, "open", "فعالیت دوباره فعال شد")}>فعال‌سازی مجدد</Button>
                )}
              </div>
            </Card>
          );
        })}
      </div>

      {manageTask && (
        <Modal open={!!manageId} size="lg" title={`درخواست‌ها و تخصیص — ${manageTask.title}`} onClose={() => setManageId("")}>
          <p className="text-sm text-stone-500">
            {workModeLabel(manageTask.work_mode)} · {manageTask.location || (manageTask.work_mode === "remote" ? "دورکار" : "—")} · تاییدشده {manageTask.reserved_count}/{manageTask.capacity}
          </p>
          <TrainingBadge task={manageTask} className="mt-2" />
          {manageTask.kind === "recurring" && (manageTask.slots || []).length > 0 && (
            <p className="mt-1 text-xs text-mahak-700">
              {(manageTask.slots || []).map((s) => `${WEEKDAYS[s.weekday]} ظرفیت ${s.capacity}`).join("، ")}
            </p>
          )}
          {msg && <p className="mt-2 text-sm text-mahak-700">{msg}</p>}
          {manageTask.status === "open" && (
            <div className="mt-4 space-y-3 rounded-2xl border border-stone-100 bg-stone-50/70 p-3">
              <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto]">
                <div className="space-y-2">
                  <Field label="جستجوی داوطلب">
                    <input
                      className={inputClass}
                      placeholder="نام، شهر یا موبایل"
                      value={volQuery}
                      onChange={(e) => setVolQuery(e.target.value)}
                    />
                  </Field>
                  {volQuery.trim() && (
                    <div className="max-h-40 overflow-y-auto rounded-2xl border border-stone-200 bg-white">
                      {volunteerChoices.length === 0 && <p className="px-3 py-2 text-xs text-stone-400">داوطلبی پیدا نشد</p>}
                      {volunteerChoices.slice(0, 20).map((v) => {
                        const selected = pick[manageTask.id] === v.id;
                        return (
                          <button
                            type="button"
                            key={v.id}
                            onClick={() => setPick({ ...pick, [manageTask.id]: v.id })}
                            className={`block w-full px-3 py-2 text-right text-sm hover:bg-mahak-50 ${selected ? "bg-mahak-50 text-mahak-800" : ""}`}
                          >
                            {v.full_name}{v.city ? ` · ${v.city}` : ""}{v.phone ? ` · ${v.phone}` : ""}
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
                <Field label="داوطلب تاییدشده">
                  <select className={inputClass} value={pick[manageTask.id] || ""} onChange={(e) => setPick({ ...pick, [manageTask.id]: e.target.value })}>
                    <option value="">انتخاب کنید</option>
                    {volunteerChoices.map((v) => (
                      <option key={v.id} value={v.id}>{v.full_name}{v.city ? ` · ${v.city}` : ""}{v.phone ? ` · ${v.phone}` : ""}</option>
                    ))}
                  </select>
                </Field>
                <div className="flex items-end">
                  <Button onClick={() => assign(manageTask.id)}>تخصیص</Button>
                </div>
              </div>
              {manageTask.kind === "recurring" && (
                <div className="space-y-2">
                  <p className="text-xs text-stone-500">روزهایی که داوطلب باید تخصیص داده شود را مشخص کنید (مثل درخواست خود داوطلب).</p>
                  <div className="max-h-48 space-y-1 overflow-y-auto rounded-2xl border border-white bg-white p-2">
                    {[...manageOccs].sort((a, b) => a.starts_at.localeCompare(b.starts_at)).map((o) => {
                      const on = assignOccs.includes(o.id);
                      return (
                        <label key={o.id} className="flex cursor-pointer items-center gap-2 rounded-xl px-2 py-1.5 text-sm hover:bg-stone-50">
                          <input
                            type="checkbox"
                            checked={on}
                            onChange={() => setAssignOccs((cur) => on ? cur.filter((x) => x !== o.id) : [...cur, o.id])}
                          />
                          <span>{weekdayLabel(o.weekday)} · {fmtDate(o.starts_at)} · ظرفیت {o.reserved_count}/{o.capacity}</span>
                        </label>
                      );
                    })}
                  </div>
                  <p className="text-xs text-stone-500">{assignOccs.length ? `${assignOccs.length} نوبت انتخاب شده` : "هنوز روزی انتخاب نشده است"}</p>
                </div>
              )}
            </div>
          )}
          <div className="mt-4 space-y-3">
            <h3 className="font-bold">درخواست‌های داوطلبان</h3>
            {(() => {
              const q = volQuery.trim();
              const apps = (applicants[manageTask.id] || []).filter((a) => {
                if (!q) return true;
                const hay = `${a.volunteer?.full_name || ""} ${a.volunteer?.phone || ""} ${a.volunteer?.city || ""}`;
                return hay.includes(q);
              });
              if (apps.length === 0) {
                return <p className="text-sm text-stone-400">{q ? "درخواستی با این جستجو نیست" : "هنوز درخواستی ثبت نشده"}</p>;
              }
              return apps.map((a) => (
              <div key={a.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <Link className="font-medium text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                    {a.volunteer?.full_name || "داوطلب"}
                  </Link>
                  {a.task?.starts_at && (
                    <div className="text-xs text-stone-500">{weekdayLabel(a.task.weekday)} · {fmtDate(a.task.starts_at)}</div>
                  )}
                  <Badge status={a.status} reason={a.admin_comment} />
                </div>
                {manageTask.requires_training && a.status === "requested" && (
                  <p className="mt-2 text-xs text-amber-800">پس از تایید درخواست، داوطلب وارد مرحله آموزش می‌شود. بعد از برگزاری، حضور در آموزش را تایید کنید تا دوره به پرونده داوطلب اضافه شود.</p>
                )}
                {a.status === "training_pending" && (
                  <p className="mt-2 text-xs text-amber-800">در انتظار تایید حضور داوطلب در آموزش. پس از تایید، دوره به فهرست آموزش‌های داوطلب اضافه می‌شود و مرحله انجام فعالیت شروع می‌شود.</p>
                )}
                {(a.delivery_note || a.delivery_file_name) && (
                  <div className="mt-2 text-sm text-stone-600">
                    {a.delivery_note && <p>نتیجه: {a.delivery_note}</p>}
                    {a.delivery_file_name && (
                      <AttachmentButton
                        name={a.delivery_file_name}
                        label="دانلود پیوست نتیجه"
                        onOpen={() => void openAuth(`/api/v1/admin/assignments/${a.id}/delivery`)}
                      />
                    )}
                  </div>
                )}
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  {a.status === "requested" && (
                    <>
                      <Button onClick={async () => {
                        try {
                          await api.approveAssignment(a.id);
                          setMsg(manageTask.requires_training ? "درخواست تایید شد؛ داوطلب در انتظار تایید آموزش است" : "درخواست تایید شد و به داوطلب اطلاع داده شد");
                          await load();
                          await loadApplicants(manageTask.id);
                        } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                      }}>تایید</Button>
                      <Button variant="danger" onClick={async () => {
                        try {
                          await api.rejectAssignment(a.id, notes[a.id] || "");
                          setMsg("درخواست رد شد و به داوطلب اطلاع داده شد");
                          await load();
                          await loadApplicants(manageTask.id);
                        } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                      }}>رد</Button>
                    </>
                  )}
                  {a.status === "training_pending" && (
                    <Button onClick={async () => {
                      try {
                        await api.confirmTraining(a.id);
                        setMsg("حضور در آموزش تایید شد و دوره به پرونده داوطلب اضافه شد");
                        await load();
                        await loadApplicants(manageTask.id);
                      } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                    }}>تایید حضور در آموزش</Button>
                  )}
                  {(a.status === "reserved" || a.status === "in_progress" || a.status === "submitted" || a.status === "attended") && manageTask.work_mode !== "remote" && (
                    <div className="w-full space-y-2">
                      <AttendancePanel assignment={a} onDone={async (ok) => { setMsg(ok); await openManage(manageTask.id); }} />
                      <Button variant="danger" onClick={async () => { await api.markAbsent(a.id); setMsg("عدم حضور ثبت شد"); await openManage(manageTask.id); }}>عدم حضور</Button>
                    </div>
                  )}
                  {manageTask.work_mode === "remote" && a.status === "submitted" && (
                    <Button variant="outline" onClick={async () => {
                      const comment = (notes[a.id] || "").trim();
                      if (!comment) {
                        setMsg("برای درخواست اصلاح، توضیح را در کادر پیام بنویسید");
                        return;
                      }
                      try {
                        await api.requestRevision(a.id, comment);
                        setMsg("درخواست اصلاح برای داوطلب ارسال شد");
                        await loadApplicants(manageTask.id);
                      } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                    }}>درخواست اصلاح / تکمیل</Button>
                  )}
                  {((manageTask.work_mode === "remote" && a.status === "submitted") || (manageTask.work_mode !== "remote" && a.status === "attended")) && (
                    <Button variant="outline" onClick={async () => {
                      try {
                        await api.complete(a.id, { discipline: 5, expertise: 5, ethics: 5, comment: "" });
                        setMsg("تکمیل شد");
                        await loadApplicants(manageTask.id);
                      } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                    }}>تایید نتیجه و تکمیل</Button>
                  )}
                  <input
                    className={inputClass + " max-w-xs"}
                    placeholder={manageTask.work_mode === "remote" && a.status === "submitted" ? "توضیح اصلاح یا پیام به داوطلب" : "پیام به داوطلب"}
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
              ));
            })()}
          </div>
          <div className="mt-4 flex justify-end">
            <Button variant="ghost" onClick={() => setManageId("")}>بستن</Button>
          </div>
        </Modal>
      )}

      {!!seriesId && (
        <Modal open size="lg" title="نوبت‌های روزانه فعالیت جاری" onClose={() => setSeriesId("")}>
          {occurrences.filter((o) => o.status !== "closed" && o.status !== "cancelled").length === 0 && (
            <p className="text-sm text-stone-400">نوبتی ساخته نشده است.</p>
          )}
          <div className="space-y-2">
            {occurrences.filter((o) => o.status !== "closed" && o.status !== "cancelled").map((o) => (
              <div key={o.id} className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-stone-100 px-3 py-2 text-sm">
                <div>
                  <div className="font-medium">{weekdayLabel(o.weekday)} · {fmtDate(o.starts_at)} تا {fmtDate(o.ends_at)}</div>
                  <div className="text-xs text-stone-500">ظرفیت {o.reserved_count}/{o.capacity}</div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge status={o.status} />
                  <Button variant="outline" onClick={() => { setSeriesId(""); void openManage(seriesId); }}>لیست داوطلبان</Button>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-4 flex justify-end">
            <Button variant="ghost" onClick={() => setSeriesId("")}>بستن</Button>
          </div>
        </Modal>
      )}
    </div>
  );
}
