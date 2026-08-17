"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api, Availability, DocumentFile, Volunteer, openAuth } from "@/lib/api";
import { skillLabel, WEEKDAYS } from "@/lib/labels";
import { Badge, Button, Card, inputClass } from "@/components/ui";

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
      <Card className="p-5 text-sm leading-8">
        <div>کد ملی: {v.national_id} · موبایل: {v.phone} · شهر: {v.city}</div>
        <div>رشته: {v.education_field} · نظام پزشکی: {v.medical_license || "—"}</div>
        <div>مهارت‌ها: {(v.skill_categories || []).map(skillLabel).join("، ")}</div>
        <div>{v.bio}</div>
        <div>ساعات: {v.total_hours} · امتیاز: {v.average_score}</div>
      </Card>
      <Card className="p-5">
        <h2 className="font-bold">مدارک</h2>
        <ul className="mt-2 space-y-1 text-sm">
          {docs.map((d) => (
            <li key={d.id}>
              <button className="text-mahak-700" onClick={() => openAuth(`/api/v1/admin/documents/${d.id}`)}>{d.kind} — {d.file_name}</button>
            </li>
          ))}
          {docs.length === 0 && <li className="text-stone-400">مدرکی بارگذاری نشده</li>}
        </ul>
      </Card>
      <Card className="p-5">
        <h2 className="font-bold">زمان‌های آزاد</h2>
        <ul className="mt-2 text-sm">
          {slots.map((s, i) => <li key={i}>{WEEKDAYS[s.weekday]} {s.start_time} تا {s.end_time}</li>)}
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
