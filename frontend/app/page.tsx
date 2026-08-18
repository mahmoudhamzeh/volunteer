import Link from "next/link";
import { MahakLogo } from "@/components/mahak-logo";

export default function Home() {
  return (
    <div className="mx-auto flex min-h-screen max-w-5xl flex-col px-6 py-10">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MahakLogo className="h-12 w-auto" />
          <div>
            <h1 className="text-xl font-black text-ink-900">سامانه متمرکز داوطلبان</h1>
            <p className="text-sm text-stone-500">موسسه خیریه حمایت از کودکان مبتلا به سرطان</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Link href="/login" className="rounded-xl bg-mahak-500 px-4 py-2 text-sm text-white">
            ورود
          </Link>
          <Link href="/register" className="rounded-xl border border-mahak-200 px-4 py-2 text-sm text-mahak-700">
            ثبت‌نام داوطلب
          </Link>
        </div>
      </header>

      <section className="mt-16 grid gap-8 md:grid-cols-2 md:items-center">
        <div>
          <p className="mb-3 text-sm font-medium text-mahak-600">از جذب تا به‌کارگیری، یک مسیر شفاف</p>
          <h2 className="text-3xl font-black leading-snug text-ink-900 md:text-4xl">
            پروفایل تخصصی بسازید، فعالیت انتخاب کنید و گواهی رسمی بگیرید.
          </h2>
          <p className="mt-4 text-stone-600">
            این سامانه جایگزین فایل‌های پراکنده Excel و هماهنگی‌های تلفنی است: احراز هویت، رزرو ظرفیت‌دار، امتیازدهی دوطرفه و صدور گواهی با شناسه یکتا و QR.
          </p>
          <div className="mt-8 grid grid-cols-3 gap-3 text-center">
            {[
              ["۵ وضعیت", "ماشین حالت داوطلب"],
              ["قفل ظرفیت", "جلوگیری از رزرو همزمان"],
              ["گواهی QR", "استعلام اصالت"],
            ].map(([t, s]) => (
              <div key={t} className="rounded-2xl bg-white p-3 shadow-card">
                <div className="font-bold text-mahak-600">{t}</div>
                <div className="text-xs text-stone-500">{s}</div>
              </div>
            ))}
          </div>
        </div>
        <div className="rounded-3xl bg-ink-900 p-8 text-white shadow-card">
          <h3 className="text-lg font-bold">حساب‌های نمونه</h3>
          <ul className="mt-4 space-y-3 text-sm text-stone-200">
            <li>داوطلب: از «ثبت‌نام داوطلب» با موبایل و کد پیامک وارد شوید</li>
            <li>ادمین: <code className="text-mahak-300">admin@mahak.ir</code> / Admin@123</li>
            <li>داوطلب نمونه (ایمیل): <code className="text-mahak-300">volunteer@mahak.ir</code> / Volunteer@123</li>
          </ul>
          <p className="mt-6 text-xs text-stone-400">برای محیط توسعه. در اتصال به Auth محک، توکن اصلی استخراج و پروفایل ساخته می‌شود.</p>
        </div>
      </section>
    </div>
  );
}
