"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { api, Mission } from "@/lib/api";
import { Badge, Button, Card, Field, inputClass } from "@/components/ui";
import { MISSION_KIND_LABEL, VERIFY_MODE_LABEL } from "@/lib/labels";

const emptyForm = {
  title: "",
  description: "",
  kind: "complete_profile",
  verify_mode: "internal",
  hour_weight: 1,
  target_count: 1,
  deadline_hours: 0,
  webhook_event: "",
  verify_url: "",
};

function defaultMode(kind: string) {
  if (kind === "complete_profile") return "internal";
  if (kind === "invite_users" || kind === "webhook") return "inbound";
  return "outbound";
}

export default function AdminMissions() {
  const [items, setItems] = useState<Mission[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [msg, setMsg] = useState("");
  const webhookPath = useMemo(
    () => (typeof window === "undefined" ? "/api/v1/webhooks/events" : `${window.location.origin}/api/v1/webhooks/events`),
    [],
  );

  async function load() {
    setItems((await api.adminMissions()) || []);
  }
  useEffect(() => {
    load();
  }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setMsg("");
    try {
      await api.createMission({
        ...form,
        deadline_hours: form.deadline_hours > 0 ? form.deadline_hours : null,
      });
      setForm({ ...emptyForm, kind: form.kind, verify_mode: form.verify_mode });
      await load();
      setMsg("ماموریت ساخته شد. اگر وب‌هوک دارد، توکن را از کارت زیر کپی کنید.");
    } catch (err) {
      setMsg(err instanceof Error ? err.message : "خطا");
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">ماموریت‌های سیستمی</h1>
      <Card className="space-y-3 p-5 text-sm leading-7 text-stone-700">
        <p className="font-bold text-ink-800">ماموریت با کلیک داوطلب تمام نمی‌شود.</p>
        <p>
          وقتی داوطلب «بررسی تأیید» می‌زند، سامانه باید از روی داده واقعی یا وب‌سرویس شما مطمئن شود کار انجام شده است.
          هنگام تعریف ماموریت یکی از این روش‌ها را انتخاب کنید:
        </p>
        <ul className="list-disc pr-5 space-y-1">
          <li>
            <b>بررسی داخلی:</b> برای تکمیل پروفایل. سامانه چک می‌کند پروفایل، مهارت و کارت ملی واقعاً ارسال شده باشد.
          </li>
          <li>
            <b>فراخوانی وب‌سرویس:</b> آدرس و توکن را بگذارید. با کلیک داوطلب، همین سامانه با Bearer توکن به آدرس شما POST می‌زند و انتظار JSON مثل
            {" "}
            <code dir="ltr">{`{ "ok": true, "progress": 5 }`}</code>
            {" "}
            دارد.
          </li>
          <li>
            <b>وب‌هوک ورودی:</b> توکن را به سرویس دیگر بدهید تا خودش پیشرفت را اعلام کند (مثلاً دعوت کاربر). داوطلب نمی‌تواند آن را دستی تمام کند.
          </li>
        </ul>
      </Card>
      <Card className="p-5">
        <form onSubmit={onSubmit} className="grid gap-3 md:grid-cols-2">
          <Field label="عنوان">
            <input className={inputClass} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
          </Field>
          <Field label="نوع ماموریت">
            <select
              className={inputClass}
              value={form.kind}
              onChange={(e) => {
                const kind = e.target.value;
                setForm({ ...form, kind, verify_mode: defaultMode(kind) });
              }}
            >
              <option value="complete_profile">تکمیل پروفایل</option>
              <option value="invite_users">دعوت کاربر</option>
              <option value="custom">سفارشی</option>
              <option value="webhook">رویداد وب‌هوک</option>
            </select>
          </Field>
          <Field label="روش تأیید">
            <select
              className={inputClass}
              value={form.verify_mode}
              onChange={(e) => setForm({ ...form, verify_mode: e.target.value })}
              disabled={form.kind === "complete_profile"}
            >
              <option value="internal">بررسی داخلی سامانه</option>
              <option value="outbound">فراخوانی وب‌سرویس شما</option>
              <option value="inbound">وب‌هوک ورودی با توکن</option>
            </select>
          </Field>
          <Field label="رویداد وب‌هوک (اختیاری)">
            <input
              className={inputClass}
              dir="ltr"
              placeholder="user.invited"
              value={form.webhook_event}
              onChange={(e) => setForm({ ...form, webhook_event: e.target.value })}
            />
          </Field>
          {(form.verify_mode === "outbound" || form.verify_mode === "inbound") && (
            <div className="md:col-span-2">
              <Field label={form.verify_mode === "outbound" ? "آدرس وب‌سرویس تأیید" : "آدرس اختیاری برای استعلام (اگر سرویس شما pull دارد)"}>
                <input
                  className={inputClass}
                  dir="ltr"
                  placeholder="https://example.com/missions/verify"
                  value={form.verify_url}
                  onChange={(e) => setForm({ ...form, verify_url: e.target.value })}
                  required={form.verify_mode === "outbound"}
                />
              </Field>
            </div>
          )}
          <div className="md:col-span-2">
            <Field label="شرح">
              <textarea className={inputClass} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            </Field>
          </div>
          <Field label="وزن ساعتی">
            <input type="number" className={inputClass} value={form.hour_weight} onChange={(e) => setForm({ ...form, hour_weight: Number(e.target.value) })} />
          </Field>
          <Field label="تعداد هدف">
            <input type="number" className={inputClass} value={form.target_count} onChange={(e) => setForm({ ...form, target_count: Number(e.target.value) })} />
          </Field>
          <Field label="مهلت (ساعت، خالی = بدون مهلت)">
            <input type="number" className={inputClass} value={form.deadline_hours || ""} onChange={(e) => setForm({ ...form, deadline_hours: Number(e.target.value) })} />
          </Field>
          <div className="flex items-end">
            <Button type="submit">ایجاد ماموریت</Button>
          </div>
        </form>
        {msg && <p className="mt-3 text-sm text-mahak-700">{msg}</p>}
      </Card>
      {items.map((m) => (
        <Card key={m.id} className="space-y-2 p-4">
          <div className="flex items-start justify-between gap-3">
            <div>
              <div className="font-bold">{m.title}</div>
              <div className="text-xs text-stone-500">
                {MISSION_KIND_LABEL[m.kind] || m.kind} · {VERIFY_MODE_LABEL[m.verify_mode || ""] || m.verify_mode} · {m.hour_weight} ساعت · هدف {m.target_count}
              </div>
            </div>
            <Badge status={m.status} />
          </div>
          {m.description && <p className="text-sm text-stone-600">{m.description}</p>}
          {m.verify_mode === "inbound" && (
            <div className="rounded-2xl bg-stone-50 p-3 text-xs leading-6 text-stone-700">
              <div>سرویس خارجی باید با این مشخصات پیشرفت را اعلام کند:</div>
              <div dir="ltr" className="mt-1 break-all font-mono">POST {webhookPath}</div>
              <div dir="ltr" className="break-all font-mono">Authorization: Bearer {m.verify_token}</div>
              {m.webhook_event && <div dir="ltr" className="font-mono">event: {m.webhook_event}</div>}
              <pre className="mt-2 overflow-x-auto rounded-xl bg-white p-2 font-mono" dir="ltr">{`{
  "event": "${m.webhook_event || "mission.progress"}",
  "phone": "0912xxxxxxx",
  "increment": 1
}`}</pre>
            </div>
          )}
          {m.verify_mode === "outbound" && (
            <div className="rounded-2xl bg-stone-50 p-3 text-xs leading-6 text-stone-700">
              <div>با کلیک داوطلب این آدرس با توکن صدا زده می‌شود:</div>
              <div dir="ltr" className="mt-1 break-all font-mono">{m.verify_url}</div>
              {m.verify_token && <div dir="ltr" className="break-all font-mono">Bearer {m.verify_token}</div>}
            </div>
          )}
          {m.verify_mode === "internal" && (
            <p className="text-xs text-stone-500">تأیید از روی داده همین سامانه انجام می‌شود؛ داوطلب با کلیک خالی نمی‌تواند آن را تمام کند.</p>
          )}
        </Card>
      ))}
    </div>
  );
}
