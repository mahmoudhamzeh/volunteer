"use client";

import { useState } from "react";
import { Button } from "@/components/ui";

export function publicCertUrl(code: string) {
  if (typeof window === "undefined") return `/verify/${code}`;
  return `${window.location.origin}/verify/${code}`;
}

export function certShareCopy(kind?: string) {
  if (kind === "official") {
    return {
      title: "گواهی‌نامه فعالیت داوطلبانه محک",
      text: "یک گواهی‌نامه فعالیت داوطلبانه محک به شما تقدیم شده است:",
      caption: "این گواهی‌نامه آماده اشتراک‌گذاری است.",
    };
  }
  return {
    title: "تقدیرنامه داوطلبی محک",
    text: "یک تقدیرنامه داوطلبی محک به شما تقدیم شده است:",
    caption: "مهربانی شما به محک رسید؛ پیام شما آماده اشتراک‌گذاری است.",
  };
}

function ShareIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
      <circle cx="18" cy="5" r="2.5" />
      <circle cx="6" cy="12" r="2.5" />
      <circle cx="18" cy="19" r="2.5" />
      <path d="M8.7 13.4l6.6 3.8M15.3 6.8l-6.6 3.8" />
    </svg>
  );
}

function LinkIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
      <path d="M10 13a5 5 0 0 0 7.07 0l1.41-1.41a5 5 0 0 0-7.07-7.07L10 5.93" />
      <path d="M14 11a5 5 0 0 0-7.07 0L5.52 12.4a5 5 0 0 0 7.07 7.07L14 18.07" />
    </svg>
  );
}

export function ShareReadyActions({
  url,
  kind,
}: {
  url: string;
  kind?: string;
}) {
  const copy = certShareCopy(kind);
  const [msg, setMsg] = useState("");

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(url);
      setMsg("لینک صفحه کپی شد");
    } catch {
      setMsg(url);
    }
  }

  async function share() {
    setMsg("");
    if (typeof navigator !== "undefined" && typeof navigator.share === "function") {
      try {
        await navigator.share({ title: copy.title, text: copy.text, url });
        return;
      } catch (e) {
        if (e instanceof DOMException && e.name === "AbortError") return;
      }
    }
    await copyLink();
    setMsg("لینک کپی شد؛ می‌توانید در پیام‌رسان بفرستید");
  }

  return (
    <div className="space-y-3">
      <p className="text-center text-sm leading-7 text-stone-600">{copy.caption}</p>
      <div className="grid grid-cols-2 gap-2 [&>button]:w-full">
        <Button onClick={() => void share()}>
          <span className="inline-flex w-full items-center justify-center gap-2">
            <ShareIcon />
            اشتراک‌گذاری
          </span>
        </Button>
        <Button variant="outline" onClick={() => void copyLink()}>
          <span className="inline-flex w-full items-center justify-center gap-2">
            <LinkIcon />
            کپی لینک صفحه
          </span>
        </Button>
      </div>
      {msg && <p className="text-center text-sm text-mahak-700">{msg}</p>}
    </div>
  );
}
