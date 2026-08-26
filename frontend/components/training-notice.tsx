"use client";

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
}: {
  task?: TrainingInfo | null;
  className?: string;
}) {
  if (!task?.requires_training) return null;
  return (
    <div className={className}>
      <div className="font-medium">این فعالیت نیاز به آموزش دارد</div>
      <ul className="mt-1 space-y-0.5 text-xs leading-6">
        <li>نوع آموزش: {trainingKindLabel(task.training_kind)}</li>
        <li>محل آموزش: {task.training_location || "—"}</li>
        <li>زمان آموزش: {fmtDate(task.training_at)}</li>
      </ul>
    </div>
  );
}
