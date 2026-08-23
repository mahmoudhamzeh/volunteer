"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { api, Availability, CertificateRequest, DocumentFile, Volunteer, openAuth } from "@/lib/api";
import { DOC_KINDS, EDUCATION_LEVELS, GENDERS, OCCUPATIONS, PROPOSAL_LABEL, STATUS_EXPLAIN, STATUS_LABEL, WEEKDAYS, docKindLabel, fmtDate, genderLabel, occupationLabel } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, AttachmentButton, inputClass } from "@/components/ui";
import { HistoryList } from "@/components/history";
import { ShamsiDateField } from "@/components/shamsi";
import { IRAN_PROVINCES, citiesOf } from "@/lib/iran";
import { onlyDigits, onlyPersianLetters } from "@/lib/persian";
import { ReactNode } from "react";

const ADMIN_STATUSES = ["draft", "pending", "approved", "rejected", "suspended"];

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
  const [statusOpen, setStatusOpen] = useState(false);
  const [statusReason, setStatusReason] = useState("");
  const [certReqs, setCertReqs] = useState<CertificateRequest[]>([]);
  const [certNote, setCertNote] = useState<Record<string, string>>({});

  const cities = useMemo(() => {
    const list = citiesOf(province);
    if (city && !list.includes(city)) return [city, ...list];
    return list;
  }, [province, city]);

  async function load() {
    const r = await api.adminVolunteer(id);
    setV(r.volunteer);
    setDocs(r.documents || []);
    setSlots(r.availability || []);
    const parts = (r.volunteer.first_name || r.volunteer.last_name)
      ? { first: r.volunteer.first_name || "", last: r.volunteer.last_name || "" }
      : { first: (r.volunteer.full_name || "").split(/\s+/)[0] || "", last: (r.volunteer.full_name || "").split(/\s+/).slice(1).join(" ") };
    setFirstName(parts.first);
    setLastName(parts.last);
    setNationalId(r.volunteer.national_id || "");
    setBirthDate(r.volunteer.birth_date || "");
    setGender(r.volunteer.gender || "");
    setOccupation(r.volunteer.occupation || "");
    setOccupationOther(r.volunteer.occupation_other || "");
    setPhone(r.volunteer.phone || "");
    setPhone2(r.volunteer.phone2 || "");
    setProvince(r.volunteer.province || "");
    setCity(r.volunteer.city || "");
    setAddress(r.volunteer.address || "");
    setPlaque(r.volunteer.plaque || "");
    setUnit(r.volunteer.unit || "");
    setEducationLevel(r.volunteer.education_level || "");
    setEducationField(r.volunteer.education_field || "");
    setMedicalLicense(r.volunteer.medical_license || "");
    setBio(r.volunteer.bio || "");
    setStatus(r.volunteer.status);
    const reqs = await api.adminCertRequests("").catch(() => [] as CertificateRequest[]);
    setCertReqs((reqs || []).filter((x) => x.volunteer_id === id));
  }
  useEffect(() => { if (id) void load(); }, [id]);

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

  async function saveProfile() {
    if (!v) return;
    if (await run(() => api.adminUpdateVolunteer(v.id, {
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
    }), "اطلاعات داوطلب ذخیره شد")) {
      setEditing(false);
    }
  }

  async function confirmStatusChange() {
    if (!v) return;
    if (!statusReason.trim()) {
      setMsg("برای تغییر وضعیت باید دلیل ثبت شود");
      return;
    }
    if (await run(() => api.setVolunteerStatus(v.id, status, statusReason.trim()), "وضعیت به‌روز شد")) {
      setStatusOpen(false);
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

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black">{v.full_name}</h1>
          <p className="mt-1 text-sm text-stone-600">{STATUS_EXPLAIN[v.status] || STATUS_LABEL[v.status] || v.status}</p>
        </div>
        <Badge status={v.status} reason={v.rejection_reason} />
      </div>

      <Card className="space-y-3 p-5">
        <h2 className="font-bold">وضعیت عضویت</h2>
        <p className="text-sm text-stone-600">{STATUS_EXPLAIN[v.status]}</p>
        {v.rejection_reason && (v.status === "rejected" || v.status === "draft" || v.status === "suspended") && (
          <p className="text-sm text-rose-700">دلیل: {v.rejection_reason}</p>
        )}
        <div className="flex flex-wrap gap-2">
          {(v.status === "pending" || v.status === "draft" || v.status === "rejected") && (
            <Button disabled={busy} onClick={() => run(() => api.review(v.id, "approve"), "تایید شد")}>تایید نهایی</Button>
          )}
          <Button variant="outline" disabled={busy} onClick={() => {
            const uploaded = new Set((docs || []).map((d) => d.kind));
            setDocKinds(uploaded.has("national_id") ? [] : ["national_id"]);
            setDocsNote("");
            setDocsOpen(true);
          }}>درخواست مدارک</Button>
          {(v.status === "pending" || v.status === "draft") && (
            <Button variant="danger" disabled={busy} onClick={() => { setRejectReason(""); setRejectOpen(true); }}>رد</Button>
          )}
          <Button variant="outline" disabled={busy} onClick={() => { setStatus(v.status); setStatusReason(""); setStatusOpen(true); }}>تغییر وضعیت</Button>
          {v.status === "approved" && <Button variant="ghost" disabled={busy} onClick={() => run(() => api.review(v.id, "suspend"), "تعلیق شد")}>تعلیق</Button>}
          {v.status === "suspended" && <Button variant="ghost" disabled={busy} onClick={() => run(() => api.review(v.id, "unsuspend"), "رفع تعلیق شد")}>رفع تعلیق</Button>}
          <Button variant="ghost" disabled={busy} onClick={() => run(() => api.issueAggregated(v.id), "گواهی تجمیعی صادر شد")}>صدور گواهی تجمیعی</Button>
        </div>
      </Card>

      <Card className="space-y-3 p-5">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="font-bold">درخواست‌های گواهی</h2>
          <Link className="text-sm text-mahak-700" href="/admin/certificates">همه درخواست‌ها</Link>
        </div>
        {certReqs.length === 0 && <p className="text-sm text-stone-400">درخواستی برای این داوطلب ثبت نشده است.</p>}
        {certReqs.map((r) => (
          <div key={r.id} className="rounded-2xl border border-stone-100 px-3 py-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <div className="font-medium">{r.assignment_title || (r.kind === "aggregated" ? "گواهی تجمیعی" : "گواهی فعالیت")}</div>
                <div className="text-xs text-stone-500">{fmtDate(r.created_at)}</div>
              </div>
              <Badge status={r.status} reason={r.admin_note} />
            </div>
            {r.status === "pending" && (
              <div className="mt-2 flex flex-wrap items-end gap-2">
                <input
                  className={inputClass + " max-w-xs"}
                  placeholder="یادداشت یا دلیل رد"
                  value={certNote[r.id] || ""}
                  onChange={(e) => setCertNote({ ...certNote, [r.id]: e.target.value })}
                />
                <Button disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "approve", certNote[r.id] || ""), "گواهی صادر شد")}>تایید و صدور</Button>
                <Button variant="danger" disabled={busy} onClick={() => run(() => api.reviewCertRequest(r.id, "reject", certNote[r.id] || ""), "رد شد")}>رد</Button>
              </div>
            )}
            {r.admin_note && r.status !== "pending" && <p className="mt-1 text-sm text-stone-600">{r.admin_note}</p>}
          </div>
        ))}
      </Card>

      <Card className="p-5">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h2 className="font-bold">اطلاعات کاربر</h2>
          {!editing && <Button variant="outline" disabled={busy} onClick={() => setEditing(true)}>ویرایش</Button>}
        </div>
        {!editing ? (
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
          <Row label="استان" value={v.province} />
          <Row label="شهر" value={v.city} />
          <Row label="پلاک" value={v.plaque} />
          <Row label="واحد" value={v.unit} />
          <Row label="مقطع تحصیلی" value={v.education_level} />
          <Row label="رشته تحصیلی" value={v.education_field} />
          <Row label="نظام پزشکی" value={v.medical_license} />
          <div className="sm:col-span-2 lg:col-span-3">
            <Row label="آدرس" value={v.address} />
          </div>
          {v.bio && (
            <div className="sm:col-span-2 lg:col-span-3">
              <Row label="درباره داوطلب" value={v.bio} />
            </div>
          )}
        </div>
        ) : (
        <>
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
          <ShamsiDateField className="max-w-[16rem]" label="تاریخ تولد" value={birthDate} onChange={setBirthDate} />
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
        <div className="mt-4 flex flex-wrap gap-2">
          <Button disabled={busy} onClick={saveProfile}>ذخیره اطلاعات</Button>
          <Button variant="ghost" disabled={busy} onClick={() => setEditing(false)}>انصراف</Button>
        </div>
        </>
        )}
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">مهارت‌ها</h2>
        <div className="flex flex-wrap gap-2">
          {(v.skills || []).length === 0 && <span className="text-sm text-stone-400">مهارتی ثبت نشده</span>}
          {(v.skills || []).map((s) => (
            <span key={s.skill_id} className="rounded-full border border-mahak-100 bg-mahak-50 px-3 py-1 text-sm text-mahak-800">
              {s.group_title} / {s.title}
            </span>
          ))}
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">آمار همکاری</h2>
        <div className="grid gap-4 sm:grid-cols-3">
          <Row label="ساعات داوطلبی" value={v.total_hours} />
          <Row label="میانگین امتیاز" value={v.average_score?.toFixed?.(1) ?? v.average_score} />
          <Row label="فعالیت‌های تکمیل‌شده" value={v.completed_tasks} />
        </div>
      </Card>

      {(v.proposals || []).length > 0 && (
        <Card className="p-5">
          <h2 className="mb-3 font-bold">پیشنهاد مهارت</h2>
          <ul className="space-y-2 text-sm">
            {(v.proposals || []).map((p) => (
              <li key={p.id} className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-stone-100 px-3 py-2">
                <span>{p.group_title} / {p.title}</span>
                <span className="text-stone-500">{PROPOSAL_LABEL[p.status] || p.status}{p.admin_note ? ` · ${p.admin_note}` : ""}</span>
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card className="p-5">
        <h2 className="mb-3 font-bold">مدارک</h2>
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
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">زمان‌های آزاد</h2>
        <ul className="text-sm">
          {(slots || []).length === 0 && <li className="text-stone-400">بازه‌ای ثبت نشده</li>}
          {(slots || []).map((s, i) => <li key={i}>{WEEKDAYS[s.weekday]} {s.start_time} تا {s.end_time}</li>)}
        </ul>
      </Card>

      <Card className="space-y-3 p-5">
        <h2 className="font-bold">پیام به داوطلب</h2>
        <p className="text-sm text-stone-500">این پیام در تاریخچه پرونده و برای داوطلب نمایش داده می‌شود.</p>
        <textarea className={inputClass} rows={3} placeholder="متن کامنت یا پیام" value={comment} onChange={(e) => setComment(e.target.value)} />
        <Button disabled={busy || !comment.trim()} onClick={() => run(async () => {
          await api.commentVolunteer(v.id, comment.trim());
          setComment("");
        }, "پیام ارسال شد")}>ارسال پیام</Button>
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">تاریخچه پرونده</h2>
        <HistoryList items={v.history} filterable />
      </Card>

      {msg && <p className="text-sm text-mahak-700">{msg}</p>}

      <Modal open={statusOpen} title="تغییر وضعیت عضویت" onClose={() => setStatusOpen(false)}>
        <p className="text-sm text-stone-600">وضعیت جدید را انتخاب کنید و دلیل را بنویسید. این متن برای داوطلب نمایش داده می‌شود.</p>
        <div className="mt-3 space-y-3">
          <Field label="وضعیت جدید">
            <select className={inputClass} value={status} onChange={(e) => setStatus(e.target.value)}>
              {ADMIN_STATUSES.map((s) => (
                <option key={s} value={s}>{STATUS_LABEL[s] || s}</option>
              ))}
            </select>
          </Field>
          <Field label="دلیل تغییر وضعیت">
            <textarea className={inputClass} rows={4} placeholder="دلیل الزامی است" value={statusReason} onChange={(e) => setStatusReason(e.target.value)} />
          </Field>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" onClick={() => setStatusOpen(false)}>انصراف</Button>
          <Button disabled={busy || !statusReason.trim()} onClick={confirmStatusChange}>ثبت تغییر وضعیت</Button>
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
