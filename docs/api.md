# API سامانه مدیریت داوطلبان محک

**Service:** `mahak-volunteer-api`  
**Product:** Mahak Volunteer Management Platform  
**Base URL:** `http://localhost:8080` (در پورتال: همان مبدأ `/api/v1`)

احراز کاربر: هدر `Authorization: Bearer <jwt>`  
احراز یکپارچگی داخلی: هدر `X-Internal-Token: <INTERNAL_API_TOKEN>`

فهرست زندهٔ مسیرها: `GET /api/v1`  
کالکشن Postman: `postman/Mahak-Volunteer-Management.postman_collection.json`  
OpenAPI: `docs/openapi.yaml`

خطای استاندارد: `{ "error": "..." }` با کدهای ۴۰۰ / ۴۰۱ / ۴۰۳ / ۴۰۴ / ۴۰۹ / ۴۲۲ / ۵۰۳.

## سلامت

| روش | مسیر | احراز | توضیح |
| --- | --- | --- | --- |
| GET | `/healthz` | — | زنده بودن فرایند |
| GET | `/readyz` | — | آماده بودن Postgres |
| GET | `/api/v1` | — | کاتالوگ مسیرها |

## احراز هویت

| روش | مسیر | احراز | توضیح |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | — | ثبت‌نام داوطلب (نقش همیشه `volunteer`) |
| POST | `/api/v1/auth/login` | — | ورود |
| POST | `/api/v1/auth/external` | Internal | نگاشت User ID سرویس Auth محک |

ثبت‌نام:

```json
{ "email": "new@mahak.ir", "password": "Secret@123", "full_name": "نام داوطلب" }
```

ورود:

```json
{ "email": "volunteer@mahak.ir", "password": "Volunteer@123" }
```

پاسخ: `{ "token": "<jwt>", "user": { "id", "email", "role", "external_user_id", "created_at" } }`

خارجی:

```json
{ "external_user_id": "mahak-sub-001", "email": "user@mahak.ir", "full_name": "نام", "role": "volunteer" }
```

## نشست و اعلان

| روش | مسیر | نقش | توضیح |
| --- | --- | --- | --- |
| GET | `/api/v1/me` | همه | کاربر جاری + پروفایل داوطلب |
| GET | `/api/v1/notifications` | همه | اعلان‌ها |
| POST | `/api/v1/notifications/{id}/read` | همه | خوانده‌شدن اعلان |

## پروفایل داوطلب

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/volunteers/me` | پروفایل |
| PUT | `/api/v1/volunteers/me` | به‌روزرسانی پیش‌نویس |
| POST | `/api/v1/volunteers/me/submit` | ارسال برای بررسی (کارت ملی الزامی) |
| GET/PUT | `/api/v1/volunteers/me/availability` | تقویم زمانی |
| GET/POST | `/api/v1/volunteers/me/documents` | مدارک؛ آپلود `multipart` فیلد `file` و `kind` |

`kind`: `national_id` · `driving_license` · `medical_license` · `education` · `other`  
MIME مجاز: JPEG / PNG / WebP / PDF — حداکثر ۵ مگابایت.

```json
{
  "full_name": "سارا محمدی",
  "national_id": "0012345678",
  "phone": "09121234567",
  "city": "تهران",
  "bio": "طراح گرافیک",
  "skill_categories": ["artistic", "administrative"],
  "education_field": "گرافیک",
  "medical_license": ""
}
```

```json
{ "slots": [{ "weekday": 6, "start_time": "09:00", "end_time": "13:00" }] }
```

`weekday`: ۰ یکشنبه … ۶ شنبه.

مهارت‌ها: `medical` · `administrative` · `artistic` · `technical` · `education` · `logistics` · `psychological`

## تسک و واگذاری (داوطلب)

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/tasks` | تسک‌های واجد شرایط (`q`, `skill`, `limit`, `offset`) |
| GET | `/api/v1/tasks/{id}` | جزئیات (فقط داوطلب تاییدشده یا ستاد) |
| POST | `/api/v1/tasks/{id}/accept` | رزرو ظرفیت |
| GET | `/api/v1/assignments/me` | کارهای من |
| POST | `/api/v1/assignments/{id}/rate` | امتیاز ۱–۵ به سازماندهی |
| POST | `/api/v1/assignments/{id}/cancel` | انصراف از رزرو (قبل از حضور) |

```json
{ "rating": 5, "comment": "هماهنگی خوب بود" }
```

## ماموریت و گواهی (داوطلب)

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/missions` | ماموریت‌های فعال |
| POST | `/api/v1/missions/{id}/start` | شروع |
| POST | `/api/v1/missions/{id}/progress` | پیشرفت `{ "increment": 1 }` |
| GET | `/api/v1/missions/me` | پیشرفت من |
| GET | `/api/v1/certificates/me` | گواهی‌های من |
| GET | `/api/v1/certificates/verify/{code}` | استعلام عمومی |
| GET | `/api/v1/certificates/{code}/pdf` | PDF عمومی |

## ادمین — داوطلبان

نقش: `admin` یا `operator`

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/admin/dashboard` | شاخص‌ها |
| GET | `/api/v1/admin/volunteers` | فیلتر `status` / `skill` / `q` / `limit` / `offset` |
| GET | `/api/v1/admin/volunteers/{id}` | پرونده + مدارک + تقویم |
| POST | `/api/v1/admin/volunteers/{id}/review` | تصمیم بررسی |
| GET | `/api/v1/admin/volunteers/{id}/documents` | مدارک |
| GET | `/api/v1/admin/volunteers/{id}/availability` | تقویم |
| GET | `/api/v1/admin/documents/{id}` | مشاهده فایل |
| POST | `/api/v1/admin/volunteers/{id}/certificates/aggregated` | گواهی تجمیعی (`from`/`to` به صورت `YYYY-MM-DD`) |

```json
{ "action": "approve" }
{ "action": "request_documents", "reason": "تصویر کارت ملی ناخوانا است" }
{ "action": "reject", "reason": "عدم تطابق مدارک" }
{ "action": "suspend", "reason": "تخلف انضباطی" }
{ "action": "unsuspend" }
```

## ادمین — تسک و حضور

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/admin/tasks` | فهرست |
| POST | `/api/v1/admin/tasks` | ایجاد |
| PUT | `/api/v1/admin/tasks/{id}` | ویرایش / بستن |
| DELETE | `/api/v1/admin/tasks/{id}` | حذف |
| GET | `/api/v1/admin/assignments` | `status` / `volunteer_id` / `task_id` |
| POST | `/api/v1/admin/assignments/{id}/attendance` | تایید حضور |
| POST | `/api/v1/admin/assignments/{id}/complete` | نمره و ساعات |
| POST | `/api/v1/admin/assignments/{id}/cancel` | لغو رزرو توسط ستاد |
| POST | `/api/v1/admin/assignments/{id}/certificate` | گواهی موردی |

```json
{
  "title": "طراحی پوستر هفته حمایت از کودک",
  "description": "طراحی پوستر دیجیتال",
  "location": "دورکاری",
  "starts_at": "2026-08-21T09:00:00Z",
  "ends_at": "2026-08-23T18:00:00Z",
  "capacity": 4,
  "hour_weight": 6,
  "required_skills": ["artistic"],
  "min_score": 0,
  "required_education": "گرافیک",
  "status": "open"
}
```

```json
{ "discipline": 5, "expertise": 4, "ethics": 5, "comment": "همکاری منظم" }
```

## ادمین — ماموریت و گزارش

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/api/v1/admin/missions` | همه ماموریت‌ها |
| POST | `/api/v1/admin/missions` | ایجاد |
| PUT | `/api/v1/admin/missions/{id}` | ویرایش |
| GET | `/api/v1/admin/reports/ranking` | رتبه‌بندی؛ `format=csv` خروجی اکسل‌پذیر |
| GET | `/api/v1/admin/reports/skills` | توزیع مهارت |

```json
{
  "title": "دعوت از ۵ کاربر جدید",
  "description": "۵ نفر را دعوت کنید",
  "kind": "invite_users",
  "hour_weight": 2,
  "deadline_hours": 72,
  "webhook_event": "user.invited",
  "target_count": 5
}
```

`kind`: `complete_profile` · `invite_users` · `custom` · `webhook`

## یکپارچگی

| روش | مسیر | احراز | توضیح |
| --- | --- | --- | --- |
| POST | `/api/v1/webhooks/events` | Internal | رویداد سیستمی ماموریت |

```json
{ "event": "user.invited", "volunteer_id": "<uuid>", "increment": 1 }
```
