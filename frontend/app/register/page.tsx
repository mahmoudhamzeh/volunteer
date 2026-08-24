"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { api, setToken } from "@/lib/api";
import { MahakLogo } from "@/components/mahak-logo";
import { needsVolunteerRegistration, toEnDigits } from "@/lib/persian";

function PhoneIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
      <path strokeLinecap="round" strokeLinejoin="round" d="M6.6 3.8h2.7c.5 0 .9.3 1 .8l.6 2.4a1 1 0 0 1-.3 1L9.3 9.3a12.4 12.4 0 0 0 5.4 5.4l1.3-1.3a1 1 0 0 1 1-.3l2.4.6c.5.1.8.5.8 1v2.7a1 1 0 0 1-1.1 1A16.4 16.4 0 0 1 3.6 4.9a1 1 0 0 1 1-1.1Z" />
    </svg>
  );
}

function LoginIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
      <path strokeLinecap="round" strokeLinejoin="round" d="M10 17v2a2 2 0 0 0 2 2h7a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2h-7a2 2 0 0 0-2 2v2" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12H3m0 0 3.5-3.5M3 12l3.5 3.5" />
    </svg>
  );
}

const fieldClass =
  "w-full rounded-2xl border border-stone-200 bg-white py-3.5 pe-12 ps-4 text-sm text-ink-800 outline-none placeholder:text-stone-400 focus:border-mahak-300 focus:ring-2 focus:ring-mahak-200";

export default function RegisterPage() {
  const router = useRouter();
  const [step, setStep] = useState<"phone" | "code">("phone");
  const [phone, setPhone] = useState("");
  const [sentPhone, setSentPhone] = useState("");
  const [code, setCode] = useState("");
  const [devCode, setDevCode] = useState("");
  const [wait, setWait] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (wait <= 0) return;
    const t = window.setTimeout(() => setWait((n) => n - 1), 1000);
    return () => window.clearTimeout(t);
  }, [wait]);

  async function sendSms(e?: FormEvent) {
    e?.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await api.sendOtp(toEnDigits(phone));
      setSentPhone(res.phone);
      setDevCode(res.dev_code || "");
      setWait(res.resend_after || 60);
      setStep("code");
      setCode("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "ارسال پیامک ناموفق بود");
    } finally {
      setBusy(false);
    }
  }

  async function verify(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await api.verifyOtp(sentPhone, toEnDigits(code));
      setToken(res.token, "volunteer");
      const me = await api.me();
      router.push(needsVolunteerRegistration(me.volunteer?.status) ? "/volunteer/profile" : "/volunteer");
    } catch (err) {
      setError(err instanceof Error ? err.message : "تایید کد ناموفق بود");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex min-h-screen justify-center bg-white">
      <div className="flex w-full max-w-md flex-col px-5 py-6">
        <header className="relative overflow-hidden rounded-[28px] bg-gradient-to-br from-mahak-400 via-mahak-500 to-mahak-700 px-6 py-8 text-white shadow-card">
          <div className="pointer-events-none absolute -left-8 -top-10 h-40 w-40 rounded-full bg-white/10" />
          <div className="pointer-events-none absolute -bottom-12 right-10 h-36 w-36 rounded-full bg-white/10" />
          <div className="pointer-events-none absolute left-16 top-4 h-16 w-16 rounded-full border border-white/15" />
          <MahakLogo variant="white" className="h-14 w-auto" />
          <h1 className="mt-5 text-[28px] font-black leading-snug">ثبت‌نام و ورود داوطلبان</h1>
          <p className="mt-2 max-w-sm text-sm leading-7 text-white/90">
            با ثبت‌نام تو سامانه داوطلبان محک، یه قدم برای حمایت از کودکان برداشتی.
          </p>
        </header>

        {step === "phone" ? (
          <form onSubmit={sendSms} className="flex flex-1 flex-col">
            <p className="mt-10 text-center text-[15px] leading-8 text-stone-600">
              برای ورود به سامانه داوطلبان محک، لطفاً شماره موبایل خود را وارد کنید. کد تأیید یکبارمصرف به همین شماره ارسال خواهد شد.
            </p>
            <label className="mt-10 block">
              <span className="mb-2 block text-sm text-stone-500">شماره موبایل</span>
              <span className="relative block">
                <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-stone-400">
                  <PhoneIcon />
                </span>
                <input
                  className={fieldClass}
                  inputMode="tel"
                  autoComplete="tel"
                  placeholder="شماره موبایل خود را وارد کنید..."
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  required
                />
              </span>
            </label>
            {error && <p className="mt-3 text-sm text-rose-600">{error}</p>}
            <button
              type="submit"
              disabled={busy}
              className="mt-8 flex w-full items-center justify-center gap-2 rounded-2xl bg-mahak-500 py-3.5 text-base font-bold text-white shadow-card transition hover:bg-mahak-600 disabled:opacity-60"
            >
              <LoginIcon />
              {busy ? "در حال ارسال پیامک..." : "ثبت‌نام و ورود"}
            </button>
          </form>
        ) : (
          <form onSubmit={verify} className="flex flex-1 flex-col">
            <p className="mt-10 text-center text-[15px] leading-8 text-stone-600">
              کد تأیید یکبارمصرف به شماره <span className="font-bold text-ink-800" dir="ltr">{sentPhone}</span> ارسال شد. لطفاً کد را وارد کنید.
            </p>
            {devCode && (
              <p className="mt-4 rounded-2xl bg-amber-50 px-3 py-2 text-center text-sm text-amber-900">
                محیط آزمایشی — کد پیامک: <b dir="ltr">{devCode}</b>
              </p>
            )}
            <label className="mt-8 block">
              <span className="mb-2 block text-sm text-stone-500">کد تأیید</span>
              <input
                className="w-full rounded-2xl border border-stone-200 bg-white px-4 py-3.5 text-center text-lg tracking-[0.4em] outline-none focus:border-mahak-300 focus:ring-2 focus:ring-mahak-200"
                dir="ltr"
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={5}
                placeholder="_____"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                required
              />
            </label>
            {error && <p className="mt-3 text-sm text-rose-600">{error}</p>}
            <button
              type="submit"
              disabled={busy}
              className="mt-8 flex w-full items-center justify-center gap-2 rounded-2xl bg-mahak-500 py-3.5 text-base font-bold text-white shadow-card transition hover:bg-mahak-600 disabled:opacity-60"
            >
              <LoginIcon />
              {busy ? "در حال بررسی..." : "تأیید و ورود"}
            </button>
            <div className="mt-4 flex flex-wrap items-center justify-center gap-4 text-sm">
              <button
                type="button"
                className="text-mahak-700 disabled:text-stone-400"
                disabled={busy || wait > 0}
                onClick={() => sendSms()}
              >
                {wait > 0 ? `ارسال مجدد تا ${wait} ثانیه` : "ارسال مجدد پیامک"}
              </button>
              <button
                type="button"
                className="text-stone-500"
                onClick={() => { setStep("phone"); setError(""); setDevCode(""); }}
              >
                تغییر شماره
              </button>
            </div>
          </form>
        )}

        <p className="mt-auto pt-8 text-center text-xs text-stone-400">
          <Link href="/" className="text-stone-500">بازگشت به صفحه معرفی</Link>
          {" · "}
          ادمین هستید؟ <Link href="/login?as=admin" className="text-mahak-700">ورود با ایمیل</Link>
        </p>
      </div>
    </div>
  );
}
