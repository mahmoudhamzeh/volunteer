# راهنمای استقرار

این سرویس برای یک نمونهٔ عملیاتی با حدود **۱۰۰۰ داوطلب فعال** و چند ده درخواست همزمان طراحی شده است. یک فرایند API به‌همراه PostgreSQL و Redis روی یک میزبان متوسط کافی است.

## پیش‌نیاز

- Go 1.22+
- Node.js 20+
- PostgreSQL 16 (`max_connections` حداقل ۲۰۰)
- Redis 7 برای قفل ظرفیت تسک
- در محیط عملیاتی: MinIO یا هر S3-compatible برای اسناد

## اجرای محلی

```bash
docker compose up -d postgres redis
cd backend && go run ./cmd/api
```

ترمینال دوم:

```bash
cd frontend && npm install && npm run dev
```

حساب‌های دمو فقط وقتی ساخته می‌شوند که `SEED_DEMO=true` باشد (پیش‌فرض محیط توسعه):

- ادمین: `admin@mahak.ir` / `Admin@123`
- داوطلب تاییدشده: `volunteer@mahak.ir` / `Volunteer@123`

## Docker Compose

توسعه:

```bash
docker compose up --build
```

نزدیک به تولید (یک میزبان):

```bash
export JWT_SECRET='یک-رمز-حداقل-۳۲-کاراکتری'
export INTERNAL_API_KEY='کلید-پل-احراز'
export WEBHOOK_SECRET='کلید-وب‌هوک'
export PUBLIC_BASE_URL='https://volunteers.example.ir'
export CORS_ORIGINS='https://volunteers.example.ir'
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
```

- وب: http://localhost:3000
- API: http://localhost:8080
- سلامت فرایند: `GET /healthz`
- آمادگی وابستگی‌ها: `GET /readyz` (پاسخ ۵۰۳ اگر Postgres/Redis قطع باشد)

## متغیرهای محیطی بک‌اند

| متغیر | پیش‌فرض | توضیح |
| --- | --- | --- |
| `APP_ENV` | `development` | در `production` سید دمو خاموش، JWT ضعیف رد، Redis الزامی |
| `HTTP_ADDR` | `:8080` | آدرس گوش‌دادن |
| `DATABASE_URL` | postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable | اتصال Postgres |
| `DB_MAX_CONNS` | `40` | سقف استخر اتصال هر فرایند API |
| `JWT_SECRET` | مقدار توسعه | در تولید حداقل ۳۲ کاراکتر و غیرپیش‌فرض |
| `JWT_TTL_HOURS` | `24` | عمر توکن |
| `REDIS_URL` | redis://127.0.0.1:6379 | قفل توزیع‌شده ظرفیت |
| `STORAGE_DIR` | ./data/uploads | پوشه مدارک |
| `PUBLIC_BASE_URL` | آدرس عمومی برای QR گواهی | |
| `CORS_ORIGINS` | `*` در توسعه | فهرست مبدأها با ویرگول |
| `SEED_DEMO` | `true` در توسعه | حساب و تسک نمونه |
| `INTERNAL_API_KEY` | خالی | هدر `X-Internal-Key` برای `POST /auth/external` |
| `WEBHOOK_SECRET` | خالی | هدر `X-Webhook-Secret` برای `POST /webhooks/events` |

اسکیمای دیتابیس در `backend/migrations/001_init.sql` و `backend/internal/adapter/postgres/schema.sql` است. مایگریشن هنگام استارت اجرا می‌شود.

در محیط محک، لایه Auth موجود باید `sub` کاربر را با `X-Internal-Key` به `POST /api/v1/auth/external` بفرستد تا پروفایل داوطلب بدون ثبت‌نام جداگانه ساخته شود. ثبت‌نام عمومی همیشه نقش `volunteer` می‌سازد.

## ظرفیت حدود ۱۰۰۰ کاربر

برای این مقیاس:

1. **یک فرایند API** با `DB_MAX_CONNS=40` کافی است. اگر دو replica می‌گذارید Redis را روشن نگه دارید؛ قفل حافظه‌ای بین فرایندها کار نمی‌کند.
2. Postgres را با `max_connections=200` و `shared_buffers` حداقل ۲۵۶MB بالا بیاورید.
3. پشت API یک reverse proxy (Nginx/Caddy) با TLS، محدودیت حجم بدنه ۶MB، و timeout ۳۰ ثانیه بگذارید.
4. ورود و ثبت‌نام روی هر IP به ۲۰ درخواست در دقیقه محدود است؛ بقیه API حدود ۳۰۰ درخواست در دقیقه.
5. قبل از بار واقعی `GET /readyz` باید `{"status":"ready"}` برگرداند.

نمونه فشار با [hey](https://github.com/rakyll/hey):

```bash
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"volunteer@mahak.ir","password":"Volunteer@123"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')

hey -n 5000 -c 100 -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/me
hey -n 2000 -c 50 -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/tasks
```

هدف عملیاتی برای این مقیاس: p95 زیر ۲۰۰ms برای `/me` و `/tasks` و بدون خطای ۵xx در رزرو همزمان ظرفیت.
