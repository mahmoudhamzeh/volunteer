"use client";

import { Suspense, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { api, Availability, DocumentFile, SkillGroup, SkillProposal, Volunteer } from "@/lib/api";
import { EDUCATION_LEVELS, GENDERS, OCCUPATIONS, PROPOSAL_LABEL, STATUS_EXPLAIN, STATUS_LABEL, WEEKDAYS, DOC_KINDS, docKindLabel } from "@/lib/labels";
import { IRAN_PROVINCES, citiesOf } from "@/lib/iran";
import { Badge, Button, Card, Field, Modal, inputClass } from "@/components/ui";
import { HistoryList } from "@/components/history";
import { ShamsiDateField } from "@/components/shamsi";
import { currentJalaliYear } from "@/lib/jalali";
import { isNationalID, isPersianName, MIN_VOLUNTEER_AGE, needsVolunteerRegistration, onlyDigits, onlyPersianLetters, volunteerBirthDateError } from "@/lib/persian";

const STEPS = ["اطلاعات فردی", "نشانی", "تحصیلات", "مهارت‌ها", "مدارک و زمان آزاد"];
const TABS = ["هویت", "نشانی", "تحصیلات", "مهارت‌ها", "مدارک و زمان آزاد"];

function namesOf(v: Partial<Volunteer>) {
  if (v.first_name || v.last_name) return { first: v.first_name || "", last: v.last_name || "" };
  const parts = (v.full_name || "").trim().split(/\s+/).filter(Boolean);
  return { first: parts[0] || "", last: parts.slice(1).join(" ") };
}

function ProfilePage() {
  const searchParams = useSearchParams();
  const [step, setStep] = useState(0);
  const [form, setForm] = useState<Partial<Volunteer>>({});
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
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
  const [skillQuery, setSkillQuery] = useState("");
  const [openGroup, setOpenGroup] = useState("");
  const [draftOpen, setDraftOpen] = useState(false);
  const [submitOpen, setSubmitOpen] = useState(false);

  const wizard = needsVolunteerRegistration(form.status);
  const identityLocked = !wizard;
  const canDeleteDoc = form.status !== "approved" && form.status !== "suspended";

  useEffect(() => {
    api.me().then((r) => {
      if (!r.volunteer) return;
      setForm(r.volunteer);
      const n = namesOf(r.volunteer);
      setFirstName(n.first);
      setLastName(n.last);
      setSelected(r.volunteer.skill_ids || r.volunteer.skills?.map((s) => s.skill_id) || []);
      setProposals(r.volunteer.proposals || []);
    });
    api.myDocs().then((x) => setDocs(x || [])).catch(() => undefined);
    api.myAvailability().then((x) => setSlots(x || [])).catch(() => undefined);
    api.skillCatalog().then((x) => {
      setCatalog(x || []);
      if (x?.[0]?.id) setOpenGroup(x[0].id);
    }).catch(() => undefined);
  }, []);

  useEffect(() => {
    const tab = searchParams.get("tab");
    if (tab === "docs") {
      setStep(TABS.length - 1);
      window.setTimeout(() => document.getElementById("docs-upload")?.scrollIntoView({ behavior: "smooth", block: "start" }), 50);
    } else if (tab === "skills") {
      setStep(3);
    }
  }, [searchParams]);

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
      if (!isPersianName(firstName)) return "نام را فقط با حروف فارسی وارد کنید";
      if (!isPersianName(lastName)) return "نام خانوادگی را فقط با حروف فارسی وارد کنید";
      if (!isNationalID(form.national_id || "")) return "کد ملی باید دقیقاً ۱۰ رقم باشد";
      if (!form.phone?.trim()) return "شماره موبایل مشخص نیست";
      const birthErr = volunteerBirthDateError(form.birth_date);
      if (birthErr) return birthErr;
      if (!form.gender) return "جنسیت را انتخاب کنید";
      if (!form.occupation) return "شغل را انتخاب کنید";
      if (form.occupation === "other" && !form.occupation_other?.trim()) return "در صورت انتخاب «سایر»، شغل خود را بنویسید";
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

  function payload() {
    return {
      ...form,
      first_name: firstName,
      last_name: lastName,
      full_name: `${firstName} ${lastName}`.trim(),
      skill_ids: selected,
    };
  }

  async function saveDraft() {
    const v = await api.updateProfile(payload());
    setForm(v);
    setSelected(v.skill_ids || selected);
    setProposals(v.proposals || proposals);
    const n = namesOf(v);
    setFirstName(n.first || firstName);
    setLastName(n.last || lastName);
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

  async function saveCurrent() {
    setErr("");
    setSaving(true);
    try {
      if (wizard) {
        const problem = validateStep(step);
        if (problem) {
          setErr(problem);
          return;
        }
      }
      await saveDraft();
      if (slots.length) await api.setAvailability(slots);
      setMsg("");
      if (wizard) {
        setDraftOpen(true);
      } else {
        setMsg("ذخیره شد");
      }
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
      setMsg("");
      setSubmitOpen(true);
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

  const labels = wizard ? STEPS : TABS;
  const identity = (
    <div className="grid gap-4 md:grid-cols-2">
      <Field label="نام">
        <input
          className={inputClass}
          value={firstName}
          disabled={identityLocked}
          onChange={(e) => setFirstName(onlyPersianLetters(e.target.value))}
          placeholder="فقط حروف فارسی"
        />
      </Field>
      <Field label="نام خانوادگی">
        <input
          className={inputClass}
          value={lastName}
          disabled={identityLocked}
          onChange={(e) => setLastName(onlyPersianLetters(e.target.value))}
          placeholder="فقط حروف فارسی"
        />
      </Field>
      <Field label="کد ملی">
        <input
          className={inputClass}
          dir="ltr"
          inputMode="numeric"
          maxLength={10}
          value={form.national_id || ""}
          disabled={identityLocked}
          onChange={(e) => setForm({ ...form, national_id: onlyDigits(e.target.value, 10) })}
          placeholder="۱۰ رقم"
        />
      </Field>
      <Field label="موبایل">
        <input className={inputClass + " bg-stone-50"} dir="ltr" value={form.phone || ""} disabled />
        <p className="mt-1 text-xs text-stone-400">همین شماره‌ای است که با آن وارد شده‌اید و قابل تغییر نیست.</p>
      </Field>
      <Field label="شماره تماس دوم">
        <input className={inputClass} dir="ltr" value={form.phone2 || ""} onChange={(e) => setForm({ ...form, phone2: onlyDigits(e.target.value, 11) })} />
      </Field>
      <div>
        <ShamsiDateField
          className="max-w-[16rem]"
          label="تاریخ تولد"
          value={form.birth_date}
          disabled={identityLocked}
          maxYear={currentJalaliYear() - MIN_VOLUNTEER_AGE}
          onChange={(birth_date) => setForm({ ...form, birth_date })}
        />
        <p className="mt-1 text-xs text-stone-500">حداقل سن داوطلبی ۱۸ سال تمام است.</p>
      </div>
      <Field label="جنسیت">
        <select
          className={inputClass}
          value={form.gender || ""}
          disabled={identityLocked}
          onChange={(e) => setForm({ ...form, gender: e.target.value })}
        >
          <option value="">انتخاب کنید</option>
          {GENDERS.map((g) => <option key={g.id} value={g.id}>{g.label}</option>)}
        </select>
      </Field>
      <Field label="شغل">
        <select
          className={inputClass}
          value={form.occupation || ""}
          disabled={identityLocked}
          onChange={(e) => setForm({
            ...form,
            occupation: e.target.value,
            occupation_other: e.target.value === "other" ? (form.occupation_other || "") : "",
          })}
        >
          <option value="">انتخاب کنید</option>
          {OCCUPATIONS.map((o) => <option key={o.id} value={o.id}>{o.label}</option>)}
        </select>
      </Field>
      {form.occupation === "other" && (
        <Field label="شرح شغل">
          <input
            className={inputClass}
            value={form.occupation_other || ""}
            disabled={identityLocked}
            maxLength={80}
            onChange={(e) => setForm({ ...form, occupation_other: e.target.value })}
            placeholder="شغل خود را بنویسید"
          />
        </Field>
      )}
      {identityLocked && (
        <p className="md:col-span-2 text-sm text-stone-500">اطلاعات هویتی پس از ثبت‌نام فقط توسط ادمین قابل تغییر است.</p>
      )}
    </div>
  );

  const address = (
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
  );

  const education = (
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
  );

  const q = skillQuery.trim();
  const skills = (
    <div className="space-y-4">
      <p className="text-sm text-stone-500">گروه را باز کنید و زیرمهارت‌های خود را انتخاب کنید. مهارت‌های انتخاب‌شده بالا نمایش داده می‌شوند.</p>
      {selected.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {catalog.flatMap((g) => (g.skills || []).filter((s) => selected.includes(s.id)).map((s) => (
            <button key={s.id} type="button" onClick={() => toggleSkill(s.id)} className="rounded-full bg-mahak-50 px-3 py-1 text-sm text-mahak-800">
              {g.title} / {s.title} ×
            </button>
          )))}
        </div>
      )}
      <input className={inputClass} placeholder="جستجوی مهارت" value={skillQuery} onChange={(e) => setSkillQuery(e.target.value)} />
      {(catalog || []).length === 0 && (
        <p className="text-sm text-rose-600">فهرست مهارت‌ها بارگذاری نشد. بک‌اند را با <code>go run .\cmd\api</code> دوباره اجرا کنید.</p>
      )}
      <div className="space-y-2">
            {(catalog || []).filter((g) => g.slug !== "general").map((g) => {
          const items = (g.skills || []).filter((s) => s.status !== "inactive" && (!q || s.title.includes(q) || g.title.includes(q)));
          if (q && items.length === 0) return null;
          const count = (g.skills || []).filter((s) => selected.includes(s.id)).length;
          const open = openGroup === g.id || !!q;
          return (
            <div key={g.id} className="overflow-hidden rounded-2xl border border-stone-100">
              <button type="button" className="flex w-full items-center justify-between bg-stone-50 px-4 py-3 text-right" onClick={() => setOpenGroup(open && !q ? "" : g.id)}>
                <span className="font-bold text-mahak-800">{g.title}</span>
                <span className="text-xs text-stone-500">{count ? `${count} انتخاب` : open ? "بستن" : "باز کردن"}</span>
              </button>
              {open && (
                <div className="grid gap-2 p-3 sm:grid-cols-2">
                  {items.map((s) => {
                    const on = selected.includes(s.id);
                    return (
                      <button
                        type="button"
                        key={s.id}
                        onClick={() => toggleSkill(s.id)}
                        className={`rounded-2xl border px-3 py-3 text-right text-sm ${on ? "border-mahak-400 bg-mahak-50 text-mahak-800" : "border-stone-200 bg-white"}`}
                      >
                        <span className="ml-2 inline-flex h-5 w-5 items-center justify-center rounded-md border text-xs">{on ? "✓" : ""}</span>
                        {s.title}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
      <div className="rounded-2xl border border-dashed border-mahak-200 p-4">
        <h3 className="font-bold">پیشنهاد مهارت جدید</h3>
        <p className="mt-1 text-sm text-stone-500">اگر مهارت شما در فهرست نیست پیشنهاد دهید؛ تا تایید ادمین با وضعیت «در انتظار تایید» می‌ماند.</p>
        <div className="mt-3 grid gap-3 md:grid-cols-3">
          <select className={inputClass} value={proposeGroup} onChange={(e) => setProposeGroup(e.target.value)}>
            <option value="">گروه مهارت</option>
            {(catalog || []).filter((g) => g.slug !== "general").map((g) => <option key={g.id} value={g.id}>{g.title}</option>)}
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
  );

  const docsTime = (
    <div className="space-y-8" id="docs-upload">
      <section>
        <h2 className="font-bold">مدارک شناسایی</h2>
        <p className="mt-1 text-sm text-stone-500">تصویر کارت ملی الزامی است. فرمت JPG، PNG یا PDF تا ۵ مگابایت.</p>
        <div className="mt-3 flex flex-wrap items-center gap-3">
          <select className={inputClass + " w-auto"} value={kind} onChange={(e) => setKind(e.target.value)}>
            {DOC_KINDS.map((d) => <option key={d.id} value={d.id}>{d.label}</option>)}
          </select>
          <input type="file" onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])} />
        </div>
        <ul className="mt-3 space-y-1 text-sm">
          {(docs || []).length === 0 && <li className="text-stone-400">هنوز مدرکی بارگذاری نشده است.</li>}
          {(docs || []).map((d) => (
            <li key={d.id} className="flex items-center justify-between gap-2 rounded-xl bg-stone-50 px-3 py-2">
              <span>{docKindLabel(d.kind)}{d.file_name ? ` · ${d.file_name.length > 32 ? `${d.file_name.slice(0, 16)}…${d.file_name.slice(-8)}` : d.file_name}` : ""}</span>
              {canDeleteDoc && (
                <Button variant="ghost" onClick={async () => {
                  try {
                    await api.deleteDoc(d.id);
                    setDocs((await api.myDocs()) || []);
                    setMsg("مدرک حذف شد");
                  } catch (e) {
                    setErr(e instanceof Error ? e.message : "حذف مدرک ممکن نیست");
                  }
                }}>حذف</Button>
              )}
            </li>
          ))}
        </ul>
      </section>
      <section className="rounded-2xl border border-mahak-100 bg-mahak-50/40 p-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="font-bold">زمان‌های آزاد برای داوطلبی</h2>
            <p className="mt-1 max-w-xl text-sm text-stone-600">
              روزهایی از هفته که می‌توانید فعالیت کنید را مشخص کنید؛ برای هر بازه، روز، ساعت شروع و ساعت پایان را جداگانه وارد کنید.
            </p>
          </div>
          <Button variant="outline" onClick={() => setSlots([...(slots || []), { weekday: 6, start_time: "09:00", end_time: "13:00" }])}>
            افزودن بازه زمانی
          </Button>
        </div>
        {(slots || []).length === 0 && (
          <p className="mt-4 rounded-2xl bg-white px-4 py-3 text-sm text-stone-500">
            هنوز بازه‌ای ثبت نشده. با «افزودن بازه زمانی» مثلاً شنبه ۹ تا ۱۳ را وارد کنید.
          </p>
        )}
        <div className="mt-4 space-y-3">
          {(slots || []).map((s, i) => (
            <div key={i} className="grid gap-2 rounded-2xl bg-white p-3 md:grid-cols-[1fr_1fr_1fr_auto]">
              <Field label="روز هفته">
                <select className={inputClass} value={s.weekday} onChange={(e) => {
                  const n = [...slots]; n[i] = { ...s, weekday: Number(e.target.value) }; setSlots(n);
                }}>
                  {WEEKDAYS.map((w, idx) => <option key={w} value={idx}>{w}</option>)}
                </select>
              </Field>
              <Field label="از ساعت">
                <input type="time" className={inputClass} value={s.start_time} onChange={(e) => {
                  const n = [...slots]; n[i] = { ...s, start_time: e.target.value }; setSlots(n);
                }} />
              </Field>
              <Field label="تا ساعت">
                <input type="time" className={inputClass} value={s.end_time} onChange={(e) => {
                  const n = [...slots]; n[i] = { ...s, end_time: e.target.value }; setSlots(n);
                }} />
              </Field>
              <div className="flex items-end">
                <Button variant="danger" onClick={() => setSlots(slots.filter((_, idx) => idx !== i))}>حذف</Button>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );

  const panels = [identity, address, education, skills, docsTime];

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black">{wizard ? "ثبت‌نام داوطلب" : "پروفایل و مدارک"}</h1>
          {!wizard && <p className="text-sm text-stone-500">اطلاعات را در تب‌ها ببینید و بخش‌های غیرهویتی را در صورت نیاز به‌روز کنید.</p>}
        </div>
        {form.status && <Badge status={form.status} reason={form.rejection_reason} />}
      </div>
      {form.status && (
        <Card className="p-4 text-sm">
          <div className="font-bold">وضعیت عضویت: {STATUS_LABEL[form.status] || form.status}</div>
          <p className="mt-1 text-stone-600">{STATUS_EXPLAIN[form.status]}</p>
        </Card>
      )}
      {form.rejection_reason && (
        <Card className="border-rose-200 p-4 text-sm text-rose-800">
          <div>دلیل ادمین: {form.rejection_reason}</div>
          <button type="button" className="mt-2 text-sm font-bold text-mahak-700" onClick={() => setStep(TABS.length - 1)}>
            رفتن به بارگذاری مدارک
          </button>
        </Card>
      )}

      <div className="flex gap-2 overflow-x-auto pb-1">
        {labels.map((label, i) => (
          <button
            key={label}
            type="button"
            onClick={() => { setErr(""); setStep(i); }}
            className={`flex min-w-fit items-center gap-2 rounded-full border px-3 py-1.5 text-sm ${
              i === step ? "border-mahak-500 bg-mahak-50 text-mahak-800" : wizard && i < step ? "border-emerald-200 bg-emerald-50 text-emerald-800" : "border-stone-200 text-stone-500"
            }`}
          >
            {wizard && (
              <span className={`grid h-6 w-6 place-items-center rounded-full text-xs font-bold ${i === step ? "bg-mahak-500 text-white" : i < step ? "bg-emerald-500 text-white" : "bg-stone-200"}`}>
                {i + 1}
              </span>
            )}
            {label}
          </button>
        ))}
      </div>

      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        {panels[step]}

        {wizard ? (
          <div className="mt-6 flex flex-wrap items-center justify-between gap-2">
            <Button variant="ghost" disabled={step === 0} onClick={() => { setErr(""); setStep((s) => Math.max(0, s - 1)); }}>قبلی</Button>
            <div className="flex flex-wrap gap-2">
              {step < STEPS.length - 1 ? (
                <Button disabled={saving} onClick={goNext}>ذخیره و بعدی</Button>
              ) : (
                <>
                  <Button variant="outline" disabled={saving} onClick={saveCurrent}>ذخیره پیش‌نویس</Button>
                  <Button disabled={saving} onClick={submit}>ارسال برای بررسی ادمین</Button>
                </>
              )}
            </div>
          </div>
        ) : (
          <div className="mt-6">
            <Button disabled={saving} onClick={saveCurrent}>ذخیره این بخش</Button>
            {form.status === "pending" && <p className="mt-2 text-xs text-stone-400">درخواست شما در حال بررسی است.</p>}
          </div>
        )}
      </Card>

      {(form.history || []).length > 0 && (
        <Card className="p-5">
          <h2 className="mb-3 font-bold">تاریخچه پرونده</h2>
          <HistoryList items={form.history} audience="volunteer" />
        </Card>
      )}

      <Modal open={draftOpen} title="پیش‌نویس ذخیره شد" onClose={() => setDraftOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست شما ذخیره شد. هر زمان بخواهید می‌توانید ادامه دهید و بعداً برای بررسی ادمین ارسال کنید.
        </p>
        <div className="mt-4 flex justify-end">
          <Button onClick={() => setDraftOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>

      <Modal open={submitOpen} title="درخواست ارسال شد" onClose={() => setSubmitOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست شما ارسال شد و در مرحله بررسی قرار گرفت. پس از بررسی ادمین، نتیجه در همین پرونده و اعلان‌ها نمایش داده می‌شود.
        </p>
        <div className="mt-4 flex justify-end">
          <Button onClick={() => setSubmitOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>
    </div>
  );
}

export default function VolunteerProfilePage() {
  return (
    <Suspense fallback={<p className="p-6 text-sm text-stone-500">در حال بارگذاری پرونده…</p>}>
      <ProfilePage />
    </Suspense>
  );
}
