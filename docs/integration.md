# تحویل وب‌سرویس به تیم پیاده‌سازی

این بسته برای شرکتی است که کلاینت (وب، موبایل، یا سرویس داخلی محک) را جداگانه پیاده می‌کند. سرور همین مخزن است؛ قرارداد HTTP اینجا ثابت است.

## فایل‌های قرارداد

| فایل | کاربرد |
| --- | --- |
| [`docs/openapi.yaml`](openapi.yaml) | مشخصات OpenAPI 3.0 — تولید کلاینت / Swagger |
| [`docs/api.md`](api.md) | مرجع فارسی همه مسیرها و بدنه‌ها |
| [`postman/Mahak-Volunteer-Management.postman_collection.json`](../postman/Mahak-Volunteer-Management.postman_collection.json) | کالکشن قابل اجرا |
| [`postman/Mahak-Volunteer-Management.postman_environment.json`](../postman/Mahak-Volunteer-Management.postman_environment.json) | متغیرهای محیط |
| `GET /api/v1/` | کاتالوگ زنده روی سرور |

Import کردن OpenAPI در Postman یا Insomnia همان قرارداد را می‌دهد. اگر مسیر جدیدی به روتر اضافه شود و در کاتالوگ نباشد، تست `TestCatalogCoversEveryRegisteredRoute` رد می‌شود.

## آدرس‌ها

| محیط | Base |
| --- | --- |
| توسعه | `http://localhost:8080` |
| از طریق پورتال Next | `/api/v1/...` (پروکسی به API) |
| سرور | `http://<host>:8080` یا دامنه پشت Nginx |

سلامت: `GET /healthz` · آمادگی دیتابیس: `GET /readyz`

## نقش‌ها

| `role` | فارسی | دسترسی |
| --- | --- | --- |
| `volunteer` | داوطلب | پروفایل، فعالیت، ماموریت، گواهی، تیکت |
| `operator` | بهره‌بردار / واحد پشتیبانی | همه `/api/v1/admin/*` |
| `admin` | ادمین سامانه | همان `/api/v1/admin/*` |

حساب نمونه (`SEED_DEMO=true`):

- ادمین: `admin@mahak.ir` / `Admin@123`
- بهره‌بردار: `operator@mahak.ir` / `Operator@123`
- داوطلب: `volunteer@mahak.ir` / `Volunteer@123`

ثبت‌نام عمومی (OTP یا ایمیل) همیشه `volunteer` می‌سازد. ساخت staff فقط از seed یا `POST /api/v1/auth/external`.

ورود پنل فعلی: داوطلب از `/register` و `/login?as=volunteer`؛ بهره‌بردار و ادمین از `/login?as=admin` (همان شل `/admin`).

## احراز JWT

```http
Authorization: Bearer <jwt>
Content-Type: application/json
```

پاسخ ورود/OTP:

```json
{ "token": "<jwt>", "user": { "id": "...", "email": "...", "phone": "...", "role": "volunteer|operator|admin" } }
```

## یکپارچگی داخلی محک

### نگاشت کاربر Auth

```http
POST /api/v1/auth/external
X-Internal-Token: <INTERNAL_API_TOKEN>
```

```json
{
  "external_user_id": "sub از سرویس Auth",
  "email": "optional",
  "full_name": "optional",
  "role": "volunteer | operator | admin"
}
```

`role` فقط هنگام **ایجاد** کاربر اعمال می‌شود؛ کاربر موجود به‌روز نمی‌شود.

### وب‌هوک ماموریت

دو لایه توکن جدا است:

1. دروازه: `X-Internal-Token` = `INTERNAL_API_TOKEN`
2. ماموریت: فیلد JSON `token` یا هدر `X-Mission-Token` = `verify_token` همان ماموریت (از `GET /admin/missions`)

```http
POST /api/v1/webhooks/events
X-Internal-Token: <INTERNAL_API_TOKEN>
Content-Type: application/json

{
  "event": "user.invited",
  "volunteer_id": "<uuid>",
  "phone": "09121234567",
  "increment": 1,
  "token": "<mission verify_token>"
}
```

اگر فقط `Authorization: Bearer <INTERNAL_API_TOKEN>` بفرستید، آن مقدار به‌عنوان توکن ماموریت قبول **نمی‌شود**.

ماموریت با `verify_mode=outbound` هنگام «بررسی تأیید» داوطلب، API شما را با Bearer `verify_token` صدا می‌زند.

## قرارداد خطا

همه خطاها JSON:

```json
{ "error": "ظرفیت این فعالیت تکمیل است" }
```

| HTTP | معنی |
| --- | --- |
| 400 | ورودی / انتقال وضعیت نامعتبر |
| 401 | توکن نیست یا غلط است |
| 403 | نقش staff نیست |
| 404 | منبع نیست |
| 409 | تکراری (ایمیل، درخواست قبلی) |
| 422 | ظرفیت پر، واجدشرایط نیست، ماموریت تأیید نشده |
| 503 | قفل شلوغ یا دیتابیس آماده نیست |

## جریان‌های اجباری کلاینت

1. **عضویت:** OTP → تکمیل پروفایل (هویت، نشانی، تحصیل، مهارت، کارت ملی) → `submit` → بررسی بهره‌بردار.
2. **فعالیت حضوری:** درخواست → تایید بهره‌بردار → حضور/غیبت دستی → امتیاز ۱–۵.
3. **فعالیت دورکار:** درخواست → تایید → `start` → `deliver` → تایید / اصلاح / رد.
4. **آموزش:** اگر `requires_training=true` داوطلب باید قبل از درخواست آموزش را تایید کند.
5. **گواهی‌نامه رسمی:** حداقل ۹۰ ساعت → درخواست `official` → آماده‌سازی → صدور → تحویل ارسال/حضوری.
6. **تقدیرنامه:** موردی روی تخصیص تکمیل‌شده، یا تجمیعی، یا درخواست داوطلب با `kind=task|aggregated`.

آرایه‌های خالی همیشه `[]` هستند نه `null`.

## محدودیت پیاده‌سازی فعلی (سمت سرور)

- OTP در محیط بررسی با `OTP_REVEAL=true` فیلد `dev_code` برمی‌گرداند؛ SMS واقعی به درگاه پیامک وصل نیست.
- اعلان‌ها درون‌برنامه‌ای‌اند (جدول `notifications`)؛ پوش/ایمیل جدا نیست.
- یادآوری آموزش هنگام لیست فعالیت شلیک می‌شود، نه با cron جدا.

این موارد مانع پیاده‌سازی کلاینت نیستند؛ قرارداد HTTP کامل است.
