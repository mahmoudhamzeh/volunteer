"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api, Availability, DocumentFile, Volunteer, openAuth } from "@/lib/api";
import { PROPOSAL_LABEL, WEEKDAYS, docKindLabel, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, inputClass } from "@/components/ui";
import { ReactNode } from "react";

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
  const [reason, setReason] = useState("");
  const [msg, setMsg] = useState("");

  async function load() {
    const r = await api.adminVolunteer(id);
    setV(r.volunteer);
    setDocs(r.documents || []);
    setSlots(r.availability || []);
  }
  useEffect(() => { if (id) void load(); }, [id]);

  async function act(action: string) {
    try {
      await api.review(id, action, reason);
      setMsg("ثبت شد");
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  if (!v) return null;
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-black">{v.full_name}</h1>
        <Badge status={v.status} />
      </div>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">هویت</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Row label="نام و نام خانوادگی" value={v.full_name} />
          <Row label="کد ملی" value={v.national_id} />
          <Row label="تاریخ تولد" value={fmtDate(v.birth_date)} />
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">تماس</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Row label="موبایل" value={v.phone} />
          <Row label="موبایل دوم" value={v.phone2} />
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">نشانی</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Row label="استان" value={v.province} />
          <Row label="شهر" value={v.city} />
          <Row label="پلاک" value={v.plaque} />
          <Row label="واحد" value={v.unit} />
          <div className="sm:col-span-2 lg:col-span-3">
            <Row label="آدرس" value={v.address} />
          </div>
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">تحصیلات</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Row label="مقطع" value={v.education_level} />
          <Row label="رشته" value={v.education_field} />
          <Row label="نظام پزشکی" value={v.medical_license} />
        </div>
        {v.bio && (
          <div className="mt-4">
            <Row label="درباره داوطلب" value={v.bio} />
          </div>
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
              <button className="text-mahak-700" onClick={() => openAuth(`/api/v1/admin/documents/${d.id}`)}>
                {docKindLabel(d.kind)} — {d.file_name}
              </button>
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

      <Card className="p-5 space-y-3">
        <textarea className={inputClass} rows={3} placeholder="دلیل رد یا نقص مدرک" value={reason} onChange={(e) => setReason(e.target.value)} />
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => act("approve")}>تایید نهایی</Button>
          <Button variant="outline" onClick={() => act("request_documents")}>نیاز به مدارک بیشتر</Button>
          <Button variant="danger" onClick={() => act("reject")}>رد</Button>
          {v.status === "approved" && <Button variant="ghost" onClick={() => act("suspend")}>تعلیق</Button>}
          {v.status === "suspended" && <Button variant="ghost" onClick={() => act("unsuspend")}>رفع تعلیق</Button>}
          <Button variant="ghost" onClick={() => api.issueAggregated(v.id).then(() => setMsg("گواهی تجمیعی صادر شد"))}>صدور گواهی تجمیعی</Button>
        </div>
        {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      </Card>
    </div>
  );
}
