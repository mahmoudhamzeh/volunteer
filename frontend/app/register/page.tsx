"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { api, setToken } from "@/lib/api";
import { Button, Card, Field, inputClass } from "@/components/ui";

function toEnDigits(value: string) {
  return value
    .replace(/[۰-۹]/g, (d) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(d)))
    .replace(/[٠-٩]/g, (d) => String("٠١٢٣٤٥٦٧٨٩".indexOf(d)));
}

export default function RegisterPage() {
  const router = useRouter();
  const [step, setStep] = useState<"phone" | "code">("phone");
  const [phone, setPhone] = useState("");
  const [sentPhone, setSentPhone] = useState("");
  const [code, setCode] = useState("");
  const [fullName, setFullName] = useState("");
  const [isNew, setIsNew] = useState(true);
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
      setIsNew(res.is_new);
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
      const res = await api.verifyOtp(sentPhone, toEnDigits(code), fullName);
      setToken(res.token);
      router.push(res.is_new ? "/volunteer/profile" : "/volunteer");
    } catch (err) {
      setError(err instanceof Error ? err.message : "تایید کد ناموفق بود");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-md items-center px-4">
      <Card className="w-full p-8">
        <h1 className="text-2xl font-black">ثبت‌نام داوطلب</h1>
        <p className="mt-1 text-sm text-stone-500">با شماره موبایل وارد شوید؛ کد تایید پیامک می‌شود.</p>

        {step === "phone" && (
          <form onSubmit={sendSms} className="mt-6 space-y-4">
            <Field label="شماره موبایل را وارد نمایید">
              <input
                className={inputClass}
                dir="ltr"
                inputMode="tel"
                autoComplete="tel"
                placeholder="09121234567"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
              />
            </Field>
            {error && <p className="text-sm text-rose-600">{error}</p>}
            <Button type="submit" disabled={busy}>{busy ? "در حال ارسال..." : "ارسال پیامک"}</Button>
          </form>
        )}

        {step === "code" && (
          <form onSubmit={verify} className="mt-6 space-y-4">
            <p className="rounded-2xl bg-mahak-50 px-3 py-2 text-sm text-mahak-800">
              کد تایید به شماره <span className="font-bold" dir="ltr">{sentPhone}</span> ارسال شد.
            </p>
            {devCode && (
              <p className="rounded-2xl bg-amber-50 px-3 py-2 text-sm text-amber-900">
                محیط آزمایشی — کد پیامک: <b dir="ltr">{devCode}</b>
              </p>
            )}
            <Field label="کد ۵ رقمی پیامک">
              <input
                className={inputClass}
                dir="ltr"
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={5}
                placeholder="_____"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                required
              />
            </Field>
            {isNew && (
              <Field label="نام و نام خانوادگی">
                <input className={inputClass} value={fullName} onChange={(e) => setFullName(e.target.value)} placeholder="برای تکمیل ثبت‌نام" />
              </Field>
            )}
            {error && <p className="text-sm text-rose-600">{error}</p>}
            <Button type="submit" disabled={busy}>{busy ? "در حال بررسی..." : "تایید و ورود"}</Button>
            <div className="flex flex-wrap items-center gap-3 text-sm">
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

        <p className="mt-4 text-sm">
          ادمین هستید؟ <Link href="/login" className="text-mahak-700">ورود با ایمیل</Link>
        </p>
      </Card>
    </div>
  );
}
