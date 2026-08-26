"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { api, Assignment, Certificate, CertificateRequest, Volunteer } from "@/lib/api";
import { CERT_REQ_LABEL, certKindLabel, certRequestTitle, deliveryMethodLabel, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, Modal, inputClass } from "@/components/ui";
import { AppreciationCard } from "@/components/appreciation-card";

const OFFICIAL_HOURS = 90;

export default function CertsPage() {
  const [items, setItems] = useState<Certificate[]>([]);
  const [requests, setRequests] = useState<CertificateRequest[]>([]);
  const [work, setWork] = useState<Assignment[]>([]);
  const [me, setMe] = useState<Volunteer | null>(null);
  const [kind, setKind] = useState("aggregated");
  const [assignmentID, setAssignmentID] = useState("");
  const [busy, setBusy] = useState("");
  const [err, setErr] = useState("");
  const [okOpen, setOkOpen] = useState(false);

  async function load() {
    const [certs, reqs, assignments, profile] = await Promise.all([
      api.myCerts().catch(() => [] as Certificate[]),
      api.myCertRequests().catch(() => [] as CertificateRequest[]),
      api.myAssignments().catch(() => [] as Assignment[]),
      api.me().catch(() => ({ volunteer: undefined })),
    ]);
    setItems(certs || []);
    setRequests(reqs || []);
    setWork(assignments || []);
    setMe(profile.volunteer || null);
  }

  useEffect(() => { void load(); }, []);

  const completed = useMemo(
    () => (work || []).filter((a) => a.status === "completed"),
    [work],
  );
  const pendingAssignment = useMemo(
    () => new Set((requests || []).filter((r) => (r.status === "pending" || r.status === "preparing") && r.assignment_id).map((r) => r.assignment_id as string)),
    [requests],
  );
  const hours = me?.total_hours || 0;
  const canOfficial = hours >= OFFICIAL_HOURS;
  const officialOpen = (requests || []).some((r) => r.kind === "official" && ["pending", "preparing", "ready"].includes(r.status));
  const appreciation = (items || []).filter((c) => c.kind !== "official");
  const officialCerts = (items || []).filter((c) => c.kind === "official");
  const appreciationReqs = (requests || []).filter((r) => r.kind !== "official");
  const officialReqs = (requests || []).filter((r) => r.kind === "official");

  async function submitAppreciation() {
    setBusy("appreciation");
    setErr("");
    try {
      await api.requestCertificate(kind, kind === "task" ? assignmentID : undefined);
      setOkOpen(true);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در ثبت درخواست");
    } finally {
      setBusy("");
    }
  }

  async function submitOfficial() {
    setBusy("official");
    setErr("");
    try {
      await api.requestCertificate("official");
      setOkOpen(true);
      await load();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا در ثبت درخواست");
    } finally {
      setBusy("");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">تقدیرنامه و گواهی‌نامه</h1>
        <p className="mt-1 text-sm text-stone-500">
          تقدیرنامه همان مدرک سامانه‌ای هر فعالیت است. گواهی‌نامه فعالیت داوطلبانه پس از حداقل {OFFICIAL_HOURS} ساعت فعالیت تاییدشده درخواست می‌شود و پس از آماده‌سازی، ارسال یا حضوری تحویل می‌گردد.
        </p>
      </div>
      {err && <p className="text-sm font-medium text-rose-600">{err}</p>}

      <Card className="space-y-3 p-5">
        <h2 className="font-bold">درخواست تقدیرنامه</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <Field label="نوع تقدیرنامه">
            <select className={inputClass} value={kind} onChange={(e) => setKind(e.target.value)}>
              <option value="aggregated">تقدیرنامه تجمیعی (همه فعالیت‌های تکمیل‌شده)</option>
              <option value="task">تقدیرنامه یک فعالیت مشخص</option>
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
          disabled={busy !== "" || completed.length === 0 || (kind === "task" && !assignmentID)}
          onClick={() => void submitAppreciation()}
        >
          ارسال درخواست تقدیرنامه
        </Button>
      </Card>

      <Card className="space-y-3 p-5">
        <h2 className="font-bold">گواهی‌نامه فعالیت داوطلبانه</h2>
        <p className="text-sm text-stone-600">
          ساعات تاییدشده شما: <b>{hours}</b> از {OFFICIAL_HOURS} ساعت.
        </p>
        <div className="h-2 overflow-hidden rounded-full bg-stone-100">
          <div className="h-full rounded-full bg-mahak-500" style={{ width: `${Math.min(100, Math.round((hours / OFFICIAL_HOURS) * 100))}%` }} />
        </div>
        {!canOfficial && (
          <p className="text-sm text-stone-500">پس از رسیدن به {OFFICIAL_HOURS} ساعت فعالیت تاییدشده می‌توانید درخواست ثبت کنید.</p>
        )}
        {officialOpen && <p className="text-sm text-mahak-700">یک درخواست باز دارید که در حال آماده‌سازی یا آماده تحویل است.</p>}
        <Button disabled={busy !== "" || !canOfficial || officialOpen} onClick={() => void submitOfficial()}>
          ثبت درخواست گواهی‌نامه
        </Button>
      </Card>

      <Modal open={okOpen} title="درخواست ثبت شد" onClose={() => setOkOpen(false)}>
        <p className="text-sm leading-7 text-stone-700">
          درخواست شما ثبت شد. وضعیت آماده‌سازی، صدور و تحویل در همین صفحه به‌روز می‌شود.
        </p>
        <div className="mt-4 flex justify-end">
          <Button onClick={() => setOkOpen(false)}>متوجه شدم</Button>
        </div>
      </Modal>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">وضعیت درخواست‌ها</h2>
        {(requests || []).length === 0 && <p className="text-sm text-stone-400">هنوز درخواستی ثبت نشده است.</p>}
        <div className="grid gap-2 md:grid-cols-2">
          {(appreciationReqs.concat(officialReqs) || []).map((r) => (
            <div key={r.id} className="rounded-2xl border border-stone-100 px-3 py-2">
              <div className="flex items-center justify-between gap-2">
                <div className="font-medium">{certRequestTitle(r)}</div>
                <Badge status={r.status} reason={r.admin_note} label={CERT_REQ_LABEL[r.status]} />
              </div>
              <p className="text-xs text-stone-500">{fmtDate(r.created_at)}</p>
              {r.delivery_method && <p className="mt-1 text-xs text-stone-600">{deliveryMethodLabel(r.delivery_method)}</p>}
              {r.admin_note && <p className="mt-1 text-sm text-stone-600">پیام پشتیبانی: {r.admin_note}</p>}
            </div>
          ))}
        </div>
      </Card>

      {appreciation.length === 0 && officialCerts.length === 0 && <Card className="p-6 text-stone-500">هنوز تقدیرنامه یا گواهی‌نامه‌ای صادر نشده است.</Card>}
      {appreciation.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          {appreciation.map((c) => (
            <AppreciationCard key={c.id} cert={{ ...c, volunteer_name: c.volunteer_name || me?.full_name }} />
          ))}
        </div>
      )}
      <div className="grid gap-3 md:grid-cols-2">
        {officialCerts.map((c) => (
          <Card key={c.id} className="p-5">
            <h2 className="font-bold">{c.title}</h2>
            <p className="text-sm text-stone-500">{c.hours} ساعت · {fmtDate(c.issued_at)}</p>
            <p className="mt-1 text-xs text-stone-500">{certKindLabel(c.kind)}</p>
            <p className="mt-3 text-sm text-stone-600">این گواهی‌نامه به‌صورت ارسال یا تحویل حضوری ارائه می‌شود و از پنل دانلود نمی‌شود.</p>
            <a className="mt-3 inline-block text-sm text-mahak-700" href={`/verify/${c.verification_code}`} target="_blank">صفحه استعلام</a>
          </Card>
        ))}
      </div>
    </div>
  );
}
