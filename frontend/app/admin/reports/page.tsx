"use client";

import { useEffect, useState } from "react";
import { api, downloadAuth, RankingRow } from "@/lib/api";
import { Badge, Card } from "@/components/ui";
import { skillLabel } from "@/lib/labels";

export default function Reports() {
  const [rows, setRows] = useState<RankingRow[]>([]);
  const [skills, setSkills] = useState<Record<string, number>>({});
  useEffect(() => {
    api.ranking().then(setRows);
    api.skills().then(setSkills);
  }, []);
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-black">گزارش‌ها و رتبه‌بندی</h1>
        <button className="text-sm text-mahak-700" onClick={() => downloadAuth("/api/v1/admin/reports/ranking?format=csv", "mahak-ranking.csv")}>
          خروجی Excel/CSV
        </button>
      </div>
      <Card className="overflow-x-auto p-0">
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
                <td>{r.full_name}</td>
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
      <Card className="p-5">
        <h2 className="font-bold">توزیع تخصص</h2>
        <ul className="mt-3 space-y-1 text-sm">
          {Object.entries(skills).map(([k, n]) => (
            <li key={k} className="flex justify-between"><span>{skillLabel(k)}</span><b>{n}</b></li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
