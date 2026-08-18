"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Dashboard } from "@/lib/api";
import { Card } from "@/components/ui";
import { skillLabel } from "@/lib/labels";

export default function AdminHome() {
  const [d, setD] = useState<Dashboard | null>(null);
  useEffect(() => { api.dashboard().then(setD); }, []);
  if (!d) return null;
  const stats = [
    ["داوطلبان", d.total_volunteers],
    ["در انتظار تایید", d.pending_volunteers],
    ["تاییدشده", d.approved_volunteers],
    ["فعالیت باز", d.open_tasks],
    ["در حال اجرا", d.active_assignments],
    ["تکمیل این ماه", d.completed_this_month],
    ["ساعات کل", d.total_hours],
  ];
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">داشبورد مدیریت داوطلبان</h1>
      <div className="grid gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {stats.map(([k, v]) => (
          <Card key={String(k)} className="p-4">
            <div className="text-xs text-stone-500">{k}</div>
            <div className="mt-1 text-2xl font-black text-ink-900">{v}</div>
          </Card>
        ))}
      </div>
      <Card className="p-5">
        <h2 className="font-bold">توزیع تخصص داوطلبان تاییدشده</h2>
        <ul className="mt-3 space-y-2">
          {Object.entries(d.skill_distribution || {}).map(([k, n]) => (
            <li key={k} className="flex justify-between text-sm">
              <span>{skillLabel(k)}</span>
              <span className="font-bold">{n}</span>
            </li>
          ))}
          {Object.keys(d.skill_distribution || {}).length === 0 && <li className="text-sm text-stone-400">هنوز داده‌ای نیست</li>}
        </ul>
      </Card>
      <div className="flex gap-3 text-sm">
        <Link className="text-mahak-700" href="/admin/volunteers?status=pending">صف تایید هویت</Link>
        <Link className="text-mahak-700" href="/admin/skills">مهارت‌ها</Link>
        <Link className="text-mahak-700" href="/admin/assignments">حضور و امتیاز</Link>
        <Link className="text-mahak-700" href="/admin/reports">رتبه‌بندی</Link>
      </div>
    </div>
  );
}
