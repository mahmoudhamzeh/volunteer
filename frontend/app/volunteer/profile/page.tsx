"use client";

import { useEffect, useMemo, useState } from "react";
import { api, Availability, DocumentFile, SkillGroup, SkillProposal, Volunteer } from "@/lib/api";
import { EDUCATION_LEVELS, PROPOSAL_LABEL, WEEKDAYS } from "@/lib/labels";
import { IRAN_PROVINCES, citiesOf } from "@/lib/iran";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

const DOC_KINDS = [
  { id: "national_id", label: "کارت ملی" },
  { id: "driving_license", label: "گواهینامه رانندگی" },
  { id: "medical_license", label: "شماره نظام پزشکی" },
  { id: "education", label: "مدرک تحصیلی" },
];

const STEPS = ["اطلاعات فردی", "نشانی", "تحصیلات", "مهارت‌ها", "مدارک و زمان"];

export default function ProfilePage() {
  const [step, setStep] = useState(0);
  const [form, setForm] = useState<Partial<Volunteer>>({});
  const [docs, setDocs] = useState<DocumentFile[]>([]);
  const [slots, setSlots] = useState<Availability[]>([]);
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [proposals, setProposals] = useState<SkillProposal[]>([]);
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [kind, setKind] = useState("national_id");
  const [proposeGroup, setProposeGroup] = useState("");
  const [proposeTitle, setProposeTitle] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api.me().then((r) => {
      if (!r.volunteer) return;
      setForm(r.volunteer);
      setSelected(r.volunteer.skill_ids || r.volunteer.skills?.map((s) => s.skill_id) || []);
      setProposals(r.volunteer.proposals || []);
    });
    api.myDocs().then((x) => setDocs(x || [])).catch(() => undefined);
    api.myAvailability().then((x) => setSlots(x || [])).catch(() => undefined);
    api.skillCatalog().then((x) => setCatalog(x || [])).catch(() => undefined);
  }, []);

  const cities = useMemo(() => {
    const list = citiesOf(form.province || "");
    if (form.city && !list.includes(form.city)) return [form.city, ...list];
    return list;
  }, [form.province, form.city]);

  function toggleSkill(id: string) {
    setSelected((cur) => (cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id]));
  }

  function validateStep(n: number): string {
    if (n === 0) {
      if (!form.full_name?.trim()) return "نام کامل را وارد کنید";
      if (!form.national_id?.trim()) return "کد ملی را وارد کنید";
      if (!form.phone?.trim()) return "شماره موبایل را وارد کنید";
    }
    if (n === 1) {
      if (!form.province) return "استان را انتخاب کنید";
      if (!form.city) return "شهر را انتخاب کنید";
    }
    if (n === 2) {
      if (!form.education_level) return "میزان تحصیلات را انتخاب کنید";
    }
    if (n === 3) {
      if (selected.length === 0 && (proposals || []).every((p) => p.status === "rejected")) {
        return "حداقل یک مهارت انتخاب کنید یا مهارت جدید پیشنهاد دهید";
      }
    }
    return "";
  }

  async function saveDraft() {
    const v = await api.updateProfile({ ...form, skill_ids: selected });
    setForm(v);
    setSelected(v.skill_ids || selected);
    setProposals(v.proposals || proposals);
  }

  async function goNext() {
    const problem = validateStep(step);
    if (problem) {
      setErr(problem);
      setMsg("");
      return;
    }
    setErr("");
    setSaving(true);
    try {
      await saveDraft();
      setMsg("این مرحله ذخیره شد");
      setStep((s) => Math.min(s + 1, STEPS.length - 1));
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در ذخیره");
    } finally {
      setSaving(false);
    }
  }

  async function submit() {
    const problem = STEPS.map((_, i) => validateStep(i)).find(Boolean);
    if (problem) {
      setErr(problem);
      return;
    }
    setErr("");
    setSaving(true);
    try {
      await saveDraft();
      if (slots.length) await api.setAvailability(slots);
      const v = await api.submitProfile();
      setForm(v);
      setMsg("برای بررسی ادمین ارسال شد");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    } finally {
      setSaving(false);
    }
  }

  async function upload(file: File) {
    try {
      await api.uploadDoc(kind, file);
      setDocs((await api.myDocs()) || []);
      setMsg("مدرک بارگذاری شد");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در بارگذاری");
    }
  }

  async function sendProposal() {
    setErr("");
    try {
      if (!proposeGroup || !proposeTitle.trim()) {
        setErr("گروه و عنوان مهارت پیشنهادی را وارد کنید");
        return;
      }
      const p = await api.proposeSkill(proposeGroup, proposeTitle.trim());
      setProposals([p, ...proposals]);
      setProposeTitle("");
      setMsg("مهارت پیشنهادی برای تایید ارسال شد");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در پیشنهاد مهارت");
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-black">ثبت‌نام داوطلب</h1>
        {form.status && <Badge status={form.status} />}
      </div>
      {form.rejection_reason && (
        <Card className="border-rose-200 p-4 text-sm text-rose-800">دلیل ادمین: {form.rejection_reason}</Card>
      )}

      <div className="flex gap-2 overflow-x-auto pb-1">
        {STEPS.map((label, i) => (
          <button
            key={label}
            type="button"
            onClick={() => { setErr(""); setStep(i); }}
            className={`flex min-w-fit items-center gap-2 rounded-full border px-3 py-1.5 text-sm ${
              i === step ? "border-mahak-500 bg-mahak-50 text-mahak-800" : i < step ? "border-emerald-200 bg-emerald-50 text-emerald-800" : "border-stone-200 text-stone-500"
            }`}
          >
            <span className={`grid h-6 w-6 place-items-center rounded-full text-xs font-bold ${i === step ? "bg-mahak-500 text-white" : i < step ? "bg-emerald-500 text-white" : "bg-stone-200"}`}>
              {i + 1}
            </span>
            {label}
          </button>
        ))}
      </div>

      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        {step === 0 && (
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="نام کامل">
              <input className={inputClass} value={form.full_name || ""} onChange={(e) => setForm({ ...form, full_name: e.target.value })} />
            </Field>
            <Field label="کد ملی">
              <input className={inputClass} value={form.national_id || ""} onChange={(e) => setForm({ ...form, national_id: e.target.value })} />
            </Field>
            <Field label="موبایل">
              <input className={inputClass} value={form.phone || ""} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="0912..." />
            </Field>
            <Field label="شماره تماس دوم">
              <input className={inputClass} value={form.phone2 || ""} onChange={(e) => setForm({ ...form, phone2: e.target.value })} />
            </Field>
          </div>
        )}

        {step === 1 && (
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="استان">
              <select className={inputClass} value={form.province || ""} onChange={(e) => setForm({ ...form, province: e.target.value, city: "" })}>
                <option value="">انتخاب استان</option>
                {IRAN_PROVINCES.map((p) => <option key={p.name} value={p.name}>{p.name}</option>)}
              </select>
            </Field>
            <Field label="شهر">
              <select className={inputClass} value={form.city || ""} onChange={(e) => setForm({ ...form, city: e.target.value })} disabled={!form.province}>
                <option value="">انتخاب شهر</option>
                {cities.map((c) => <option key={c} value={c}>{c}</option>)}
              </select>
            </Field>
            <div className="md:col-span-2">
              <Field label="آدرس">
                <input className={inputClass} value={form.address || ""} onChange={(e) => setForm({ ...form, address: e.target.value })} />
              </Field>
            </div>
            <Field label="پلاک">
              <input className={inputClass} value={form.plaque || ""} onChange={(e) => setForm({ ...form, plaque: e.target.value })} />
            </Field>
            <Field label="واحد">
              <input className={inputClass} value={form.unit || ""} onChange={(e) => setForm({ ...form, unit: e.target.value })} />
            </Field>
          </div>
        )}

        {step === 2 && (
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="تحصیلات">
              <select className={inputClass} value={form.education_level || ""} onChange={(e) => setForm({ ...form, education_level: e.target.value })}>
                <option value="">انتخاب کنید</option>
                {EDUCATION_LEVELS.map((x) => <option key={x} value={x}>{x}</option>)}
              </select>
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
          </div>
        )}

        {step === 3 && (
          <div className="space-y-4">
            <p className="text-sm text-stone-500">یک یا چند زیرمهارت را از هر گروه انتخاب کنید.</p>
            {(catalog || []).length === 0 && (
              <p className="text-sm text-rose-600">فهرست مهارت‌ها بارگذاری نشد. بک‌اند را با <code>go run .\cmd\api</code> دوباره اجرا کنید.</p>
            )}
            {(catalog || []).map((g) => (
              <div key={g.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-4">
                <div className="mb-2 font-bold text-mahak-700">{g.title}</div>
                <div className="flex flex-wrap gap-2">
                  {(g.skills || []).filter((s) => s.status !== "inactive").map((s) => (
                    <button
                      type="button"
                      key={s.id}
                      onClick={() => toggleSkill(s.id)}
                      className={`rounded-full border px-3 py-1 text-sm ${selected.includes(s.id) ? "border-mahak-500 bg-mahak-50 text-mahak-800" : "border-stone-200 bg-white"}`}
                    >
                      {s.title}
                    </button>
                  ))}
                </div>
              </div>
            ))}
            <div className="rounded-2xl border border-dashed border-mahak-200 p-4">
              <h3 className="font-bold">پیشنهاد مهارت جدید</h3>
              <p className="mt-1 text-sm text-stone-500">اگر مهارت شما در فهرست نیست پیشنهاد دهید؛ تا تایید ادمین با وضعیت «در انتظار تایید» می‌ماند.</p>
              <div className="mt-3 grid gap-3 md:grid-cols-3">
                <select className={inputClass} value={proposeGroup} onChange={(e) => setProposeGroup(e.target.value)}>
                  <option value="">گروه مهارت</option>
                  {(catalog || []).map((g) => <option key={g.id} value={g.id}>{g.title}</option>)}
                </select>
                <input className={inputClass} placeholder="مثلاً شنا یا نقاشی" value={proposeTitle} onChange={(e) => setProposeTitle(e.target.value)} />
                <Button variant="outline" onClick={sendProposal}>ارسال پیشنهاد</Button>
              </div>
              <ul className="mt-3 space-y-1 text-sm">
                {(proposals || []).map((p) => (
                  <li key={p.id} className="flex flex-wrap items-center gap-2">
                    <span>{p.group_title} / {p.title}</span>
                    <span className={`rounded-full border px-2 py-0.5 text-xs ${p.status === "pending" ? "border-amber-200 bg-amber-50 text-amber-800" : p.status === "approved" ? "border-emerald-200 bg-emerald-50 text-emerald-700" : "border-rose-200 bg-rose-50 text-rose-700"}`}>
                      {PROPOSAL_LABEL[p.status] || p.status}
                    </span>
                    {p.admin_note && <span className="text-stone-500">— {p.admin_note}</span>}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}

        {step === 4 && (
          <div className="space-y-6">
            <div>
              <h2 className="font-bold">مدارک (حداکثر ۵ مگابایت، JPG/PNG/PDF)</h2>
              <div className="mt-3 flex flex-wrap items-center gap-3">
                <select className={inputClass + " w-auto"} value={kind} onChange={(e) => setKind(e.target.value)}>
                  {DOC_KINDS.map((d) => <option key={d.id} value={d.id}>{d.label}</option>)}
                </select>
                <input type="file" onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])} />
              </div>
              <ul className="mt-3 text-sm">
                {(docs || []).map((d) => <li key={d.id}>{d.kind} — {d.file_name}</li>)}
              </ul>
            </div>
            <div>
              <div className="flex items-center justify-between">
                <h2 className="font-bold">تقویم زمانی آزاد</h2>
                <Button variant="ghost" onClick={() => setSlots([...(slots || []), { weekday: 6, start_time: "09:00", end_time: "13:00" }])}>افزودن بازه</Button>
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
            </div>
          </div>
        )}

        <div className="mt-6 flex flex-wrap items-center justify-between gap-2">
          <Button variant="ghost" disabled={step === 0} onClick={() => { setErr(""); setStep((s) => Math.max(0, s - 1)); }}>قبلی</Button>
          <div className="flex flex-wrap gap-2">
            {step < STEPS.length - 1 ? (
              <Button disabled={saving} onClick={goNext}>ذخیره و بعدی</Button>
            ) : (
              <>
                <Button variant="outline" disabled={saving} onClick={async () => {
                  setErr("");
                  try {
                    await saveDraft();
                    if (slots.length) await api.setAvailability(slots);
                    setMsg("پیش‌نویس ذخیره شد");
                  } catch (e) {
                    setErr(e instanceof Error ? e.message : "خطا در ذخیره");
                  }
                }}>ذخیره پیش‌نویس</Button>
                <Button disabled={saving} onClick={submit}>ارسال برای بررسی ادمین</Button>
              </>
            )}
          </div>
        </div>
      </Card>
    </div>
  );
}
