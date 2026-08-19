"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { api, setToken } from "@/lib/api";
import { Button, Card, Field, inputClass } from "@/components/ui";
import { MahakLogo } from "@/components/mahak-logo";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("volunteer@mahak.ir");
  const [password, setPassword] = useState("Volunteer@123");
  const [error, setError] = useState("");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await api.login(email, password);
      setToken(res.token);
      router.push(res.user.role === "volunteer" ? "/volunteer" : "/admin");
    } catch (err) {
      setError(err instanceof Error ? err.message : "خطا در ورود");
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-md items-center px-4">
      <Card className="w-full p-8">
        <MahakLogo className="mb-4 h-10 w-auto" />
        <h1 className="text-2xl font-black text-ink-900">ورود به سامانه داوطلبان</h1>
        <p className="mt-1 text-sm text-stone-500">ادمین با ایمیل وارد می‌شود؛ داوطلب با شماره موبایل</p>
        <form onSubmit={onSubmit} className="mt-6 space-y-4">
          <Field label="ایمیل">
            <input className={inputClass} value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <Field label="رمز عبور">
            <input type="password" className={inputClass} value={password} onChange={(e) => setPassword(e.target.value)} />
          </Field>
          {error && <p className="text-sm text-rose-600">{error}</p>}
          <Button type="submit">ورود</Button>
        </form>
        <p className="mt-4 text-sm text-stone-500">
          داوطلب هستید؟ <Link href="/register" className="text-mahak-700">ورود / ثبت‌نام با موبایل</Link>
        </p>
      </Card>
    </div>
  );
}
