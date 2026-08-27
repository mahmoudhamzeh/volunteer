"use client";

import Link from "next/link";
import { Ticket, TicketMessage } from "@/lib/api";
import { TICKET_LABEL, TICKET_SUBJECTS, fmtDate } from "@/lib/labels";
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
  return (
    <button
      type="button"
      onClick={() => onOpen(t.id)}
      className={`w-full rounded-2xl border px-3 py-2.5 text-right transition ${
        selected ? "border-mahak-300 bg-mahak-50" : "border-stone-100 bg-white hover:border-mahak-200 hover:bg-stone-50"
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="truncate font-medium">{t.subject || "بدون موضوع"}</span>
        <Badge status={t.status} />
      </div>
      <div className="mt-1 text-xs text-stone-500">
        {showVolunteer && t.volunteer_name ? `${t.volunteer_name} · ` : ""}
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
      className="grid gap-3 md:grid-cols-[220px_1fr_auto] md:items-end"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <Field label="موضوع">
        <select className={inputClass} value={subject} onChange={(e) => onSubject(e.target.value)} required>
          <option value="">انتخاب موضوع…</option>
          {TICKET_SUBJECTS.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </Field>
      <Field label="متن پرسش">
        <textarea className={inputClass} rows={2} value={body} onChange={(e) => onBody(e.target.value)} placeholder="شرح کوتاه پرسش" required />
      </Field>
      <Button type="submit" disabled={busy}>ارسال تیکت</Button>
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
  return (
    <div className="flex h-full min-h-[420px] flex-col">
      <div className="flex flex-wrap items-start justify-between gap-2 border-b border-stone-100 pb-3">
        <div>
          <div className="text-xs text-stone-500">موضوع</div>
          <h2 className="font-bold">{ticket.subject}</h2>
          {ticket.volunteer_name && (
            volunteerHref
              ? <Link className="text-sm text-mahak-700" href={volunteerHref}>{ticket.volunteer_name}</Link>
              : <p className="text-sm text-stone-500">{ticket.volunteer_name}</p>
          )}
        </div>
        <Badge status={ticket.status} />
      </div>
      <div className="mt-3 min-h-0 flex-1 space-y-2 overflow-y-auto">
        {(ticket.messages || []).map((m) => (
          <TicketBubble key={m.id} message={m} />
        ))}
        {(ticket.messages || []).length === 0 && (
          <p className="text-sm text-stone-400">هنوز پیامی نیست.</p>
        )}
      </div>
      {error && <p className="mt-2 text-sm text-rose-600">{error}</p>}
      {closed ? (
        <p className="mt-3 rounded-2xl bg-stone-50 px-3 py-2 text-sm text-stone-600">
          این تیکت {TICKET_LABEL.closed} است و امکان ارسال پیام وجود ندارد.
        </p>
      ) : (
        <div className="mt-3 space-y-2 border-t border-stone-100 pt-3">
          <textarea className={inputClass} rows={3} value={reply} onChange={(e) => onReply(e.target.value)} placeholder="پیام خود را بنویسید" />
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
    <div className={`flex ${staff ? "justify-end" : "justify-start"}`}>
      <div className={`max-w-[85%] rounded-2xl px-3 py-2 text-sm ${staff ? "bg-mahak-50 text-ink-900" : "bg-stone-100 text-ink-900"}`}>
        <div className="text-[11px] text-stone-500">
          {staff ? "پشتیبانی" : "داوطلب"} · {fmtDate(m.created_at)}
        </div>
        <p className="mt-0.5 whitespace-pre-wrap leading-6">{m.body}</p>
      </div>
    </div>
  );
}
