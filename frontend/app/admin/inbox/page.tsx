"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Assignment, CertificateRequest, SkillProposal, Ticket, Volunteer } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Button, Card } from "@/components/ui";

export default function AdminInbox() {
  const [requests, setRequests] = useState<Assignment[]>([]);
  const [volunteers, setVolunteers] = useState<Volunteer[]>([]);
  const [skills, setSkills] = useState<SkillProposal[]>([]);
  const [certs, setCerts] = useState<CertificateRequest[]>([]);
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [msg, setMsg] = useState("");

  async function load() {
    const [a, v, s, c, t] = await Promise.all([
      api.adminAssignments("?status=requested&limit=200").catch(() => ({ items: [] as Assignment[] })),
      api.adminVolunteers("?status=pending&limit=50").catch(() => ({ items: [] as Volunteer[] })),
      api.adminSkillProposals("pending").catch(() => [] as SkillProposal[]),
      api.adminCertRequests("pending").catch(() => [] as CertificateRequest[]),
      api.adminTickets("open").catch(() => [] as Ticket[]),
    ]);
    setRequests(a.items || []);
    setVolunteers(v.items || []);
    setSkills(s || []);
    setCerts(c || []);
    setTickets(t || []);
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
        <h1 className="text-2xl font-black">صندوق درخواست‌ها</h1>
        <p className="mt-1 text-sm text-stone-500">همه موارد نیازمند اقدام در یک صفحه جمع شده‌اند تا لازم نباشد فعالیت‌ها را یکی‌یکی باز کنید.</p>
      </div>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      <div className="grid gap-3 sm:grid-cols-4">
        {[
          ["درخواست فعالیت", requests.length, "/admin/inbox"],
          ["تایید هویت", volunteers.length, "/admin/volunteers?status=pending"],
          ["مهارت پیشنهادی", skills.length, "/admin/skills"],
          ["گواهی / تیکت", certs.length + tickets.length, "/admin/certificates"],
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
                    <td className="px-3 py-2">{a.task?.title}</td>
                    <td className="px-3 py-2 text-stone-500">{fmtDate(a.created_at)}</td>
                    <td className="px-3 py-2">
                      <div className="flex flex-wrap gap-2">
                        <Button onClick={() => void approve(a.id)}>تایید</Button>
                        <Button variant="danger" onClick={async () => { await api.rejectAssignment(a.id); await load(); }}>رد</Button>
                        <Link className="rounded-2xl border border-mahak-200 px-3 py-2 text-sm text-mahak-700" href="/admin/tasks">جزئیات فعالیت</Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
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
          <div className="mb-3 flex justify-between"><h2 className="font-bold">تیکت‌های باز</h2><Link className="text-sm text-mahak-700" href="/admin/tickets">همه</Link></div>
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
