"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { api, setToken } from "@/lib/api";
import { Button, Card, Field, inputClass } from "@/components/ui";

export default function RegisterPage() {
  const router = useRouter();
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await api.register(email, password, fullName);
      setToken(res.token);
      router.push("/volunteer/profile");
    } catch (err) {
      setError(err instanceof Error ? err.message : "خطا در ثبت‌نام");
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-md items-center px-4">
      <Card className="w-full p-8">
        <h1 className="text-2xl font-black">ثبت‌نام داوطلب</h1>
        <form onSubmit={onSubmit} className="mt-6 space-y-4">
          <Field label="نام و نام خانوادگی">
            <input className={inputClass} value={fullName} onChange={(e) => setFullName(e.target.value)} required />
          </Field>
          <Field label="ایمیل">
            <input type="email" className={inputClass} value={email} onChange={(e) => setEmail(e.target.value)} required />
          </Field>
          <Field label="رمز عبور (حداقل ۸ کاراکتر)">
            <input type="password" className={inputClass} value={password} onChange={(e) => setPassword(e.target.value)} minLength={8} required />
          </Field>
          {error && <p className="text-sm text-rose-600">{error}</p>}
          <Button type="submit">ایجاد حساب</Button>
        </form>
        <p className="mt-4 text-sm">
          حساب دارید؟ <Link href="/login" className="text-mahak-700">ورود</Link>
        </p>
      </Card>
    </div>
  );
}
