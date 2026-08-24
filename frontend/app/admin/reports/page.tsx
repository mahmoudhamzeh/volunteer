"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, RankingRow, ReportOverview, downloadAuth } from "@/lib/api";
import { Badge, Card } from "@/components/ui";
import { STATUS_LABEL, catalogLabelMap, skillLabel } from "@/lib/labels";

function Bars({ data, labels }: { data: Record<string, number>; labels?: Record<string, string> }) {
  const entries = Object.entries(data || {});
  const max = Math.max(1, ...entries.map(([, n]) => n));
  if (entries.length === 0) return <p className="text-sm text-stone-400">داده‌ای نیست</p>;
  return (
    <ul className="space-y-2">
      {entries.map(([k, n]) => (
        <li key={k}>
          <div className="mb-1 flex justify-between text-sm">
            <span>{labels?.[k] || STATUS_LABEL[k] || skillLabel(k, labels) || k}</span>
            <b>{n}</b>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-stone-100">
            <div className="h-full rounded-full bg-mahak-500" style={{ width: `${Math.round((n / max) * 100)}%` }} />
          </div>
        </li>
      ))}
    </ul>
  );
}

export default function Reports() {
  const [rows, setRows] = useState<RankingRow[]>([]);
  const [ov, setOv] = useState<ReportOverview | null>(null);
  const [skillNames, setSkillNames] = useState<Record<string, string>>({});

  useEffect(() => {
    api.ranking().then(setRows).catch(() => undefined);
    api.reportOverview().then(setOv).catch(() => undefined);
    api.skillCatalog().then((g) => setSkillNames(catalogLabelMap(g))).catch(() => undefined);
  }, []);

  const cards = ov ? [
    ["داوطلبان", ov.total_volunteers],
    ["فعال", ov.approved_volunteers],
    ["در انتظار هویت", ov.pending_volunteers],
    ["مدارک اصلاح‌شده", ov.resubmitted_documents || 0],
    ["ساعات کل", ov.total_hours],
    ["ساعات این ماه", ov.hours_this_month || 0],
    ["فعالیت باز", ov.open_tasks],
    ["تکمیل این ماه", ov.completed_this_month],
    ["گواهی صادرشده", ov.certificates_issued || 0],
    ["تیکت باز", ov.open_tickets || 0],
  ] : [];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 className="text-2xl font-black">گزارش‌ها</h1>
          <p className="mt-1 text-sm text-stone-500">وضعیت نیرو، حضور، مهارت و رتبه‌بندی داوطلبان در یک نگاه.</p>
        </div>
        <button className="text-sm text-mahak-700" onClick={() => downloadAuth("/api/v1/admin/reports/ranking?format=csv", "mahak-ranking.csv")}>
          خروجی Excel/CSV رتبه‌بندی
        </button>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        {cards.map(([k, v]) => (
          <Card key={String(k)} className="p-4">
            <div className="text-xs text-stone-500">{k}</div>
            <div className="mt-1 text-2xl font-black">{v}</div>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card className="p-5">
          <h2 className="mb-3 font-bold">وضعیت داوطلبان</h2>
          <Bars data={ov?.volunteers_by_status || {}} />
        </Card>
        <Card className="p-5">
          <h2 className="mb-3 font-bold">وضعیت تخصیص‌ها</h2>
          <Bars data={ov?.assignments_by_status || {}} />
        </Card>
        <Card className="p-5">
          <h2 className="mb-3 font-bold">فعالیت‌ها</h2>
          <Bars data={ov?.tasks_by_status || {}} />
          <h3 className="mb-2 mt-4 text-sm font-bold">نوع فعالیت</h3>
          <Bars data={ov?.tasks_by_kind || {}} labels={{ one_off: "موردی", recurring: "جاری", occurrence: "نوبت روزانه" }} />
        </Card>
        <Card className="p-5">
          <h2 className="mb-3 font-bold">توزیع مهارت داوطلبان فعال</h2>
          <Bars data={ov?.skill_distribution || {}} labels={skillNames} />
        </Card>
      </div>

      <Card className="p-5">
        <h2 className="mb-3 font-bold">شهرهای با بیشترین داوطلب فعال</h2>
        {(ov?.top_cities || []).length === 0 && <p className="text-sm text-stone-400">هنوز داده‌ای نیست</p>}
        <ul className="grid gap-2 sm:grid-cols-2">
          {(ov?.top_cities || []).map((c) => (
            <li key={c.city} className="flex justify-between rounded-2xl bg-stone-50 px-3 py-2 text-sm">
              <span>{c.city}</span><b>{c.count}</b>
            </li>
          ))}
        </ul>
      </Card>

      <Card className="overflow-x-auto p-0">
        <div className="flex items-center justify-between px-4 py-3">
          <h2 className="font-bold">رتبه‌بندی داوطلبان</h2>
          <Link className="text-sm text-mahak-700" href="/admin/volunteers">پرونده‌ها</Link>
        </div>
        <table className="w-full text-sm">
          <thead className="bg-stone-50 text-right">
            <tr>
              <th className="p-3">رتبه</th>
              <th>نام</th>
              <th>شهر</th>
              <th>ساعات</th>
              <th>امتیاز</th>
              <th>فعالیت</th>
              <th>وضعیت</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={r.volunteer_id} className="border-t">
                <td className="p-3">{i + 1}</td>
                <td>
                  <Link className="text-mahak-700" href={`/admin/volunteers/${r.volunteer_id}`}>{r.full_name}</Link>
                </td>
                <td>{r.city}</td>
                <td>{r.total_hours}</td>
                <td>{r.average_score.toFixed(2)}</td>
                <td>{r.completed_tasks}</td>
                <td className="p-3"><Badge status={r.status} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
