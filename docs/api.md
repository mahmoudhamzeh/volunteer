# API ماژول داوطلبان

Base URL: `http://localhost:8080/api/v1`  
احراز: هدر `Authorization: Bearer <jwt>`

## عمومی

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/healthz` | سلامت سرویس |
| POST | `/api/v1/auth/register` | ثبت‌نام داوطلب |
| POST | `/api/v1/auth/login` | ورود |
| POST | `/api/v1/auth/external` | نگاشت User ID سرویس Auth محک |
| GET | `/api/v1/certificates/verify/{code}` | استعلام اصالت |
| GET | `/api/v1/certificates/{code}/pdf` | دانلود PDF گواهی |
| POST | `/api/v1/webhooks/events` | رویداد سیستمی (ماموریت) |

## داوطلب

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/me` | کاربر جاری |
| PUT | `/volunteers/me` | پروفایل |
| POST | `/volunteers/me/submit` | ارسال برای بررسی |
| PUT | `/volunteers/me/availability` | تقویم زمانی |
| POST | `/volunteers/me/documents` | آپلود مدرک (`multipart`) |
| GET | `/tasks` | تسک‌های واجد شرایط |
| POST | `/tasks/{id}/accept` | پذیرش / رزرو ظرفیت |
| GET | `/assignments/me` | کارهای من |
| POST | `/assignments/{id}/rate` | امتیاز داوطلب به سازماندهی |
| GET/POST | `/missions` و `/missions/{id}/start` | ماموریت‌ها |

## ادمین

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/admin/dashboard` | مانیتورینگ |
| GET | `/admin/volunteers` | فیلتر status/skill/q |
| POST | `/admin/volunteers/{id}/review` | `approve` / `reject` / `request_documents` / `suspend` |
| CRUD | `/admin/tasks` | تعریف تسک با شرط مهارت و امتیاز |
| POST | `/admin/assignments/{id}/attendance` | حضور |
| POST | `/admin/assignments/{id}/complete` | امتیاز ۱–۵ و ساعات |
| POST | `/admin/assignments/{id}/certificate` | گواهی موردی |
| GET | `/admin/reports/ranking?format=csv` | رتبه‌بندی و خروجی اکسل |

نمونه بررسی:

```json
{ "action": "approve" }
{ "action": "request_documents", "reason": "تصویر کارت ملی ناخوانا است" }
{ "action": "reject", "reason": "عدم تطابق مدارک" }
```
