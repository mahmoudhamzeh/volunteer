# راهنمای استقرار

## پیش‌نیاز

- Go 1.22+
- Node.js 20+
- PostgreSQL 16
- Redis 7 (اختیاری ولی توصیه‌شده برای قفل ظرفیت)
- MinIO یا هر S3-compatible برای اسناد در محیط عملیاتی

## اجرای محلی

```bash
# دیتابیس
createdb mahak_volunteers

# بک‌اند — مایگریشن هنگام استارت اجرا می‌شود و داده نمونه ساخته می‌شود
cd backend && go run ./cmd/api

# فرانت
cd frontend && npm install && npm run dev
```

حساب‌های دمو:

- ادمین: `admin@mahak.ir` / `Admin@123`
- داوطلب تاییدشده: `volunteer@mahak.ir` / `Volunteer@123`

## Docker Compose

```bash
docker compose up --build
```

- وب: http://localhost:3000
- API: http://localhost:8080
- MinIO console: http://localhost:9001

## متغیرهای محیطی بک‌اند

| متغیر | پیش‌فرض |
| --- | --- |
| `HTTP_ADDR` | `:8080` |
| `DATABASE_URL` | postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable |
| `JWT_SECRET` | مقدار توسعه |
| `REDIS_URL` | redis://127.0.0.1:6379 |
| `STORAGE_DIR` | ./data/uploads |
| `PUBLIC_BASE_URL` | آدرس عمومی برای QR گواهی |

اسکیمای دیتابیس در `backend/migrations/001_init.sql` و `backend/internal/adapter/postgres/schema.sql` است.

در محیط محک، لایه Auth موجود باید `sub` کاربر را به `POST /api/v1/auth/external` بفرستد تا پروفایل داوطلب بدون ثبت‌نام جداگانه ساخته شود.
