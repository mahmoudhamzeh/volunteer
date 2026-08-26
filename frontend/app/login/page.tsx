"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useEffect, useState } from "react";
import { api, hasToken, setToken } from "@/lib/api";
import { Button, Card, Field, inputClass } from "@/components/ui";
import { MahakLogo } from "@/components/mahak-logo";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const as = params.get("as") === "admin" ? "admin" : "volunteer";
  const [email, setEmail] = useState(as === "admin" ? "admin@mahak.ir" : "volunteer@mahak.ir");
  const [password, setPassword] = useState(as === "admin" ? "Admin@123" : "Volunteer@123");
  const [error, setError] = useState("");
  const [adminSession, setAdminSession] = useState(false);
  const [volunteerSession, setVolunteerSession] = useState(false);

  useEffect(() => {
    setAdminSession(hasToken("admin"));
    setVolunteerSession(hasToken("volunteer"));
  }, []);

  useEffect(() => {
    if (as === "admin") {
      setEmail("admin@mahak.ir");
      setPassword("Admin@123");
    } else {
      setEmail("volunteer@mahak.ir");
      setPassword("Volunteer@123");
    }
    setError("");
  }, [as]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await api.login(email, password);
      const role = res.user.role === "volunteer" ? "volunteer" : "admin";
      setToken(res.token, role);
      router.push(role === "volunteer" ? "/volunteer" : "/admin");
    } catch (err) {
      setError(err instanceof Error ? err.message : "خطا در ورود");
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-md items-center px-4">
      <Card className="w-full p-8">
        <MahakLogo className="mb-4 h-10 w-auto" />
        <h1 className="text-2xl font-black text-ink-900">ورود به سامانه داوطلبان</h1>
        <p className="mt-1 text-sm text-stone-500">
          نشست پشتیبانی و داوطلب جدا هستند؛ می‌توانید هر کدام را در یک تب جدا باز بگذارید.
        </p>

        <div className="mt-5 grid grid-cols-2 gap-2">
          <Link
            href="/login?as=admin"
            className={`rounded-2xl border px-3 py-3 text-center text-sm ${as === "admin" ? "border-mahak-400 bg-mahak-50 font-bold text-mahak-700" : "border-stone-200 text-stone-600"}`}
          >
            پنل پشتیبانی
          </Link>
          <Link
            href="/login?as=volunteer"
            className={`rounded-2xl border px-3 py-3 text-center text-sm ${as === "volunteer" ? "border-mahak-400 bg-mahak-50 font-bold text-mahak-700" : "border-stone-200 text-stone-600"}`}
          >
            پنل داوطلب
          </Link>
        </div>

        {(adminSession || volunteerSession) && (
          <div className="mt-3 flex flex-wrap gap-2 text-sm">
            {adminSession && (
              <Link href="/admin" className="rounded-xl bg-stone-100 px-3 py-1.5 text-mahak-700">ادامه نشست پشتیبانی</Link>
            )}
            {volunteerSession && (
              <Link href="/volunteer" className="rounded-xl bg-stone-100 px-3 py-1.5 text-mahak-700">ادامه نشست داوطلب</Link>
            )}
          </div>
        )}

        <form onSubmit={onSubmit} className="mt-6 space-y-4">
          <Field label="ایمیل">
            <input className={inputClass} value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <Field label="رمز عبور">
            <input type="password" className={inputClass} value={password} onChange={(e) => setPassword(e.target.value)} />
          </Field>
          {error && <p className="text-sm text-rose-600">{error}</p>}
          <Button type="submit">{as === "admin" ? "ورود پشتیبانی" : "ورود داوطلب"}</Button>
        </form>
        <p className="mt-4 text-sm text-stone-500">
          {as === "admin"
            ? "ادمین و بهره‌بردار با همین فرم وارد می‌شوند."
            : <>داوطلب هستید؟ <Link href="/register" className="text-mahak-700">ورود / ثبت‌نام با موبایل</Link></>}
        </p>
      </Card>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="p-8 text-center text-sm text-stone-500">در حال بارگذاری…</div>}>
      <LoginForm />
    </Suspense>
  );
}
