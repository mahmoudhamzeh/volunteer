"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Ticket } from "@/lib/api";
import { TICKET_LABEL, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, inputClass } from "@/components/ui";

export default function AdminTicketsPage() {
  const [items, setItems] = useState<Ticket[]>([]);
  const [filter, setFilter] = useState("open");
  const [openId, setOpenId] = useState("");
  const [detail, setDetail] = useState<Ticket | null>(null);
  const [reply, setReply] = useState("");
  const [err, setErr] = useState("");

  async function load(status = filter) {
    setItems((await api.adminTickets(status)) || []);
  }
  useEffect(() => { void load("open"); }, []);

  async function openTicket(id: string) {
    setOpenId(id);
    setDetail(await api.adminTicket(id));
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">تیکت‌های داوطلبان</h1>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      <div className="flex flex-wrap gap-2">
        {[["open", "باز"], ["answered", "پاسخ‌داده‌شده"], ["closed", "بسته"], ["", "همه"]].map(([id, label]) => (
          <button key={id || "all"} onClick={() => { setFilter(id); void load(id); }}
            className={`rounded-full px-3 py-1 text-sm ${filter === id ? "bg-mahak-500 text-white" : "bg-white"}`}>{label}</button>
        ))}
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          {(items || []).length === 0 && <p className="text-sm text-stone-400">موردی نیست</p>}
          <ul className="space-y-2">
            {(items || []).map((t) => (
              <li key={t.id}>
                <button type="button" className="w-full rounded-2xl border border-stone-100 px-3 py-2 text-right text-sm hover:bg-mahak-50" onClick={() => void openTicket(t.id)}>
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-medium">{t.subject}</span>
                    <Badge status={t.status} />
                  </div>
                  <div className="text-xs text-stone-500">{t.volunteer_name} · {fmtDate(t.updated_at)}</div>
                </button>
              </li>
            ))}
          </ul>
        </Card>
        <Card className="p-5">
          {!detail && <p className="text-sm text-stone-400">یک تیکت را انتخاب کنید.</p>}
          {detail && (
            <div className="space-y-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="font-bold">{detail.subject}</h2>
                <Link className="text-sm text-mahak-700" href={`/admin/volunteers/${detail.volunteer_id}`}>{detail.volunteer_name}</Link>
              </div>
              <p className="text-xs text-stone-500">{TICKET_LABEL[detail.status]}</p>
              <div className="max-h-80 space-y-2 overflow-y-auto">
                {(detail.messages || []).map((m) => (
                  <div key={m.id} className={`rounded-2xl px-3 py-2 text-sm ${m.author_role === "admin" ? "bg-mahak-50" : "bg-stone-50"}`}>
                    <div className="text-xs text-stone-400">{m.author_role === "admin" ? "ادمین" : "داوطلب"} · {fmtDate(m.created_at)}</div>
                    <p className="whitespace-pre-wrap">{m.body}</p>
                  </div>
                ))}
              </div>
              {detail.status !== "closed" && (
                <div className="space-y-2">
                  <textarea className={inputClass} rows={3} value={reply} onChange={(e) => setReply(e.target.value)} placeholder="پاسخ ادمین" />
                  <div className="flex flex-wrap gap-2">
                    <Button onClick={async () => {
                      try {
                        const t = await api.replyAdminTicket(openId, reply);
                        setReply("");
                        setDetail(t);
                        await load(filter);
                      } catch (e) { setErr(e instanceof Error ? e.message : "خطا"); }
                    }}>ارسال پاسخ</Button>
                    <Button variant="ghost" onClick={async () => {
                      await api.setTicketStatus(openId, "closed");
                      await openTicket(openId);
                      await load(filter);
                    }}>بستن تیکت</Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
