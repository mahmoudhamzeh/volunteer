# راهنمای استقرار لایو — سامانه مدیریت داوطلبان محک

هدف این راهنما رساندن پلتفرم به یک سرور برای بررسی ذینفعان و بار واقعی است.

## نام‌گذاری سرویس‌ها روی سرور

| کانتینر | نقش | پورت پیش‌فرض |
| --- | --- | --- |
| `mahak-volunteer-portal` | رابط کاربری Next.js | 3000 |
| `mahak-volunteer-api` | REST API | 8080 |
| `mahak-volunteer-db` | PostgreSQL | 5432 |
| `mahak-volunteer-redis` | قفل توزیع‌شده ظرفیت تسک | 6379 |

Compose project name: `mahak-volunteers`

## استقرار از ویندوز (PowerShell)

روی ویندوز، فقط `root@IP` را عوض کنید و کل بلوک را در PowerShell بزنید. `.env` و پوشهٔ `data` روی سرور دست نمی‌خورند. `docker compose down -v` نزنید.

شاخهٔ کامل فعلی: `cursor/structure-webservice-review-bc50`

```powershell
$Server = "root@YOUR_SERVER_IP"

$cmd = @'
set -euo pipefail
BRANCH="cursor/structure-webservice-review-bc50"
TAR=/tmp/mahak-volunteers.tgz
SRC=/tmp/volunteer-src
APP=/opt/mahak-volunteers
FOLDER=$(echo "volunteer-${BRANCH}" | tr "/" "-")

timeout 90 curl -4 -fL -o "$TAR" \
  "https://github.com/mahmoudhamzeh/volunteer/archive/refs/heads/${BRANCH}.tar.gz"
rm -rf "$SRC" && mkdir -p "$SRC"
tar -xzf "$TAR" -C "$SRC"
mkdir -p "$APP"
if [ -f "$APP/.env" ]; then cp -a "$APP/.env" "$APP/.env.bak"; fi
rsync -a --delete \
  --exclude ".env" --exclude ".env.bak" --exclude ".git" --exclude "data" \
  "$SRC/$FOLDER/" "$APP/"
if [ ! -f "$APP/.env" ]; then cp "$APP/.env.example" "$APP/.env"; fi
cd "$APP"
docker compose --env-file .env up -d --build --force-recreate api web
docker compose ps
curl -fsS http://127.0.0.1:8080/readyz
curl -fsS http://127.0.0.1:8080/api/v1/ | head -c 200
'@

$cmd | ssh $Server bash
```

بعد از اتمام، در مرورگر **Ctrl+F5**. پورتال: `http://IP:3000` — API: `http://IP:8080/api/v1/`

## استقرار با Docker Compose

روی سرور:

```bash
git clone https://github.com/mahmoudhamzeh/volunteer.git
cd volunteer
cp .env.example .env
```

`.env` را برای محیط بررسی ویرایش کنید:

```bash
APP_ENV=development          # برای بررسی ذینفعان؛ در بهره‌برداری نهایی production
POSTGRES_PASSWORD=<قوی>
JWT_SECRET=<حداقل ۳۲ کاراکتر تصادفی>
INTERNAL_API_TOKEN=<توکن سرویس‌های داخلی محک>
PUBLIC_BASE_URL=https://volunteers.example.ir
CORS_ORIGINS=https://volunteers.example.ir
SEED_DEMO=true               # حساب‌های نمونه برای بررسی ذینفعان
WEB_PORT=3000
API_PORT=8080
```

اگر `APP_ENV=production` باشد و `JWT_SECRET` مقدار پیش‌فرض بماند، API بالا نمی‌آید.

```bash
docker compose --env-file .env up -d --build
docker compose ps
curl -fsS http://127.0.0.1:8080/readyz
```

پورتال را پشت Nginx/Caddy با TLS قرار دهید و `PUBLIC_BASE_URL` را روی همان دامنه بگذارید (QR گواهی به این آدرس اشاره می‌کند).

نمونهٔ معکوس‌کننده Nginx:

```nginx
server {
    listen 443 ssl;
    server_name volunteers.example.ir;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

پورتال Next درخواست‌های `/api/*` را به API داخلی پروکسی می‌کند؛ نیازی به افشای پورت 8080 روی اینترنت نیست مگر برای Postman مستقیم.

## متغیرهای محیطی API (`mahak-volunteer-api`)

| متغیر | پیش‌فرض | توضیح |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | آدرس گوش‌دادن |
| `APP_ENV` | `development` | در `production` رمز JWT پیش‌فرض رد می‌شود |
| `DATABASE_URL` | postgres://mahak:mahak@127.0.0.1:5432/mahak_volunteers?sslmode=disable | اتصال Postgres |
| `DB_MAX_CONNS` | `40` | سقف اتصال استخر |
| `JWT_SECRET` | مقدار توسعه | کلید امضای JWT |
| `JWT_TTL_HOURS` | `24` | عمر توکن |
| `REDIS_URL` | redis://127.0.0.1:6379 | قفل ظرفیت؛ اگر در دسترس نباشد قفل حافظه‌ای |
| `STORAGE_DIR` | ./data/uploads | مدارک داوطلب |
| `PUBLIC_BASE_URL` | http://localhost:3000 | پایه URL برای QR گواهی |
| `INTERNAL_API_TOKEN` | خالی | الزامی برای `/auth/external` و `/webhooks/events` |
| `CORS_ORIGINS` | `*` | لیست مبدأها با ویرگول |
| `SEED_DEMO` | در توسعه `true` | ساخت حساب‌های نمونه |

اسکیمای دیتابیس هنگام استارت از `backend/internal/adapter/postgres/schema.sql` اعمال می‌شود.

## یکپارچگی Auth و Notification محک

لایه Auth موجود باید `sub` کاربر را با هدر `X-Internal-Token` به `POST /api/v1/auth/external` بفرستد.

وب‌هوک ماموریت‌ها: `POST /api/v1/webhooks/events` با همان توکن داخلی.

MinIO در این نسخه اختیاری است (`docker compose --profile object-storage up`)؛ مدارک فعلاً روی فایل‌سیستم ذخیره می‌شوند.

## بررسی سلامت زیر بار

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
docker compose logs -f api web
```

لاگ JSON کانتینرها با سقف ۲۰ مگابایت نگهداری می‌شود. برای بار همزمان، `DB_MAX_CONNS` و منابع CPU/RAM سرور را متناسب با تعداد داوطلب همزمان تنظیم کنید.
