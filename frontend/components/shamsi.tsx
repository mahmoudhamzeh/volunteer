"use client";

import { Field, inputClass } from "@/components/ui";
import {
  JALALI_MONTHS,
  currentJalaliYear,
  isoDateToJalali,
  isoToParts,
  jalaliMonthLength,
  jalaliToIsoDate,
  jalaliToIsoDateTime,
} from "@/lib/jalali";

function years(from: number, to: number) {
  const out: number[] = [];
  for (let y = to; y >= from; y--) out.push(y);
  return out;
}

function hours() {
  return Array.from({ length: 24 }, (_, i) => i);
}

function minutes() {
  return Array.from({ length: 12 }, (_, i) => i * 5);
}

export function ShamsiDateField({
  label,
  value,
  onChange,
}: {
  label: string;
  value?: string;
  onChange: (isoDate: string) => void;
}) {
  const now = currentJalaliYear();
  const j = isoDateToJalali(value);
  const jy = j?.jy || "";
  const jm = j?.jm || "";
  const dim = j ? jalaliMonthLength(j.jy, j.jm) : 31;
  const day = j ? Math.min(j.jd, dim) : "";

  function set(partial: { jy?: number; jm?: number; jd?: number }) {
    const nextJy = partial.jy ?? j?.jy ?? now - 25;
    const nextJm = partial.jm ?? j?.jm ?? 1;
    const nextJd = partial.jd ?? j?.jd ?? 1;
    const max = jalaliMonthLength(nextJy, nextJm);
    onChange(jalaliToIsoDate(nextJy, nextJm, Math.min(nextJd, max)));
  }

  return (
    <Field label={label}>
      <div className="grid grid-cols-3 gap-2">
        <select className={inputClass} value={jy} onChange={(e) => set({ jy: Number(e.target.value) })}>
          <option value="">سال</option>
          {years(now - 90, now).map((y) => (
            <option key={y} value={y}>{y.toLocaleString("fa-IR")}</option>
          ))}
        </select>
        <select className={inputClass} value={jm} onChange={(e) => set({ jm: Number(e.target.value) })}>
          <option value="">ماه</option>
          {JALALI_MONTHS.map((m, i) => (
            <option key={m} value={i + 1}>{m}</option>
          ))}
        </select>
        <select className={inputClass} value={day} onChange={(e) => set({ jd: Number(e.target.value) })}>
          <option value="">روز</option>
          {Array.from({ length: dim }, (_, i) => i + 1).map((d) => (
            <option key={d} value={d}>{d.toLocaleString("fa-IR")}</option>
          ))}
        </select>
      </div>
    </Field>
  );
}

export function ShamsiDateTimeField({
  label,
  value,
  onChange,
}: {
  label: string;
  value?: string;
  onChange: (iso: string) => void;
}) {
  const nowY = currentJalaliYear();
  const p = isoToParts(value) || { jy: nowY, jm: 1, jd: 1, hour: 9, minute: 0 };
  const dim = jalaliMonthLength(p.jy, p.jm);
  const day = Math.min(p.jd, dim);

  function set(partial: Partial<typeof p>) {
    const next = { ...p, ...partial };
    const max = jalaliMonthLength(next.jy, next.jm);
    onChange(jalaliToIsoDateTime(next.jy, next.jm, Math.min(next.jd, max), next.hour, next.minute));
  }

  return (
    <Field label={label}>
      <div className="grid grid-cols-3 gap-2">
        <select className={inputClass} value={p.jy} onChange={(e) => set({ jy: Number(e.target.value) })}>
          {years(nowY - 1, nowY + 3).map((y) => (
            <option key={y} value={y}>{y.toLocaleString("fa-IR")}</option>
          ))}
        </select>
        <select className={inputClass} value={p.jm} onChange={(e) => set({ jm: Number(e.target.value) })}>
          {JALALI_MONTHS.map((m, i) => (
            <option key={m} value={i + 1}>{m}</option>
          ))}
        </select>
        <select className={inputClass} value={day} onChange={(e) => set({ jd: Number(e.target.value) })}>
          {Array.from({ length: dim }, (_, i) => i + 1).map((d) => (
            <option key={d} value={d}>{d.toLocaleString("fa-IR")}</option>
          ))}
        </select>
      </div>
      <div className="mt-2 grid grid-cols-2 gap-2">
        <select className={inputClass} value={p.hour} onChange={(e) => set({ hour: Number(e.target.value) })}>
          {hours().map((h) => (
            <option key={h} value={h}>{`${h.toLocaleString("fa-IR")} ساعت`}</option>
          ))}
        </select>
        <select className={inputClass} value={p.minute - (p.minute % 5)} onChange={(e) => set({ minute: Number(e.target.value) })}>
          {minutes().map((m) => (
            <option key={m} value={m}>{`${m.toLocaleString("fa-IR")} دقیقه`}</option>
          ))}
        </select>
      </div>
    </Field>
  );
}
