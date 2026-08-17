"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api, Certificate } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Card } from "@/components/ui";

export default function VerifyPage() {
  const params = useParams<{ code: string }>();
  const [data, setData] = useState<Certificate | null>(null);
  const [err, setErr] = useState("");
  useEffect(() => {
    if (!params.code) return;
    api.verify(params.code).then(setData).catch((e) => setErr(e.message));
  }, [params.code]);
  return (
    <div className="mx-auto max-w-lg px-4 py-16">
      <Card className="p-8">
        <h1 className="text-2xl font-black">استعلام اصالت گواهی محک</h1>
        {err && <p className="mt-4 text-rose-600">این گواهی معتبر نیست یا شناسه نادرست است.</p>}
        {data && (
          <div className="mt-4 space-y-2 text-sm">
            <p>نام داوطلب: <b>{data.volunteer_name || ""}</b></p>
            <p>عنوان: {data.title}</p>
            <p>ساعات همکاری: {data.hours}</p>
            <p className="font-mono text-xs">{data.verification_code}</p>
            <p className="text-stone-500">{fmtDate(data.issued_at)}</p>
            <p className="text-emerald-700">گواهی اصیل و در سامانه ثبت شده است.</p>
            <a className="inline-block text-mahak-700" href={`/api/v1/certificates/${data.verification_code}/pdf`} target="_blank">دانلود PDF</a>
          </div>
        )}
      </Card>
    </div>
  );
}
