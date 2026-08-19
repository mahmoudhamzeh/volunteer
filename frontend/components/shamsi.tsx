"use client";

import { useEffect, useRef, useState } from "react";
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

const WEEKDAYS = ["شن", "یک", "دو", "سه", "چهار", "پنج", "جم"];

function faDigits(value: number | string) {
  return String(value).replace(/\d/g, (d) => "۰۱۲۳۴۵۶۷۸۹"[Number(d)]);
}

function formatJalaliDate(jy: number, jm: number, jd: number) {
  return `${faDigits(jy)}/${faDigits(String(jm).padStart(2, "0"))}/${faDigits(String(jd).padStart(2, "0"))}`;
}

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
  className = "",
}: {
  label: string;
  value?: string;
  onChange: (isoDate: string) => void;
  disabled?: boolean;
  className?: string;
}) {
  const nowY = currentJalaliYear();
  const minYear = nowY - 90;
  const maxYear = nowY;
  const today = gregorianToJalali(new Date().getFullYear(), new Date().getMonth() + 1, new Date().getDate());
  const selected = isoDateToJalali(value);
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState<"day" | "month" | "year">("day");
  const [viewJy, setViewJy] = useState(selected?.jy || today.jy);
  const [viewJm, setViewJm] = useState(selected?.jm || today.jm);
  const [yearPage, setYearPage] = useState(minYear + Math.floor(((selected?.jy || today.jy) - minYear) / 12) * 12);
  const rootRef = useRef<HTMLDivElement>(null);

  function yearPageFor(y: number) {
    const clamped = Math.min(maxYear, Math.max(minYear, y));
    return minYear + Math.floor((clamped - minYear) / 12) * 12;
  }

  useEffect(() => {
    if (!open) return;
    const now = gregorianToJalali(new Date().getFullYear(), new Date().getMonth() + 1, new Date().getDate());
    const next = isoDateToJalali(value) || now;
    setViewJy(next.jy);
    setViewJm(next.jm);
    setYearPage(minYear + Math.floor((Math.min(maxYear, Math.max(minYear, next.jy)) - minYear) / 12) * 12);
    setMode("day");
  }, [open, value, minYear, maxYear]);

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

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
    jy = Math.min(maxYear, Math.max(minYear, jy));
    setViewJy(jy);
    setViewJm(jm);
  }

  function shiftYear(delta: number) {
    setViewJy(Math.min(maxYear, Math.max(minYear, viewJy + delta)));
  }

  function onHeaderPrev() {
    if (mode === "year") setYearPage(yearPageFor(yearPage - 12));
    else if (mode === "month") shiftYear(-1);
    else shiftMonth(-1);
  }

  function onHeaderNext() {
    if (mode === "year") setYearPage(yearPageFor(Math.min(maxYear, yearPage + 12)));
    else if (mode === "month") shiftYear(1);
    else shiftMonth(1);
  }

  function pickDay(day: number) {
    onChange(jalaliToIsoDate(viewJy, viewJm, day));
    setOpen(false);
  }

  const dim = jalaliMonthLength(viewJy, viewJm);
  const offset = jalaliSatOffset(viewJy, viewJm);
  const yearCells = Array.from({ length: 12 }, (_, i) => yearPage + i).filter((y) => y >= minYear && y <= maxYear);
  const display = selected ? formatJalaliDate(selected.jy, selected.jm, selected.jd) : "";

  return (
    <div className={`block space-y-1.5 ${className}`}>
      <span className="text-sm text-stone-600">{label}</span>
      <div className="relative" ref={rootRef}>
      <button
        type="button"
        disabled={disabled}
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={() => !disabled && setOpen((v) => !v)}
        className={`${inputClass} flex items-center justify-between text-right disabled:bg-stone-50 disabled:opacity-70 ${open ? "ring-2" : ""}`}
      >
        <span className={display ? "text-ink-800" : "text-stone-400"}>{display || "انتخاب تاریخ"}</span>
        <svg viewBox="0 0 24 24" className="h-4 w-4 shrink-0 text-mahak-500" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <rect x="3" y="5" width="18" height="16" rx="2" />
          <path d="M3 9h18M8 3v4M16 3v4" />
        </svg>
      </button>
      {open && !disabled && (
        <div className="absolute start-0 z-40 mt-1 w-72 rounded-2xl border border-stone-200 bg-white p-3 shadow-xl">
          <div className="mb-3 flex items-center justify-between gap-1">
            <button type="button" className="rounded-lg p-1.5 text-mahak-600 hover:bg-mahak-50" onClick={onHeaderPrev} aria-label="قبل">
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 18l6-6-6-6" />
              </svg>
            </button>
            <div className="flex items-center gap-0.5 text-sm font-semibold text-stone-800">
              <button
                type="button"
                className={`rounded-lg px-2 py-1 hover:bg-mahak-50 ${mode === "month" ? "bg-mahak-50 text-mahak-700" : ""}`}
                onClick={() => setMode(mode === "month" ? "day" : "month")}
              >
                {JALALI_MONTHS[viewJm - 1]}
              </button>
              <span>،</span>
              <button
                type="button"
                className={`rounded-lg px-2 py-1 hover:bg-mahak-50 ${mode === "year" ? "bg-mahak-50 text-mahak-700" : ""}`}
                onClick={() => {
                  setYearPage(yearPageFor(viewJy));
                  setMode(mode === "year" ? "day" : "year");
                }}
              >
                {faDigits(viewJy)}
              </button>
            </div>
            <button type="button" className="rounded-lg p-1.5 text-mahak-600 hover:bg-mahak-50" onClick={onHeaderNext} aria-label="بعد">
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M15 18l-6-6 6-6" />
              </svg>
            </button>
          </div>

          {mode === "year" && (
            <div className="grid grid-cols-3 gap-2">
              {yearCells.map((y) => (
                <button
                  key={y}
                  type="button"
                  onClick={() => {
                    setViewJy(y);
                    setMode("month");
                  }}
                  className={`rounded-xl py-2 text-sm ${y === viewJy ? "bg-mahak-500 text-white" : "hover:bg-mahak-50"}`}
                >
                  {faDigits(y)}
                </button>
              ))}
            </div>
          )}

          {mode === "month" && (
            <div className="grid grid-cols-3 gap-2">
              {JALALI_MONTHS.map((name, i) => (
                <button
                  key={name}
                  type="button"
                  onClick={() => {
                    setViewJm(i + 1);
                    setMode("day");
                  }}
                  className={`rounded-xl py-2 text-sm ${i + 1 === viewJm ? "bg-mahak-500 text-white" : "hover:bg-mahak-50"}`}
                >
                  {name}
                </button>
              ))}
            </div>
          )}

          {mode === "day" && (
            <div className="grid grid-cols-7 gap-1 text-center text-xs text-mahak-600">
              {WEEKDAYS.map((d) => (
                <div key={d} className="py-1 font-medium">
                  {d}
                </div>
              ))}
              {Array.from({ length: offset }).map((_, i) => (
                <div key={`e${i}`} />
              ))}
              {Array.from({ length: dim }, (_, i) => i + 1).map((d) => {
                const active = selected && selected.jy === viewJy && selected.jm === viewJm && selected.jd === d;
                return (
                  <button
                    type="button"
                    key={d}
                    onClick={() => pickDay(d)}
                    className={`rounded-full py-1.5 text-sm ${active ? "bg-mahak-500 text-white" : "text-stone-800 hover:bg-mahak-50"}`}
                  >
                    {faDigits(d)}
                  </button>
                );
              })}
            </div>
          )}
        </div>
      )}
      </div>
    </div>
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
