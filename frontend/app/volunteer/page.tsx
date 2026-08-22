"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api, Assignment, Mission, Notification, Volunteer } from "@/lib/api";
import { Badge, Button, Card, StarRating } from "@/components/ui";
import { notificationHref } from "@/lib/labels";

export default function VolunteerHome() {
  const [me, setMe] = useState<Volunteer | null>(null);
  const [work, setWork] = useState<Assignment[]>([]);
  const [notes, setNotes] = useState<Notification[]>([]);
  const [missions, setMissions] = useState<Mission[]>([]);

  async function loadNotes() {
    const x = await api.notifications().catch(() => [] as Notification[]);
    setNotes(x || []);
  }

  useEffect(() => {
    api.me().then((r) => setMe(r.volunteer || null)).catch(() => undefined);
    api.myAssignments().then((x) => setWork(x || [])).catch(() => undefined);
    void loadNotes();
    api.missions().then((x) => setMissions(x || [])).catch(() => undefined);
  }, []);

  const unread = useMemo(() => (notes || []).filter((n) => !n.read), [notes]);
  const needsDocs = me?.status === "draft" && Boolean(me?.rejection_reason);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-black">سلام {me?.full_name || ""}</h1>
        <p className="text-stone-500">وضعیت عضویت شما: {me ? <Badge status={me.status} /> : "—"}</p>
        {me?.email && <p className="mt-1 text-sm text-stone-500">ایمیل: {me.email}</p>}
      </div>
      {needsDocs && (
        <Card className="border-rose-200 p-5">
          <p className="font-medium text-rose-800">ادمین مدارک تکمیلی خواسته است.</p>
          <p className="mt-1 whitespace-pre-wrap text-sm text-rose-700">{me?.rejection_reason}</p>
          <Link href="/volunteer/profile?tab=docs" className="mt-3 inline-block">
            <Button>رفتن به بارگذاری مدارک</Button>
          </Link>
        </Card>
      )}
      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="p-5">
          <div className="text-sm text-stone-500">ساعات داوطلبی</div>
          <div className="mt-1 text-3xl font-black text-mahak-600">{me?.total_hours ?? 0}</div>
        </Card>
        <Card className="p-5">
          <div className="text-sm text-stone-500">میانگین امتیاز</div>
          <div className="mt-2">
            <StarRating value={me?.average_score || 0} readOnly />
          </div>
        </Card>
        <Card className="p-5">
          <div className="text-sm text-stone-500">فعالیت‌های تکمیل‌شده</div>
          <div className="mt-1 text-3xl font-black">{me?.completed_tasks ?? 0}</div>
        </Card>
      </div>
      {me?.status !== "approved" && !needsDocs && (
        <Card className="p-5">
          <p className="font-medium">برای مشاهده فعالیت‌های عملیاتی باید پروفایل تایید شود.</p>
          <Link href="/volunteer/profile" className="mt-2 inline-block text-sm text-mahak-700">
            تکمیل پروفایل و ارسال مدارک
          </Link>
        </Card>
      )}
      <Card className="p-5">
        <div className="mb-3 flex items-center justify-between gap-2">
          <h2 className="font-bold">اعلان‌های جدید</h2>
          <Link href="/volunteer/notifications" className="text-sm text-mahak-700">همه اعلان‌ها</Link>
        </div>
        {unread.length === 0 && <p className="text-sm text-stone-400">اعلان خوانده‌نشده‌ای نیست.</p>}
        <ul className="space-y-2">
          {unread.slice(0, 5).map((n) => (
            <li key={n.id} className="rounded-2xl border border-mahak-100 bg-mahak-50/70 px-3 py-2 text-sm">
              <Link href={notificationHref(n.title)} className="block" onClick={() => { void api.markRead(n.id); }}>
                <div className="flex items-center gap-2 font-medium">
                  <span className="h-2 w-2 rounded-full bg-rose-500" />
                  {n.title}
                </div>
                <div className="text-stone-600">{n.body}</div>
              </Link>
            </li>
          ))}
        </ul>
      </Card>
      <div className="grid gap-4 md:grid-cols-2">
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
          <Link href="/volunteer/missions" className="mt-3 inline-block text-sm text-mahak-700">مشاهده همه</Link>
        </Card>
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
        </Card>
      </div>
    </div>
  );
}
