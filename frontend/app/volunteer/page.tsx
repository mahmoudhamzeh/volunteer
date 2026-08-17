"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Assignment, Mission, Notification, Volunteer } from "@/lib/api";
import { Badge, Card } from "@/components/ui";

export default function VolunteerHome() {
  const [me, setMe] = useState<Volunteer | null>(null);
  const [work, setWork] = useState<Assignment[]>([]);
  const [notes, setNotes] = useState<Notification[]>([]);
  const [missions, setMissions] = useState<Mission[]>([]);

  useEffect(() => {
    api.me().then((r) => setMe(r.volunteer || null)).catch(() => undefined);
    api.myAssignments().then(setWork).catch(() => undefined);
    api.notifications().then(setNotes).catch(() => undefined);
    api.missions().then(setMissions).catch(() => undefined);
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-black">سلام {me?.full_name || ""}</h1>
        <p className="text-stone-500">وضعیت عضویت شما: {me ? <Badge status={me.status} /> : "—"}</p>
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
          <div className="text-sm text-stone-500">تسک‌های تکمیل‌شده</div>
          <div className="mt-1 text-3xl font-black">{me?.completed_tasks ?? 0}</div>
        </Card>
      </div>
      {me?.status !== "approved" && (
        <Card className="p-5">
          <p className="font-medium">برای مشاهده تسک‌های عملیاتی باید پروفایل تایید شود.</p>
          <Link href="/volunteer/profile" className="mt-2 inline-block text-sm text-mahak-700">
            تکمیل پروفایل و ارسال مدارک
          </Link>
        </Card>
      )}
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <h2 className="font-bold">اعلان‌ها</h2>
          <ul className="mt-3 space-y-2 text-sm">
            {notes.length === 0 && <li className="text-stone-400">اعلانی نیست</li>}
            {notes.slice(0, 5).map((n) => (
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
            {missions.slice(0, 4).map((m) => (
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
        <h2 className="font-bold">آخرین پذیرش‌ها</h2>
        <ul className="mt-3 space-y-2">
          {work.slice(0, 5).map((a) => (
            <li key={a.id} className="flex items-center justify-between text-sm">
              <span>{a.task?.title}</span>
              <Badge status={a.status} />
            </li>
          ))}
          {work.length === 0 && <li className="text-sm text-stone-400">هنوز تسکی نپذیرفته‌اید</li>}
        </ul>
      </Card>
    </div>
  );
}
