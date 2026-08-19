"use client";

import { FormEvent, useEffect, useState } from "react";
import { api, Task } from "@/lib/api";
import { SKILLS, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminTasks() {
  const [items, setItems] = useState<Task[]>([]);
  const [form, setForm] = useState({
    title: "", description: "", location: "", starts_at: "", ends_at: "",
    capacity: 5, hour_weight: 4, min_score: 0, required_education: "", required_skills: [] as string[],
  });

  async function load() {
    const r = await api.adminTasks();
    setItems(r.items || []);
  }
  useEffect(() => { load(); }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    await api.createTask({
      ...form,
      starts_at: new Date(form.starts_at).toISOString(),
      ends_at: new Date(form.ends_at).toISOString(),
    });
    setForm({ ...form, title: "", description: "" });
    await load();
  }

  function toggle(id: string) {
    setForm({
      ...form,
      required_skills: form.required_skills.includes(id)
        ? form.required_skills.filter((x) => x !== id)
        : [...form.required_skills, id],
    });
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">تعریف تسک عملیاتی</h1>
      <Card className="p-5">
        <form onSubmit={onSubmit} className="grid gap-3 md:grid-cols-2">
          <Field label="عنوان"><input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></Field>
          <Field label="مکان"><input className={inputClass} value={form.location} onChange={(e) => setForm({ ...form, location: e.target.value })} /></Field>
          <div className="md:col-span-2">
            <Field label="شرح کار"><textarea className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} required /></Field>
          </div>
          <Field label="شروع"><input type="datetime-local" className={inputClass} value={form.starts_at} onChange={(e) => setForm({ ...form, starts_at: e.target.value })} required /></Field>
          <Field label="پایان"><input type="datetime-local" className={inputClass} value={form.ends_at} onChange={(e) => setForm({ ...form, ends_at: e.target.value })} required /></Field>
          <Field label="ظرفیت"><input type="number" className={inputClass} value={form.capacity} onChange={(e) => setForm({ ...form, capacity: Number(e.target.value) })} /></Field>
          <Field label="وزن ساعتی"><input type="number" step="0.5" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} /></Field>
          <Field label="حداقل امتیاز"><input type="number" step="0.1" className={inputClass} value={form.min_score} onChange={(e) => setForm({ ...form, min_score: Number(e.target.value) })} /></Field>
          <Field label="رشته تحصیلی الزامی"><input className={inputClass} value={form.required_education} onChange={(e) => setForm({ ...form, required_education: e.target.value })} /></Field>
          <div className="md:col-span-2 flex flex-wrap gap-2">
            {SKILLS.map((s) => (
              <button type="button" key={s.id} onClick={() => toggle(s.id)}
                className={`rounded-full border px-3 py-1 text-xs ${form.required_skills.includes(s.id) ? "border-mahak-400 bg-mahak-50" : "border-stone-200"}`}>
                {s.label}
              </button>
            ))}
          </div>
          <Button type="submit">ایجاد تسک</Button>
        </form>
      </Card>
      {items.map((t) => (
        <Card key={t.id} className="p-4">
          <div className="flex items-start justify-between">
            <div>
              <div className="font-bold">{t.title}</div>
              <div className="text-xs text-stone-500">{t.location} · {fmtDate(t.starts_at)} · {t.reserved_count}/{t.capacity} · {t.hour_weight} ساعت</div>
            </div>
            <div className="flex items-center gap-2">
              <Badge status={t.status} />
              {t.status === "open" && (
                <button className="text-xs text-stone-500" onClick={async () => {
                  await api.updateTask(t.id, { ...t, status: "closed" });
                  await load();
                }}>بستن</button>
              )}
              <button className="text-xs text-rose-600" onClick={async () => {
                await api.deleteTask(t.id);
                await load();
              }}>حذف</button>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
