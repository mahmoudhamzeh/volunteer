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
    ["داوطلبان", d.total_volunteers, "/admin/volunteers"],
    ["در انتظار تایید هویت", d.pending_volunteers, "/admin/volunteers?status=pending"],
    ["درخواست فعالیت", d.pending_task_requests || 0, "/admin/inbox"],
    ["مهارت پیشنهادی", d.pending_skill_proposals || 0, "/admin/skills"],
    ["درخواست گواهی", d.pending_certificates || 0, "/admin/certificates"],
    ["تیکت باز", d.open_tickets || 0, "/admin/tickets"],
    ["فعالیت باز", d.open_tasks, "/admin/tasks"],
    ["ساعات کل", d.total_hours, "/admin/reports"],
  ];
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">داشبورد مدیریت داوطلبان</h1>
      <div className="grid gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {stats.map(([k, v, href]) => (
          <Link key={String(k)} href={String(href)}>
            <Card className="p-4 hover:border-mahak-200">
              <div className="text-xs text-stone-500">{k}</div>
              <div className="mt-1 text-2xl font-black text-ink-900">{v}</div>
            </Card>
          </Link>
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
    </div>
  );
}
