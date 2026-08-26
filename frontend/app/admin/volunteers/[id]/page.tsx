"use client";

import { ReactNode, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { api, Assignment, Availability, CertificateRequest, DocumentFile, MissionProgress, SkillGroup, Volunteer, openAuth } from "@/lib/api";
import { CERT_REQ_LABEL, DOC_KINDS, EDUCATION_LEVELS, GENDERS, OCCUPATIONS, PROPOSAL_LABEL, STATUS_EXPLAIN, STATUS_LABEL, WEEKDAYS, catalogLabelMap, certRequestTitle, docKindLabel, fmtDate, genderLabel, occupationLabel, skillLabel, workModeLabel } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, AttachmentButton, inputClass } from "@/components/ui";
import { HistoryList } from "@/components/history";
import { ShamsiDateField } from "@/components/shamsi";
import { TabBar } from "@/components/tabs";
import { IRAN_PROVINCES, citiesOf } from "@/lib/iran";
import { currentJalaliYear } from "@/lib/jalali";
import { MIN_VOLUNTEER_AGE, onlyDigits, onlyPersianLetters, volunteerBirthDateError } from "@/lib/persian";

const ADMIN_STATUSES = ["draft", "pending", "approved", "rejected", "suspended"];

const FILE_TABS = [
  { id: "identity", label: "هویت" },
  { id: "address", label: "نشانی" },
  { id: "education", label: "تحصیلات" },
  { id: "skills", label: "مهارت‌ها" },
  { id: "docs", label: "مدارک و زمان آزاد" },
  { id: "requests", label: "درخواست‌ها" },
  { id: "stats", label: "آمار" },
  { id: "activity", label: "فعالیت و مأموریت" },
  { id: "history", label: "تاریخچه" },
] as const;

type FileTab = (typeof FILE_TABS)[number]["id"];
const EDITABLE_TABS: FileTab[] = ["identity", "address", "education"];

function Row({ label, value }: { label: string; value?: ReactNode }) {
  const empty = value === undefined || value === null || value === "";
  return (
    <div className="min-w-0">
      <div className="text-xs text-stone-500">{label}</div>
      <div className="mt-0.5 break-words font-medium text-ink-900">{empty ? "—" : value}</div>
    </div>
  );
}

export default function VolunteerReview() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const [v, setV] = useState<Volunteer | null>(null);
  const [docs, setDocs] = useState<DocumentFile[]>([]);
  const [slots, setSlots] = useState<Availability[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [missions, setMissions] = useState<MissionProgress[]>([]);
  const [msg, setMsg] = useState("");
  const [status, setStatus] = useState("pending");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [nationalId, setNationalId] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [gender, setGender] = useState("");
  const [occupation, setOccupation] = useState("");
  const [occupationOther, setOccupationOther] = useState("");
  const [phone, setPhone] = useState("");
  const [phone2, setPhone2] = useState("");
  const [province, setProvince] = useState("");
  const [city, setCity] = useState("");
  const [address, setAddress] = useState("");
  const [plaque, setPlaque] = useState("");
  const [unit, setUnit] = useState("");
  const [educationLevel, setEducationLevel] = useState("");
  const [educationField, setEducationField] = useState("");
  const [medicalLicense, setMedicalLicense] = useState("");
  const [bio, setBio] = useState("");
  const [comment, setComment] = useState("");
  const [rejectOpen, setRejectOpen] = useState(false);
  const [rejectReason, setRejectReason] = useState("");
  const [docsOpen, setDocsOpen] = useState(false);
  const [docKinds, setDocKinds] = useState<string[]>([]);
  const [docsNote, setDocsNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [editing, setEditing] = useState(false);
  const [statusActionsOpen, setStatusActionsOpen] = useState(false);
  const [statusReason, setStatusReason] = useState("");
  const [certReqs, setCertReqs] = useState<CertificateRequest[]>([]);
  const [certNote, setCertNote] = useState<Record<string, string>>({});
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [skillQuery, setSkillQuery] = useState("");
  const [openGroup, setOpenGroup] = useState("");
  const [skillOpen, setSkillOpen] = useState(false);
  const [draftSkills, setDraftSkills] = useState<string[]>([]);
  const [tab, setTab] = useState<FileTab>("identity");

  const cities = useMemo(() => {
    const list = citiesOf(province);
    if (city && !list.includes(city)) return [city, ...list];
    return list;
  }, [province, city]);
  const skillNames = useMemo(() => catalogLabelMap(catalog), [catalog]);

  function skillName(sid: string) {
    const own = (v?.skills || []).find((s) => s.skill_id === sid);
    if (own?.title) return own.group_title ? `${own.group_title} / ${own.title}` : own.title;
    for (const g of catalog) {
      const s = (g.skills || []).find((x) => x.id === sid);
      if (s) return `${g.title} / ${s.title}`;
    }
    return skillLabel(sid, skillNames);
  }

  function applyVolunteer(vol: Volunteer) {
    const parts = (vol.first_name || vol.last_name)
      ? { first: vol.first_name || "", last: vol.last_name || "" }
      : { first: (vol.full_name || "").split(/\s+/)[0] || "", last: (vol.full_name || "").split(/\s+/).slice(1).join(" ") };
    setFirstName(parts.first);
    setLastName(parts.last);
    setNationalId(vol.national_id || "");
    setBirthDate(vol.birth_date || "");
    setGender(vol.gender || "");
    setOccupation(vol.occupation || "");
    setOccupationOther(vol.occupation_other || "");
    setPhone(vol.phone || "");
    setPhone2(vol.phone2 || "");
    setProvince(vol.province || "");
    setCity(vol.city || "");
    setAddress(vol.address || "");
    setPlaque(vol.plaque || "");
    setUnit(vol.unit || "");
    setEducationLevel(vol.education_level || "");
    setEducationField(vol.education_field || "");
    setMedicalLicense(vol.medical_license || "");
    setBio(vol.bio || "");
    setStatus(vol.status);
    setSelectedSkills((vol.skill_ids || (vol.skills || []).map((s) => s.skill_id)).filter(Boolean));
  }

  async function load() {
    const r = await api.adminVolunteer(id);
    setV(r.volunteer);
    setDocs(r.documents || []);
    setSlots(r.availability || []);
    setAssignments(r.assignments || []);
    setMissions(r.missions || []);
    applyVolunteer(r.volunteer);
    const reqs = await api.adminCertRequests("").catch(() => [] as CertificateRequest[]);
    setCertReqs((reqs || []).filter((x) => x.volunteer_id === id));
  }
  useEffect(() => { if (id) void load(); }, [id]);
  useEffect(() => {
    api.adminSkillCatalog().then((g) => setCatalog(g || [])).catch(() => undefined);
  }, []);

  async function run(fn: () => Promise<unknown>, ok = "ثبت شد") {
    setBusy(true);
    setMsg("");
    try {
      await fn();
      setMsg(ok);
      await load();
      return true;
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
      return false;
    } finally {
      setBusy(false);
    }
  }

  function profileBody(skillIds?: string[]) {
    const body: Parameters<typeof api.adminUpdateVolunteer>[1] = {
      first_name: firstName,
      last_name: lastName,
      national_id: nationalId,
      phone,
      phone2,
      birth_date: birthDate,
      gender,
      occupation,
      occupation_other: occupation === "other" ? occupationOther : "",
      province,
      city,
      address,
      plaque,
      unit,
      education_level: educationLevel,
      education_field: educationField,
      medical_license: medicalLicense,
      bio,
    };
    if (skillIds) body.skill_ids = skillIds;
    return body;
  }

  async function saveProfile() {
    if (!v) return;
    const birthErr = volunteerBirthDateError(birthDate);
    if (birthErr) {
      setMsg(birthErr);
      return;
    }
    if (await run(() => api.adminUpdateVolunteer(v.id, profileBody()), "اطلاعات داوطلب ذخیره شد")) {
      setEditing(false);
    }
  }

  function cancelEdit() {
    if (v) applyVolunteer(v);
    setEditing(false);
    setMsg("");
  }

  function changeTab(next: FileTab) {
    if (editing) cancelEdit();
    setMsg("");
    setTab(next);
  }

  function openSkills() {
    setDraftSkills(selectedSkills);
    setSkillQuery("");
    setOpenGroup("");
    setSkillOpen(true);
  }

  async function saveSkills() {
    if (!v) return;
    if (await run(() => api.adminUpdateVolunteer(v.id, profileBody(draftSkills)), "مهارت‌ها ذخیره شد")) {
      setSkillOpen(false);
    }
  }

  function toggleStatusActions() {
    if (!v) return;
    const next = !statusActionsOpen;
    setStatusActionsOpen(next);
    if (next) {
      setStatus(v.status);
      setStatusReason("");
    }
  }

  async function confirmStatusChange() {
    if (!v) return;
    if (!statusReason.trim()) {
      setMsg("برای تغییر وضعیت باید دلیل ثبت شود");
      return;
    }
    if (await run(() => api.setVolunteerStatus(v.id, status, statusReason.trim()), "وضعیت به‌روز شد")) {
      setStatusActionsOpen(false);
      setStatusReason("");
    }
  }

  async function confirmReject() {
    if (!v) return;
    if (!rejectReason.trim()) {
      setMsg("برای رد کردن درخواست باید دلیل ثبت شود");
      return;
    }
    await run(async () => {
      await api.setVolunteerStatus(v.id, "rejected", rejectReason.trim());
    }, "درخواست رد شد");
    setRejectOpen(false);
    setRejectReason("");
    setStatusActionsOpen(false);
  }

  async function confirmDocs() {
    if (!v) return;
    const labels = DOC_KINDS.filter((d) => docKinds.includes(d.id)).map((d) => d.label);
    const parts = [
      labels.length ? `مدارک مورد نیاز: ${labels.join("، ")}` : "",
      docsNote.trim(),
    ].filter(Boolean);
    const reason = parts.join("\n");
    if (!reason) {
      setMsg("مدارک درخواستی یا توضیح را وارد کنید");
      return;
    }
    await run(() => api.review(v.id, "request_documents", reason), "درخواست مدارک ثبت شد");
    setDocsOpen(false);
    setDocKinds([]);
    setDocsNote("");
  }

  if (!v) return null;
  const placeholderEmail = (v.email || "").endsWith("@otp.mahak.local");
  const tabItems = FILE_TABS.map((t) => {
    if (t.id === "requests" && certReqs.length) return { id: t.id, label: `درخواست‌ها (${certReqs.length})` };
    if (t.id === "activity") {
      const n = assignments.length + missions.length;
      if (n) return { id: t.id, label: `فعالیت و مأموریت (${n})` };
    }
    return { id: t.id, label: t.label };
  });

  const identityView = (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Row label="ایمیل" value={placeholderEmail ? `${v.email} (ورود با موبایل)` : v.email} />
      <Row label="موبایل" value={v.phone} />
      <Row label="موبایل دوم" value={v.phone2} />
      <Row label="نام" value={v.first_name || v.full_name} />
      <Row label="نام خانوادگی" value={v.last_name} />
      <Row label="کد ملی" value={v.national_id} />
      <Row label="تاریخ تولد" value={fmtDate(v.birth_date)} />
      <Row label="جنسیت" value={genderLabel(v.gender)} />
      <Row label="شغل" value={occupationLabel(v.occupation, v.occupation_other)} />
    </div>
  );

  const identityForm = (
    <div className="grid gap-3 md:grid-cols-2">
      <Field label="نام">
        <input className={inputClass} value={firstName} onChange={(e) => setFirstName(onlyPersianLetters(e.target.value))} />
      </Field>
      <Field label="نام خانوادگی">
        <input className={inputClass} value={lastName} onChange={(e) => setLastName(onlyPersianLetters(e.target.value))} />
      </Field>
      <Field label="کد ملی">
        <input className={inputClass} dir="ltr" maxLength={10} value={nationalId} onChange={(e) => setNationalId(onlyDigits(e.target.value, 10))} />
      </Field>
      <Field label="موبایل">
        <input className={inputClass} dir="ltr" value={phone} onChange={(e) => setPhone(onlyDigits(e.target.value, 11))} />
      </Field>
      <Field label="موبایل دوم">
        <input className={inputClass} dir="ltr" value={phone2} onChange={(e) => setPhone2(onlyDigits(e.target.value, 11))} />
      </Field>
      <div>
        <ShamsiDateField className="max-w-[16rem]" label="تاریخ تولد" value={birthDate} maxYear={currentJalaliYear() - MIN_VOLUNTEER_AGE} onChange={setBirthDate} />
        <p className="mt-1 text-xs text-stone-500">حداقل سن داوطلبی ۱۸ سال تمام است.</p>
      </div>
      <Field label="جنسیت">
        <select className={inputClass} value={gender} onChange={(e) => setGender(e.target.value)}>
          <option value="">انتخاب کنید</option>
          {GENDERS.map((g) => <option key={g.id} value={g.id}>{g.label}</option>)}
        </select>
      </Field>
      <Field label="شغل">
        <select className={inputClass} value={occupation} onChange={(e) => {
          setOccupation(e.target.value);
          if (e.target.value !== "other") setOccupationOther("");
        }}>
          <option value="">انتخاب کنید</option>
          {OCCUPATIONS.map((o) => <option key={o.id} value={o.id}>{o.label}</option>)}
        </select>
      </Field>
      {occupation === "other" && (
        <Field label="شرح شغل">
          <input className={inputClass} value={occupationOther} maxLength={80} onChange={(e) => setOccupationOther(e.target.value)} placeholder="شغل را بنویسید" />
        </Field>
      )}
    </div>
  );

  const addressView = (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Row label="استان" value={v.province} />
      <Row label="شهر" value={v.city} />
      <Row label="پلاک" value={v.plaque} />
      <Row label="واحد" value={v.unit} />
      <div className="sm:col-span-2 lg:col-span-3">
        <Row label="آدرس" value={v.address} />
      </div>
    </div>
  );

  const addressForm = (
    <div className="grid gap-3 md:grid-cols-2">
      <Field label="استان">
        <select className={inputClass} value={province} onChange={(e) => { setProvince(e.target.value); setCity(""); }}>
          <option value="">انتخاب استان</option>
          {IRAN_PROVINCES.map((p) => <option key={p.name} value={p.name}>{p.name}</option>)}
        </select>
      </Field>
      <Field label="شهر">
        <select className={inputClass} value={city} onChange={(e) => setCity(e.target.value)} disabled={!province}>
          <option value="">انتخاب شهر</option>
          {cities.map((c) => <option key={c} value={c}>{c}</option>)}
        </select>
      </Field>
      <div className="md:col-span-2">
        <Field label="آدرس">
          <input className={inputClass} value={address} onChange={(e) => setAddress(e.target.value)} />
        </Field>
      </div>
      <Field label="پلاک">
        <input className={inputClass} value={plaque} onChange={(e) => setPlaque(e.target.value)} />
      </Field>
      <Field label="واحد">
        <input className={inputClass} value={unit} onChange={(e) => setUnit(e.target.value)} />
      </Field>
    </div>
  );

  const educationView = (
    <div className="grid gap-4 sm:grid-cols-2">
      <Row label="مقطع تحصیلی" value={v.education_level} />
      <Row label="رشته تحصیلی" value={v.education_field} />
      <Row label="نظام پزشکی" value={v.medical_license} />
      <div className="sm:col-span-2">
        <Row label="درباره داوطلب" value={v.bio} />
      </div>
    </div>
  );

  const educationForm = (
    <div className="grid gap-3 md:grid-cols-2">
      <Field label="مقطع تحصیلی">
        <select className={inputClass} value={educationLevel} onChange={(e) => setEducationLevel(e.target.value)}>
          <option value="">انتخاب کنید</option>
          {EDUCATION_LEVELS.map((x) => <option key={x} value={x}>{x}</option>)}
        </select>
      </Field>
      <Field label="رشته تحصیلی">
        <input className={inputClass} value={educationField} onChange={(e) => setEducationField(e.target.value)} />
      </Field>
      <Field label="نظام پزشکی">
        <input className={inputClass} value={medicalLicense} onChange={(e) => setMedicalLicense(e.target.value)} />
      </Field>
      <div className="md:col-span-2">
        <Field label="درباره داوطلب">
          <textarea className={inputClass} rows={3} value={bio} onChange={(e) => setBio(e.target.value)} />
        </Field>
      </div>
    </div>
  );

  const panel: Record<FileTab, { title: string; view: ReactNode; edit?: ReactNode }> = {
    identity: { title: "هویت", view: identityView, edit: identityForm },
    address: { title: "نشانی", view: addressView, edit: addressForm },
    education: { title: "تحصیلات", view: educationView, edit: educationForm },
    skills: {
      title: "مهارت‌ها",
      view: (
        <div className="space-y-4">
          {(v.skills || []).length === 0 ? (
            <p className="text-sm text-stone-400">مهارتی ثبت نشده</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {(v.skills || []).map((s) => (
                <span key={s.skill_id} className="rounded-full bg-mahak-50 px-3 py-1 text-sm text-mahak-800">
                  {s.group_title ? `${s.group_title} / ${s.title}` : s.title}
                </span>
              ))}
            </div>
          )}
          {(v.proposals || []).length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-bold text-stone-600">پیشنهاد مهارت</h3>
              <ul className="space-y-2 text-sm">
                {(v.proposals || []).map((p) => (
                  <li key={p.id} className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-stone-100 px-3 py-2">
                    <span>{p.group_title} / {p.title}</span>
                    <span className="text-stone-500">{PROPOSAL_LABEL[p.status] || p.status}{p.admin_note ? ` · ${p.admin_note}` : ""}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      ),
    },
    docs: {
      title: "مدارک و زمان آزاد",
      view: (
        <div className="space-y-8">
          <section>
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-bold">مدارک شناسایی</h3>
              <Button variant="outline" disabled={busy} onClick={() => {
                const uploaded = new Set((docs || []).map((d) => d.kind));
                setDocKinds(uploaded.has("national_id") ? [] : ["national_id"]);
                setDocsNote("");
                setDocsOpen(true);
              }}>درخواست مدارک</Button>
            </div>
            <ul className="space-y-2 text-sm">
              {(docs || []).map((d) => (
                <li key={d.id}>
                  <AttachmentButton
                    name={d.file_name}
                    label={docKindLabel(d.kind)}
                    onOpen={() => void openAuth(`/api/v1/admin/documents/${d.id}`)}
                  />
                </li>
              ))}
              {(docs || []).length === 0 && <li className="text-stone-400">مدرکی بارگذاری نشده</li>}
            </ul>
          </section>
          <section>
            <h3 className="mb-3 font-bold">زمان‌های آزاد برای داوطلبی</h3>
            <ul className="space-y-1 text-sm">
              {(slots || []).length === 0 && <li className="text-stone-400">بازه‌ای ثبت نشده</li>}
              {(slots || []).map((s, i) => (
                <li key={i} className="rounded-xl bg-stone-50 px-3 py-2">
                  {WEEKDAYS[s.weekday]} · {s.start_time} تا {s.end_time}
                </li>
              ))}
            </ul>
          </section>
        </div>
      ),
    },
    requests: {
      title: "درخواست‌های تقدیرنامه و گواهی",
      view: (
        <div className="space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Link className="text-sm text-mahak-700" href="/admin/certificates">همه درخواست‌ها</Link>
            <Button variant="ghost" disabled={busy} onClick={() => run(() => api.issueAggregated(v.id), "تقدیرنامه تجمیعی صادر شد")}>
              صدور تقدیرنامه تجمیعی
            </Button>
          </div>
          {certReqs.length === 0 && <p className="text-sm text-stone-400">درخواستی برای این داوطلب ثبت نشده است.</p>}
          {certReqs.map((r) => (
            <div key={r.id} className="rounded-2xl border border-stone-100 px-3 py-2">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <div className="font-medium">{certRequestTitle(r)}</div>
                  <div className="text-xs text-stone-500">{fmtDate(r.created_at)}</div>
                </div>
                <Badge status={r.status} reason={r.admin_note} label={CERT_REQ_LABEL[r.status]} />
              </div>
              {(r.status === "pending" || r.status === "preparing") && (
                <div className="mt-2 flex flex-wrap items-end gap-2">
                  <input
                    className={inputClass + " max-w-xs"}
                    placeholder="یادداشت یا دلیل رد"
                    value={certNote[r.id] || ""}
                    onChange={(e) => setCertNote({ ...certNote, [r.id]: e.target.value })}
                  />
                  <Button disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "approve", certNote[r.id] || ""), r.kind === "official" ? "آماده تحویل شد" : "تقدیرنامه صادر شد")}>
                    {r.kind === "official" ? "بررسی و صدور" : "تایید و صدور"}
                  </Button>
                  <Button variant="danger" disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "reject", certNote[r.id] || ""), "رد شد")}>رد</Button>
                </div>
              )}
              {r.status === "ready" && (
                <div className="mt-2 flex flex-wrap gap-2">
                  <Button disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "deliver", "", "send"), "ارسال ثبت شد")}>ارسال برای داوطلب</Button>
                  <Button variant="outline" disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "deliver", "", "in_person"), "تحویل حضوری ثبت شد")}>تحویل حضوری</Button>
                </div>
              )}
              {r.admin_note && r.status !== "pending" && r.status !== "preparing" && <p className="mt-1 text-sm text-stone-600">{r.admin_note}</p>}
            </div>
          ))}
        </div>
      ),
    },
    stats: {
      title: "آمار همکاری",
      view: (
        <div className="grid gap-4 sm:grid-cols-3">
          <Row label="ساعات داوطلبی" value={v.total_hours} />
          <Row label="میانگین امتیاز" value={v.average_score?.toFixed?.(1) ?? v.average_score} />
          <Row label="فعالیت‌های تکمیل‌شده" value={v.completed_tasks} />
        </div>
      ),
    },
    activity: {
      title: "فعالیت و مأموریت",
      view: (
        <div className="space-y-8">
          <section>
            <h3 className="mb-3 font-bold">فعالیت‌ها</h3>
            {assignments.length === 0 && <p className="text-sm text-stone-400">فعالیتی ثبت نشده است.</p>}
            <ul className="space-y-2">
              {assignments.map((a) => (
                <li key={a.id} className="rounded-2xl border border-stone-100 px-3 py-2 text-sm">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <div className="font-medium">{a.task?.title || "فعالیت"}</div>
                      <div className="mt-0.5 text-xs text-stone-500">
                        {workModeLabel(a.task?.work_mode)}
                        {a.task?.kind === "occurrence" || a.task?.kind === "recurring" ? " · جاری" : ""}
                        {a.task?.starts_at ? ` · ${fmtDate(a.task.starts_at)}` : ""}
                        {a.hours_awarded ? ` · ${a.hours_awarded} ساعت` : ""}
                      </div>
                    </div>
                    <Badge status={a.status} reason={a.admin_comment} />
                  </div>
                  {a.composite_score ? <p className="mt-1 text-xs text-stone-600">امتیاز پشتیبانی: {a.composite_score}</p> : null}
                  {a.volunteer_rating ? (
                    <p className="mt-1 text-xs text-stone-600">
                      امتیاز داوطلب: {a.volunteer_rating}
                      {a.volunteer_comment ? ` — ${a.volunteer_comment}` : ""}
                    </p>
                  ) : null}
                  {a.delivery_note && <p className="mt-1 text-xs text-stone-500">نتیجه ارسالی: {a.delivery_note}</p>}
                </li>
              ))}
            </ul>
          </section>
          <section>
            <h3 className="mb-3 font-bold">مأموریت‌ها</h3>
            {missions.length === 0 && <p className="text-sm text-stone-400">مأموریتی ثبت نشده است.</p>}
            <ul className="space-y-2">
              {missions.map((m) => (
                <li key={m.id} className="flex flex-wrap items-start justify-between gap-2 rounded-2xl border border-stone-100 px-3 py-2 text-sm">
                  <div>
                    <div className="font-medium">{m.mission?.title || "مأموریت"}</div>
                    <div className="mt-0.5 text-xs text-stone-500">
                      پیشرفت {m.progress}{m.mission?.target_count ? ` از ${m.mission.target_count}` : ""}
                      {m.started_at ? ` · شروع ${fmtDate(m.started_at)}` : ""}
                      {m.completed_at ? ` · پایان ${fmtDate(m.completed_at)}` : ""}
                      {m.mission?.hour_weight ? ` · ${m.mission.hour_weight} ساعت` : ""}
                    </div>
                  </div>
                  <Badge status={m.status} />
                </li>
              ))}
            </ul>
          </section>
        </div>
      ),
    },
    history: {
      title: "تاریخچه پرونده",
      view: (
        <div className="space-y-6">
          <section className="space-y-3">
            <h3 className="font-bold">پیام به داوطلب</h3>
            <p className="text-sm text-stone-500">این پیام در تاریخچه پرونده و برای داوطلب نمایش داده می‌شود.</p>
            <textarea className={inputClass} rows={3} placeholder="متن کامنت یا پیام" value={comment} onChange={(e) => setComment(e.target.value)} />
            <Button disabled={busy || !comment.trim()} onClick={() => run(async () => {
              await api.commentVolunteer(v.id, comment.trim());
              setComment("");
            }, "پیام ارسال شد")}>ارسال پیام</Button>
          </section>
          <HistoryList items={v.history} filterable />
        </div>
      ),
    },
  };

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black">{v.full_name}</h1>
          <p className="mt-1 text-sm text-stone-600">{STATUS_EXPLAIN[v.status] || STATUS_LABEL[v.status] || v.status}</p>
        </div>
        <Badge status={v.status} reason={v.rejection_reason} />
      </div>

      <Card className="space-y-3 p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="font-bold">وضعیت عضویت: {STATUS_LABEL[v.status] || v.status}</h2>
            <p className="mt-1 text-sm text-stone-600">{STATUS_EXPLAIN[v.status]}</p>
            {v.rejection_reason && (v.status === "rejected" || v.status === "draft" || v.status === "suspended") && (
              <p className="mt-1 text-sm text-rose-700">دلیل: {v.rejection_reason}</p>
            )}
          </div>
          <Button variant="outline" disabled={busy} onClick={toggleStatusActions}>
            {statusActionsOpen ? "بستن" : "تغییر وضعیت"}
          </Button>
        </div>
        {statusActionsOpen && (
          <div className="space-y-4 border-t border-stone-100 pt-4">
            <div className="flex flex-wrap gap-2">
              {(v.status === "pending" || v.status === "draft" || v.status === "rejected") && (
                <Button disabled={busy} onClick={() => run(() => api.review(v.id, "approve"), "تایید شد").then((ok) => { if (ok) setStatusActionsOpen(false); })}>تایید نهایی</Button>
              )}
              {(v.status === "pending" || v.status === "draft") && (
                <Button variant="danger" disabled={busy} onClick={() => { setRejectReason(""); setRejectOpen(true); }}>رد</Button>
              )}
              {v.status === "approved" && <Button variant="ghost" disabled={busy} onClick={() => run(() => api.review(v.id, "suspend"), "تعلیق شد").then((ok) => { if (ok) setStatusActionsOpen(false); })}>تعلیق</Button>}
              {v.status === "suspended" && <Button variant="ghost" disabled={busy} onClick={() => run(() => api.review(v.id, "unsuspend"), "رفع تعلیق شد").then((ok) => { if (ok) setStatusActionsOpen(false); })}>رفع تعلیق</Button>}
            </div>
            <div className="grid gap-3 md:grid-cols-2">
              <Field label="وضعیت جدید">
                <select className={inputClass} value={status} onChange={(e) => setStatus(e.target.value)}>
                  {ADMIN_STATUSES.map((s) => (
                    <option key={s} value={s}>{STATUS_LABEL[s] || s}</option>
                  ))}
                </select>
              </Field>
              <div className="md:col-span-2">
                <Field label="دلیل تغییر وضعیت">
                  <textarea className={inputClass} rows={3} placeholder="دلیل الزامی است" value={statusReason} onChange={(e) => setStatusReason(e.target.value)} />
                </Field>
              </div>
            </div>
            <Button disabled={busy || !statusReason.trim()} onClick={confirmStatusChange}>ثبت تغییر وضعیت</Button>
          </div>
        )}
      </Card>

      <TabBar items={tabItems} active={tab} onChange={(next) => changeTab(next as FileTab)} />
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-black">{panel[tab].title}</h2>
          {EDITABLE_TABS.includes(tab) && !editing && (
            <Button variant="outline" disabled={busy} onClick={() => setEditing(true)}>ویرایش</Button>
          )}
          {tab === "skills" && (
            <Button variant="outline" disabled={busy} onClick={openSkills}>افزودن</Button>
          )}
        </div>
        {EDITABLE_TABS.includes(tab) && editing ? panel[tab].edit : panel[tab].view}
        {editing && EDITABLE_TABS.includes(tab) && (
          <div className="mt-6 flex flex-wrap gap-2">
            <Button disabled={busy} onClick={saveProfile}>ذخیره تغییرات</Button>
            <Button variant="ghost" disabled={busy} onClick={cancelEdit}>انصراف</Button>
          </div>
        )}
      </Card>

      <Modal open={skillOpen} title="افزودن و حذف مهارت" onClose={() => setSkillOpen(false)} size="lg">
        <p className="text-sm text-stone-500">مهارت‌های فعلی را بردارید یا مهارت جدید انتخاب کنید، سپس ذخیره کنید.</p>
        {draftSkills.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-2">
            {draftSkills.map((sid) => (
              <button
                key={sid}
                type="button"
                onClick={() => setDraftSkills(draftSkills.filter((x) => x !== sid))}
                className="rounded-full bg-mahak-50 px-3 py-1 text-sm text-mahak-800"
              >
                {skillName(sid)} ×
              </button>
            ))}
          </div>
        )}
        <input className={inputClass + " mt-3"} placeholder="جستجوی مهارت" value={skillQuery} onChange={(e) => setSkillQuery(e.target.value)} />
        <div className="mt-3 max-h-[50vh] space-y-2 overflow-y-auto">
          {(catalog || []).filter((g) => g.slug !== "general").map((g) => {
            const q = skillQuery.trim();
            const items = (g.skills || []).filter((s) => s.status !== "inactive" && (!q || s.title.includes(q) || g.title.includes(q)));
            if (q && items.length === 0) return null;
            const count = (g.skills || []).filter((s) => draftSkills.includes(s.id)).length;
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
                      const on = draftSkills.includes(s.id);
                      return (
                        <button
                          type="button"
                          key={s.id}
                          onClick={() => setDraftSkills(on ? draftSkills.filter((x) => x !== s.id) : [...draftSkills, s.id])}
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
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setSkillOpen(false)}>انصراف</Button>
          <Button disabled={busy} onClick={() => void saveSkills()}>ذخیره مهارت‌ها</Button>
        </div>
      </Modal>

      <Modal open={rejectOpen} title="دلیل رد درخواست" onClose={() => setRejectOpen(false)}>
        <p className="text-sm text-stone-600">برای رد کردن پرونده باید دلیل ثبت شود. این متن برای داوطلب نمایش داده می‌شود.</p>
        <textarea className={inputClass + " mt-3"} rows={4} placeholder="دلیل رد" value={rejectReason} onChange={(e) => setRejectReason(e.target.value)} />
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setRejectOpen(false)}>انصراف</Button>
          <Button variant="danger" disabled={busy || !rejectReason.trim()} onClick={confirmReject}>ثبت رد</Button>
        </div>
      </Modal>

      <Modal open={docsOpen} title="درخواست مدارک" onClose={() => setDocsOpen(false)}>
        <p className="text-sm text-stone-600">فقط مدارکی که هنوز بارگذاری نشده‌اند در فهرست می‌آیند.</p>
        <div className="mt-3 space-y-2">
          {DOC_KINDS.filter((d) => !(docs || []).some((x) => x.kind === d.id)).map((d) => (
            <label key={d.id} className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={docKinds.includes(d.id)}
                onChange={(e) => setDocKinds((cur) => e.target.checked ? [...cur, d.id] : cur.filter((x) => x !== d.id))}
              />
              {d.label}
            </label>
          ))}
          {DOC_KINDS.filter((d) => (docs || []).some((x) => x.kind === d.id)).length > 0 && (
            <p className="text-xs text-stone-400">
              قبلاً بارگذاری شده: {DOC_KINDS.filter((d) => (docs || []).some((x) => x.kind === d.id)).map((d) => d.label).join("، ")}
            </p>
          )}
        </div>
        <textarea className={inputClass + " mt-3"} rows={3} placeholder="توضیح برای داوطلب" value={docsNote} onChange={(e) => setDocsNote(e.target.value)} />
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setDocsOpen(false)}>انصراف</Button>
          <Button disabled={busy} onClick={confirmDocs}>ثبت درخواست مدارک</Button>
        </div>
      </Modal>
    </div>
  );
}
