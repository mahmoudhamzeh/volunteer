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
| POST | `/api/v1/webhooks/events` | پیشرفت ماموریت از سرویس خارجی. هدر `Authorization: Bearer <verify_token>` و بدنه `{ "event", "phone" یا "volunteer_id", "increment" }` |

## داوطلب

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/me` | کاربر جاری |
| PUT | `/volunteers/me` | پروفایل (استان، شهر، آدرس، پلاک، واحد، شماره دوم، تحصیلات، skill_ids) |
| POST | `/volunteers/me/submit` | ارسال برای بررسی (پیام خطای فارسی) |
| POST | `/volunteers/me/documents` | آپلود مدرک (`multipart`) |
| DELETE | `/volunteers/me/documents/{id}` | حذف مدرک قبل از تایید/تعلیق ادمین |
| GET | `/skills` | کاتالوگ گروه و زیرمهارت |
| POST | `/volunteers/me/skill-proposals` | پیشنهاد مهارت جدید |
| PUT | `/volunteers/me/availability` | تقویم زمانی |
| GET | `/tasks` | تسک‌های واجد شرایط |
| POST | `/tasks/{id}/accept` | پذیرش / رزرو ظرفیت |
| GET | `/assignments/me` | کارهای من |
| POST | `/assignments/{id}/rate` | امتیاز داوطلب به سازماندهی |
| GET/POST | `/missions` ، `/missions/{id}/start` ، `/missions/{id}/progress` | لیست، شروع، و **بررسی تأیید** (نه ثبت دستی پیشرفت) |

## ادمین

| روش | مسیر | توضیح |
| --- | --- | --- |
| GET | `/admin/dashboard` | مانیتورینگ |
| GET | `/admin/volunteers` | فیلتر status/skill/q (ایمیل کاربر هم برمی‌گردد) |
| PUT | `/admin/volunteers/{id}` | ویرایش اطلاعات پرونده توسط ادمین |
| POST | `/admin/volunteers/{id}/review` | `approve` / `reject` / `request_documents` / `suspend` — برای رد و درخواست مدرک فیلد `reason` الزامی است (پیام فارسی) |
| POST | `/admin/volunteers/{id}/status` | تغییر مستقیم وضعیت؛ برای `rejected` فیلد `reason` الزامی است |
| POST | `/admin/volunteers/{id}/comments` | ثبت پیام/کامنت در تاریخچه و ارسال اعلان به داوطلب |
| CRUD | `/admin/tasks` | تعریف تسک با شرط مهارت و امتیاز |
| POST | `/admin/assignments/{id}/attendance` | حضور |
| POST | `/admin/assignments/{id}/complete` | امتیاز ۱–۵ و ساعات |
| POST | `/admin/assignments/{id}/certificate` | گواهی موردی |
| GET | `/admin/skill-catalog` | گروه‌ها و مهارت‌ها |
| POST | `/admin/skill-catalog/groups` | افزودن گروه |
| POST | `/admin/skill-catalog/skills` | افزودن زیرمهارت |
| PUT | `/admin/skill-catalog/skills/{id}` | ویرایش مهارت |
| GET | `/admin/skill-proposals` | پیشنهادهای مهارت (`?status=pending`) |
| POST | `/admin/skill-proposals/{id}/review` | `approve` / `edit_approve` / `reject` |

نمونه بررسی:

```json
{ "action": "approve" }
{ "action": "request_documents", "reason": "تصویر کارت ملی ناخوانا است" }
{ "action": "reject", "reason": "عدم تطابق مدارک" }
```

تاریخچه پرونده (`history`) در پاسخ `GET /volunteers/me` و `GET /admin/volunteers/{id}` برمی‌گردد.
