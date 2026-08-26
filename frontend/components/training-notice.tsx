"use client";

import { useState } from "react";
import { fmtDate, trainingKindLabel } from "@/lib/labels";

export type TrainingInfo = {
  requires_training?: boolean;
  training_kind?: string;
  training_location?: string;
  training_at?: string;
};

export function TrainingNotice({
  task,
  className = "mt-2 rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-950",
  title = "این فعالیت نیاز به آموزش دارد",
}: {
  task?: TrainingInfo | null;
  className?: string;
  title?: string;
}) {
  if (!task?.requires_training) return null;
  return (
    <div className={className}>
      {title ? <div className="font-medium">{title}</div> : null}
      <ul className={`${title ? "mt-1" : ""} space-y-0.5 text-xs leading-6`}>
        <li>نوع آموزش: {trainingKindLabel(task.training_kind)}</li>
        <li>محل آموزش: {task.training_location || "—"}</li>
        <li>زمان آموزش: {fmtDate(task.training_at)}</li>
      </ul>
    </div>
  );
}

export function TrainingBadge({
  task,
  completed = false,
  className = "",
}: {
  task?: TrainingInfo | null;
  completed?: boolean;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  if (!task?.requires_training) return null;
  return (
    <div className={className}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
          completed
            ? "border-emerald-200 bg-emerald-50 text-emerald-800"
            : "border-amber-200 bg-amber-50 text-amber-900"
        }`}
      >
        {completed ? "آموزش گذرانده‌شده" : "نیاز به آموزش"}
      </button>
      {open && (
        <TrainingNotice
          task={task}
          title={completed ? "جزئیات آموزش گذرانده‌شده" : "جزئیات آموزش"}
          className="mt-2 rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-950"
        />
      )}
    </div>
  );
}
