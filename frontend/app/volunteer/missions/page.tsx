"use client";

import { useEffect, useState } from "react";
import { api, Mission, MissionProgress } from "@/lib/api";
import { Badge, Button, Card } from "@/components/ui";

function canSelfVerify(m: Mission) {
  if (m.can_check) return true;
  return m.verify_mode === "internal" || m.verify_mode === "outbound";
}

export default function MissionsPage() {
  const [missions, setMissions] = useState<Mission[]>([]);
  const [mine, setMine] = useState<MissionProgress[]>([]);
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState("");

  async function load() {
    setMissions((await api.missions()) || []);
    setMine((await api.myMissions().catch(() => [])) || []);
  }
  useEffect(() => {
    load();
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">ماموریت‌های دیجیتال</h1>
      <p className="text-sm text-stone-600">تکمیل ماموریت فقط بعد از تأیید سامانه یا وب‌سرویس مربوطه ثبت می‌شود.</p>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {(missions || []).map((m) => {
        const p = mine.find((x) => x.mission_id === m.id);
        return (
          <Card key={m.id} className="p-5">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h2 className="font-bold">{m.title}</h2>
                <p className="text-sm text-stone-600">{m.description}</p>
                <p className="mt-1 text-xs text-stone-500">
                  معادل {m.hour_weight} ساعت · هدف {m.target_count}
                  {m.deadline_hours ? ` · مهلت ${m.deadline_hours} ساعت` : ""}
                  {p ? ` · پیشرفت ${p.progress}/${m.target_count}` : ""}
                </p>
              </div>
              {p ? <Badge status={p.status} /> : <Badge status="active" />}
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              {!p && (
                <Button
                  disabled={busy === m.id}
                  onClick={async () => {
                    setBusy(m.id);
                    try {
                      await api.startMission(m.id);
                      await load();
                    } catch (e) {
                      setMsg(e instanceof Error ? e.message : "خطا");
                    } finally {
                      setBusy("");
                    }
                  }}
                >
                  شروع
                </Button>
              )}
              {p && p.status === "in_progress" && canSelfVerify(m) && (
                <Button
                  variant="outline"
                  disabled={busy === m.id}
                  onClick={async () => {
                    setBusy(m.id);
                    try {
                      await api.verifyMission(m.id);
                      setMsg("ماموریت تأیید و تکمیل شد");
                    } catch (e) {
                      setMsg(e instanceof Error ? e.message : "خطا");
                    } finally {
                      await load();
                      setBusy("");
                    }
                  }}
                >
                  بررسی تأیید ({p.progress}/{m.target_count})
                </Button>
              )}
            </div>
            {p && p.status === "in_progress" && !canSelfVerify(m) && (
              <p className="mt-3 text-xs text-stone-500">
                این ماموریت را نمی‌توان دستی تمام کرد. وقتی سرویس مربوطه انجام کار را با وب‌هوک تأیید کند، پیشرفت اینجا به‌روز می‌شود.
              </p>
            )}
          </Card>
        );
      })}
    </div>
  );
}
