"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api, Certificate } from "@/lib/api";
import { certKindLabel, fmtDate } from "@/lib/labels";
import { Card } from "@/components/ui";
import { AppreciationCard } from "@/components/appreciation-card";

export default function VerifyPage() {
  const params = useParams<{ code: string }>();
  const [data, setData] = useState<Certificate | null>(null);
  const [err, setErr] = useState("");
  useEffect(() => {
    if (!params.code) return;
    api.verify(params.code).then(setData).catch((e) => setErr(e.message));
  }, [params.code]);
  return (
    <div className="mx-auto max-w-3xl px-4 py-16">
      <Card className="p-8">
        <h1 className="text-2xl font-black">استعلام اصالت مدرک محک</h1>
        {err && <p className="mt-4 text-rose-600">این مدرک معتبر نیست یا شناسه نادرست است.</p>}
        {data && data.kind !== "official" && (
          <div className="mt-4 space-y-3">
            <p className="text-sm text-emerald-700">مدرک اصیل و در سامانه ثبت شده است.</p>
            <AppreciationCard cert={data} embedded />
          </div>
        )}
        {data && data.kind === "official" && (
          <div className="mt-4 space-y-2 text-sm">
            <p>نام داوطلب: <b>{data.volunteer_name || ""}</b></p>
            <p>عنوان: {data.title}</p>
            <p className="text-xs text-stone-500">{certKindLabel(data.kind)}</p>
            <p>ساعات همکاری: {data.hours}</p>
            <p className="font-mono text-xs">{data.verification_code}</p>
            <p className="text-stone-500">{fmtDate(data.issued_at)}</p>
            <p className="text-emerald-700">مدرک اصیل و در سامانه ثبت شده است.</p>
            <p className="text-stone-600">این گواهی‌نامه به‌صورت ارسال یا تحویل حضوری ارائه می‌شود و از پنل دانلود نمی‌شود.</p>
          </div>
        )}
      </Card>
    </div>
  );
}
