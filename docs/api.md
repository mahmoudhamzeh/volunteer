# وب‌سرویس سامانه داوطلبان محک

Base URL: `http://localhost:8080`  
کاتالوگ زنده: `GET /api/v1/`  
OpenAPI: [`docs/openapi.yaml`](openapi.yaml)  
راهنمای تحویل به شرکت: [`docs/integration.md`](integration.md)  
Postman: [`postman/Mahak-Volunteer-Management.postman_collection.json`](../postman/Mahak-Volunteer-Management.postman_collection.json)

خطاها: `{ "error": "پیام" }`  
کدها: `400` ورودی، `401` بدون توکن، `403` نقش ناکافی، `404` یافت نشد، `409` تعارض، `422` ظرفیت/واجدشرایط، `503` مشغول.

## احراز هویت

| نوع | هدر | نقش |
| --- | --- | --- |
| JWT داوطلب / بهره‌بردار / ادمین | `Authorization: Bearer <jwt>` | همه مسیرهای غیرعمومی |
| توکن داخلی محک | `X-Internal-Token: <INTERNAL_API_TOKEN>` | `/auth/external` و `/webhooks/events` |
| توکن ماموریت | بدنه `token` یا `X-Mission-Token` | فقط وب‌هوک ماموریت |

نقش JWT: `volunteer` · `operator` (بهره‌بردار / واحد پشتیبانی) · `admin`. مسیرهای `/api/v1/admin/*` برای `admin` و `operator` یکسان است.

## عمومی

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/healthz` | سلامت |
| GET | `/readyz` | آمادگی Postgres |
| GET | `/api/v1/` | فهرست کامل مسیرها |
| POST | `/api/v1/auth/otp/send` | `{ "phone" }` → `{ ttl_seconds, dev_code? }` |
| POST | `/api/v1/auth/otp/verify` | `{ "phone", "code", "full_name?" }` → JWT |
| POST | `/api/v1/auth/login` | `{ "email", "password" }` |
| POST | `/api/v1/auth/register` | ثبت‌نام ایمیل؛ `role` نادیده گرفته می‌شود و همیشه داوطلب است |
| POST | `/api/v1/auth/external` | نگاشت Auth محک؛ نیاز به `X-Internal-Token` |
| GET | `/api/v1/certificates/verify/{code}` | استعلام اصالت |
| GET | `/api/v1/certificates/{code}/pdf` | PDF تقدیرنامه |
| POST | `/api/v1/webhooks/events` | پیشرفت ماموریت از سیستم بالادستی |

نمونه وب‌هوک:

```http
POST /api/v1/webhooks/events
X-Internal-Token: <INTERNAL_API_TOKEN>
X-Mission-Token: <verify_token ماموریت>
Content-Type: application/json

{ "event": "user.invited", "volunteer_id": "<uuid>", "phone": "0912…", "increment": 1, "token": "<verify_token>" }
```

`Authorization: Bearer` را برای توکن ماموریت نفرستید اگر همان توکن داخلی است.

## نشست و اعلان

| روش | مسیر |
| --- | --- |
| GET | `/api/v1/me` |
| GET | `/api/v1/notifications` |
| POST | `/api/v1/notifications/{id}/read` |
| POST | `/api/v1/notifications/read-all` |

`GET /me` برای داوطلب فیلد `volunteer` را هم برمی‌گرداند. وضعیت `suspended` برای خود داوطلب به‌صورت `approved` دیده می‌شود تا برچسب تعلیق نمایش داده نشود؛ فعالیت جدید نمی‌بیند.

## داوطلب — پروفایل

| روش | مسیر | بدنه |
| --- | --- | --- |
| GET | `/volunteers/me` | — |
| PUT | `/volunteers/me` | `first_name`, `last_name`, `national_id`, `phone2`, `province`, `city`, `address`, `plaque`, `unit`, `bio`, `skill_ids`, `education_level`, `education_field`, `medical_license`, `birth_date` (حداقل ۱۸ سال), `gender` (`male`/`female`), `occupation`, `occupation_other` |
| POST | `/volunteers/me/submit` | ارسال برای بررسی |
| GET/PUT | `/volunteers/me/availability` | `{ "slots": [{ "weekday", "start_time", "end_time" }] }` — `weekday`: ۰=یکشنبه |
| GET/POST | `/volunteers/me/documents` | `multipart`: `kind`, `file` (JPG/PNG/PDF تا ۵MB) |
| DELETE | `/volunteers/me/documents/{id}` | قبل از تایید |
| GET | `/volunteers/me/trainings` | دوره‌های آموزشی گذرانده‌شده |
| GET | `/skills` | کاتالوگ گروه و مهارت |
| POST | `/volunteers/me/skill-proposals` | `{ "group_id", "title" }` |
| GET | `/volunteers/me/skill-proposals` | پیشنهادهای من |

مدارک: `national_id` · `driving_license` · `medical_license` · `education` · `other`

## داوطلب — فعالیت و کار

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/tasks` | واجد شرایط (`q`, `skill`, `limit`, `offset`) |
| GET | `/tasks/{id}` | جزئیات شامل آموزش |
| POST | `/tasks/{id}/accept` | درخواست / رزرو ظرفیت |
| GET | `/assignments/me` | کارهای من (پس از تایید بهره‌بردار) |
| POST | `/assignments/{id}/start` | شروع دورکار |
| POST | `/assignments/{id}/deliver` | `multipart`: `note`, `file` |
| POST | `/assignments/{id}/rate` | `{ "rating": 1-5, "comment" }` پس از تکمیل |
| POST | `/assignments/{id}/cancel` | انصراف |

وضعیت تخصیص: `requested` → `reserved` / `in_progress` → `attended` | `submitted` | `revision_requested` → `completed` | `rejected` | `absent` | `cancelled`

- **حضوری (`onsite`)**: داوطلب شروع/تحویل ندارد؛ بهره‌بردار حضور یا غیبت می‌زند.
- **دورکار (`remote`)**: داوطلب شروع و بارگذاری نتیجه؛ بهره‌بردار تایید، رد، یا درخواست اصلاح (`revision`).
- **نیاز به آموزش**: هنگام تعریف فعالیت، `requires_training`، `training_course_id` (انتخاب از فهرست دوره‌های ازپیش‌تعریف‌شده) و `training_at` (زمان جلسه این فعالیت؛ باید قبل از شروع فعالیت باشد). داوطلبی که همان دوره را گذرانده باشد برای هر فعالیت با همان دوره نیاز به آموزش مجدد ندارد. تا تایید آموزش، شروع/حضور/تکمیل فعالیت ممکن نیست.

GET `/volunteers/me/trainings` دوره‌های تاییدشده را در پروفایل داوطلب نشان می‌دهد.

## داوطلب — ماموریت و گواهی

| روش | مسیر |
| --- | --- |
| GET | `/missions` |
| POST | `/missions/{id}/start` |
| POST | `/missions/{id}/progress` | بررسی تأیید (نه ثبت دستی) |
| GET | `/missions/me` |
| GET | `/certificates/me` |
| GET/POST | `/certificates/requests` | `{ "kind": "task"\|"aggregated"\|"official", "assignment_id?" }` |

`kind=official` فقط با حداقل ۹۰ ساعت تاییدشده. فرآیند: `preparing` → `ready` → `delivered` (`send` یا `in_person`).

## داوطلب — تیکت

| روش | مسیر | بدنه |
| --- | --- | --- |
| GET | `/tickets/me` | — |
| POST | `/tickets` | `{ "subject", "body" }` |
| GET | `/tickets/{id}` | — |
| POST | `/tickets/{id}/messages` | `{ "body" }` |

## بهره‌بردار (`/admin/*`) — داوطلب

نیاز به JWT با نقش `admin` یا `operator`.

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/admin/dashboard` | صف‌ها و آمار |
| GET | `/admin/volunteers` | `status`, `skill`, `q`, `attention`, `limit`, `offset` |
| GET | `/admin/volunteers/{id}` | پرونده کامل + مدارک + زمان آزاد + تخصیص + ماموریت |
| PUT | `/admin/volunteers/{id}` | ویرایش پروفایل و `skill_ids` |
| POST | `/admin/volunteers/{id}/review` | `approve` / `reject` / `request_documents` / `suspend` / `unsuspend` — برای رد و نقص مدرک `reason` الزامی |
| POST | `/admin/volunteers/{id}/status` | تغییر مستقیم وضعیت |
| POST | `/admin/volunteers/{id}/comments` | `{ "comment" }` |
| GET | `/admin/volunteers/{id}/documents` | — |
| GET | `/admin/volunteers/{id}/availability` | — |
| GET | `/admin/documents/{id}` | استریم فایل |

```json
{ "action": "approve" }
{ "action": "request_documents", "reason": "تصویر کارت ملی ناخوانا است" }
{ "action": "reject", "reason": "عدم تطابق مدارک" }
```

## بهره‌بردار — فعالیت

| روش | مسیر |
| --- | --- |
| GET/POST | `/admin/tasks` |
| PUT/DELETE | `/admin/tasks/{id}` |
| POST | `/admin/tasks/{id}/status` `{ "status": "open"\|"closed"\|"cancelled"\|"inactive" }` |
| POST | `/admin/tasks/{id}/assign` `{ "volunteer_id" }` |
| GET | `/admin/tasks/{id}/assignments` |

بدنه تعریف فعالیت: `title`, `description`, `location`, `starts_at`, `ends_at` (RFC3339), `capacity`, `hour_weight`, `required_skills`, `required_skill_ids`, `min_score`, `work_mode` (`onsite`/`remote`), `delivery_hint`, `requires_training`, `training_course_id`, `training_at`, `kind` (`one_off`/`recurring`), `slots`.

اگر `requires_training` باشد باید `training_course_id` از فهرست دوره‌های آموزشی انتخاب شود و `training_at` قبل از شروع فعالیت باشد. نوع و محل آموزش از روی دوره کپی می‌شود؛ زمان جلسه روی خود فعالیت ذخیره می‌شود.

مهارت `general` یعنی همه داوطلبان فعال می‌توانند درخواست بدهند.

## بهره‌بردار — تخصیص و حضور

| روش | مسیر | بدنه |
| --- | --- | --- |
| GET | `/admin/assignments` | فیلتر `status`, `volunteer_id`, `task_id`, `series_id` |
| POST | `/admin/assignments/{id}/approve` | — |
| POST | `/admin/assignments/{id}/confirm-training` | تایید حضور در دوره؛ تا تایید، فعالیت ادامه پیدا نمی‌کند |
| POST | `/admin/assignments/{id}/reject` | `{ "comment" }` |
| POST | `/admin/assignments/{id}/revision` | `{ "comment" }` الزامی برای دورکار |
| POST | `/admin/assignments/{id}/message` | `{ "body" }` |
| POST | `/admin/assignments/{id}/attendance` | `{ "check_in_at?", "check_out_at?" }` RFC3339؛ خالی = الان |
| POST | `/admin/assignments/{id}/absent` | — |
| POST | `/admin/assignments/{id}/complete` | `{ "discipline", "expertise", "ethics" }` هر کدام ۱–۵، `comment` |
| POST | `/admin/assignments/{id}/cancel` | — |
| POST | `/admin/assignments/{id}/certificate` | تقدیرنامه موردی |
| GET | `/admin/assignments/{id}/delivery` | فایل نتیجه دورکار |

## بهره‌بردار — آموزش

| روش | مسیر |
| --- | --- |
| GET | `/admin/training-courses` `?active=1` |
| POST | `/admin/training-courses` `{ "title", "kind", "location", "description?", "status?" }` |
| GET | `/admin/training-courses/{id}` |
| PUT | `/admin/training-courses/{id}` |

`kind`: `in_person` / `online` / `hybrid` / `workshop` / `other`. `status`: `active` / `inactive`. نام دوره یکتا است. زمان جلسه روی دوره ذخیره نمی‌شود.

پس از تایید درخواست فعالیت نیازمند آموزش، تخصیص با وضعیت `training_pending` در صف تایید آموزش می‌ماند تا `confirm-training` زده شود. تا آن لحظه حتی بهره‌بردار هم نمی‌تواند فرایند فعالیت را ادامه دهد.

## بهره‌بردار — ماموریت، تیکت، گواهی، گزارش، مهارت

| روش | مسیر |
| --- | --- |
| GET/POST | `/admin/missions` |
| PUT | `/admin/missions/{id}` |
| GET | `/admin/tickets` `?status=` |
| GET | `/admin/tickets/{id}` |
| POST | `/admin/tickets/{id}/messages` |
| POST | `/admin/tickets/{id}/status` `{ "status": "open"\|"answered"\|"closed" }` |
| POST | `/admin/volunteers/{id}/certificates/aggregated` | بازه ثابت ۱۲ ماه اخیر |
| GET | `/admin/certificate-requests` |
| POST | `/admin/certificate-requests/{id}/review` | تقدیرنامه: `approve`/`reject` · گواهی‌نامه: `approve`/`reject`/`deliver` + `delivery_method` |
| GET | `/admin/reports/ranking` `?format=csv` |
| GET | `/admin/reports/skills` |
| GET | `/admin/reports/overview` |
| GET | `/admin/skills/` |
| POST | `/admin/skills/groups` `{ "slug", "title", "sort_order" }` |
| PUT/DELETE | `/admin/skills/groups/{id}` |
| POST | `/admin/skills/` `{ "group_id", "title" }` |
| PUT/DELETE | `/admin/skills/{id}` |
| GET | `/admin/skills/proposals` |
| POST | `/admin/skills/proposals/{id}/review` | `approve` / `edit` / `reject` / `edit_approve` |

مسیرهای قدیمی `/admin/skill-catalog/*` و `/admin/skill-proposals` همان handlerها هستند.

نمونه بررسی مهارت:

```json
{ "action": "approve" }
{ "action": "reject", "admin_note": "عنوان مبهم است" }
{ "action": "edit_approve", "title": "گرافیک رایانه‌ای", "group_id": "<uuid>" }
```
