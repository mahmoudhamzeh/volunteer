"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment, Certificate, CertificateRequest } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, inputClass } from "@/components/ui";

export default function CertsPage() {
  const [items, setItems] = useState<Certificate[]>([]);
  const [requests, setRequests] = useState<CertificateRequest[]>([]);
  const [work, setWork] = useState<Assignment[]>([]);
  const [kind, setKind] = useState("aggregated");
  const [assignmentID, setAssignmentID] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const [okOpen, setOkOpen] = useState(false);

  async function load() {
    const [certs, reqs, assignments] = await Promise.all([
      api.myCerts().catch(() => [] as Certificate[]),
      api.myCertRequests().catch(() => [] as CertificateRequest[]),
      api.myAssignments().catch(() => [] as Assignment[]),
    ]);
    setItems(certs || []);
    setRequests(reqs || []);
    setWork(assignments || []);
  }

  useEffect(() => { void load(); }, []);

  const completed = useMemo(
    () => (work || []).filter((a) => a.status === "completed"),
    [work],
  );
  const pendingAssignment = useMemo(
    () => new Set((requests || []).filter((r) => r.status === "pending" && r.assignment_id).map((r) => r.assignment_id as string)),
    [requests],
  );

  async function submit() {
    setBusy(true);
    setErr("");
    try {
      await api.requestCertificate(kind, kind === "task" ? assignmentID : undefined);
      setOkOpen(true);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در ثبت درخواست");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">گواهی‌های داوطلبی</h1>
        <p className="mt-1 text-sm text-stone-500">
          پس از تکمیل فعالیت می‌توانید صدور گواهی را درخواست کنید. ادمین تایید یا رد می‌کند و نتیجه در اعلان‌ها می‌آید.
        </p>
      </div>
      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}

      <Card className="space-y-3 p-5">
        <h2 className="font-bold">درخواست صدور گواهی‌نامه</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <Field label="نوع گواهی">
            <select className={inputClass} value={kind} onChange={(e) => setKind(e.target.value)}>
              <option value="aggregated">گواهی تجمیعی (همه فعالیت‌های تکمیل‌شده)</option>
              <option value="task">گواهی یک فعالیت مشخص</option>
            </select>
          </Field>
          {kind === "task" && (
            <Field label="فعالیت تکمیل‌شده">
              <select className={inputClass} value={assignmentID} onChange={(e) => setAssignmentID(e.target.value)}>
                <option value="">انتخاب کنید</option>
                {completed.map((a) => (
                  <option key={a.id} value={a.id} disabled={pendingAssignment.has(a.id)}>
                    {a.task?.title || "فعالیت"} — {a.hours_awarded || a.task?.hour_weight || 0} ساعت
                    {pendingAssignment.has(a.id) ? " (در حال بررسی)" : ""}
                  </option>
                ))}
              </select>
            </Field>
          )}
        </div>
        {completed.length === 0 && (
          <p className="text-sm text-stone-500">
            هنوز فعالیت تکمیل‌شده‌ای ندارید. پس از تایید نتیجه در{" "}
            <Link className="text-mahak-700" href="/volunteer/work">کارهای من</Link>
            {" "}این گزینه فعال می‌شود.
          </p>
        )}
        <Button
          disabled={busy || completed.length === 0 || (kind === "task" && !assignmentID)}
          onClick={() => void submit()}
        >
          ارسال درخواست صدور گواهی
        </Button>
      </Card>

      <Modal open={okOpen} title="درخواست ارسال شد" onClose={() => setOkOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست صدور گواهی‌نامه شما ارسال شد و در حال بررسی است. پس از تایید یا رد ادمین، نتیجه در اعلان‌ها و همین صفحه نمایش داده می‌شود.
        </p>
        <div className="mt-4 flex justify-end">
          <Button onClick={() => setOkOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">وضعیت درخواست‌ها</h2>
        {(requests || []).length === 0 && <p className="text-sm text-stone-400">هنوز درخواستی ثبت نشده است.</p>}
        <div className="grid gap-2 md:grid-cols-2">
          {(requests || []).map((r) => (
            <div key={r.id} className="rounded-2xl border border-stone-100 px-3 py-2">
              <div className="flex items-center justify-between gap-2">
                <div className="font-medium">{r.assignment_title || (r.kind === "aggregated" ? "گواهی تجمیعی" : "گواهی فعالیت")}</div>
                <Badge status={r.status} reason={r.admin_note} />
              </div>
              <p className="text-xs text-stone-500">{fmtDate(r.created_at)}</p>
              {r.admin_note && <p className="mt-1 text-sm text-stone-600">پیام ادمین: {r.admin_note}</p>}
            </div>
          ))}
        </div>
      </Card>

      {items.length === 0 && <Card className="p-6 text-stone-500">هنوز گواهی صادر نشده است.</Card>}
      <div className="grid gap-3 md:grid-cols-2">
        {items.map((c) => (
          <Card key={c.id} className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-bold">{c.title}</h2>
                <p className="text-sm text-stone-500">{c.hours} ساعت · {fmtDate(c.issued_at)}</p>
                <p className="mt-1 font-mono text-xs">{c.verification_code}</p>
              </div>
              <Badge status={c.kind} />
            </div>
            <a className="mt-3 inline-block text-sm text-mahak-700" href={`/api/v1/certificates/${c.verification_code}/pdf`} target="_blank">دانلود PDF</a>
            {" · "}
            <a className="text-sm text-mahak-700" href={`/verify/${c.verification_code}`} target="_blank">صفحه استعلام</a>
          </Card>
        ))}
      </div>
    </div>
  );
}
