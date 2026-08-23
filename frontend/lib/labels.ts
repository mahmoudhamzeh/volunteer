export const SKILLS: { id: string; label: string }[] = [
  { id: "sports", label: "ورزش" },
  { id: "artistic", label: "هنر" },
  { id: "medical", label: "پزشکی" },
  { id: "administrative", label: "اداری" },
  { id: "technical", label: "فنی" },
  { id: "education", label: "آموزشی" },
  { id: "logistics", label: "لجستیک" },
  { id: "psychological", label: "روان‌شناختی" },
  { id: "field_ops", label: "فعالیت‌های جاری" },
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
  const hit = SKILLS.find((s) => s.id === id);
  if (hit) return hit.label;
  if (/^g-[0-9a-f]+$/i.test(id) || /^[0-9a-f-]{36}$/i.test(id)) return "مهارت سفارشی";
  return id;
}

export function fmtDate(iso?: string) {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString("fa-IR");
  } catch {
    return iso;
  }
}
