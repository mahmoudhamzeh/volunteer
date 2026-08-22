"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Assignment, CertificateRequest, SkillProposal, Ticket, Volunteer } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Button, Card } from "@/components/ui";

function activityHref(a: Assignment) {
  const sid = a.task?.series_id;
  const id = a.task?.kind === "occurrence" && sid ? sid : a.task_id;
  return `/admin/tasks?manage=${id}`;
}

export default function AdminInbox() {
  const [requests, setRequests] = useState<Assignment[]>([]);
  const [deliveries, setDeliveries] = useState<Assignment[]>([]);
  const [volunteers, setVolunteers] = useState<Volunteer[]>([]);
  const [resubmitted, setResubmitted] = useState<Volunteer[]>([]);
  const [skills, setSkills] = useState<SkillProposal[]>([]);
  const [certs, setCerts] = useState<CertificateRequest[]>([]);
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [msg, setMsg] = useState("");

  async function load() {
    const [a, v, s, c, t, rs, d] = await Promise.all([
      api.adminAssignments("?status=requested&limit=200").catch(() => ({ items: [] as Assignment[] })),
      api.adminVolunteers("?status=pending&limit=50").catch(() => ({ items: [] as Volunteer[] })),
      api.adminSkillProposals("pending").catch(() => [] as SkillProposal[]),
      api.adminCertRequests("pending").catch(() => [] as CertificateRequest[]),
      api.adminTickets("open").catch(() => [] as Ticket[]),
      api.adminVolunteers("?attention=resubmitted&limit=50").catch(() => ({ items: [] as Volunteer[] })),
      api.adminAssignments("?status=submitted&limit=200").catch(() => ({ items: [] as Assignment[] })),
    ]);
    setRequests(a.items || []);
    setVolunteers(v.items || []);
    setSkills(s || []);
    setCerts(c || []);
    setTickets(t || []);
    setResubmitted(rs.items || []);
    setDeliveries(d.items || []);
  }
  useEffect(() => { void load(); }, []);

  async function approve(id: string) {
    try {
      await api.approveAssignment(id);
      setMsg("درخواست تایید شد");
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">درخواست‌ها</h1>
        <p className="mt-1 text-sm text-stone-500">همه موارد نیازمند اقدام در یک صفحه جمع شده‌اند.</p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      <div className="grid gap-3 sm:grid-cols-3 lg:grid-cols-6">
        {[
          ["درخواست فعالیت", requests.length, "/admin/inbox"],
          ["نتیجه ارسال‌شده", deliveries.length, "/admin/assignments"],
          ["تایید هویت", volunteers.length, "/admin/volunteers?status=pending"],
          ["مدارک اصلاح‌شده", resubmitted.length, "/admin/volunteers?attention=resubmitted"],
          ["مهارت پیشنهادی", skills.length, "/admin/skills"],
          ["درخواست گواهی", certs.length, "/admin/certificates"],
        ].map(([k, n, href]) => (
          <Link key={String(k)} href={String(href)}>
            <Card className="p-4">
              <div className="text-xs text-stone-500">{k}</div>
              <div className="mt-1 text-2xl font-black text-mahak-700">{n as number}</div>
            </Card>
          </Link>
        ))}
      </div>

      <Card className="overflow-hidden">
        <div className="border-b border-stone-100 px-4 py-3 font-bold">درخواست فعالیت ({requests.length})</div>
        {requests.length === 0 && <p className="px-4 py-5 text-sm text-stone-400">درخواست فعالی در انتظار نیست.</p>}
        {requests.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[720px] text-sm">
              <thead className="bg-stone-50 text-right text-xs text-stone-500">
                <tr>
                  <th className="px-3 py-2">داوطلب</th>
                  <th className="px-3 py-2">فعالیت</th>
                  <th className="px-3 py-2">زمان</th>
                  <th className="px-3 py-2">اقدام</th>
                </tr>
              </thead>
              <tbody>
                {requests.map((a) => (
                  <tr key={a.id} className="border-t border-stone-100">
                    <td className="px-3 py-2">
                      <Link className="text-mahak-700" href={`/admin/volunteers/${a.volunteer_id}`}>{a.volunteer?.full_name || "داوطلب"}</Link>
                      <div className="text-xs text-stone-400">{a.volunteer?.phone}</div>
                    </td>
                    <td className="px-3 py-2">
                      <div>{a.task?.title}</div>
                      {a.task?.starts_at && (
                        <div className="text-xs text-stone-400">{fmtDate(a.task.starts_at)}</div>
                      )}
                    </td>
                    <td className="px-3 py-2 text-stone-500">{fmtDate(a.created_at)}</td>
                    <td className="px-3 py-2">
                      <div className="flex flex-wrap gap-2">
                        <Button onClick={() => void approve(a.id)}>تایید</Button>
                        <Button variant="danger" onClick={async () => { await api.rejectAssignment(a.id); await load(); }}>رد</Button>
                        <Link className="rounded-2xl border border-mahak-200 px-3 py-2 text-sm text-mahak-700" href={activityHref(a)}>جزئیات فعالیت</Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b border-stone-100 px-4 py-3">
          <h2 className="font-bold">نتیجه ارسال‌شده توسط داوطلب ({deliveries.length})</h2>
          <Link className="text-sm text-mahak-700" href="/admin/assignments">حضور و امتیاز</Link>
        </div>
        {deliveries.length === 0 && <p className="px-4 py-5 text-sm text-stone-400">نتیجه‌ای در انتظار بررسی نیست.</p>}
        {deliveries.map((a) => (
          <Link key={a.id} href="/admin/assignments" className="flex items-center justify-between border-t border-stone-100 px-4 py-3 text-sm hover:bg-stone-50">
            <div>
              <div className="font-medium">{a.volunteer?.full_name || "داوطلب"} — {a.task?.title}</div>
              <div className="text-xs text-stone-500">{a.delivery_note ? `شرح: ${a.delivery_note}` : "نتیجه ارسال شده"} · {fmtDate(a.created_at)}</div>
            </div>
            <Badge status={a.status} />
          </Link>
        ))}
      </Card>

      <Card className="p-5">
        <div className="mb-3 flex justify-between">
          <h2 className="font-bold">مدارک اصلاح‌شده پس از رد / نقص مدرک ({resubmitted.length})</h2>
          <Link className="text-sm text-mahak-700" href="/admin/volunteers?attention=resubmitted">همه</Link>
        </div>
        {resubmitted.length === 0 && <p className="text-sm text-stone-400">بارگذاری مجددی در انتظار بررسی نیست.</p>}
        {resubmitted.map((v) => (
          <Link key={v.id} href={`/admin/volunteers/${v.id}`} className="mb-2 flex items-center justify-between rounded-xl border border-amber-100 bg-amber-50 px-3 py-2 text-sm">
            <div>
              <div className="font-medium">{v.full_name}</div>
              <div className="text-xs text-stone-500">مدارک دوباره بارگذاری شده — برای بررسی پرونده را باز کنید</div>
            </div>
            <Badge status={v.status} />
          </Link>
        ))}
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <div className="mb-3 flex justify-between"><h2 className="font-bold">تایید هویت</h2><Link className="text-sm text-mahak-700" href="/admin/volunteers?status=pending">همه</Link></div>
          {(volunteers || []).slice(0, 6).map((v) => (
            <Link key={v.id} href={`/admin/volunteers/${v.id}`} className="mb-2 flex items-center justify-between rounded-xl bg-stone-50 px-3 py-2 text-sm">
              <span>{v.full_name}</span>
              <Badge status={v.status} />
            </Link>
          ))}
          {volunteers.length === 0 && <p className="text-sm text-stone-400">موردی نیست</p>}
        </Card>
        <Card className="p-5">
          <div className="mb-3 flex justify-between">
            <h2 className="font-bold">درخواست گواهی ({certs.length})</h2>
            <Link className="text-sm text-mahak-700" href="/admin/certificates">همه</Link>
          </div>
          {certs.length === 0 && <p className="mb-3 text-sm text-stone-400">درخواست گواهی در انتظار نیست</p>}
          {certs.slice(0, 6).map((r) => (
            <Link key={r.id} href="/admin/certificates" className="mb-2 block rounded-xl bg-amber-50 px-3 py-2 text-sm">
              <div className="font-medium">{r.volunteer_name || "داوطلب"}</div>
              <div className="text-xs text-stone-500">{r.assignment_title || (r.kind === "aggregated" ? "گواهی تجمیعی" : "گواهی فعالیت")} · {fmtDate(r.created_at)}</div>
            </Link>
          ))}
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <div className="mb-3 flex justify-between"><h2 className="font-bold">مهارت پیشنهادی</h2><Link className="text-sm text-mahak-700" href="/admin/skills">همه</Link></div>
          {skills.length === 0 && <p className="text-sm text-stone-400">موردی نیست</p>}
          {skills.slice(0, 6).map((p) => (
            <Link key={p.id} href="/admin/skills" className="mb-2 block rounded-xl bg-stone-50 px-3 py-2 text-sm">
              {p.title}
            </Link>
          ))}
        </Card>
        <Card className="p-5">
          <div className="mb-3 flex justify-between"><h2 className="font-bold">تیکت باز ({tickets.length})</h2><Link className="text-sm text-mahak-700" href="/admin/tickets">همه</Link></div>
          {(tickets || []).slice(0, 6).map((t) => (
            <Link key={t.id} href="/admin/tickets" className="mb-2 block rounded-xl bg-stone-50 px-3 py-2 text-sm">
              <div className="font-medium">{t.subject}</div>
              <div className="text-xs text-stone-500">{t.volunteer_name} · {fmtDate(t.updated_at)}</div>
            </Link>
          ))}
          {tickets.length === 0 && <p className="text-sm text-stone-400">تیکت بازی نیست</p>}
        </Card>
      </div>
    </div>
  );
}
