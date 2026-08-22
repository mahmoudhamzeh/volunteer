"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, Notification } from "@/lib/api";
import { fmtDate, notificationHref } from "@/lib/labels";
import { Button, Card } from "@/components/ui";

export default function VolunteerNotifications() {
  const [items, setItems] = useState<Notification[]>([]);

  async function load() {
    setItems((await api.notifications()) || []);
  }
  useEffect(() => { void load(); }, []);

  const unread = useMemo(() => (items || []).filter((n) => !n.read), [items]);
  const read = useMemo(() => (items || []).filter((n) => n.read), [items]);

  async function openItem(n: Notification) {
    if (!n.read) await api.markRead(n.id).catch(() => undefined);
    await load();
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h1 className="text-2xl font-black">اعلان‌ها</h1>
          <p className="mt-1 text-sm text-stone-500">اعلان‌های جدید جدا از پیام‌های قبلی نمایش داده می‌شوند.</p>
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
                <p className="text-xs text-stone-400">{fmtDate(n.created_at)}</p>
              </Link>
            </li>
          ))}
        </ul>
      </Card>
      <Card className="p-5">
        <h2 className="mb-3 font-bold">قبلی</h2>
        {read.length === 0 && <p className="text-sm text-stone-400">اعلان قدیمی‌تری نیست.</p>}
        <ul className="space-y-2">
          {read.map((n) => (
            <li key={n.id} className="rounded-2xl border border-stone-100 px-3 py-2 text-sm text-stone-600">
              <Link href={notificationHref(n.title)} className="block">
                <div className="font-medium text-ink-800">{n.title}</div>
                <p>{n.body}</p>
                <p className="text-xs text-stone-400">{fmtDate(n.created_at)}</p>
              </Link>
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
