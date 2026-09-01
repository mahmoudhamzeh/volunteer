"use client";

import { useState } from "react";
import { Certificate } from "@/lib/api";
import { certKindLabel, fmtDate } from "@/lib/labels";
import { Button, Modal } from "@/components/ui";
import { ShareReadyActions, publicCertUrl } from "@/components/share-ready";

function loadImage(src: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("بارگذاری تصویر تقدیرنامه ممکن نشد"));
    img.src = src;
  });
}

function wrapText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number) {
  const words = text.split(" ");
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    const test = line ? `${line} ${word}` : word;
    if (ctx.measureText(test).width > maxWidth && line) {
      lines.push(line);
      line = word;
    } else {
      line = test;
    }
  }
  if (line) lines.push(line);
  return lines;
}

async function renderAppreciation(c: Certificate) {
  await document.fonts.ready;
  const img = await loadImage("/appreciation-frame.png");
  const canvas = document.createElement("canvas");
  canvas.width = img.naturalWidth;
  canvas.height = img.naturalHeight;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("مرورگر از ذخیره تصویر پشتیبانی نمی‌کند");
  ctx.drawImage(img, 0, 0);
  ctx.textAlign = "center";
  ctx.direction = "rtl";
  const cx = canvas.width / 2;
  const max = canvas.width * 0.62;
  ctx.fillStyle = "#880E4F";
  ctx.font = `700 ${Math.round(canvas.width * 0.038)}px Vazirmatn, Tahoma, sans-serif`;
  ctx.fillText("تقدیرنامه داوطلبی محک", cx, canvas.height * 0.24);
  ctx.fillStyle = "#6b7280";
  ctx.font = `500 ${Math.round(canvas.width * 0.018)}px Vazirmatn, Tahoma, sans-serif`;
  ctx.fillText("موسسه خیریه حمایت از کودکان مبتلا به سرطان", cx, canvas.height * 0.3);
  ctx.fillStyle = "#141B22";
  ctx.font = `700 ${Math.round(canvas.width * 0.032)}px Vazirmatn, Tahoma, sans-serif`;
  ctx.fillText(c.volunteer_name || "داوطلب محک", cx, canvas.height * 0.42);
  ctx.fillStyle = "#374151";
  ctx.font = `500 ${Math.round(canvas.width * 0.02)}px Vazirmatn, Tahoma, sans-serif`;
  const body = `بدین‌وسیله از همراهی داوطلبانه در «${c.title}» به مدت ${c.hours} ساعت قدردانی می‌شود.`;
  const lines = wrapText(ctx, body, max);
  lines.forEach((line, i) => ctx.fillText(line, cx, canvas.height * 0.5 + i * canvas.height * 0.04));
  ctx.fillStyle = "#6b7280";
  ctx.font = `500 ${Math.round(canvas.width * 0.016)}px Vazirmatn, Tahoma, sans-serif`;
  ctx.fillText(`${certKindLabel(c.kind)}  ·  ${fmtDate(c.issued_at)}`, cx, canvas.height * 0.72);
  ctx.font = `400 ${Math.round(canvas.width * 0.013)}px Vazirmatn, Tahoma, sans-serif`;
  ctx.fillText(c.verification_code, cx, canvas.height * 0.78);
  return canvas;
}

function CertificateArt({ cert, compact }: { cert: Certificate; compact?: boolean }) {
  return (
    <div className={`relative overflow-hidden bg-[#f7efe6] ${compact ? "h-full w-full" : "aspect-[4/3]"}`}>
      <img src="/appreciation-frame.png" alt="" className="absolute inset-0 h-full w-full object-cover" />
      <div className={`absolute inset-[11%] flex flex-col items-center justify-center text-center ${compact ? "px-1" : "px-4"}`}>
        <p className={`font-black text-mahak-800 ${compact ? "text-[7px] leading-tight" : "text-lg md:text-2xl"}`}>تقدیرنامه داوطلبی محک</p>
        {!compact && (
          <p className="mt-1 text-[11px] text-stone-500 md:text-xs">موسسه خیریه حمایت از کودکان مبتلا به سرطان</p>
        )}
        <p className={`font-black text-ink-900 ${compact ? "mt-1 text-[8px] leading-tight" : "mt-5 text-base md:text-xl"}`}>
          {cert.volunteer_name || "داوطلب محک"}
        </p>
        {!compact && (
          <>
            <p className="mt-3 max-w-md text-sm leading-7 text-stone-700">
              بدین‌وسیله از همراهی داوطلبانه در «{cert.title}» به مدت {cert.hours} ساعت قدردانی می‌شود.
            </p>
            <p className="mt-4 text-xs text-stone-500">{certKindLabel(cert.kind)} · {fmtDate(cert.issued_at)}</p>
          </>
        )}
      </div>
    </div>
  );
}

export function AppreciationCard({
  cert,
  embedded = false,
}: {
  cert: Certificate;
  embedded?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState("");

  async function downloadPng() {
    setBusy("png");
    setMsg("");
    try {
      const canvas = await renderAppreciation(cert);
      await new Promise<void>((resolve, reject) => {
        canvas.toBlob((blob) => {
          if (!blob) {
            reject(new Error("ساخت تصویر ممکن نشد"));
            return;
          }
          const url = URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = url;
          a.download = `tashakor-mahak-${cert.verification_code.slice(0, 8)}.png`;
          a.click();
          URL.revokeObjectURL(url);
          resolve();
        }, "image/png");
      });
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "دانلود ممکن نشد");
    } finally {
      setBusy("");
    }
  }

  const share = (
    <div className="mt-4 space-y-3">
      <ShareReadyActions url={publicCertUrl(cert.verification_code)} kind={cert.kind} />
      {msg && <p className="text-center text-sm text-mahak-700">{msg}</p>}
      <div className="flex flex-wrap justify-center gap-3 text-xs">
        <button type="button" className="text-mahak-700" disabled={busy !== ""} onClick={() => void downloadPng()}>
          {busy === "png" ? "در حال ساخت تصویر…" : "دانلود تصویر"}
        </button>
        <a className="text-mahak-700" href={`/api/v1/certificates/${cert.verification_code}/pdf`} target="_blank" rel="noreferrer">
          دانلود PDF
        </a>
        {!embedded && (
          <a className="text-mahak-700" href={`/verify/${cert.verification_code}`} target="_blank" rel="noreferrer">
            صفحه استعلام
          </a>
        )}
      </div>
    </div>
  );

  return (
    <>
      {embedded ? (
        <div>
          <CertificateArt cert={cert} />
          {share}
        </div>
      ) : (
        <button
          type="button"
          onClick={() => { setOpen(true); setMsg(""); }}
          className="flex w-full items-center gap-3 rounded-2xl border border-stone-100 bg-white p-2 text-right hover:border-mahak-200 hover:bg-mahak-50/40"
        >
          <div className="h-14 w-[5.5rem] shrink-0 overflow-hidden rounded-xl border border-stone-100">
            <CertificateArt cert={cert} compact />
          </div>
          <div className="min-w-0 flex-1">
            <div className="truncate font-bold text-ink-900">{cert.title}</div>
            <div className="mt-0.5 text-xs text-stone-500">
              {certKindLabel(cert.kind)} · {cert.hours} ساعت · {fmtDate(cert.issued_at)}
            </div>
          </div>
          <span className="shrink-0 text-xs font-bold text-mahak-700">مشاهده</span>
        </button>
      )}

      <Modal open={open} title="تقدیرنامه داوطلبی محک" onClose={() => setOpen(false)} size="lg">
        <CertificateArt cert={cert} />
        {share}
        <div className="mt-3 flex justify-end">
          <Button variant="ghost" onClick={() => setOpen(false)}>بستن</Button>
        </div>
      </Modal>
    </>
  );
}

