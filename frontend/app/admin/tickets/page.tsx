"use client";

import { useEffect, useMemo, useState } from "react";
import { api, Ticket } from "@/lib/api";
import { TICKET_SUBJECTS } from "@/lib/labels";
import { Card, inputClass } from "@/components/ui";
import { TicketListItem, TicketThread } from "@/components/tickets";

export default function AdminTicketsPage() {
  const [items, setItems] = useState<Ticket[]>([]);
  const [status, setStatus] = useState("open");
  const [q, setQ] = useState("");
  const [qDebounced, setQDebounced] = useState("");
  const [subject, setSubject] = useState("");
  const [openId, setOpenId] = useState("");
  const [detail, setDetail] = useState<Ticket | null>(null);
  const [reply, setReply] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const t = window.setTimeout(() => setQDebounced(q), 280);
    return () => window.clearTimeout(t);
  }, [q]);

  async function load(nextStatus = status, nextQ = qDebounced) {
    setItems((await api.adminTickets(nextStatus, nextQ)) || []);
  }
  useEffect(() => {
    void load(status, qDebounced);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, qDebounced]);

  async function openTicket(id: string) {
    setOpenId(id);
    setDetail(await api.adminTicket(id));
  }

  const visible = useMemo(() => {
    const rank = (s?: string) => (s === "open" ? 0 : s === "answered" ? 1 : 2);
    const list = [...items].sort((a, b) => {
      const r = rank(a.status) - rank(b.status);
      if (r !== 0) return r;
      return (b.updated_at || "").localeCompare(a.updated_at || "");
    });
    if (!subject) return list;
    return list.filter((t) => t.subject === subject);
  }, [items, subject]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">تیکت‌های داوطلبان</h1>
        <p className="mt-1 text-sm text-stone-500">جستجو بر اساس شماره تیکت، نام داوطلب یا موبایل. گفتگو در ستون مقابل باز می‌شود.</p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      <div className="flex flex-wrap items-center gap-2">
        {[["open", "باز"], ["answered", "پاسخ‌داده‌شده"], ["closed", "بسته"], ["", "همه"]].map(([id, label]) => (
          <button
            key={id || "all"}
            onClick={() => setStatus(id)}
            className={`rounded-full px-3 py-1 text-sm ${status === id ? "bg-mahak-500 text-white" : "bg-white"}`}
          >
            {label}
          </button>
        ))}
        <select className={inputClass + " max-w-xs"} value={subject} onChange={(e) => setSubject(e.target.value)}>
          <option value="">همه موضوع‌ها</option>
          {TICKET_SUBJECTS.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>
      <div className="grid gap-4 lg:grid-cols-[minmax(0,22rem)_1fr] lg:items-stretch">
        <Card className="flex h-[min(70vh,580px)] flex-col p-4">
          <div className="mb-3 flex shrink-0 items-center justify-between gap-2">
            <h2 className="font-bold">فهرست</h2>
            <span className="text-xs text-stone-400">{visible.length}</span>
          </div>
          <input
            className={inputClass + " mb-3 shrink-0"}
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="جستجو شماره تیکت، نام یا موبایل"
          />
          {visible.length === 0 && <p className="text-sm text-stone-400">موردی نیست</p>}
          <ul className="min-h-0 flex-1 space-y-2 overflow-y-auto">
            {visible.map((t) => (
              <li key={t.id}>
                <TicketListItem
                  ticket={t}
                  selected={t.id === openId}
                  showVolunteer
                  onOpen={(id) => void openTicket(id)}
                />
              </li>
            ))}
          </ul>
        </Card>
        <Card className="flex h-[min(70vh,580px)] flex-col p-5">
          {!detail && <p className="m-auto py-16 text-center text-sm text-stone-400">یک تیکت را از فهرست انتخاب کنید.</p>}
          {detail && (
            <TicketThread
              ticket={detail}
              volunteerHref={`/admin/volunteers/${detail.volunteer_id}`}
              reply={reply}
              onReply={setReply}
              sending={busy}
              error={err}
              onSend={async () => {
                setErr("");
                setBusy(true);
                try {
                  const t = await api.replyAdminTicket(openId, reply);
                  setReply("");
                  setDetail(t);
                  await load(status, qDebounced);
                } catch (e) {
                  setErr(e instanceof Error ? e.message : "خطا");
                } finally {
                  setBusy(false);
                }
              }}
              onClose={async () => {
                setBusy(true);
                try {
                  await api.setTicketStatus(openId, "closed");
                  await openTicket(openId);
                  await load(status, qDebounced);
                } catch (e) {
                  setErr(e instanceof Error ? e.message : "خطا");
                } finally {
                  setBusy(false);
                }
              }}
            />
          )}
        </Card>
      </div>
    </div>
  );
}
