"use client";

import { useEffect, useMemo, useState } from "react";
import { api, Ticket } from "@/lib/api";
import { TICKET_SUBJECTS } from "@/lib/labels";
import { Card, inputClass } from "@/components/ui";
import { TicketComposer, TicketListItem, TicketThread } from "@/components/tickets";

export default function VolunteerTickets() {
  const [items, setItems] = useState<Ticket[]>([]);
  const [openId, setOpenId] = useState("");
  const [detail, setDetail] = useState<Ticket | null>(null);
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [reply, setReply] = useState("");
  const [filter, setFilter] = useState("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function load() {
    setItems((await api.myTickets()) || []);
  }
  useEffect(() => { void load(); }, []);

  async function openTicket(id: string) {
    setOpenId(id);
    setDetail(await api.getTicket(id));
  }

  async function create() {
    setErr("");
    if (!subject) {
      setErr("موضوع را از فهرست انتخاب کنید");
      return;
    }
    setBusy(true);
    try {
      const t = await api.createTicket(subject, body);
      setSubject("");
      setBody("");
      setMsg("تیکت ارسال شد");
      await load();
      await openTicket(t.id);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    } finally {
      setBusy(false);
    }
  }

  const visible = useMemo(() => {
    const rank = (s?: string) => (s === "open" ? 0 : s === "answered" ? 1 : 2);
    const list = [...items].sort((a, b) => {
      const r = rank(a.status) - rank(b.status);
      if (r !== 0) return r;
      return (b.updated_at || "").localeCompare(a.updated_at || "");
    });
    if (!filter) return list;
    return list.filter((t) => t.subject === filter);
  }, [items, filter]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">پشتیبانی</h1>
        <p className="mt-1 text-sm text-stone-500">موضوع را از فهرست انتخاب کنید و پرسش را بفرستید. پاسخ در همین صفحه می‌آید.</p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        <h2 className="mb-3 font-bold">تیکت جدید</h2>
        <TicketComposer
          subject={subject}
          body={body}
          onSubject={setSubject}
          onBody={setBody}
          onSubmit={() => void create()}
          busy={busy}
        />
      </Card>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,18rem)_1fr] lg:items-stretch">
        <Card className="flex h-[min(70vh,580px)] flex-col p-4">
          <div className="mb-3 flex shrink-0 items-center justify-between gap-2">
            <h2 className="font-bold">تیکت‌های من</h2>
            <span className="text-xs text-stone-400">{visible.length}</span>
          </div>
          <select className={inputClass + " mb-3 shrink-0"} value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="">همه موضوع‌ها</option>
            {TICKET_SUBJECTS.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
          {visible.length === 0 && <p className="text-sm text-stone-400">تیکتی با این موضوع نیست.</p>}
          <ul className="min-h-0 flex-1 space-y-2 overflow-y-auto">
            {visible.map((t) => (
              <li key={t.id}>
                <TicketListItem ticket={t} selected={t.id === openId} onOpen={(id) => void openTicket(id)} />
              </li>
            ))}
          </ul>
        </Card>
        <Card className="flex h-[min(70vh,580px)] flex-col p-5">
          {!detail && <p className="m-auto py-16 text-center text-sm text-stone-400">یک تیکت را از فهرست انتخاب کنید.</p>}
          {detail && (
            <TicketThread
              ticket={detail}
              reply={reply}
              onReply={setReply}
              sending={busy}
              error={err}
              onSend={async () => {
                setErr("");
                setBusy(true);
                try {
                  const t = await api.replyTicket(openId, reply);
                  setReply("");
                  setDetail(t);
                  await load();
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
