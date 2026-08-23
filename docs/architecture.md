# معماری ماژول داوطلبان محک

سامانه به صورت یک **میکروسرویس مستقل** طراحی شده است. لایه‌ها بر اساس Clean Architecture جدا شده‌اند تا منطق تجاری به PostgreSQL، Redis یا MinIO وابسته نباشد.

```
frontend (Next.js App Router)
        │  REST / JWT
        ▼
backend/cmd/api  ── HTTP adapters (chi)
        │
        ├── usecase/   منطق تجاری: احراز، پروفایل، تسک، ماموریت، امتیاز، گواهی
        ├── domain/    موجودیت‌ها، ماشین وضعیت، پورت‌ها
        └── adapter/
              ├── postgres
              ├── lock (Redis / memory)
              └── storage (filesystem امروز، S3/MinIO فردا)
```

## ماشین وضعیت داوطلب

`draft → pending → approved | rejected | draft(نقص مدرک)`  
`approved ⇄ suspended`

فقط وضعیت `approved` تسک‌های عملیاتی را می‌بیند.

## رزرو تسک و جلوگیری از Race

1. قفل توزیع‌شده Redis روی `task:{id}`
2. تراکنش PostgreSQL با `SELECT … FOR UPDATE`
3. افزایش اتمی `reserved_count` فقط اگر `reserved_count < capacity`

تست واحد `internal/usecase/taskuc/reserve_test.go` این تضمین را با ۱۲ درخواست همزمان روی ظرفیت ۳ بررسی می‌کند.

## امتیازدهی

- هر تسک/ماموریت یک `hour_weight` دارد (مثلاً طراحی پوستر = ۶ ساعت).
- مدیر سه نمره ۱ تا ۵ می‌دهد: انضباط، تخصص، اخلاق. میانگین = امتیاز ترکیبی.
- میانگین داوطلب با ورود هر تسک تکمیل‌شده به‌روزرسانی می‌شود.
- رتبه‌بندی: ساعات نزولی، سپس امتیاز.

## گواهی

- موردی: پس از تکمیل و امتیاز تسک
- تجمیعی: مجموع ساعات در بازه
- PDF با واترمارک، UUID و QR که به `/verify/{code}` می‌رود

## یکپارچگی Auth و Notification

- `POST /api/v1/auth/external` توکن/شناسه Auth بالادستی را به پروفایل محلی نگاشت می‌کند.
- اعلان‌ها در جدول `notifications` ذخیره می‌شوند و پورت `Notifier` برای اتصال به سرویس Notification محک آماده است.
- وب‌هوک `POST /api/v1/webhooks/events` با توکن ماموریت، رویدادهایی مثل `user.invited` را ثبت می‌کند. داوطلب نمی‌تواند ماموریت را دستی تمام کند؛ «بررسی تأیید» سرویس داخلی یا وب‌سرویس تعریف‌شده در پنل ادمین را صدا می‌زند.
