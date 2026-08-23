"use client";

import { FormEvent, useEffect, useState } from "react";
import { api, Ticket } from "@/lib/api";
import { TICKET_LABEL, fmtDate } from "@/lib/labels";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";

export default function VolunteerTickets() {
  const [items, setItems] = useState<Ticket[]>([]);
  const [openId, setOpenId] = useState("");
  const [detail, setDetail] = useState<Ticket | null>(null);
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [reply, setReply] = useState("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");

  async function load() {
    setItems((await api.myTickets()) || []);
  }
  useEffect(() => { void load(); }, []);

  async function openTicket(id: string) {
    setOpenId(id);
    setDetail(await api.getTicket(id));
  }

  async function create(e: FormEvent) {
    e.preventDefault();
    setErr("");
    try {
      const t = await api.createTicket(subject, body);
      setSubject("");
      setBody("");
      setMsg("تیکت ارسال شد");
      await load();
      await openTicket(t.id);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">پشتیبانی</h1>
        <p className="mt-1 text-sm text-stone-500">سوال خود را برای ادمین بفرستید؛ پاسخ در همین صفحه و اعلان‌ها می‌آید.</p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      <Card className="p-5">
        <h2 className="mb-3 font-bold">تیکت جدید</h2>
        <form onSubmit={create} className="space-y-3">
          <Field label="موضوع">
            <input className={inputClass} value={subject} onChange={(e) => setSubject(e.target.value)} />
          </Field>
          <Field label="متن پرسش">
            <textarea className={inputClass} rows={4} value={body} onChange={(e) => setBody(e.target.value)} />
          </Field>
          <Button type="submit">ارسال تیکت</Button>
        </form>
      </Card>
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <h2 className="mb-3 font-bold">تیکت‌های من</h2>
          {(items || []).length === 0 && <p className="text-sm text-stone-400">هنوز تیکتی نیست.</p>}
          <ul className="space-y-2">
            {(items || []).map((t) => (
              <li key={t.id}>
                <button type="button" className="w-full rounded-2xl border border-stone-100 px-3 py-2 text-right text-sm hover:bg-mahak-50" onClick={() => void openTicket(t.id)}>
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-medium">{t.subject}</span>
                    <Badge status={t.status} />
                  </div>
                  <div className="text-xs text-stone-400">{fmtDate(t.updated_at)}</div>
                </button>
              </li>
            ))}
          </ul>
        </Card>
        <Card className="p-5">
          {!detail && <p className="text-sm text-stone-400">یک تیکت را انتخاب کنید.</p>}
          {detail && (
            <div className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <h2 className="font-bold">{detail.subject}</h2>
                <span className="text-xs text-stone-500">{TICKET_LABEL[detail.status] || detail.status}</span>
              </div>
              <div className="max-h-80 space-y-2 overflow-y-auto">
                {(detail.messages || []).map((m) => (
                  <div key={m.id} className={`rounded-2xl px-3 py-2 text-sm ${m.author_role === "admin" ? "bg-mahak-50" : "bg-stone-50"}`}>
                    <div className="text-xs text-stone-400">{m.author_role === "admin" ? "ادمین" : "شما"} · {fmtDate(m.created_at)}</div>
                    <p className="whitespace-pre-wrap">{m.body}</p>
                  </div>
                ))}
              </div>
              {detail.status === "closed" ? (
                <p className="rounded-2xl bg-rose-50 px-3 py-2 text-sm text-rose-800">
                  این تیکت بسته شده است و امکان ارسال پیام وجود ندارد.
                </p>
              ) : (
                <div className="space-y-2">
                  <textarea className={inputClass} rows={3} value={reply} onChange={(e) => setReply(e.target.value)} placeholder="پاسخ شما" />
                  <Button onClick={async () => {
                    setErr("");
                    try {
                      const t = await api.replyTicket(openId, reply);
                      setReply("");
                      setDetail(t);
                      await load();
                    } catch (e) {
                      setErr(e instanceof Error ? e.message : "خطا");
                      try { setDetail(await api.getTicket(openId)); } catch { /* ignore */ }
                    }
                  }}>ارسال پاسخ</Button>
                </div>
              )}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
