"use client";

import { FormEvent, useEffect, useState } from "react";
import { api, Availability, DocumentFile, Volunteer } from "@/lib/api";
import { SKILLS, WEEKDAYS } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

const DOC_KINDS = [
  { id: "national_id", label: "کارت ملی" },
  { id: "driving_license", label: "گواهینامه رانندگی" },
  { id: "medical_license", label: "شماره نظام پزشکی" },
  { id: "education", label: "مدرک تحصیلی" },
];

export default function ProfilePage() {
  const [form, setForm] = useState<Partial<Volunteer>>({});
  const [docs, setDocs] = useState<DocumentFile[]>([]);
  const [slots, setSlots] = useState<Availability[]>([]);
  const [msg, setMsg] = useState("");
  const [kind, setKind] = useState("national_id");

  useEffect(() => {
    api.me().then((r) => r.volunteer && setForm(r.volunteer));
    api.myDocs().then((x) => setDocs(x || [])).catch(() => undefined);
    api.myAvailability().then((x) => setSlots(x || [])).catch(() => undefined);
  }, []);

  function toggleSkill(id: string) {
    const cur = form.skill_categories || [];
    setForm({ ...form, skill_categories: cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id] });
  }

  async function save(e: FormEvent) {
    e.preventDefault();
    const v = await api.updateProfile(form);
    setForm(v);
    setMsg("پروفایل ذخیره شد");
  }

  async function submit() {
    try {
      const v = await api.submitProfile();
      setForm(v);
      setMsg("برای بررسی ادمین ارسال شد");
    } catch (err) {
      setMsg(err instanceof Error ? err.message : "خطا");
    }
  }

  async function upload(file: File) {
    await api.uploadDoc(kind, file);
    setDocs((await api.myDocs()) || []);
  }

  function addSlot() {
    setSlots([...(slots || []), { weekday: 6, start_time: "09:00", end_time: "13:00" }]);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-black">پروفایل تخصصی</h1>
        {form.status && <Badge status={form.status} />}
      </div>
      {form.rejection_reason && (
        <Card className="border-rose-200 p-4 text-sm text-rose-800">دلیل ادمین: {form.rejection_reason}</Card>
      )}
      <form onSubmit={save} className="grid gap-4 md:grid-cols-2">
        <Field label="نام کامل">
          <input className={inputClass} value={form.full_name || ""} onChange={(e) => setForm({ ...form, full_name: e.target.value })} />
        </Field>
        <Field label="کد ملی">
          <input className={inputClass} value={form.national_id || ""} onChange={(e) => setForm({ ...form, national_id: e.target.value })} />
        </Field>
        <Field label="موبایل">
          <input className={inputClass} value={form.phone || ""} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
        </Field>
        <Field label="شهر">
          <input className={inputClass} value={form.city || ""} onChange={(e) => setForm({ ...form, city: e.target.value })} />
        </Field>
        <Field label="رشته تحصیلی">
          <input className={inputClass} value={form.education_field || ""} onChange={(e) => setForm({ ...form, education_field: e.target.value })} />
        </Field>
        <Field label="شماره نظام پزشکی (در صورت وجود)">
          <input className={inputClass} value={form.medical_license || ""} onChange={(e) => setForm({ ...form, medical_license: e.target.value })} />
        </Field>
        <div className="md:col-span-2">
          <Field label="درباره توانمندی‌ها">
            <textarea className={inputClass} rows={3} value={form.bio || ""} onChange={(e) => setForm({ ...form, bio: e.target.value })} />
          </Field>
        </div>
        <div className="md:col-span-2">
          <div className="mb-2 text-sm text-stone-600">دسته‌بندی توانمندی‌ها</div>
          <div className="flex flex-wrap gap-2">
            {SKILLS.map((s) => (
              <button
                type="button"
                key={s.id}
                onClick={() => toggleSkill(s.id)}
                className={`rounded-full border px-3 py-1 text-sm ${(form.skill_categories || []).includes(s.id) ? "border-mahak-400 bg-mahak-50 text-mahak-800" : "border-stone-200"}`}
              >
                {s.label}
              </button>
            ))}
          </div>
        </div>
        <div className="flex gap-2 md:col-span-2">
          <Button type="submit">ذخیره پیش‌نویس</Button>
          <Button variant="outline" onClick={submit}>ارسال برای بررسی ادمین</Button>
        </div>
      </form>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        <h2 className="font-bold">مدارک (حداکثر ۵ مگابایت، JPG/PNG/PDF)</h2>
        <div className="mt-3 flex flex-wrap items-center gap-3">
          <select className={inputClass + " w-auto"} value={kind} onChange={(e) => setKind(e.target.value)}>
            {DOC_KINDS.map((d) => (
              <option key={d.id} value={d.id}>{d.label}</option>
            ))}
          </select>
          <input type="file" onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])} />
        </div>
        <ul className="mt-3 text-sm">
          {(docs || []).map((d) => (
            <li key={d.id}>{d.kind} — {d.file_name}</li>
          ))}
        </ul>
      </Card>

      <Card className="p-5">
        <div className="flex items-center justify-between">
          <h2 className="font-bold">تقویم زمانی آزاد</h2>
          <Button variant="ghost" onClick={addSlot}>افزودن بازه</Button>
        </div>
        <div className="mt-3 space-y-2">
          {(slots || []).map((s, i) => (
            <div key={i} className="grid grid-cols-3 gap-2">
              <select className={inputClass} value={s.weekday} onChange={(e) => {
                const n = [...slots]; n[i] = { ...s, weekday: Number(e.target.value) }; setSlots(n);
              }}>
                {WEEKDAYS.map((w, idx) => <option key={w} value={idx}>{w}</option>)}
              </select>
              <input className={inputClass} value={s.start_time} onChange={(e) => {
                const n = [...slots]; n[i] = { ...s, start_time: e.target.value }; setSlots(n);
              }} />
              <input className={inputClass} value={s.end_time} onChange={(e) => {
                const n = [...slots]; n[i] = { ...s, end_time: e.target.value }; setSlots(n);
              }} />
            </div>
          ))}
        </div>
        <div className="mt-3">
          <Button onClick={() => api.setAvailability(slots).then(() => setMsg("زمان‌بندی ذخیره شد"))}>ذخیره زمان‌بندی</Button>
        </div>
      </Card>
    </div>
  );
}
