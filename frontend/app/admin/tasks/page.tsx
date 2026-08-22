"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment, SkillGroup, Task, Volunteer, openAuth } from "@/lib/api";
import { fmtDate, skillLabel, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, inputClass } from "@/components/ui";
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

const emptyForm = () => ({
  title: "", description: "", location: "", ...defaultTaskTimes(),
  capacity: 5, hour_weight: 4, min_score: 0, required_education: "",
  work_mode: "onsite",
  delivery_hint: "",
  required_skills: [] as string[],
  required_skill_ids: [] as string[],
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

  async function load() {
    const r = await api.adminTasks();
    setItems(r.items || []);
    try {
      const a = await api.adminAssignments("?limit=200");
      const byTask: Record<string, Assignment[]> = {};
      for (const x of a.items || []) {
        (byTask[x.task_id] ||= []).push(x);
      }
      setApplicants(byTask);
    } catch {
      /* list still usable */
    }
  }

  async function loadApplicants(taskId: string) {
    const r = await api.adminAssignments(`?task_id=${taskId}&limit=100`);
    setApplicants((prev) => ({ ...prev, [taskId]: r.items || [] }));
  }
  useEffect(() => {
    load();
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
    api.adminVolunteers("?status=approved&limit=100").then((r) => setVolunteers(r.items || [])).catch(() => undefined);
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
      required_skills: t.required_skills || [],
      required_skill_ids: t.required_skill_ids || [],
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
    if (!(end.getTime() > start.getTime())) {
      setMsg("تاریخ پایان باید بعد از تاریخ شروع باشد");
      return;
    }
    if (!Number.isFinite(form.capacity) || form.capacity < 1) {
      setMsg("ظرفیت باید حداقل ۱ نفر باشد");
      return;
    }
    if (!Number.isFinite(form.hour_weight) || form.hour_weight <= 0) {
      setMsg("وزن ساعتی باید بزرگ‌تر از صفر باشد");
      return;
    }
    const body = { ...form, status: editingId ? items.find((x) => x.id === editingId)?.status : "open" };
    try {
      if (editingId) await api.updateTask(editingId, body);
      else await api.createTask(body);
      setMsg(editingId ? "فعالیت ویرایش شد" : "فعالیت ایجاد شد");
      resetForm();
      await load();
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
      const hay = `${v.full_name} ${v.city || ""} ${v.phone || ""} ${v.national_id || ""}`;
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
    try {
      await api.assignVolunteer(taskId, vid);
      setMsg("داوطلب به فعالیت تخصیص داده شد");
      setPick({ ...pick, [taskId]: "" });
      await load();
      await loadApplicants(taskId);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  async function openManage(id: string) {
    setManageId(id);
    await loadApplicants(id);
  }

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
                  <option value="onsite">حضوری — نیاز به حضور در محل</option>
                  <option value="remote">دورکار — داوطلب نتیجه را در پنل ارسال می‌کند</option>
                </select>
              </Field>
              <Field label="راهنمای تحویل نتیجه (اختیاری)">
                <input className={inputClass} placeholder="مثلا فایل پوستر، گزارش تست، یا لینک" value={form.delivery_hint} onChange={(e) => setForm({ ...form, delivery_hint: e.target.value })} />
              </Field>
              <ShamsiDateTimeField label="شروع (شمسی)" value={form.starts_at} onChange={(starts_at) => setForm({ ...form, starts_at })} />
              <ShamsiDateTimeField label="پایان (شمسی)" value={form.ends_at} onChange={(ends_at) => setForm({ ...form, ends_at })} />
              <Field label="ظرفیت"><input type="number" className={inputClass} value={form.capacity} onChange={(e) => setForm({ ...form, capacity: Number(e.target.value) })} /></Field>
              <Field label="وزن ساعتی"><input type="number" step="0.5" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} /></Field>
              <Field label="حداقل امتیاز"><input type="number" step="0.1" className={inputClass} value={form.min_score} onChange={(e) => setForm({ ...form, min_score: Number(e.target.value) })} /></Field>
              <Field label="رشته تحصیلی الزامی"><input className={inputClass} value={form.required_education} onChange={(e) => setForm({ ...form, required_education: e.target.value })} /></Field>
            </section>

            <section className="space-y-3 rounded-2xl border border-stone-100 bg-stone-50/50 p-4">
              <h3 className="text-sm font-bold text-stone-700">مهارت مورد نیاز</h3>
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
          const pending = apps.filter((a) => a.status === "requested").length;
          return (
            <Card key={t.id} className="p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="font-bold">{t.title}</div>
                  <div className="text-xs text-stone-500">
                    {workModeLabel(t.work_mode)} · {t.location || (t.work_mode === "remote" ? "دورکار" : "—")} · {fmtDate(t.starts_at)} تا {fmtDate(t.ends_at)} · تاییدشده {t.reserved_count}/{t.capacity} · {t.hour_weight} ساعت
                  </div>
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
          {msg && <p className="mt-2 text-sm text-mahak-700">{msg}</p>}
          {manageTask.status === "open" && (
            <div className="mt-4 grid gap-2 rounded-2xl border border-stone-100 bg-stone-50/70 p-3 md:grid-cols-[1fr_1fr_auto]">
              <Field label="جستجوی داوطلب">
                <input
                  className={inputClass}
                  placeholder="نام، شهر یا موبایل"
                  value={volQuery}
                  onChange={(e) => setVolQuery(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      const qs = new URLSearchParams({ status: "approved", limit: "100" });
                      if (volQuery.trim()) qs.set("q", volQuery.trim());
                      api.adminVolunteers(`?${qs.toString()}`).then((r) => setVolunteers(r.items || [])).catch(() => undefined);
                    }
                  }}
                />
              </Field>
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
          )}
          <div className="mt-4 space-y-3">
            <h3 className="font-bold">درخواست‌های داوطلبان</h3>
            {(applicants[manageTask.id] || []).length === 0 && <p className="text-sm text-stone-400">هنوز درخواستی ثبت نشده</p>}
            {(applicants[manageTask.id] || []).map((a) => (
              <div key={a.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <Link className="font-medium text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                    {a.volunteer?.full_name || "داوطلب"}
                  </Link>
                  <Badge status={a.status} />
                </div>
                {(a.delivery_note || a.delivery_file_name) && (
                  <div className="mt-2 text-sm text-stone-600">
                    {a.delivery_note && <p>نتیجه: {a.delivery_note}</p>}
                    {a.delivery_file_name && (
                      <button className="text-mahak-700" onClick={() => openAuth(`/api/v1/admin/assignments/${a.id}/delivery`)}>
                        فایل: {a.delivery_file_name}
                      </button>
                    )}
                  </div>
                )}
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  {a.status === "requested" && (
                    <>
                      <Button onClick={async () => {
                        try {
                          await api.approveAssignment(a.id);
                          setMsg("درخواست تایید شد و به داوطلب اطلاع داده شد");
                          await load();
                          await loadApplicants(manageTask.id);
                        } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                      }}>تایید</Button>
                      <Button variant="danger" onClick={async () => {
                        try {
                          await api.rejectAssignment(a.id);
                          setMsg("درخواست رد شد و به داوطلب اطلاع داده شد");
                          await load();
                          await loadApplicants(manageTask.id);
                        } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                      }}>رد</Button>
                    </>
                  )}
                  {(a.status === "reserved" || a.status === "in_progress") && manageTask.work_mode !== "remote" && (
                    <Button onClick={async () => { await api.attendance(a.id); await loadApplicants(manageTask.id); }}>تایید حضور</Button>
                  )}
                  {(a.status === "submitted" || a.status === "attended" || (manageTask.work_mode !== "remote" && (a.status === "in_progress" || a.status === "reserved"))) && (
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
          <div className="mt-4 flex justify-end">
            <Button variant="ghost" onClick={() => setManageId("")}>بستن</Button>
          </div>
        </Modal>
      )}
    </div>
  );
}
