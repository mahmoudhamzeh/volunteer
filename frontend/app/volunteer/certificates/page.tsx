"use client";

import { useEffect, useState } from "react";
import { api, Certificate } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Badge, Card } from "@/components/ui";

export default function CertsPage() {
  const [items, setItems] = useState<Certificate[]>([]);
  useEffect(() => { api.myCerts().then((x) => setItems(x || [])); }, []);
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">گواهی‌های داوطلبی</h1>
      {items.length === 0 && <Card className="p-6 text-stone-500">هنوز گواهی صادر نشده است.</Card>}
      {items.map((c) => (
        <Card key={c.id} className="p-5">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="font-bold">{c.title}</h2>
              <p className="text-sm text-stone-500">{c.hours} ساعت · {fmtDate(c.issued_at)}</p>
              <p className="mt-1 font-mono text-xs">{c.verification_code}</p>
            </div>
            <Badge status={c.kind} />
          </div>
          <a className="mt-3 inline-block text-sm text-mahak-700" href={`/api/v1/certificates/${c.verification_code}/pdf`} target="_blank">دانلود PDF</a>
          {" · "}
          <a className="text-sm text-mahak-700" href={`/verify/${c.verification_code}`} target="_blank">صفحه استعلام</a>
        </Card>
      ))}
    </div>
  );
}
