# کالکشن Postman — سامانه مدیریت داوطلبان محک

فایل‌ها:

- `Mahak-Volunteer-Management.postman_collection.json`
- `Mahak-Volunteer-Management.postman_environment.json`

## ورود به Postman

1. Import → هر دو فایل را انتخاب کنید.
2. Environment را روی **Mahak Volunteers — Local / Live Review** بگذارید.
3. برای سرور لایو مقدار `base_url` را به `https://volunteers.example.ir` یا `http://SERVER:8080` تغییر دهید.
4. از پوشه **01. Auth** به‌ترتیب **Login admin** و **Login volunteer** را بزنید تا `admin_token` و `volunteer_token` ذخیره شوند.

`INTERNAL_API_TOKEN` در محیط باید با مقدار سرور یکی باشد (پیش‌فرض توسعه: `mahak-internal-demo-token`).
