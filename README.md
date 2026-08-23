# سامانه مدیریت داوطلبان محک

**Mahak Volunteer Management Platform (MVMP)**

جریان کامل جذب تا به‌کارگیری: ثبت‌نام با موبایل، ویزارد پروفایل، کاتالوگ مهارت، درخواست فعالیت، تایید ادمین، کار حضوری/دورکار، امتیاز و گواهی.

## استقرار روی سرور (نسخهٔ کامل فرآیندها)

شاخهٔ درست این است — نه `main` (نسخهٔ اول):

```bash
cd /opt/mahak-volunteers
git fetch origin
git checkout cursor/live-readiness-postman-642b
git pull origin cursor/live-readiness-postman-642b
docker compose --env-file .env up -d --build
```

پورتال: `http://IP:3000`  
API: `http://IP:8080/api/v1`

## فرآیند داوطلب

1. ثبت‌نام با شماره موبایل و کد پیامکی
2. ویزارد پروفایل (هویت، نشانی، تحصیل، مهارت، مدارک) — هویت بعد از ارسال قفل می‌شود
3. ادمین تایید / نقص مدرک / رد
4. درخواست فعالیت → تایید ادمین → شروع کار
5. حضوری: تایید حضور ادمین · دورکار: ارسال فایل/توضیح
6. امتیاز ۱–۵ و صدور گواهی QR

## حساب‌های نمونه (`SEED_DEMO=true`)

| نقش | ورود |
| --- | --- |
| ادمین | `admin@mahak.ir` / `Admin@123` |
| داوطلب | `volunteer@mahak.ir` / `Volunteer@123` |

ثبت‌نام جدید از صفحهٔ ثبت‌نام با موبایل است. در محیط بررسی، کد OTP در پاسخ API فیلد `dev_code` برمی‌گردد (`OTP_REVEAL=true`).

کالکشن Postman: `postman/Mahak-Volunteer-Management.postman_collection.json`

مستندات: [معماری](docs/architecture.md) · [API](docs/api.md) · [استقرار](docs/deployment.md)
