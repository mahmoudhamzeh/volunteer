"use client";

import { EVENT_LABEL, STATUS_LABEL, fmtDate } from "@/lib/labels";
import { VolunteerEvent } from "@/lib/api";

export function HistoryList({ items }: { items?: VolunteerEvent[] }) {
  if (!items?.length) {
    return <p className="text-sm text-stone-400">هنوز رویدادی در تاریخچه ثبت نشده است.</p>;
  }
  return (
    <ol className="space-y-3">
      {items.map((e) => (
        <li key={e.id} className="rounded-2xl border border-stone-100 bg-stone-50 px-3 py-2 text-sm">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="font-medium text-ink-900">{EVENT_LABEL[e.event_type] || e.event_type}</span>
            <span className="text-xs text-stone-500">{fmtDate(e.created_at)}</span>
          </div>
          {e.from_status && e.to_status && e.from_status !== e.to_status && (
            <div className="mt-1 text-xs text-stone-500">
              از {STATUS_LABEL[e.from_status] || e.from_status} به {STATUS_LABEL[e.to_status] || e.to_status}
            </div>
          )}
          {e.comment && <p className="mt-1 whitespace-pre-wrap text-stone-700">{e.comment}</p>}
        </li>
      ))}
    </ol>
  );
}
