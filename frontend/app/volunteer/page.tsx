"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Assignment, Mission, Notification, Volunteer } from "@/lib/api";
import { Badge, Card } from "@/components/ui";
import { HistoryList } from "@/components/history";
import { STATUS_EXPLAIN } from "@/lib/labels";

export default function VolunteerHome() {
  const [me, setMe] = useState<Volunteer | null>(null);
  const [work, setWork] = useState<Assignment[]>([]);
  const [notes, setNotes] = useState<Notification[]>([]);
  const [missions, setMissions] = useState<Mission[]>([]);

  useEffect(() => {
    api.me().then((r) => setMe(r.volunteer || null)).catch(() => undefined);
    api.myAssignments().then((x) => setWork(x || [])).catch(() => undefined);
    api.notifications().then((x) => setNotes(x || [])).catch(() => undefined);
    api.missions().then((x) => setMissions(x || [])).catch(() => undefined);
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-black">سلام {me?.full_name || ""}</h1>
        <p className="text-stone-500">وضعیت عضویت شما: {me ? <Badge status={me.status} /> : "—"}</p>
        {me?.status && <p className="mt-1 text-sm text-stone-600">{STATUS_EXPLAIN[me.status]}</p>}
        {me?.email && <p className="mt-1 text-sm text-stone-500">ایمیل: {me.email}</p>}
        {me?.rejection_reason && <p className="mt-2 text-sm text-rose-700">پیام ادمین: {me.rejection_reason}</p>}
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="p-5">
          <div className="text-sm text-stone-500">ساعات داوطلبی</div>
          <div className="mt-1 text-3xl font-black text-mahak-600">{me?.total_hours ?? 0}</div>
        </Card>
        <Card className="p-5">
          <div className="text-sm text-stone-500">میانگین امتیاز</div>
          <div className="mt-1 text-3xl font-black">{me?.average_score?.toFixed(1) ?? "0.0"}</div>
        </Card>
        <Card className="p-5">
          <div className="text-sm text-stone-500">فعالیت‌های تکمیل‌شده</div>
          <div className="mt-1 text-3xl font-black">{me?.completed_tasks ?? 0}</div>
        </Card>
      </div>
      {me?.status !== "approved" && (
        <Card className="p-5">
          <p className="font-medium">برای مشاهده فعالیت‌های عملیاتی باید پروفایل تایید شود.</p>
          <Link href="/volunteer/profile" className="mt-2 inline-block text-sm text-mahak-700">
            تکمیل پروفایل و ارسال مدارک
          </Link>
        </Card>
      )}
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <h2 className="font-bold">اعلان‌ها</h2>
          <ul className="mt-3 space-y-2 text-sm">
            {(notes || []).length === 0 && <li className="text-stone-400">اعلانی نیست</li>}
            {(notes || []).slice(0, 5).map((n) => (
              <li key={n.id}>
                <div className="font-medium">{n.title}</div>
                <div className="text-stone-500">{n.body}</div>
              </li>
            ))}
          </ul>
        </Card>
        <Card className="p-5">
          <h2 className="font-bold">ماموریت‌های فعال</h2>
          <ul className="mt-3 space-y-2 text-sm">
            {(missions || []).slice(0, 4).map((m) => (
              <li key={m.id} className="flex justify-between gap-2">
                <span>{m.title}</span>
                <span className="text-mahak-700">{m.hour_weight} ساعت</span>
              </li>
            ))}
          </ul>
          <Link href="/volunteer/missions" className="mt-3 inline-block text-sm text-mahak-700">
            مشاهده همه
          </Link>
        </Card>
      </div>
      <Card className="p-5">
        <div className="flex items-center justify-between gap-2">
          <h2 className="font-bold">آخرین درخواست‌ها</h2>
          <Link href="/volunteer/work" className="text-sm text-mahak-700">کارهای من</Link>
        </div>
        <ul className="mt-3 space-y-2">
          {(work || []).slice(0, 5).map((a) => (
            <li key={a.id} className="flex items-center justify-between text-sm">
              <Link href="/volunteer/work" className="hover:text-mahak-700">{a.task?.title}</Link>
              <Badge status={a.status} />
            </li>
          ))}
          {(work || []).length === 0 && <li className="text-sm text-stone-400">هنوز درخواستی ثبت نکرده‌اید</li>}
        </ul>
        {(work || []).some((a) => a.status === "reserved" || a.status === "in_progress") && (
          <p className="mt-3 rounded-2xl bg-mahak-50 px-3 py-2 text-sm text-mahak-800">
            فعالیت تاییدشده دارید. برای شروع کار و ارسال نتیجه به «کارهای من» بروید.
          </p>
        )}
      </Card>
      {(me?.history || []).length > 0 && (
        <Card className="p-5">
          <h2 className="mb-3 font-bold">تاریخچه پرونده</h2>
          <HistoryList items={me?.history} />
        </Card>
      )}
    </div>
  );
}
