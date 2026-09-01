"use client";

import { AssignmentEvent, openAuth } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { AttachmentButton } from "@/components/ui";

const KIND_LABEL: Record<string, string> = {
  delivery: "ارسال نتیجه توسط داوطلب",
  revision: "درخواست اصلاح توسط پشتیبانی",
  message: "پیام پشتیبانی",
};

export function DeliveryHistory({
  items,
  assignmentId,
  fileHref,
}: {
  items?: AssignmentEvent[];
  assignmentId: string;
  fileHref: (assignmentId: string, fileId: string) => string;
}) {
  if (!items?.length) return null;
  return (
    <div className="space-y-2">
      <p className="text-sm font-medium text-ink-800">تاریخچه نتیجه و اصلاحات</p>
      <ol className="space-y-2">
        {items.map((e) => (
          <li key={e.id} className={`rounded-2xl px-3 py-2 text-sm ${e.kind === "revision" || e.kind === "message" ? "bg-amber-50 text-amber-950" : "bg-white border border-stone-100"}`}>
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <span className="font-medium">{KIND_LABEL[e.kind] || e.kind}</span>
              <span className="text-xs text-stone-500">{fmtDate(e.created_at)}</span>
            </div>
            {e.note ? <p className="mt-1 whitespace-pre-wrap text-stone-700">{e.note}</p> : null}
            {(e.files || []).length > 0 && (
              <div className="mt-1 space-y-1">
                {(e.files || []).map((f) => (
                  <AttachmentButton
                    key={f.id}
                    name={f.file_name}
                    label="دانلود پیوست"
                    onOpen={() => void openAuth(fileHref(assignmentId, f.id))}
                  />
                ))}
              </div>
            )}
          </li>
        ))}
      </ol>
    </div>
  );
}
