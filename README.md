# سامانه داوطلبان محک

میکروسرویس مدیریت **جذب تا به‌کارگیری** داوطلبان موسسه خیریه حمایت از کودکان مبتلا به سرطان (محک).

- بک‌اند: **Go** با معماری Clean
- فرانت: **Next.js (App Router) + Tailwind**، واکنش‌گرا و RTL
- دیتابیس: **PostgreSQL**
- قفل همزمانی رزرو تسک: **Redis**
- اسناد: ذخیره‌ساز فایل (قابل جایگزینی با MinIO/S3)

برای حدود **۱۰۰۰ کاربر** یک فرایند API به‌همراه Postgres و Redis کافی است. جزئیات در [استقرار](docs/deployment.md).

## اجرا روی لپ‌تاپ شخصی

پیش‌نیاز: [Docker Desktop](https://www.docker.com/products/docker-desktop/)، [Go 1.22+](https://go.dev/dl/)، [Node.js 20+](https://nodejs.org/)

### روش ۱ — پیشنهادی (Docker فقط برای دیتابیس)

```bash
git clone https://github.com/mahmoudhamzeh/volunteer.git
cd volunteer
docker compose up -d postgres redis
cd backend && go run ./cmd/api
```

ترمینال دوم:

```bash
cd frontend && npm install && npm run dev
```

سپس مرورگر: [http://localhost:3000](http://localhost:3000)

بک‌اند با پیش‌فرض `postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers` وصل می‌شود. جدول‌ها در استارت اول ساخته می‌شوند. دادهٔ نمونه فقط با `SEED_DEMO=true` (پیش‌فرض توسعه) ساخته می‌شود.

### روش ۲ — همه چیز با Docker

```bash
docker compose up --build
```

- وب: http://localhost:3000
- API: http://localhost:8080
- آمادگی: http://localhost:8080/readyz

نزدیک به تولید:

```bash
export JWT_SECRET='یک-رمز-حداقل-۳۲-کاراکتری'
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
```

### حساب‌های تست

| نقش | ایمیل | رمز |
| --- | --- | --- |
| ادمین | `admin@mahak.ir` | `Admin@123` |
| داوطلب تاییدشده | `volunteer@mahak.ir` | `Volunteer@123` |
| در انتظار بررسی | `pending@mahak.ir` | `Volunteer@123` |

مسیر تست سریع:

1. با داوطلب وارد شوید → تسک‌ها → «پذیرش»
2. در «کارهای من» می‌توانید رزرو را لغو کنید
3. خروج، ورود با ادمین → «حضور و امتیاز» → تایید حضور، نمره ۱–۵، صدور گواهی
4. ادمین → داوطلبان → وضعیت pending → تایید / نقص مدرک / رد
5. داوطلب → گواهی‌ها → دانلود PDF و صفحهٔ استعلام

## جریان محصول

1. داوطلب پروفایل و مدارک را ثبت می‌کند (`draft`).
2. ادمین تایید می‌کند، نقص مدرک می‌گیرد، یا رد می‌کند.
3. پس از `approved` تسک‌های واجد شرایط (مهارت / امتیاز / رشته) نمایش داده می‌شود.
4. پذیرش تسک با کنترل ظرفیت (قفل Redis + `SELECT FOR UPDATE`).
5. ادمین حضور را تایید و نمره ۱ تا ۵ (انضباط، تخصص، اخلاق) می‌دهد.
6. ساعات معادل ثبت می‌شود و در صورت تایید، گواهی PDF با UUID و QR صادر می‌گردد.

مستندات: [معماری](docs/architecture.md) · [API](docs/api.md) · [استقرار](docs/deployment.md)
