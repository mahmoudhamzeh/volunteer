"use client";

import { useEffect, useState } from "react";
import { api, Assignment } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Button, Card, inputClass } from "@/components/ui";

export default function WorkPage() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [rating, setRating] = useState<Record<string, number>>({});
  const [msg, setMsg] = useState("");

  async function load() {
    setItems((await api.myAssignments()) || []);
  }
  useEffect(() => { void load(); }, []);

  async function cancel(id: string) {
    try {
      await api.cancelMyAssignment(id);
      setMsg("انصراف ثبت شد");
      await load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">کارهای من</h1>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {items.length === 0 && <Card className="p-6 text-stone-500">هنوز درخواستی ثبت نکرده‌اید.</Card>}
      {items.map((a) => (
        <Card key={a.id} className="p-5">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="font-bold">{a.task?.title}</h2>
              <p className="text-sm text-stone-500">{a.task?.location} · {fmtDate(a.task?.starts_at)}</p>
              {a.composite_score && <p className="text-sm">امتیاز مدیر: {a.composite_score.toFixed(1)}</p>}
            </div>
            <Badge status={a.status} />
          </div>
          {(a.status === "requested" || a.status === "reserved") && (
            <div className="mt-3">
              <Button variant="danger" onClick={() => cancel(a.id)}>انصراف</Button>
            </div>
          )}
          {(a.status === "completed" || a.status === "attended") && !a.volunteer_rating && (
            <div className="mt-3 flex items-center gap-2">
              <input className={inputClass + " w-20"} type="number" min={1} max={5} placeholder="1-5"
                onChange={(e) => setRating({ ...rating, [a.id]: Number(e.target.value) })} />
              <Button variant="outline" onClick={async () => {
                await api.rateAssignment(a.id, rating[a.id] || 5, "");
                await load();
              }}>امتیاز به سازماندهی</Button>
            </div>
          )}
        </Card>
      ))}
    </div>
  );
}
