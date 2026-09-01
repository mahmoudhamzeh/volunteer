"use client";

import { useEffect, useState } from "react";
import { api, Assignment } from "@/lib/api";
import { fmtDate } from "@/lib/labels";
import { Button } from "@/components/ui";
import { ShamsiDateTimeField } from "@/components/shamsi";

function formatSystemNow(d: Date) {
  return d.toLocaleString("fa-IR-u-ca-persian", {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function AttendancePanel({
  assignment,
  onDone,
}: {
  assignment: Assignment;
  onDone: (ok: string) => Promise<void> | void;
}) {
  const a = assignment;
  const [now, setNow] = useState(() => new Date());
  const [checkIn, setCheckIn] = useState(a.check_in_at || new Date().toISOString());
  const [checkOut, setCheckOut] = useState(a.check_out_at || "");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const t = window.setInterval(() => setNow(new Date()), 1000);
    return () => window.clearInterval(t);
  }, []);

  useEffect(() => {
    setCheckIn(a.check_in_at || new Date().toISOString());
    setCheckOut(a.check_out_at || "");
  }, [a.id, a.check_in_at, a.check_out_at]);

  const attended = a.status === "attended";

  async function save() {
    setBusy(true);
    try {
      await api.attendance(a.id, {
        check_in_at: checkIn || undefined,
        check_out_at: checkOut || undefined,
      });
      await onDone(attended ? "ساعت ورود و خروج به‌روز شد" : "حضور ثبت شد");
    } catch (e) {
      await onDone(e instanceof Error ? e.message : "خطا");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-3 rounded-2xl border border-mahak-100 bg-mahak-50/40 p-3">
      <div>
        <div className="text-xs font-bold text-mahak-800">تاریخ و ساعت سیستم</div>
        <p className="mt-1 text-sm font-medium text-ink-800">{formatSystemNow(now)}</p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <ShamsiDateTimeField label="ساعت ورود" value={checkIn} onChange={setCheckIn} />
          <button type="button" className="mt-1 text-xs text-mahak-700" onClick={() => setCheckIn(new Date().toISOString())}>
            ثبت ساعت فعلی ورود
          </button>
        </div>
        <div>
          <ShamsiDateTimeField label="ساعت خروج" value={checkOut} onChange={setCheckOut} />
          <button type="button" className="mt-1 text-xs text-mahak-700" onClick={() => setCheckOut(new Date().toISOString())}>
            ثبت ساعت فعلی خروج
          </button>
        </div>
      </div>
      {(a.check_in_at || a.check_out_at) && (
        <p className="text-xs text-stone-500">
          آخرین ثبت: ورود {fmtDate(a.check_in_at)} · خروج {fmtDate(a.check_out_at)}
        </p>
      )}
      <Button disabled={busy} onClick={() => void save()}>
        {attended ? "به‌روزرسانی ورود و خروج" : "ثبت حضور"}
      </Button>
    </div>
  );
}
