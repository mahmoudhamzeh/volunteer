"use client";

import Link from "next/link";
import { Assignment, MissionProgress, VolunteerTraining, openAuth } from "@/lib/api";
import {
  ASSIGNMENT_STATUS_HINT,
  STATUS_LABEL,
  VERIFY_MODE_LABEL,
  adminActivityHref,
  fmtDate,
  isActiveWork,
  missionKindLabel,
  trainingKindLabel,
  weekdayLabel,
  workModeLabel,
} from "@/lib/labels";
import { AttachmentButton, Badge, StarRating } from "@/components/ui";
import { TrainingNotice } from "@/components/training-notice";
import { DeliveryHistory } from "@/components/delivery-history";

function statusHint(a: Assignment) {
  if (a.status === "reserved" && a.task?.work_mode === "remote") {
    return "تایید شده؛ داوطلب باید از پنل کارها فعالیت را شروع و نتیجه را بارگذاری کند.";
  }
  if (a.status === "reserved") {
    return "تایید شده؛ واحد پشتیبانی حضور یا عدم حضور را ثبت می‌کند. داوطلب نیازی به شروع ندارد.";
  }
  if (a.status === "revision_requested" && a.admin_comment) {
    return `${ASSIGNMENT_STATUS_HINT.revision_requested} ${a.admin_comment}`;
  }
  if (a.status === "rejected" && a.admin_comment) {
    return `${ASSIGNMENT_STATUS_HINT.rejected} ${a.admin_comment}`;
  }
  return ASSIGNMENT_STATUS_HINT[a.status] || STATUS_LABEL[a.status] || a.status;
}

export function AssignmentDetail({
  assignment: a,
  fileHref,
}: {
  assignment: Assignment;
  fileHref: (assignmentId: string, fileId: string) => string;
}) {
  const remote = a.task?.work_mode === "remote";
  const hours = a.hours_awarded || a.task?.hour_weight || 0;
  return (
    <div className="space-y-4 text-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <div className="text-xs text-mahak-700">فعالیت</div>
          <h3 className="text-lg font-black">{a.task?.title || "فعالیت"}</h3>
        </div>
        <Badge status={a.status} reason={a.admin_comment} />
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <DetailRow label="نوع انجام" value={workModeLabel(a.task?.work_mode)} />
        <DetailRow label="محل" value={a.task?.location || (remote ? "دورکار" : "—")} />
        <DetailRow
          label="زمان شروع"
          value={a.task?.starts_at ? `${a.task.kind === "occurrence" ? `${weekdayLabel(a.task.weekday)} · ` : ""}${fmtDate(a.task.starts_at)}` : "—"}
        />
        <DetailRow label="زمان پایان" value={fmtDate(a.task?.ends_at)} />
        <DetailRow label="ساعت معادل" value={hours ? `${hours} ساعت` : "—"} />
        <DetailRow label="ثبت درخواست" value={fmtDate(a.created_at)} />
        {a.completed_at ? <DetailRow label="تکمیل" value={fmtDate(a.completed_at)} /> : null}
        {a.delivered_at ? <DetailRow label="ارسال نتیجه" value={fmtDate(a.delivered_at)} /> : null}
      </div>

      {a.task?.description ? (
        <section>
          <h4 className="mb-1 font-bold">شرح فعالیت</h4>
          <p className="whitespace-pre-wrap leading-7 text-stone-700">{a.task.description}</p>
        </section>
      ) : null}

      {a.task?.delivery_hint ? (
        <p className="rounded-2xl bg-mahak-50 px-3 py-2 text-mahak-800">تحویل مورد انتظار: {a.task.delivery_hint}</p>
      ) : null}

      {a.task?.requires_training ? (
        <TrainingNotice
          task={a.task}
          title={isActiveWork(a.status) ? "آموزش این فعالیت" : "این فعالیت نیاز به آموزش دارد"}
        />
      ) : null}

      <div className="rounded-2xl bg-stone-50 px-3 py-2 leading-7 text-stone-700">
        <p>{statusHint(a)}</p>
        {(a.status === "attended" || a.status === "completed") && (a.check_in_at || a.check_out_at || a.attended_at) && (
          <p className="mt-1 text-xs text-stone-500">
            {a.check_in_at ? `ورود ${fmtDate(a.check_in_at)}` : a.attended_at ? `حضور ${fmtDate(a.attended_at)}` : ""}
            {a.check_out_at ? ` · خروج ${fmtDate(a.check_out_at)}` : ""}
          </p>
        )}
        {a.admin_comment && a.status !== "rejected" && a.status !== "revision_requested" ? (
          <p className="mt-1">نظر پشتیبانی: {a.admin_comment}</p>
        ) : null}
      </div>

      {(a.admin_discipline || a.admin_expertise || a.admin_ethics || a.composite_score) ? (
        <section className="space-y-2">
          <h4 className="font-bold">امتیاز پشتیبانی</h4>
          <div className="grid gap-2 sm:grid-cols-3">
            <StarRating label="انضباط" value={a.admin_discipline || 0} readOnly size="sm" />
            <StarRating label="تخصص" value={a.admin_expertise || 0} readOnly size="sm" />
            <StarRating label="اخلاق" value={a.admin_ethics || 0} readOnly size="sm" />
          </div>
          {a.composite_score ? <p className="text-xs text-stone-500">میانگین: {a.composite_score}</p> : null}
        </section>
      ) : null}

      {a.volunteer_rating ? (
        <section className="rounded-2xl border border-stone-100 px-3 py-2">
          <StarRating label="امتیاز داوطلب به سازماندهی" value={a.volunteer_rating} readOnly size="sm" />
          {a.volunteer_comment ? <p className="mt-1 text-stone-600">{a.volunteer_comment}</p> : null}
        </section>
      ) : null}

      <DeliveryHistory items={a.history} assignmentId={a.id} fileHref={fileHref} />
      {!a.history?.length && (a.delivery_note || a.delivery_file_name) && (
        <section className="space-y-1">
          <h4 className="font-bold">نتیجه ارسالی</h4>
          {a.delivery_note ? <p className="whitespace-pre-wrap text-stone-700">{a.delivery_note}</p> : null}
          {a.delivery_file_name ? (
            <AttachmentButton
              name={a.delivery_file_name}
              label="دانلود پیوست نتیجه"
              onOpen={() => void openAuth(`/api/v1/admin/assignments/${a.id}/delivery`)}
            />
          ) : null}
        </section>
      )}

      <Link className="inline-block text-sm text-mahak-700" href={adminActivityHref(a)}>
        باز کردن این فعالیت در صفحه درخواست‌ها و تخصیص
      </Link>
    </div>
  );
}

export function MissionProgressDetail({ item }: { item: MissionProgress }) {
  const m = item.mission;
  return (
    <div className="space-y-4 text-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <div className="text-xs text-mahak-700">مأموریت</div>
          <h3 className="text-lg font-black">{m?.title || "مأموریت"}</h3>
        </div>
        <Badge status={item.status} />
      </div>
      {m?.description ? <p className="whitespace-pre-wrap leading-7 text-stone-700">{m.description}</p> : null}
      <div className="grid gap-3 sm:grid-cols-2">
        <DetailRow label="نوع" value={m?.kind ? missionKindLabel(m.kind) : "—"} />
        <DetailRow label="روش تأیید" value={VERIFY_MODE_LABEL[m?.verify_mode || ""] || m?.verify_mode || "—"} />
        <DetailRow label="پیشرفت" value={`${item.progress}${m?.target_count ? ` از ${m.target_count}` : ""}`} />
        <DetailRow label="ساعت معادل" value={m?.hour_weight ? `${m.hour_weight} ساعت` : "—"} />
        <DetailRow label="شروع" value={fmtDate(item.started_at)} />
        <DetailRow label="مهلت" value={fmtDate(item.due_at)} />
        {item.completed_at ? <DetailRow label="پایان" value={fmtDate(item.completed_at)} /> : null}
        {m?.deadline_hours ? <DetailRow label="مهلت تعریف‌شده" value={`${m.deadline_hours} ساعت`} /> : null}
      </div>
      <Link className="inline-block text-sm text-mahak-700" href="/admin/missions">
        فهرست مأموریت‌ها
      </Link>
    </div>
  );
}

export function TrainingCourseDetail({ item }: { item: VolunteerTraining }) {
  return (
    <div className="space-y-4 text-sm">
      <div>
        <div className="text-xs text-emerald-700">دوره آموزشی</div>
        <h3 className="text-lg font-black">{item.source_task_title || "دوره آموزشی"}</h3>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <DetailRow label="نوع آموزش" value={trainingKindLabel(item.training_kind)} />
        <DetailRow label="محل" value={item.training_location || "—"} />
        <DetailRow label="زمان آموزش" value={fmtDate(item.training_at)} />
        <DetailRow label="تایید حضور" value={fmtDate(item.confirmed_at)} />
      </div>
      <p className="rounded-2xl bg-emerald-50 px-3 py-2 text-emerald-900">
        این دوره در پرونده داوطلب ثبت شده است. برای فعالیت‌هایی با همین آموزش، حضور مجدد در کلاس لازم نیست.
      </p>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value?: string | number | null }) {
  const empty = value === undefined || value === null || value === "";
  return (
    <div className="min-w-0">
      <div className="text-xs text-stone-500">{label}</div>
      <div className="mt-0.5 break-words font-medium text-ink-900">{empty ? "—" : value}</div>
    </div>
  );
}
