export const SKILLS: { id: string; label: string }[] = [
  { id: "general", label: "عمومی" },
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

export const EDUCATION_LEVELS = [
  "زیر دیپلم",
  "دیپلم",
  "کاردانی",
  "کارشناسی",
  "کارشناسی ارشد",
  "دکتری",
  "حوزوی",
];

export const GENDERS = [
  { id: "male", label: "مذکر" },
  { id: "female", label: "مونث" },
];

export const OCCUPATIONS = [
  { id: "student", label: "دانشجو" },
  { id: "employee", label: "کارمند" },
  { id: "teacher", label: "معلم / فرهنگی" },
  { id: "medical", label: "پزشک / کادر درمان" },
  { id: "engineer", label: "مهندس" },
  { id: "worker", label: "کارگر" },
  { id: "self_employed", label: "آزاد" },
  { id: "homemaker", label: "خانه‌دار" },
  { id: "retired", label: "بازنشسته" },
  { id: "unemployed", label: "بیکار" },
  { id: "soldier", label: "سرباز" },
  { id: "seminary", label: "حوزوی" },
  { id: "other", label: "سایر" },
];

export function genderLabel(id?: string) {
  return GENDERS.find((x) => x.id === id)?.label || "";
}

export function occupationLabel(id?: string, other?: string) {
  if (!id) return "";
  if (id === "other") return other?.trim() ? `سایر — ${other.trim()}` : "سایر";
  return OCCUPATIONS.find((x) => x.id === id)?.label || id;
}

export const WEEKDAYS = ["یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه", "شنبه"];

export function weekdayLabel(wd?: number) {
  if (typeof wd === "number" && wd >= 0 && wd < WEEKDAYS.length) return WEEKDAYS[wd];
  return "—";
}

export function fileKindLabel(name?: string) {
  const ext = (name || "").split(".").pop()?.toLowerCase() || "";
  if (["jpg", "jpeg", "png", "webp", "gif"].includes(ext)) return "تصویر";
  if (ext === "pdf") return "PDF";
  if (["doc", "docx"].includes(ext)) return "سند";
  if (ext) return ext.toUpperCase();
  return "فایل";
}

export const STATUS_EXPLAIN: Record<string, string> = {
  draft: "پیش‌نویس — هنوز ثبت درخواست نشده است. می‌توانید هر زمان ادامه دهید و بعداً ارسال کنید.",
  pending: "در انتظار بررسی پشتیبانی — درخواست شما ثبت شده و در صف بررسی قرار دارد.",
  rejected: "رد شده یا نقص مدرک — پیام پشتیبانی را بخوانید، اطلاعات یا مدارک را اصلاح کنید و دوباره ارسال کنید.",
  approved: "فعال — عضویت شما پذیرفته شده و می‌توانید فعالیت‌های عملیاتی را ببینید و درخواست دهید.",
  suspended: "تعلیق شده — حساب شما موقتاً غیرفعال است تا پشتیبانی رفع تعلیق کند.",
};

export const EVENT_LABEL: Record<string, string> = {
  submitted: "ارسال برای بررسی",
  approved: "تایید عضویت",
  rejected: "رد درخواست",
  documents_requested: "درخواست مدارک",
  suspended: "تعلیق",
  unsuspended: "رفع تعلیق",
  status_changed: "تغییر وضعیت",
  comment: "پیام / کامنت",
  profile_updated: "ویرایش اطلاعات توسط پشتیبانی",
  document_deleted: "حذف مدرک",
  document_uploaded: "بارگذاری مجدد مدرک",
  skill_proposal: "مهارت پیشنهادی",
  certificate: "تقدیرنامه/گواهی",
  ticket: "تیکت پشتیبانی",
};

export const STATUS_LABEL: Record<string, string> = {
  draft: "پیش‌نویس",
  pending: "در انتظار بررسی",
  rejected: "رد شده",
  approved: "فعال",
  suspended: "تعلیق",
  open: "باز",
  closed: "اتمام‌یافته",
  cancelled: "لغو شده",
  inactive: "غیرفعال",
  requested: "در انتظار تایید واحد پشتیبانی",
  training_pending: "در انتظار تایید آموزش",
  reserved: "تایید شده — در انتظار انجام",
  attended: "حضور تایید شد",
  absent: "عدم حضور",
  submitted: "نتیجه ارسال شد",
  revision_requested: "نیاز به اصلاح نتیجه",
  preparing: "در حال آماده‌سازی",
  ready: "آماده تحویل",
  delivered: "تحویل شد",
  completed: "تکمیل شده",
  in_progress: "در حال انجام",
  expired: "منقضی",
  active: "فعال",
  archived: "بایگانی",
  task: "موردی",
  aggregated: "تجمیعی",
  pending_skill: "در انتظار تایید",
  answered: "پاسخ داده‌شده",
  recurring: "فعالیت جاری",
  occurrence: "نوبت",
  one_off: "موردی",
};

export const DOC_KIND_LABEL: Record<string, string> = {
  national_id: "کارت ملی",
  driving_license: "گواهینامه رانندگی",
  medical_license: "شماره نظام پزشکی",
  education: "مدرک تحصیلی",
};

export const DOC_KINDS = [
  { id: "national_id", label: "کارت ملی" },
  { id: "driving_license", label: "گواهینامه رانندگی" },
  { id: "medical_license", label: "شماره نظام پزشکی" },
  { id: "education", label: "مدرک تحصیلی" },
];

export const MISSION_KIND_LABEL: Record<string, string> = {
  complete_profile: "تکمیل پروفایل",
  invite_users: "دعوت کاربر",
  custom: "سفارشی",
  webhook: "رویداد وب‌هوک",
};

export const VERIFY_MODE_LABEL: Record<string, string> = {
  internal: "بررسی داخلی سامانه",
  outbound: "فراخوانی وب‌سرویس",
  inbound: "وب‌هوک ورودی",
};

export function docKindLabel(id: string) {
  return DOC_KIND_LABEL[id] || id;
}

export function missionKindLabel(id: string) {
  return MISSION_KIND_LABEL[id] || id;
}

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
    case "delivered":
      return "bg-emerald-50 text-emerald-700 border-emerald-200";
    case "pending":
    case "requested":
    case "training_pending":
    case "reserved":
    case "in_progress":
    case "pending_skill":
    case "revision_requested":
    case "preparing":
      return "bg-amber-50 text-amber-800 border-amber-200";
    case "rejected":
    case "suspended":
    case "expired":
    case "cancelled":
    case "absent":
      return "bg-rose-50 text-rose-700 border-rose-200";
    case "attended":
    case "submitted":
    case "ready":
      return "bg-sky-50 text-sky-800 border-sky-200";
    case "inactive":
    case "closed":
      return "bg-stone-100 text-stone-700 border-stone-200";
    default:
      return "bg-stone-100 text-stone-700 border-stone-200";
  }
}

export const TICKET_LABEL: Record<string, string> = {
  open: "باز",
  answered: "پاسخ داده‌شده",
  closed: "بسته",
};

export const TICKET_SUBJECTS = [
  "سوال درباره فعالیت",
  "آموزش",
  "حضور و غیاب",
  "مدارک و پرونده",
  "تقدیرنامه و گواهی",
  "زمان‌بندی و انصراف",
  "پیشنهاد و انتقاد",
  "سایر",
];

export const WORK_STATUS_FILTERS = [
  { id: "", label: "همه وضعیت‌ها" },
  { id: "open", label: "باز و در حال انجام" },
  { id: "completed", label: "تکمیل شده" },
  { id: "absent", label: "عدم حضور" },
  { id: "rejected", label: "رد شده" },
  { id: "cancelled", label: "انصراف / لغو" },
];

const OPEN_ASSIGNMENT_STATUSES = [
  "requested",
  "training_pending",
  "reserved",
  "in_progress",
  "revision_requested",
  "submitted",
  "attended",
];

export function isOpenAssignment(status?: string) {
  return OPEN_ASSIGNMENT_STATUSES.includes(status || "");
}

export function assignmentSortRank(status?: string) {
  const i = OPEN_ASSIGNMENT_STATUSES.indexOf(status || "");
  if (i >= 0) return i;
  const done = ["completed", "absent", "rejected", "cancelled"];
  const j = done.indexOf(status || "");
  return j < 0 ? 80 : 50 + j;
}

export function sortAssignmentsOpenFirst<T extends { status?: string; created_at?: string; task?: { starts_at?: string } }>(items: T[]) {
  return [...items].sort((a, b) => {
    const ra = assignmentSortRank(a.status);
    const rb = assignmentSortRank(b.status);
    if (ra !== rb) return ra - rb;
    return (b.task?.starts_at || b.created_at || "").localeCompare(a.task?.starts_at || a.created_at || "");
  });
}

export function matchesWorkStatusFilter(status: string, filter: string) {
  if (!filter) return true;
  if (filter === "open") return isOpenAssignment(status) && status !== "requested";
  if (filter === "cancelled") return status === "cancelled";
  return status === filter;
}

export function notificationHref(title: string) {
  if (title.includes("مدارک")) return "/volunteer/profile?tab=docs";
  if (title.includes("گواهی") || title.includes("تقدیرنامه")) return "/volunteer/certificates";
  if (title.includes("تیکت")) return "/volunteer/tickets";
  if (title.includes("آموزش")) return "/volunteer/work";
  if (title.includes("درخواست فعالیت ثبت")) return "/volunteer/tasks";
  if (title.includes("فعالیت") || title.includes("کار")) return "/volunteer/work";
  if (title.includes("مهارت")) return "/volunteer/profile";
  return "/volunteer";
}

export function workModeLabel(mode?: string) {
  return mode === "remote" ? "دورکار" : "حضوری";
}

export function adminActivityHref(a: { task_id?: string; task?: { kind?: string; series_id?: string } }) {
  const sid = a.task?.series_id;
  const id = a.task?.kind === "occurrence" && sid ? sid : a.task_id;
  return `/admin/tasks?manage=${id || ""}`;
}

export const ASSIGNMENT_STATUS_HINT: Record<string, string> = {
  requested: "درخواست داده شده؛ هنوز توسط واحد پشتیبانی تایید نشده است.",
  training_pending: "درخواست تایید شده؛ تا تایید آموزش در بخش آموزش، امکان ادامه فرایند فعالیت نیست.",
  reserved: "تایید شده و در انتظار انجام فعالیت است.",
  in_progress: "داوطلب کار دورکار را شروع کرده است.",
  submitted: "نتیجه دورکار ارسال شده و آماده بررسی است.",
  revision_requested: "واحد پشتیبانی درخواست اصلاح یا تکمیل نتیجه کرده است.",
  attended: "حضور داوطلب در محل ثبت شده است.",
  completed: "این فعالیت تکمیل شده است.",
  absent: "عدم حضور برای این فعالیت ثبت شده است.",
  rejected: "این درخواست رد شده است.",
  cancelled: "این تخصیص لغو شده است.",
};

export const TRAINING_KINDS = [
  { id: "in_person", label: "حضوری" },
  { id: "online", label: "آنلاین" },
  { id: "hybrid", label: "ترکیبی" },
  { id: "workshop", label: "کارگاه" },
  { id: "other", label: "سایر" },
];

export function trainingKindLabel(kind?: string) {
  return TRAINING_KINDS.find((x) => x.id === kind)?.label || kind || "—";
}

export function certKindLabel(kind?: string) {
  switch (kind) {
    case "task":
      return "تقدیرنامه موردی";
    case "aggregated":
      return "تقدیرنامه تجمیعی";
    case "official":
      return "گواهی‌نامه فعالیت داوطلبانه";
    default:
      return kind || "";
  }
}

export function certRequestTitle(r: { kind?: string; assignment_title?: string }) {
  if (r.kind === "official") return "گواهی‌نامه فعالیت داوطلبانه";
  if (r.assignment_title) return r.assignment_title;
  if (r.kind === "aggregated") return "تقدیرنامه تجمیعی";
  return "تقدیرنامه فعالیت";
}

export const CERT_REQ_LABEL: Record<string, string> = {
  pending: "در انتظار بررسی",
  preparing: "در حال آماده‌سازی",
  ready: "آماده تحویل",
  delivered: "تحویل شد",
  approved: "صادر شد",
  rejected: "رد شده",
};

export function deliveryMethodLabel(method?: string) {
  if (method === "send") return "ارسال برای داوطلب";
  if (method === "in_person") return "تحویل حضوری";
  return "";
}

export function isActiveWork(status?: string) {
  return ["training_pending", "reserved", "in_progress", "attended", "submitted", "revision_requested", "completed", "absent"].includes(status || "");
}

export function trainingSatisfied(
  task?: {
    training_course_id?: string;
    training_course?: { title?: string };
    series_id?: string;
    training_kind?: string;
    training_location?: string;
  } | null,
  courses?: {
    course_id?: string;
    course_title?: string;
    series_id?: string;
    training_kind?: string;
    training_location?: string;
  }[],
) {
  if (!task || !courses?.length) return false;
  const courseId = (task.training_course_id || "").trim();
  const title = (task.training_course?.title || "").trim().toLowerCase();
  if (courseId || title) {
    return courses.some((c) => {
      if (courseId && c.course_id && courseId === c.course_id) return true;
      return Boolean(title && (c.course_title || "").trim().toLowerCase() === title);
    });
  }
  const kind = (task.training_kind || "").trim().toLowerCase();
  const loc = (task.training_location || "").trim().toLowerCase();
  return courses.some((c) => {
    if (task.series_id && c.series_id && task.series_id === c.series_id) return true;
    const ck = (c.training_kind || "").trim().toLowerCase();
    const cl = (c.training_location || "").trim().toLowerCase();
    return Boolean(kind && loc && kind === ck && loc === cl);
  });
}

export function trainingCourseTitle(
  c?: { course_title?: string; source_task_title?: string; training_course?: { title?: string } } | null,
) {
  return c?.course_title?.trim() || c?.training_course?.title?.trim() || c?.source_task_title?.trim() || "دوره آموزشی";
}

export function skillLabel(id: string, extra?: Record<string, string>) {
  if (extra?.[id]) return extra[id];
  const hit = SKILLS.find((s) => s.id === id);
  if (hit) return hit.label;
  if (/^g-[0-9a-f]+$/i.test(id)) return "مهارت سفارشی";
  return id;
}

export function catalogLabelMap(groups: { id?: string; slug?: string; title?: string; skills?: { id: string; title: string }[] }[]) {
  const m: Record<string, string> = {};
  for (const g of groups || []) {
    if (g.slug && g.title) m[g.slug] = g.title;
    if (g.id && g.title) m[g.id] = g.title;
    for (const s of g.skills || []) {
      if (s.id && s.title) m[s.id] = s.title;
    }
  }
  return m;
}

export function fmtDate(iso?: string) {
  if (!iso || iso.startsWith("0001") || iso.startsWith("0000")) return "—";
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

export function fmtDay(iso?: string) {
  if (!iso || iso.startsWith("0001") || iso.startsWith("0000")) return "—";
  try {
    const hasTime = iso.length > 10;
    const d = hasTime ? new Date(iso) : new Date(iso + "T12:00:00");
    return d.toLocaleDateString("fa-IR-u-ca-persian", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export function fmtTime(iso?: string) {
  if (!iso || iso.startsWith("0001") || iso.startsWith("0000") || iso.length <= 10) return "";
  try {
    return new Date(iso).toLocaleTimeString("fa-IR", { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}
