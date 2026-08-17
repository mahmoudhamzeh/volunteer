"use client";

import { useEffect, useState } from "react";
import { api, Mission, MissionProgress } from "@/lib/api";
import { Badge, Button, Card } from "@/components/ui";

export default function MissionsPage() {
  const [missions, setMissions] = useState<Mission[]>([]);
  const [mine, setMine] = useState<MissionProgress[]>([]);
  const [msg, setMsg] = useState("");

  async function load() {
    setMissions((await api.missions()) || []);
    setMine((await api.myMissions().catch(() => [])) || []);
  }
  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">ماموریت‌های دیجیتال</h1>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {(missions || []).map((m) => {
        const p = mine.find((x) => x.mission_id === m.id);
        return (
          <Card key={m.id} className="p-5">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h2 className="font-bold">{m.title}</h2>
                <p className="text-sm text-stone-600">{m.description}</p>
                <p className="mt-1 text-xs text-stone-500">معادل {m.hour_weight} ساعت · هدف {m.target_count}
                  {m.deadline_hours ? ` · مهلت ${m.deadline_hours} ساعت` : ""}</p>
              </div>
              {p ? <Badge status={p.status} /> : <Badge status={m.status} />}
            </div>
            <div className="mt-3 flex gap-2">
              {!p && <Button onClick={async () => { await api.startMission(m.id); await load(); }}>شروع</Button>}
              {p && p.status === "in_progress" && (
                <Button variant="outline" onClick={async () => {
                  try {
                    await api.missionProgress(m.id);
                    setMsg("پیشرفت ثبت شد");
                    await load();
                  } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
                }}>ثبت پیشرفت ({p.progress}/{m.target_count})</Button>
              )}
            </div>
          </Card>
        );
      })}
    </div>
  );
}
