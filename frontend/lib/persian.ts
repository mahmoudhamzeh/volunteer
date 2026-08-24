export function toEnDigits(value: string) {
  return value
    .replace(/[۰-۹]/g, (d) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(d)))
    .replace(/[٠-٩]/g, (d) => String("٠١٢٣٤٥٦٧٨٩".indexOf(d)));
}

export function onlyDigits(value: string, max = 32) {
  return toEnDigits(value).replace(/\D/g, "").slice(0, max);
}

const PERSIAN_NAME = /^[\u0600-\u06FF\s\u200c]+$/;

export function onlyPersianLetters(value: string) {
  return value.replace(/[^\u0600-\u06FF\s\u200c]/g, "");
}

export function isPersianName(value: string) {
  const s = value.trim();
  return s.length > 0 && PERSIAN_NAME.test(s);
}

export function isNationalID(value: string) {
  return /^\d{10}$/.test(onlyDigits(value, 10)) && onlyDigits(value, 10).length === 10;
}

export const MIN_VOLUNTEER_AGE = 18;

export function volunteerAge(isoDate?: string, now = new Date()): number | null {
  const raw = (isoDate || "").slice(0, 10);
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(raw);
  if (!m) return null;
  const year = Number(m[1]);
  const month = Number(m[2]);
  const day = Number(m[3]);
  const birth = new Date(year, month - 1, day);
  if (birth.getFullYear() !== year || birth.getMonth() !== month - 1 || birth.getDate() !== day) return null;
  let age = now.getFullYear() - year;
  if (now.getMonth() + 1 < month || (now.getMonth() + 1 === month && now.getDate() < day)) age -= 1;
  return age;
}

export function volunteerBirthDateError(isoDate?: string, now = new Date()): string {
  if (!isoDate) return "تاریخ تولد را از تقویم انتخاب کنید";
  const age = volunteerAge(isoDate, now);
  if (age === null) return "تاریخ تولد نامعتبر است";
  if (age < 0) return "تاریخ تولد نمی‌تواند در آینده باشد";
  if (age < MIN_VOLUNTEER_AGE) return "حداقل سن داوطلبی ۱۸ سال تمام است";
  return "";
}

export function needsVolunteerRegistration(status?: string) {
  return !status || status === "draft" || status === "rejected";
}
