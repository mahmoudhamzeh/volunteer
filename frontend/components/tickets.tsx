"use client";

import { useEffect, useRef } from "react";
import Link from "next/link";
import { Ticket, TicketMessage } from "@/lib/api";
import { TICKET_LABEL, TICKET_SUBJECTS, fmtDate, ticketCode } from "@/lib/labels";
import { Badge, Button, Field, inputClass } from "@/components/ui";

export function TicketListItem({
  ticket: t,
  selected,
  showVolunteer,
  onOpen,
}: {
  ticket: Ticket;
  selected?: boolean;
  showVolunteer?: boolean;
  onOpen: (id: string) => void;
}) {
  const code = ticketCode(t.number);
  return (
    <button
      type="button"
      onClick={() => onOpen(t.id)}
      className={`w-full overflow-hidden rounded-2xl border px-3 py-2.5 text-right transition ${
        selected ? "border-mahak-300 bg-mahak-50" : "border-stone-100 bg-white hover:border-mahak-200 hover:bg-stone-50"
      }`}
    >
      <div className="flex min-w-0 items-center justify-between gap-2">
        <span className="flex min-w-0 items-center gap-2">
          {code ? (
            <span className="shrink-0 rounded-full bg-mahak-50 px-2 py-0.5 text-[11px] font-bold text-mahak-800">{code}</span>
          ) : null}
          <span className="min-w-0 truncate font-medium">{t.subject || "بدون موضوع"}</span>
        </span>
        <span className="shrink-0"><Badge status={t.status} /></span>
      </div>
      <div className="mt-1 truncate text-xs text-stone-500">
        {showVolunteer && t.volunteer_name ? `${t.volunteer_name}${t.volunteer_phone ? ` · ${t.volunteer_phone}` : ""} · ` : ""}
        {fmtDate(t.updated_at)}
      </div>
    </button>
  );
}

export function TicketComposer({
  subject,
  body,
  onSubject,
  onBody,
  onSubmit,
  busy,
}: {
  subject: string;
  body: string;
  onSubject: (v: string) => void;
  onBody: (v: string) => void;
  onSubmit: () => void;
  busy?: boolean;
}) {
  return (
    <form
      className="space-y-4"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <Field label="موضوع">
        <select className={inputClass} value={subject} onChange={(e) => onSubject(e.target.value)} required>
          <option value="">موضوع را انتخاب کنید</option>
          {TICKET_SUBJECTS.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </Field>
      <Field label="متن پرسش">
        <textarea
          className={inputClass + " min-h-[128px] max-w-full leading-7"}
          rows={5}
          value={body}
          onChange={(e) => onBody(e.target.value)}
          placeholder="پرسش خود را کامل بنویسید"
          required
        />
      </Field>
      <div className="flex justify-end">
        <Button type="submit" disabled={busy}>ارسال تیکت</Button>
      </div>
    </form>
  );
}

export function TicketThread({
  ticket,
  volunteerHref,
  reply,
  onReply,
  onSend,
  onClose,
  sending,
  error,
}: {
  ticket: Ticket;
  volunteerHref?: string;
  reply: string;
  onReply: (v: string) => void;
  onSend: () => void;
  onClose?: () => void;
  sending?: boolean;
  error?: string;
}) {
  const closed = ticket.status === "closed";
  const scroller = useRef<HTMLDivElement>(null);
  const code = ticketCode(ticket.number);
  useEffect(() => {
    const el = scroller.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [ticket.id, ticket.messages?.length]);

  return (
    <div className="flex h-full min-h-0 w-full flex-col overflow-hidden">
      <div className="flex w-full shrink-0 items-start justify-between gap-2 border-b border-stone-100 pb-3">
        <div className="min-w-0 flex-1 overflow-hidden">
          <div className="flex min-w-0 items-start gap-2">
            {code ? (
              <span className="mt-0.5 shrink-0 rounded-full bg-mahak-50 px-2.5 py-0.5 text-xs font-bold text-mahak-800">{code}</span>
            ) : null}
            <h2 className="ticket-break min-w-0 flex-1 text-base font-bold">{ticket.subject}</h2>
          </div>
          {ticket.volunteer_name && (
            volunteerHref
              ? (
                <Link className="mt-1 block truncate text-sm text-mahak-700" href={volunteerHref}>
                  {ticket.volunteer_name}{ticket.volunteer_phone ? ` · ${ticket.volunteer_phone}` : ""}
                </Link>
              )
              : <p className="mt-1 truncate text-sm text-stone-500">{ticket.volunteer_name}</p>
          )}
        </div>
        <span className="shrink-0"><Badge status={ticket.status} /></span>
      </div>
      <div ref={scroller} className="mt-3 min-h-0 w-full flex-1 space-y-2 overflow-x-hidden overflow-y-auto pe-1 [scrollbar-width:thin]">
        {(ticket.messages || []).map((m) => (
          <TicketBubble key={m.id} message={m} />
        ))}
        {(ticket.messages || []).length === 0 && (
          <p className="text-sm text-stone-400">هنوز پیامی نیست.</p>
        )}
      </div>
      {error && <p className="mt-2 shrink-0 text-sm text-rose-600">{error}</p>}
      {closed ? (
        <p className="mt-3 shrink-0 rounded-2xl bg-stone-50 px-3 py-2 text-sm text-stone-600">
          این تیکت {TICKET_LABEL.closed} است و امکان ارسال پیام وجود ندارد.
        </p>
      ) : (
        <div className="mt-3 shrink-0 space-y-2 border-t border-stone-100 pt-3">
          <textarea className={inputClass + " max-w-full"} rows={3} value={reply} onChange={(e) => onReply(e.target.value)} placeholder="پیام خود را بنویسید" />
          <div className="flex flex-wrap gap-2">
            <Button disabled={sending} onClick={onSend}>ارسال پیام</Button>
            {onClose && (
              <Button variant="ghost" disabled={sending} onClick={onClose}>بستن تیکت</Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function TicketBubble({ message: m }: { message: TicketMessage }) {
  const staff = m.author_role === "admin" || m.author_role === "operator";
  return (
    <div className={`flex w-full min-w-0 ${staff ? "justify-end" : "justify-start"}`}>
      <div className={`min-w-0 max-w-[85%] overflow-hidden rounded-2xl px-3.5 py-2.5 text-sm leading-7 ${staff ? "bg-mahak-50 text-ink-900" : "bg-stone-100 text-ink-900"}`}>
        <div className="text-[11px] text-stone-500">
          {staff ? "پشتیبانی" : "داوطلب"} · {fmtDate(m.created_at)}
        </div>
        <p className="ticket-break mt-1 whitespace-pre-wrap">{m.body}</p>
      </div>
    </div>
  );
}
