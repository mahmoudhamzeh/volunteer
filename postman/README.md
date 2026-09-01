# کالکشن Postman — سامانه مدیریت داوطلبان محک

فایل‌ها:

- `Mahak-Volunteer-Management.postman_collection.json`
- `Mahak-Volunteer-Management.postman_environment.json`

قرارداد کامل HTTP: [`docs/openapi.yaml`](../docs/openapi.yaml) · [`docs/api.md`](../docs/api.md) · [`docs/integration.md`](../docs/integration.md)

## ورود به Postman

1. Import → هر دو فایل را انتخاب کنید.
2. Environment را روی **Mahak Volunteers — Local / Live Review** بگذارید.
3. برای سرور لایو مقدار `base_url` را به دامنه یا `http://SERVER:8080` تغییر دهید.
4. از پوشه **01. Auth** به‌ترتیب **Login admin**، **Login operator (بهره‌بردار)** و **Login volunteer** را بزنید تا توکن‌ها ذخیره شوند.

حساب نمونه:

| نقش | ایمیل | رمز |
| --- | --- | --- |
| ادمین | `admin@mahak.ir` | `Admin@123` |
| بهره‌بردار | `operator@mahak.ir` | `Operator@123` |
| داوطلب | `volunteer@mahak.ir` | `Volunteer@123` |

`INTERNAL_API_TOKEN` در محیط باید با مقدار سرور یکی باشد (پیش‌فرض توسعه: `mahak-internal-demo-token`).

وب‌هوک ماموریت: هدر `X-Internal-Token` برای دروازه، فیلد `token` یا `X-Mission-Token` برای راز همان ماموریت.
