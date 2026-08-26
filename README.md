# سامانه مدیریت داوطلبان محک

**Mahak Volunteer Management Platform (MVMP)**

جریان کامل جذب تا به‌کارگیری: ثبت‌نام با موبایل، ویزارد پروفایل، کاتالوگ مهارت، درخواست فعالیت، تایید ادمین، کار حضوری/دورکار، امتیاز و گواهی.

## استقرار روی سرور (نسخهٔ کامل فرآیندها)

شاخهٔ درست این است — نه `main` (نسخهٔ اول):

```bash
cd /opt/mahak-volunteers
git fetch origin
git checkout cursor/certs-availability-work-6e69
git pull origin cursor/certs-availability-work-6e69
docker compose --env-file .env up -d --build
```

پورتال: `http://IP:3000`  
API: `http://IP:8080/api/v1`

## فرآیند داوطلب

1. ثبت‌نام با شماره موبایل و کد پیامکی
2. ویزارد پروفایل (هویت، نشانی، تحصیل، مهارت، مدارک) — هویت بعد از ارسال قفل می‌شود
3. واحد پشتیبانی / بهره‌بردار تایید / نقص مدرک / رد
4. درخواست فعالیت → تایید پشتیبانی → شروع کار
5. حضوری: تایید حضور پشتیبانی · دورکار: ارسال فایل/توضیح
6. امتیاز ۱–۵ و صدور گواهی QR

## حساب‌های نمونه (`SEED_DEMO=true`)

| نقش | ورود |
| --- | --- |
| ادمین | `admin@mahak.ir` / `Admin@123` |
| بهره‌بردار | `operator@mahak.ir` / `Operator@123` |
| داوطلب | `volunteer@mahak.ir` / `Volunteer@123` |

ثبت‌نام جدید از صفحهٔ ثبت‌نام با موبایل است. در محیط بررسی، کد OTP در پاسخ API فیلد `dev_code` برمی‌گردد (`OTP_REVEAL=true`). بهره‌بردار و ادمین از `/login?as=admin` وارد پنل پشتیبانی می‌شوند.

## بسته وب‌سرویس برای تیم دیگر

- OpenAPI 3.0: [`docs/openapi.yaml`](docs/openapi.yaml)
- مرجع فارسی: [`docs/api.md`](docs/api.md)
- راهنمای تحویل: [`docs/integration.md`](docs/integration.md)
- Postman: [`postman/Mahak-Volunteer-Management.postman_collection.json`](postman/Mahak-Volunteer-Management.postman_collection.json)
- کاتالوگ زنده: `GET /api/v1/`

کالکشن Postman: `postman/Mahak-Volunteer-Management.postman_collection.json`

مستندات: [معماری](docs/architecture.md) · [API](docs/api.md) · [تحویل وب‌سرویس](docs/integration.md) · [استقرار](docs/deployment.md)
