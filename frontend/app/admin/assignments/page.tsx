"use client";

import { useEffect, useState } from "react";
import { api, Assignment } from "@/lib/api";
import { Badge, Button, Card, inputClass } from "@/components/ui";

export default function AssignmentsAdmin() {
  const [items, setItems] = useState<Assignment[]>([]);
  const [scores, setScores] = useState<Record<string, { d: number; e: number; t: number; c: string }>>({});
  const [msg, setMsg] = useState("");

  async function load() {
    const r = await api.adminAssignments("?limit=100");
    setItems(r.items || []);
  }
  useEffect(() => { load(); }, []);

  function sc(id: string) {
    return scores[id] || { d: 5, e: 5, t: 5, c: "" };
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">حضور، امتیاز و گواهی</h1>
      {msg && <p className="text-sm text-mahak-700">{msg}</p>}
      {items.map((a) => (
        <Card key={a.id} className="p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div className="font-bold">{a.task?.title}</div>
              <div className="text-sm text-stone-500">{a.volunteer?.full_name}</div>
            </div>
            <Badge status={a.status} />
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            {a.status === "reserved" && <Button onClick={async () => { await api.attendance(a.id); await load(); }}>تایید حضور</Button>}
            {a.status === "reserved" && (
              <Button variant="danger" onClick={async () => { await api.adminCancelAssignment(a.id); await load(); }}>لغو رزرو</Button>
            )}
            {(a.status === "attended" || a.status === "reserved") && (
              <>
                <input className={inputClass + " w-16"} type="number" min={1} max={5} value={sc(a.id).d} title="انضباط"
                  onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), d: Number(e.target.value) } })} />
                <input className={inputClass + " w-16"} type="number" min={1} max={5} value={sc(a.id).e} title="تخصص"
                  onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), e: Number(e.target.value) } })} />
                <input className={inputClass + " w-16"} type="number" min={1} max={5} value={sc(a.id).t} title="اخلاق"
                  onChange={(e) => setScores({ ...scores, [a.id]: { ...sc(a.id), t: Number(e.target.value) } })} />
                <Button variant="outline" onClick={async () => {
                  const s = sc(a.id);
                  await api.complete(a.id, { discipline: s.d, expertise: s.e, ethics: s.t, comment: s.c });
                  await load();
                }}>ثبت امتیاز و تکمیل</Button>
              </>
            )}
            {a.status === "completed" && (
              <Button onClick={async () => {
                try {
                  await api.issueCert(a.id);
                  setMsg("گواهی موردی صادر شد");
                } catch (e) { setMsg(e instanceof Error ? e.message : "خطا"); }
              }}>صدور گواهی</Button>
            )}
          </div>
        </Card>
      ))}
    </div>
  );
}
