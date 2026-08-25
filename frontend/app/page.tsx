"use client";

import Link from "next/link";
import { MahakLogo } from "@/components/mahak-logo";
import { hasToken } from "@/lib/api";
import { useEffect, useState } from "react";

export default function Home() {
  const [volunteerIn, setVolunteerIn] = useState(false);

  useEffect(() => {
    setVolunteerIn(hasToken("volunteer"));
  }, []);

  return (
    <div className="min-h-screen">
      <section className="relative overflow-hidden bg-gradient-to-bl from-mahak-600 via-rose-600 to-ink-900 text-white">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.18),transparent_42%),radial-gradient(circle_at_bottom_right,rgba(0,188,212,0.22),transparent_40%)]" />
        <div className="relative mx-auto flex max-w-5xl flex-col gap-10 px-6 py-12 md:py-16">
          <header className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <MahakLogo variant="white" className="h-14 w-auto" />
              <div>
                <p className="text-sm text-white/80">موسسه خیریه حمایت از کودکان مبتلا به سرطان</p>
                <h1 className="text-xl font-black md:text-2xl">سامانه داوطلبان محک</h1>
              </div>
            </div>
            <Link href="/login?as=admin" className="hidden rounded-xl border border-white/30 px-3 py-2 text-xs text-white/80 hover:bg-white/10 md:inline-block">
              ورود پشتیبانی
            </Link>
          </header>

          <div className="max-w-2xl space-y-5">
            <p className="inline-block rounded-full bg-white/15 px-3 py-1 text-sm">همراهی داوطلبانه، کنار کودکان محک</p>
            <h2 className="text-3xl font-black leading-snug md:text-5xl">زمان و مهارت شما می‌تواند مسیر درمان را مهربان‌تر کند.</h2>
            <p className="text-base leading-8 text-white/90 md:text-lg">
              داوطلبان محک در فعالیت‌های حضوری و دورکار کنار کودکان مبتلا به سرطان و خانواده‌هایشان هستند؛ از همراهی در بیمارستان و رویدادها تا کارهای تخصصی مثل آموزش، هنر، پشتیبانی اداری و تولید محتوا.
            </p>
            <div className="flex flex-wrap gap-3 pt-2">
              {volunteerIn ? (
                <Link href="/volunteer" className="rounded-2xl bg-white px-6 py-3 text-sm font-bold text-mahak-700 shadow-lg">
                  ورود به پنل داوطلب
                </Link>
              ) : (
                <>
                  <Link href="/register" className="rounded-2xl bg-white px-6 py-3 text-sm font-bold text-mahak-700 shadow-lg">
                    ثبت‌نام داوطلب
                  </Link>
                  <Link href="/register" className="rounded-2xl border border-white/50 px-6 py-3 text-sm font-bold text-white hover:bg-white/10">
                    ورود داوطلب
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-5xl px-6 py-12">
        <div className="grid gap-6 md:grid-cols-3">
          {[
            ["پروفایل و مهارت", "مهارت، مدارک و زمان‌های آزاد خود را ثبت کنید تا واحد پشتیبانی فعالیت مناسب را به شما بسپارد."],
            ["فعالیت حضوری و دورکار", "در فعالیت‌های بیمارستانی، رویدادها یا کارهای قابل انجام از راه دور مشارکت کنید."],
            ["تقدیرنامه و گواهی", "پس از فعالیت، تقدیرنامه سامانه‌ای صادر می‌شود. گواهی‌نامه فعالیت داوطلبانه پس از ۹۰ ساعت تاییدشده درخواست و ارسال یا حضوری تحویل می‌گردد."],
          ].map(([title, body]) => (
            <div key={title} className="rounded-3xl bg-white p-5 shadow-card">
              <h3 className="font-black text-ink-900">{title}</h3>
              <p className="mt-2 text-sm leading-7 text-stone-600">{body}</p>
            </div>
          ))}
        </div>
        <p className="mt-10 text-center text-sm text-stone-500">
          کار داوطلبانه در محک بدون چشمداشت مالی است و بر پایه تعهد، رازداری و احترام به کودک و خانواده انجام می‌شود.
        </p>
        <p className="mt-6 text-center text-xs text-stone-400">
          کارکنان پشتیبانی؟ <Link className="text-mahak-700" href="/login?as=admin">ورود به پنل واحد پشتیبانی</Link>
        </p>
      </section>
    </div>
  );
}
