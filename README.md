# سامانه مدیریت داوطلبان محک

**Mahak Volunteer Management Platform (MVMP)**

میکروسرویس جذب تا به‌کارگیری داوطلبان موسسه خیریه حمایت از کودکان مبتلا به سرطان (محک).

| جزء | نام فنی | فناوری |
| --- | --- | --- |
| محصول | سامانه مدیریت داوطلبان محک | — |
| API | `mahak-volunteer-api` | Go 1.22، Clean Architecture، Chi |
| پورتال | `mahak-volunteer-portal` | Next.js 15، Tailwind، RTL |
| دیتابیس | `mahak-volunteer-db` | PostgreSQL 16 |
| قفل ظرفیت | `mahak-volunteer-redis` | Redis 7 |

## اجرای سریع (محیط بررسی ذینفعان)

```bash
git clone https://github.com/mahmoudhamzeh/volunteer.git
cd volunteer
cp .env.example .env
docker compose up --build -d
```

- پورتال: http://localhost:3000
- API: http://localhost:8080
- سلامت: http://localhost:8080/healthz
- آمادگی: http://localhost:8080/readyz
- فهرست API: http://localhost:8080/api/v1
- کالکشن Postman: [`postman/Mahak-Volunteer-Management.postman_collection.json`](postman/Mahak-Volunteer-Management.postman_collection.json)

### حساب‌های نمونه (با `SEED_DEMO=true`)

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

## توسعه روی لپ‌تاپ (بدون Docker برای API/Web)

پیش‌نیاز: Docker برای Postgres/Redis، [Go 1.22+](https://go.dev/dl/)، [Node.js 20+](https://nodejs.org/)

```bash
docker compose up -d postgres redis
cd backend && go run ./cmd/api
```

ترمینال دوم:

```bash
cd frontend && npm install && npm run dev
```

مرورگر: [http://localhost:3000](http://localhost:3000)

## جریان محصول

1. داوطلب پروفایل و مدارک را ثبت می‌کند (`draft`).
2. ادمین تایید می‌کند، نقص مدرک می‌گیرد، یا رد می‌کند.
3. پس از `approved` تسک‌های واجد شرایط (مهارت / امتیاز / رشته) نمایش داده می‌شود.
4. پذیرش تسک با کنترل ظرفیت (قفل Redis + `SELECT FOR UPDATE`).
5. ادمین حضور را تایید و نمره ۱ تا ۵ (انضباط، تخصص، اخلاق) می‌دهد.
6. ساعات معادل ثبت می‌شود و در صورت تایید، گواهی PDF با UUID و QR صادر می‌گردد.

مستندات: [معماری](docs/architecture.md) · [API](docs/api.md) · [استقرار لایو](docs/deployment.md) · [OpenAPI](docs/openapi.yaml)
