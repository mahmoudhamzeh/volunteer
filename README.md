# سامانه داوطلبان محک

میکروسرویس مدیریت **جذب تا به‌کارگیری** داوطلبان موسسه خیریه حمایت از کودکان مبتلا به سرطان (محک).

- بک‌اند: **Go** با معماری Clean
- فرانت: **Next.js (App Router) + Tailwind**، واکنش‌گرا و RTL
- دیتابیس: **PostgreSQL**
- قفل همزمانی رزرو تسک: **Redis**
- اسناد: ذخیره‌ساز فایل (قابل جایگزینی با MinIO/S3)

## اجرا روی لپ‌تاپ شخصی

کد هنوز روی شاخهٔ `main` نیست؛ از شاخهٔ همین کار کلون کنید:

```bash
git clone https://github.com/mahmoudhamzeh/volunteer.git
cd volunteer
git checkout cursor/mahak-volunteer-platform-fbfe
```

### روش ۱ — پیشنهادی (Docker فقط برای دیتابیس)

پیش‌نیاز: [Docker Desktop](https://www.docker.com/products/docker-desktop/)، [Go 1.22+](https://go.dev/dl/)، [Node.js 20+](https://nodejs.org/)

```bash
docker compose up -d postgres redis
cd backend && go run ./cmd/api
```

ترمینال دوم:

```bash
cd frontend && npm install && npm run dev
```

سپس مرورگر: [http://localhost:3000](http://localhost:3000)

اگر Postgres محلی دارید و Docker نمی‌خواهید:

```bash
# macOS: brew install postgresql@16 redis && brew services start postgresql@16 redis
createdb mahak_volunteers
# یا:
psql postgres -c "CREATE USER mahak WITH PASSWORD 'mahak' SUPERUSER;"
psql postgres -c "CREATE DATABASE mahak_volunteers OWNER mahak;"
```

بک‌اند با پیش‌فرض `postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers` وصل می‌شود. جدول‌ها و دادهٔ نمونه در استارت اول ساخته می‌شوند.

### روش ۲ — همه چیز با Docker

```bash
docker compose up --build
```

- وب: http://localhost:3000
- API: http://localhost:8080

### حساب‌های تست

| نقش | ایمیل | رمز |
| --- | --- | --- |
| ادمین | `admin@mahak.ir` | `Admin@123` |
| داوطلب تاییدشده | `volunteer@mahak.ir` | `Volunteer@123` |
| در انتظار بررسی | `pending@mahak.ir` | `Volunteer@123` |

مسیر تست سریع:

1. با داوطلب وارد شوید → تسک‌ها → «پذیرش»
2. خروج، ورود با ادمین → «حضور و امتیاز» → تایید حضور، نمره ۱–۵، صدور گواهی
3. ادمین → داوطلبان → وضعیت pending → تایید / نقص مدرک / رد
4. داوطلب → گواهی‌ها → دانلود PDF و صفحهٔ استعلام

## جریان محصول

1. داوطلب پروفایل و مدارک را ثبت می‌کند (`draft`).
2. ادمین تایید می‌کند، نقص مدرک می‌گیرد، یا رد می‌کند.
3. پس از `approved` تسک‌های واجد شرایط (مهارت / امتیاز / رشته) نمایش داده می‌شود.
4. پذیرش تسک با کنترل ظرفیت (قفل Redis + `SELECT FOR UPDATE`).
5. ادمین حضور را تایید و نمره ۱ تا ۵ (انضباط، تخصص، اخلاق) می‌دهد.
6. ساعات معادل ثبت می‌شود و در صورت تایید، گواهی PDF با UUID و QR صادر می‌گردد.

مستندات: [معماری](docs/architecture.md) · [API](docs/api.md) · [استقرار](docs/deployment.md)
