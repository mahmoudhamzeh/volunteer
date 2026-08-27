"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment, TrainingCourse } from "@/lib/api";
import { TRAINING_KINDS, fmtDate, trainingCourseTitle, trainingKindLabel } from "@/lib/labels";
import { Button, Card, Field, inputClass } from "@/components/ui";

const emptyForm = () => ({
  title: "",
  kind: "in_person",
  location: "",
  description: "",
  status: "active",
});

export default function AdminTrainingsPage() {
  const [tab, setTab] = useState<"pending" | "catalog">("pending");
  const [courses, setCourses] = useState<TrainingCourse[]>([]);
  const [pending, setPending] = useState<Assignment[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [q, setQ] = useState("");

  async function load() {
    const [c, p] = await Promise.all([
      api.trainingCourses().catch(() => ({ items: [] as TrainingCourse[] })),
      api.adminAssignments("?status=training_pending&limit=200").catch(() => ({ items: [] as Assignment[] })),
    ]);
    setCourses(c.items || []);
    setPending(p.items || []);
  }

  useEffect(() => {
    load().catch((e) => setErr(e instanceof Error ? e.message : "خطا"));
  }, []);

  function resetForm() {
    setEditingId("");
    setForm(emptyForm());
  }

  function startEdit(c: TrainingCourse) {
    setEditingId(c.id);
    setForm({
      title: c.title,
      kind: c.kind || "in_person",
      location: c.location || "",
      description: c.description || "",
      status: c.status || "active",
    });
    setTab("catalog");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setErr("");
    if (!form.title.trim()) {
      setErr("عنوان دوره را وارد کنید");
      return;
    }
    if (!form.location.trim()) {
      setErr("محل برگزاری را وارد کنید");
      return;
    }
    try {
      const body = {
        title: form.title.trim(),
        kind: form.kind,
        location: form.location.trim(),
        description: form.description.trim(),
        status: form.status,
      };
      if (editingId) await api.updateTrainingCourse(editingId, body);
      else await api.createTrainingCourse(body);
      setMsg(editingId ? "دوره ویرایش شد" : "دوره آموزشی ثبت شد");
      resetForm();
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    }
  }

  async function confirm(id: string) {
    setErr("");
    try {
      await api.confirmTraining(id);
      setMsg("حضور در آموزش تایید شد و دوره به پرونده داوطلب اضافه شد");
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    }
  }

  const filtered = useMemo(() => {
    const needle = q.trim();
    if (!needle) return courses;
    return courses.filter((c) => `${c.title} ${c.location}`.includes(needle));
  }, [courses, q]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">آموزش</h1>
        <p className="mt-1 text-sm text-stone-500">
          دوره‌ها را اینجا با نام تعریف کنید؛ زمان برگزاری جلسه هنگام تعریف هر فعالیت مشخص می‌شود. پس از تایید اولیه درخواست، داوطلب در این صف می‌ماند تا حضور در همان جلسه تایید شود. تا تایید آموزش، امکان ادامه فرایند فعالیت نیست.
        </p>
      </div>
      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <div className="flex flex-wrap gap-2">
        <Button variant={tab === "pending" ? "primary" : "outline"} onClick={() => setTab("pending")}>
          تایید آموزش ({pending.length})
        </Button>
        <Button variant={tab === "catalog" ? "primary" : "outline"} onClick={() => setTab("catalog")}>
          فهرست دوره‌ها ({courses.length})
        </Button>
      </div>

      {tab === "pending" && (
        <Card className="overflow-hidden">
          {pending.length === 0 && (
            <p className="px-4 py-6 text-sm text-stone-400">درخواستی در انتظار تایید آموزش نیست.</p>
          )}
          {pending.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[720px] text-sm">
                <thead className="bg-stone-50 text-right text-xs text-stone-500">
                  <tr>
                    <th className="px-3 py-2">داوطلب</th>
                    <th className="px-3 py-2">فعالیت</th>
                    <th className="px-3 py-2">دوره آموزشی</th>
                    <th className="px-3 py-2">زمان آموزش</th>
                    <th className="px-3 py-2">اقدام</th>
                  </tr>
                </thead>
                <tbody>
                  {pending.map((a) => (
                    <tr key={a.id} className="border-t border-stone-100">
                      <td className="px-3 py-2">
                        <Link className="text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                          {a.volunteer?.full_name || "داوطلب"}
                        </Link>
                        <div className="text-xs text-stone-400">{a.volunteer?.phone}</div>
                      </td>
                      <td className="px-3 py-2">
                        <div>{a.task?.title}</div>
                        {a.task?.starts_at && <div className="text-xs text-stone-400">{fmtDate(a.task.starts_at)}</div>}
                      </td>
                      <td className="px-3 py-2">
                        <div className="font-medium">{trainingCourseTitle(a.task)}</div>
                        <div className="text-xs text-stone-500">
                          {trainingKindLabel(a.task?.training_kind)} · {a.task?.training_location || "—"}
                        </div>
                      </td>
                      <td className="px-3 py-2 text-stone-500">{fmtDate(a.task?.training_at)}</td>
                      <td className="px-3 py-2">
                        <div className="flex flex-wrap gap-2">
                          <Button onClick={() => void confirm(a.id)}>تایید آموزش</Button>
                          <Link className="rounded-2xl border border-mahak-200 px-3 py-2 text-sm text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>
                            پرونده
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {tab === "catalog" && (
        <>
          <Card className="p-5">
            <h2 className="mb-3 font-bold">{editingId ? "ویرایش دوره" : "تعریف دوره آموزشی"}</h2>
            <p className="mb-3 text-xs text-stone-500">
              دوره قابل استفاده مجدد است؛ مثلاً «مددکاری» چند بار در سال برگزار می‌شود. زمان هر جلسه را هنگام تعریف فعالیت وارد کنید.
            </p>
            <form className="grid gap-3 md:grid-cols-2" onSubmit={onSubmit}>
              <Field label="نام دوره">
                <input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="مثلاً آموزش مددکاری" />
              </Field>
              <Field label="نوع">
                <select className={inputClass} value={form.kind} onChange={(e) => setForm({ ...form, kind: e.target.value })}>
                  {TRAINING_KINDS.map((k) => (
                    <option key={k.id} value={k.id}>{k.label}</option>
                  ))}
                </select>
              </Field>
              <Field label="محل برگزاری">
                <input className={inputClass} value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} placeholder="محک، لینک کلاس آنلاین، یا آدرس" />
              </Field>
              <Field label="وضعیت">
                <select className={inputClass} value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                  <option value="active">فعال</option>
                  <option value="inactive">غیرفعال</option>
                </select>
              </Field>
              <div className="md:col-span-2">
                <Field label="توضیح (اختیاری)">
                  <textarea className={inputClass} rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
                </Field>
              </div>
              <div className="flex flex-wrap gap-2 md:col-span-2">
                <Button type="submit">{editingId ? "ذخیره تغییرات" : "ثبت دوره"}</Button>
                {editingId && <Button type="button" variant="ghost" onClick={resetForm}>انصراف</Button>}
              </div>
            </form>
          </Card>
          <Card className="p-5">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h2 className="font-bold">دوره‌های تعریف‌شده</h2>
              <input className={inputClass + " max-w-xs"} placeholder="جستجوی نام یا محل" value={q} onChange={(e) => setQ(e.target.value)} />
            </div>
            {filtered.length === 0 && <p className="text-sm text-stone-400">دوره‌ای ثبت نشده است.</p>}
            <ul className="space-y-2">
              {filtered.map((c) => (
                <li key={c.id} className="flex flex-wrap items-start justify-between gap-2 rounded-2xl border border-stone-100 px-3 py-2">
                  <div>
                    <div className="font-medium">{c.title}</div>
                    <div className="text-xs text-stone-500">
                      {trainingKindLabel(c.kind)} · {c.location || "—"}
                    </div>
                    {c.description ? <p className="mt-1 text-xs text-stone-500">{c.description}</p> : null}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`rounded-full border px-2 py-0.5 text-xs ${c.status === "inactive" ? "border-stone-200 bg-stone-50 text-stone-600" : "border-emerald-200 bg-emerald-50 text-emerald-800"}`}>
                      {c.status === "inactive" ? "غیرفعال" : "فعال"}
                    </span>
                    <Button variant="outline" onClick={() => startEdit(c)}>ویرایش</Button>
                  </div>
                </li>
              ))}
            </ul>
          </Card>
        </>
      )}
    </div>
  );
}
