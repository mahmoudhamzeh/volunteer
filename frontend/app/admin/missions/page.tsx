"use client";

import { FormEvent, useEffect, useState } from "react";
import { api, Mission } from "@/lib/api";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminMissions() {
  const [items, setItems] = useState<Mission[]>([]);
  const [form, setForm] = useState({ title: "", description: "", kind: "custom", hour_weight: 1, target_count: 1, deadline_hours: 72, webhook_event: "" });
  async function load() { setItems(await api.adminMissions()); }
  useEffect(() => { load(); }, []);
  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    await api.createMission(form);
    await load();
  }
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">ماموریت‌های سیستمی</h1>
      <Card className="p-5">
        <form onSubmit={onSubmit} className="grid gap-3 md:grid-cols-2">
          <Field label="عنوان"><input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></Field>
          <Field label="نوع">
            <select className={inputClass} value={form.kind} onChange={(e) => setForm({ ...form, kind: e.target.value })}>
              <option value="complete_profile">تکمیل پروفایل</option>
              <option value="invite_users">دعوت کاربر</option>
              <option value="custom">سفارشی</option>
              <option value="webhook">رویداد وب‌هوک</option>
            </select>
          </Field>
          <div className="md:col-span-2"><Field label="شرح"><textarea className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></Field></div>
          <Field label="وزن ساعتی"><input type="number" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} /></Field>
          <Field label="تعداد هدف"><input type="number" className={inputClass} value={form.target_count} onChange={(e) => setForm({ ...form, target_count: Number(e.target.value) })} /></Field>
          <Field label="مهلت (ساعت)"><input type="number" className={inputClass} value={form.deadline_hours} onChange={(e) => setForm({ ...form, deadline_hours: Number(e.target.value) })} /></Field>
          <Field label="رویداد وب‌هوک"><input className={inputClass} value={form.webhook_event} onChange={(e) => setForm({ ...form, webhook_event: e.target.value })} /></Field>
          <Button type="submit">ایجاد ماموریت</Button>
        </form>
      </Card>
      {items.map((m) => (
        <Card key={m.id} className="p-4 flex items-center justify-between">
          <div>
            <div className="font-bold">{m.title}</div>
            <div className="text-xs text-stone-500">{m.kind} · {m.hour_weight} ساعت · هدف {m.target_count}</div>
          </div>
          <Badge status={m.status} />
        </Card>
      ))}
    </div>
  );
}
