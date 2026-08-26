"use client";

import { useMemo, useState } from "react";
import { EVENT_LABEL, STATUS_LABEL, fmtDate } from "@/lib/labels";
import { VolunteerEvent } from "@/lib/api";

const COMMENT_TYPES = new Set(["rejected", "documents_requested", "comment", "suspended", "unsuspended", "document_uploaded"]);
const VOLUNTEER_TYPES = new Set(["submitted", "approved", "rejected", "documents_requested", "comment", "skill_proposal", "certificate", "ticket"]);

function dayKey(iso: string) {
  try {
    return new Date(iso).toLocaleDateString("fa-IR-u-ca-persian", { year: "numeric", month: "long", day: "numeric" });
  } catch {
    return iso.slice(0, 10);
  }
}

export function HistoryList({
  items,
  filterable,
  audience,
}: {
  items?: VolunteerEvent[];
  filterable?: boolean;
  audience?: "admin" | "volunteer";
}) {
  const [type, setType] = useState("");
  const source = useMemo(() => {
    const list = items || [];
    if (audience === "volunteer") return list.filter((e) => VOLUNTEER_TYPES.has(e.event_type));
    return list;
  }, [items, audience]);
  const types = useMemo(() => {
    const set = new Set(source.map((e) => e.event_type).filter(Boolean));
    return Array.from(set);
  }, [source]);
  const visible = useMemo(() => {
    if (!type) return source;
    return source.filter((e) => e.event_type === type);
  }, [source, type]);
  const groups = useMemo(() => {
    const map = new Map<string, VolunteerEvent[]>();
    for (const e of visible) {
      const key = dayKey(e.created_at);
      const cur = map.get(key) || [];
      cur.push(e);
      map.set(key, cur);
    }
    return Array.from(map.entries());
  }, [visible]);

  if (!source.length) {
    return <p className="text-sm text-stone-400">هنوز رویدادی در تاریخچه ثبت نشده است.</p>;
  }
  return (
    <div className="space-y-3">
      {filterable && (
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setType("")}
            className={`rounded-full px-3 py-1 text-xs ${type === "" ? "bg-mahak-500 text-white" : "bg-stone-100"}`}
          >
            همه
          </button>
          {types.map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setType(t)}
              className={`rounded-full px-3 py-1 text-xs ${type === t ? "bg-mahak-500 text-white" : "bg-stone-100"}`}
            >
              {EVENT_LABEL[t] || t}
            </button>
          ))}
        </div>
      )}
      {groups.map(([day, rows]) => (
        <div key={day}>
          <div className="mb-1 text-xs font-medium text-stone-400">{day}</div>
          <ol className="space-y-2">
            {rows.map((e) => (
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
                {e.comment && COMMENT_TYPES.has(e.event_type) && (
                  <p className="mt-1 whitespace-pre-wrap text-stone-700">{e.comment}</p>
                )}
              </li>
            ))}
          </ol>
        </div>
      ))}
    </div>
  );
}
