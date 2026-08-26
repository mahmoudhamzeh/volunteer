"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, Notification } from "@/lib/api";
import { fmtDate, notificationHref } from "@/lib/labels";
import { Button, Card } from "@/components/ui";

function daysAgo(iso: string) {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return 999;
  return (Date.now() - t) / 86400000;
}

export default function VolunteerNotifications() {
  const [items, setItems] = useState<Notification[]>([]);
  const [open, setOpen] = useState<Record<string, boolean>>({ week: false, month: false, older: false });

  async function load() {
    setItems((await api.notifications()) || []);
  }
  useEffect(() => { void load(); }, []);

  const unread = useMemo(() => (items || []).filter((n) => !n.read), [items]);
  const read = useMemo(() => (items || []).filter((n) => n.read), [items]);
  const buckets = useMemo(() => ({
    week: read.filter((n) => daysAgo(n.created_at) <= 7),
    month: read.filter((n) => daysAgo(n.created_at) > 7 && daysAgo(n.created_at) <= 30),
    older: read.filter((n) => daysAgo(n.created_at) > 30),
  }), [read]);

  async function openItem(n: Notification) {
    if (!n.read) await api.markRead(n.id).catch(() => undefined);
    await load();
  }

  function Accordion({ id, title, list }: { id: string; title: string; list: Notification[] }) {
    const shown = open[id];
    return (
      <Card className="overflow-hidden">
        <button
          type="button"
          className="flex w-full items-center justify-between px-5 py-3 text-right"
          onClick={() => setOpen({ ...open, [id]: !shown })}
        >
          <span className="font-bold">{title}</span>
          <span className="text-sm text-stone-400">{list.length} مورد {shown ? "▴" : "▾"}</span>
        </button>
        {shown && (
          <ul className="space-y-2 border-t border-stone-100 px-5 py-3">
            {list.length === 0 && <li className="text-sm text-stone-400">موردی نیست</li>}
            {list.map((n) => (
              <li key={n.id} className="rounded-2xl border border-stone-100 px-3 py-2 text-sm text-stone-600">
                <Link href={notificationHref(n.title)} className="block">
                  <div className="font-medium text-ink-800">{n.title}</div>
                  <p>{n.body}</p>
                {n.kind === "reminder" && n.remind_at && (
                    <p className="text-xs text-amber-800">
                      {n.fired_at || new Date(n.remind_at).getTime() <= Date.now() ? "موعد آموزش: " : "زمان یادآوری: "}
                      {fmtDate(n.remind_at)}
                    </p>
                  )}
                  <p className="text-xs text-stone-400">{fmtDate(n.created_at)}</p>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h1 className="text-2xl font-black">اعلان‌ها</h1>
          <p className="mt-1 text-sm text-stone-500">اعلان‌های جدید جدا هستند. پیام‌های خوانده‌شده به‌صورت کشویی بر اساس تاریخ جمع شده‌اند.</p>
        </div>
        {unread.length > 0 && (
          <Button variant="outline" onClick={async () => { await api.markAllRead(); await load(); }}>
            همه را خوانده‌شده کن
          </Button>
        )}
      </div>
      <Card className="p-5">
        <h2 className="mb-3 font-bold">جدید {unread.length ? `(${unread.length})` : ""}</h2>
        {unread.length === 0 && <p className="text-sm text-stone-400">اعلان جدیدی نیست.</p>}
        <ul className="space-y-2">
          {unread.map((n) => (
            <li key={n.id} className="rounded-2xl border border-mahak-100 bg-mahak-50/80 px-3 py-2">
              <Link href={notificationHref(n.title)} onClick={() => void openItem(n)} className="block text-sm">
                <div className="flex items-center gap-2 font-medium">
                  <span className="h-2 w-2 rounded-full bg-rose-500" />
                  {n.title}
                </div>
                <p className="text-stone-600">{n.body}</p>
                {n.kind === "reminder" && n.remind_at && (
                  <p className="text-xs text-amber-800">
                    {n.fired_at || new Date(n.remind_at).getTime() <= Date.now() ? "موعد آموزش: " : "زمان یادآوری: "}
                    {fmtDate(n.remind_at)}
                  </p>
                )}
                <p className="text-xs text-stone-400">{fmtDate(n.created_at)}</p>
              </Link>
            </li>
          ))}
        </ul>
      </Card>
      <Accordion id="week" title="۷ روز گذشته" list={buckets.week} />
      <Accordion id="month" title="ماه گذشته" list={buckets.month} />
      <Accordion id="older" title="قدیمی‌تر" list={buckets.older} />
    </div>
  );
}
