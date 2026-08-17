export const SKILLS: { id: string; label: string }[] = [
  { id: "sports", label: "ورزش" },
  { id: "artistic", label: "هنر" },
  { id: "medical", label: "پزشکی" },
  { id: "administrative", label: "اداری" },
  { id: "technical", label: "فنی" },
  { id: "education", label: "آموزشی" },
  { id: "logistics", label: "لجستیک" },
  { id: "psychological", label: "روان‌شناختی" },
];

export const EDUCATION_LEVELS = [
  "زیر دیپلم",
  "دیپلم",
  "کاردانی",
  "کارشناسی",
  "کارشناسی ارشد",
  "دکتری",
  "حوزوی",
];

export const WEEKDAYS = ["یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه", "شنبه"];

export const STATUS_LABEL: Record<string, string> = {
  draft: "پیش‌نویس",
  pending: "در انتظار بررسی",
  rejected: "رد شده / نقص مدرک",
  approved: "تایید شده",
  suspended: "تعلیق",
  open: "باز",
  closed: "بسته",
  cancelled: "لغو",
  reserved: "رزرو شده",
  attended: "حضور تایید شد",
  completed: "تکمیل شده",
  in_progress: "در حال انجام",
  expired: "منقضی",
  active: "فعال",
  archived: "بایگانی",
  task: "موردی",
  aggregated: "تجمیعی",
  pending_skill: "در انتظار تایید",
};

export const PROPOSAL_LABEL: Record<string, string> = {
  pending: "در انتظار تایید",
  approved: "تایید شده",
  rejected: "رد شده",
};

export function statusClass(status: string) {
  switch (status) {
    case "approved":
    case "completed":
    case "active":
      return "bg-emerald-50 text-emerald-700 border-emerald-200";
    case "pending":
    case "reserved":
    case "in_progress":
    case "pending_skill":
      return "bg-amber-50 text-amber-800 border-amber-200";
    case "rejected":
    case "suspended":
    case "expired":
    case "cancelled":
      return "bg-rose-50 text-rose-700 border-rose-200";
    case "attended":
      return "bg-sky-50 text-sky-800 border-sky-200";
    default:
      return "bg-stone-100 text-stone-700 border-stone-200";
  }
}

export function skillLabel(id: string) {
  return SKILLS.find((s) => s.id === id)?.label || id;
}

export function fmtDate(iso?: string) {
  if (!iso) return "—";
  try {
    const hasTime = iso.length > 10;
    const d = hasTime ? new Date(iso) : new Date(iso + "T12:00:00");
    return d.toLocaleString("fa-IR-u-ca-persian", {
      year: "numeric",
      month: "long",
      day: "numeric",
      ...(hasTime ? { hour: "2-digit", minute: "2-digit" } : {}),
    });
  } catch {
    return iso;
  }
}
