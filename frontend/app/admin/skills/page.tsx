"use client";

import { useEffect, useState } from "react";
import { api, SkillGroup, SkillProposal } from "@/lib/api";
import { PROPOSAL_LABEL } from "@/lib/labels";
import { Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminSkillsPage() {
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [proposals, setProposals] = useState<SkillProposal[]>([]);
  const [filter, setFilter] = useState("pending");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [groupTitle, setGroupTitle] = useState("");
  const [groupSlug, setGroupSlug] = useState("");
  const [newSkill, setNewSkill] = useState<Record<string, string>>({});
  const [edit, setEdit] = useState<Record<string, { title: string; group_id: string; note: string }>>({});

  async function load(status = filter) {
    const [c, p] = await Promise.all([
      api.adminSkillCatalog(),
      api.adminSkillProposals(status),
    ]);
    setCatalog(c || []);
    setProposals(p || []);
  }

  useEffect(() => {
    load("pending").catch((e) => setErr(e instanceof Error ? e.message : "خطا"));
  }, []);

  async function run(fn: () => Promise<unknown>, ok: string) {
    setErr("");
    try {
      await fn();
      setMsg(ok);
      await load(filter);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "خطا");
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-black">مهارت‌ها</h1>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <Card className="p-5">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-bold">پیشنهادهای داوطلبان</h2>
          <select
            className={inputClass + " w-auto"}
            value={filter}
            onChange={(e) => {
              const v = e.target.value;
              setFilter(v);
              load(v).catch((x) => setErr(x instanceof Error ? x.message : "خطا"));
            }}
          >
            <option value="pending">در انتظار تایید</option>
            <option value="approved">تایید شده</option>
            <option value="rejected">رد شده</option>
            <option value="">همه</option>
          </select>
        </div>
        <div className="space-y-4">
          {(proposals || []).length === 0 && <p className="text-sm text-stone-400">موردی نیست</p>}
          {(proposals || []).map((p) => {
            const draft = edit[p.id] || { title: p.title, group_id: p.group_id, note: p.admin_note || "" };
            return (
              <div key={p.id} className="rounded-2xl border border-stone-100 p-4">
                <div className="text-sm text-stone-500">{p.volunteer_name} · {p.group_title}</div>
                <div className="mt-2 grid gap-2 md:grid-cols-3">
                  <Field label="عنوان مهارت">
                    <input className={inputClass} value={draft.title} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, title: e.target.value } })} />
                  </Field>
                  <Field label="گروه">
                    <select className={inputClass} value={draft.group_id} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, group_id: e.target.value } })}>
                      {(catalog || []).map((g) => (
                        <option key={g.id} value={g.id}>{g.title}</option>
                      ))}
                    </select>
                  </Field>
                  <Field label="یادداشت ادمین">
                    <input className={inputClass} value={draft.note} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, note: e.target.value } })} />
                  </Field>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <span className="rounded-full border border-stone-200 px-2 py-0.5 text-xs">{PROPOSAL_LABEL[p.status] || p.status}</span>
                  {p.status === "pending" && (
                    <>
                      <Button onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "approve", title: draft.title, group_id: draft.group_id, admin_note: draft.note }), "تایید شد")}>تایید</Button>
                      <Button variant="outline" onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "edit_approve", title: draft.title, group_id: draft.group_id, admin_note: draft.note }), "ویرایش و تایید شد")}>ویرایش و تایید</Button>
                      <Button variant="danger" onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "reject", admin_note: draft.note || "رد شد" }), "رد شد")}>رد</Button>
                    </>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="mb-4 font-bold">کاتالوگ گروه‌ها و زیرمهارت‌ها</h2>
        <div className="mb-5 grid gap-2 md:grid-cols-3">
          <input className={inputClass} placeholder="شناسه لاتین مثلا sports" value={groupSlug} onChange={(e) => setGroupSlug(e.target.value)} />
          <input className={inputClass} placeholder="عنوان گروه مثلا ورزش" value={groupTitle} onChange={(e) => setGroupTitle(e.target.value)} />
          <Button variant="outline" onClick={() => run(() => api.createSkillGroup(groupSlug, groupTitle).then(() => { setGroupSlug(""); setGroupTitle(""); }), "گروه افزوده شد")}>افزودن گروه</Button>
        </div>
        <div className="space-y-4">
          {(catalog || []).map((g) => (
            <div key={g.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-4">
              <div className="font-bold text-mahak-700">{g.title}</div>
              <div className="mt-2 flex flex-wrap gap-2">
                {(g.skills || []).map((s) => (
                  <span key={s.id} className="inline-flex items-center gap-2 rounded-full border border-stone-200 bg-white px-3 py-1 text-sm">
                    <input
                      className="w-28 border-0 bg-transparent text-sm outline-none"
                      defaultValue={s.title}
                      onBlur={(e) => {
                        const t = e.target.value.trim();
                        if (t && t !== s.title) void run(() => api.updateCatalogSkill(s.id, { title: t }), "ویرایش شد");
                      }}
                    />
                    {s.status === "inactive" && <span className="text-xs text-stone-400">غیرفعال</span>}
                  </span>
                ))}
              </div>
              <div className="mt-3 flex gap-2">
                <input className={inputClass} placeholder="زیرمهارت جدید" value={newSkill[g.id] || ""} onChange={(e) => setNewSkill({ ...newSkill, [g.id]: e.target.value })} />
                <Button variant="ghost" onClick={() => run(() => api.createCatalogSkill(g.id, newSkill[g.id] || "").then(() => setNewSkill({ ...newSkill, [g.id]: "" })), "مهارت افزوده شد")}>افزودن</Button>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
