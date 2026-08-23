# API ماژول داوطلبان

Base URL: `http://localhost:8080/api/v1`  
احراز: هدر `Authorization: Bearer <jwt>`

خطاها JSON هستند: `{ "error": "پیام فارسی", "code": "invalid_input" }`.

## عمومی

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/healthz` | سلامت فرایند (liveness) |
| GET | `/readyz` | آمادگی Postgres/Redis (readiness) |
| POST | `/api/v1/auth/register` | ثبت‌نام داوطلب (نقش همیشه `volunteer` است) |
| POST | `/api/v1/auth/login` | ورود |
| POST | `/api/v1/auth/external` | نگاشت User ID سرویس Auth محک (`X-Internal-Key`) |
| GET | `/api/v1/certificates/verify/{code}` | استعلام اصالت |
| GET | `/api/v1/certificates/{code}/pdf` | دانلود PDF گواهی |
| POST | `/api/v1/webhooks/events` | رویداد سیستمی (`X-Webhook-Secret`) |

ورود و ثبت‌نام روی هر IP حداکثر ۲۰ درخواست در دقیقه است.

## داوطلب

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/me` | کاربر جاری |
| PUT | `/volunteers/me` | پروفایل |
| POST | `/volunteers/me/submit` | ارسال برای بررسی |
| PUT | `/volunteers/me/availability` | تقویم زمانی |
| POST | `/volunteers/me/documents` | آپلود مدرک (`multipart`، حداکثر ۵MB) |
| GET | `/tasks` | تسک‌های واجد شرایط (صفحه‌بندی صحیح پس از فیلتر مهارت) |
| POST | `/tasks/{id}/accept` | پذیرش / رزرو ظرفیت |
| GET | `/assignments/me` | کارهای من |
| POST | `/assignments/{id}/rate` | امتیاز داوطلب به سازماندهی |
| POST | `/assignments/{id}/cancel` | لغو رزرو توسط داوطلب |
| GET/POST | `/missions` و `/missions/{id}/start` | ماموریت‌ها |
| POST | `/notifications/{id}/read` | خوانده‌شدن اعلان |

## ادمین

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/admin/dashboard` | مانیتورینگ |
| GET | `/admin/volunteers` | فیلتر status/skill/q |
| POST | `/admin/volunteers/{id}/review` | `approve` / `reject` / `request_documents` / `suspend` |
| CRUD | `/admin/tasks` | تعریف تسک با شرط مهارت و امتیاز |
| POST | `/admin/assignments/{id}/attendance` | حضور |
| POST | `/admin/assignments/{id}/complete` | امتیاز ۱–۵ و ساعات |
| POST | `/admin/assignments/{id}/cancel` | رد رزرو و آزادسازی ظرفیت |
| POST | `/admin/assignments/{id}/certificate` | گواهی موردی |
| GET | `/admin/reports/ranking?format=csv` | رتبه‌بندی و خروجی اکسل |

نمونه بررسی:

```json
{ "action": "approve" }
{ "action": "request_documents", "reason": "تصویر کارت ملی ناخوانا است" }
{ "action": "reject", "reason": "عدم تطابق مدارک" }
```

نمونه پل Auth:

```http
POST /api/v1/auth/external
X-Internal-Key: <INTERNAL_API_KEY>
Content-Type: application/json

{"external_user_id":"sub-123","email":"a@mahak.ir","full_name":"سارا محمدی"}
```
