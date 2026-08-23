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

export function needsVolunteerRegistration(status?: string) {
  return !status || status === "draft" || status === "rejected";
}
