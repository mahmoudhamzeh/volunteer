"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, Volunteer } from "@/lib/api";
import { Badge, Card, inputClass } from "@/components/ui";
import { STATUS_LABEL } from "@/lib/labels";

export default function VolunteersAdmin() {
  const [status, setStatus] = useState("pending");
  const [attention, setAttention] = useState("");
  const [q, setQ] = useState("");
  const [items, setItems] = useState<Volunteer[]>([]);

  async function load(st = status, query = q, att = attention) {
    const qs = new URLSearchParams();
    if (att) qs.set("attention", att);
    else if (st) qs.set("status", st);
    if (query) qs.set("q", query);
    const r = await api.adminVolunteers(`?${qs.toString()}`);
    setItems(r.items || []);
  }
  useEffect(() => {
    const sp = new URLSearchParams(window.location.search);
    const att = sp.get("attention") || "";
    const st = sp.get("status");
    if (att) {
      setAttention(att);
      setStatus("");
      void load("", "", att);
      return;
    }
    if (st !== null) setStatus(st);
    void load(st !== null ? st : "pending", "", "");
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-black">تایید هویت داوطلبان</h1>
      <div className="flex flex-wrap gap-2">
        {["", "pending", "draft", "approved", "rejected", "suspended"].map((s) => (
          <button key={s || "all"} onClick={() => { setAttention(""); setStatus(s); load(s, q, ""); }}
            className={`rounded-full px-3 py-1 text-sm ${!attention && status === s ? "bg-mahak-500 text-white" : "bg-white"}`}>
            {s ? (STATUS_LABEL[s] || s) : "همه"}
          </button>
        ))}
        <button
          onClick={() => { setAttention("resubmitted"); setStatus(""); load("", q, "resubmitted"); }}
          className={`rounded-full px-3 py-1 text-sm ${attention === "resubmitted" ? "bg-amber-600 text-white" : "bg-white"}`}
        >
          مدارک اصلاح‌شده
        </button>
        <input className={inputClass + " max-w-xs"} placeholder="جستجو" value={q} onChange={(e) => setQ(e.target.value)} onKeyDown={(e) => e.key === "Enter" && load(status, q, attention)} />
      </div>
      {items.map((v) => (
        <Link key={v.id} href={`/admin/volunteers/${v.id}`}>
          <Card className="mb-3 p-4 hover:border-mahak-200">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-bold">{v.full_name}</div>
                <div className="text-xs text-stone-500">{v.email ? `${v.email} · ` : ""}{v.city} · {v.phone} · {v.education_field}</div>
              </div>
              <Badge status={v.status} />
            </div>
          </Card>
        </Link>
      ))}
    </div>
  );
}
