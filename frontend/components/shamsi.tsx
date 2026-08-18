"use client";

import { Field, inputClass } from "@/components/ui";
import {
  JALALI_MONTHS,
  currentJalaliYear,
  gregorianToJalali,
  isoDateToJalali,
  isoToParts,
  jalaliMonthLength,
  jalaliSatOffset,
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
  disabled,
}: {
  label: string;
  value?: string;
  onChange: (isoDate: string) => void;
  disabled?: boolean;
}) {
  const nowY = currentJalaliYear();
  const today = gregorianToJalali(new Date().getFullYear(), new Date().getMonth() + 1, new Date().getDate());
  const selected = isoDateToJalali(value);
  const viewJy = selected?.jy || today.jy;
  const viewJm = selected?.jm || today.jm;
  const dim = jalaliMonthLength(viewJy, viewJm);
  const offset = jalaliSatOffset(viewJy, viewJm);
  const week = ["ش", "ی", "د", "س", "چ", "پ", "ج"];

  function shiftMonth(delta: number) {
    let jy = viewJy;
    let jm = viewJm + delta;
    if (jm < 1) {
      jm = 12;
      jy -= 1;
    }
    if (jm > 12) {
      jm = 1;
      jy += 1;
    }
    const max = jalaliMonthLength(jy, jm);
    onChange(jalaliToIsoDate(jy, jm, Math.min(selected?.jd || 1, max)));
  }

  return (
    <Field label={label}>
      <div className={`rounded-2xl border border-stone-200 bg-white p-3 ${disabled ? "opacity-70" : ""}`}>
        <div className="mb-3 flex items-center justify-between gap-2">
          <button type="button" className="rounded-xl px-2 py-1 text-sm text-mahak-700 disabled:text-stone-300" disabled={disabled} onClick={() => shiftMonth(-1)}>ماه قبل</button>
          <div className="text-sm font-bold">
            {(JALALI_MONTHS[(viewJm || 1) - 1] || "")} {viewJy.toLocaleString("fa-IR")}
          </div>
          <button type="button" className="rounded-xl px-2 py-1 text-sm text-mahak-700 disabled:text-stone-300" disabled={disabled} onClick={() => shiftMonth(1)}>ماه بعد</button>
        </div>
        <div className="grid grid-cols-7 gap-1 text-center text-xs text-stone-400">
          {week.map((w) => <div key={w} className="py-1">{w}</div>)}
        </div>
        <div className="grid grid-cols-7 gap-1">
          {Array.from({ length: offset }).map((_, i) => <div key={`e${i}`} />)}
          {Array.from({ length: dim }, (_, i) => i + 1).map((d) => {
            const active = selected && selected.jy === viewJy && selected.jm === viewJm && selected.jd === d;
            return (
              <button
                type="button"
                key={d}
                disabled={disabled}
                onClick={() => onChange(jalaliToIsoDate(viewJy, viewJm, d))}
                className={`rounded-xl py-2 text-sm ${active ? "bg-mahak-500 text-white" : "hover:bg-mahak-50"}`}
              >
                {d.toLocaleString("fa-IR")}
              </button>
            );
          })}
        </div>
        <div className="mt-2 flex justify-center">
          <select
            className={inputClass + " max-w-[8rem]"}
            disabled={disabled}
            value={viewJy}
            onChange={(e) => onChange(jalaliToIsoDate(Number(e.target.value), viewJm, Math.min(selected?.jd || 1, jalaliMonthLength(Number(e.target.value), viewJm))))}
          >
            {years(nowY - 90, nowY).map((y) => (
              <option key={y} value={y}>{y.toLocaleString("fa-IR")}</option>
            ))}
          </select>
        </div>
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
