"use client";

import { useEffect, useMemo, useState } from "react";
import { api, SkillGroup, SkillProposal } from "@/lib/api";
import { PROPOSAL_LABEL } from "@/lib/labels";
import { Button, Card, Field, inputClass } from "@/components/ui";

export default function AdminSkillsPage() {
  const [tab, setTab] = useState<"proposals" | "catalog">("proposals");
  const [catalog, setCatalog] = useState<SkillGroup[]>([]);
  const [proposals, setProposals] = useState<SkillProposal[]>([]);
  const [filter, setFilter] = useState("pending");
  const [q, setQ] = useState("");
  const [msg, setMsg] = useState("");
  const [err, setErr] = useState("");
  const [groupTitle, setGroupTitle] = useState("");
  const [newSkill, setNewSkill] = useState<Record<string, string>>({});
  const [edit, setEdit] = useState<Record<string, { title: string; group_id: string; note: string }>>({});
  const [notes, setNotes] = useState<Record<string, string>>({});

  async function loadCatalog() {
    try {
      return await api.adminSkillCatalog();
    } catch {
      return api.skillCatalog();
    }
  }

  async function load(status = filter) {
    const [c, p] = await Promise.all([
      loadCatalog(),
      api.adminSkillProposals(status).catch(() => [] as SkillProposal[]),
    ]);
    setCatalog(c || []);
    setProposals(p || []);
  }

  useEffect(() => {
    load("pending").catch((e) => {
      const m = e instanceof Error ? e.message : "خطا";
      setErr(m.includes("یافت نشد") || m === "Not Found"
        ? "API مهارت در دسترس نیست. در پوشه backend دستور go run .\\cmd\\api را دوباره اجرا کنید."
        : m);
    });
  }, []);

  async function run(fn: () => Promise<unknown>, ok: string) {
    setErr("");
    try {
      await fn();
      setMsg(ok);
      await load(filter);
    } catch (e) {
      const m = e instanceof Error ? e.message : "خطا";
      setErr(m.includes("یافت نشد") || m === "Not Found"
        ? "API مهارت در دسترس نیست. بک‌اند را با go run .\\cmd\\api از پوشه backend دوباره اجرا کنید."
        : m);
    }
  }

  const filteredProposals = useMemo(() => {
    const needle = q.trim();
    if (!needle) return proposals || [];
    return (proposals || []).filter((p) => `${p.volunteer_name} ${p.title} ${p.group_title}`.includes(needle));
  }, [proposals, q]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-black">مهارت‌ها</h1>
        <p className="mt-1 text-sm text-stone-500">پیشنهادها و کاتالوگ در دو بخش جدا هستند تا صفحه کوتاه بماند.</p>
      </div>
      {err && <p className="text-sm text-rose-600">{err}</p>}
      {msg && !err && <p className="text-sm text-mahak-700">{msg}</p>}

      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setTab("proposals")}
          className={`rounded-full px-4 py-1.5 text-sm ${tab === "proposals" ? "bg-mahak-500 text-white" : "bg-white"}`}
        >
          پیشنهادهای داوطلبان {(proposals || []).length ? `(${(proposals || []).length})` : ""}
        </button>
        <button
          onClick={() => setTab("catalog")}
          className={`rounded-full px-4 py-1.5 text-sm ${tab === "catalog" ? "bg-mahak-500 text-white" : "bg-white"}`}
        >
          کاتالوگ سرگروه و زیرمهارت
        </button>
      </div>

      {tab === "proposals" && (
        <Card className="overflow-hidden">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-stone-100 px-4 py-3">
            <h2 className="font-bold">پیشنهادهای داوطلبان</h2>
            <div className="flex flex-wrap items-center gap-2">
              <input className={inputClass + " w-48"} placeholder="جستجو نام یا مهارت" value={q} onChange={(e) => setQ(e.target.value)} />
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
          </div>
          {filteredProposals.length === 0 && <p className="px-4 py-6 text-sm text-stone-400">موردی نیست</p>}
          {filteredProposals.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[760px] text-sm">
                <thead className="bg-stone-50 text-right text-xs text-stone-500">
                  <tr>
                    <th className="px-3 py-2 font-medium">داوطلب</th>
                    <th className="px-3 py-2 font-medium">مهارت</th>
                    <th className="px-3 py-2 font-medium">گروه</th>
                    <th className="px-3 py-2 font-medium">وضعیت</th>
                    <th className="px-3 py-2 font-medium">اقدام</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredProposals.map((p) => {
                    const draft = edit[p.id] || { title: p.title, group_id: p.group_id, note: notes[p.id] ?? p.admin_note ?? "" };
                    const editingProposal = Boolean(edit[p.id]);
                    const note = notes[p.id] ?? p.admin_note ?? "";
                    return (
                      <tr key={p.id} className="border-t border-stone-100 align-top">
                        <td className="px-3 py-2 whitespace-nowrap">{p.volunteer_name}</td>
                        <td className="px-3 py-2">
                          {editingProposal ? (
                            <input className={inputClass} value={draft.title} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, title: e.target.value } })} />
                          ) : (
                            <span className="font-medium">{p.title}</span>
                          )}
                        </td>
                        <td className="px-3 py-2">
                          {editingProposal ? (
                            <select className={inputClass} value={draft.group_id} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, group_id: e.target.value } })}>
                              {(catalog || []).map((g) => (
                                <option key={g.id} value={g.id}>{g.title}</option>
                              ))}
                            </select>
                          ) : (
                            p.group_title
                          )}
                        </td>
                        <td className="px-3 py-2 whitespace-nowrap">{PROPOSAL_LABEL[p.status] || p.status}</td>
                        <td className="px-3 py-2">
                          {p.status === "pending" && !editingProposal && (
                            <div className="flex flex-wrap items-center gap-2">
                              <input
                                className={inputClass + " min-w-[10rem] max-w-[14rem]"}
                                placeholder="یادداشت یا دلیل رد"
                                value={note}
                                onChange={(e) => setNotes({ ...notes, [p.id]: e.target.value })}
                              />
                              <Button onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "approve", title: p.title, group_id: p.group_id, admin_note: note }), "تایید شد")}>تایید</Button>
                              <Button variant="outline" onClick={() => setEdit({ ...edit, [p.id]: { title: p.title, group_id: p.group_id, note } })}>ویرایش</Button>
                              <Button variant="danger" onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "reject", admin_note: note || "رد شد" }), "رد شد")}>رد</Button>
                            </div>
                          )}
                          {p.status === "pending" && editingProposal && (
                            <div className="space-y-2">
                              <Field label="یادداشت ادمین">
                                <input className={inputClass} value={draft.note} onChange={(e) => setEdit({ ...edit, [p.id]: { ...draft, note: e.target.value } })} />
                              </Field>
                              <div className="flex flex-wrap gap-2">
                                <Button variant="outline" onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "edit", title: draft.title, group_id: draft.group_id, admin_note: draft.note }).then(() => {
                                  const next = { ...edit };
                                  delete next[p.id];
                                  setEdit(next);
                                }), "ویرایش شد")}>ذخیره ویرایش</Button>
                                <Button onClick={() => run(() => api.reviewSkillProposal(p.id, { action: "approve", title: draft.title, group_id: draft.group_id, admin_note: draft.note }).then(() => {
                                  const next = { ...edit };
                                  delete next[p.id];
                                  setEdit(next);
                                }), "تایید شد")}>تایید</Button>
                                <Button variant="ghost" onClick={() => {
                                  const next = { ...edit };
                                  delete next[p.id];
                                  setEdit(next);
                                }}>انصراف</Button>
                              </div>
                            </div>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {tab === "catalog" && (
        <Card className="p-5">
          <h2 className="mb-4 font-bold">کاتالوگ سرگروه‌ها و زیرمهارت‌ها</h2>
          <div className="mb-5 flex flex-wrap gap-2">
            <input className={inputClass + " max-w-sm"} placeholder="عنوان سرگروه مثلا ورزش" value={groupTitle} onChange={(e) => setGroupTitle(e.target.value)} />
            <Button variant="outline" onClick={() => {
              if (!groupTitle.trim()) {
                setErr("عنوان سرگروه را وارد کنید");
                return;
              }
              void run(() => api.createSkillGroup(groupTitle.trim()).then(() => setGroupTitle("")), "سرگروه افزوده شد");
            }}>افزودن سرگروه</Button>
          </div>
          {(catalog || []).length === 0 && <p className="text-sm text-stone-400">هنوز سرگروهی نیست. پس از راه‌اندازی بک‌اند، گروه‌های پیش‌فرض (ورزش، هنر، …) ساخته می‌شوند.</p>}
          <div className="grid gap-3 md:grid-cols-2">
            {(catalog || []).map((g) => (
              <div key={g.id} className="rounded-2xl border border-stone-100 bg-stone-50/70 p-3">
                <div className="flex flex-wrap items-center gap-2">
                  <input
                    className={inputClass + " max-w-[12rem] font-bold"}
                    defaultValue={g.title}
                    onBlur={(e) => {
                      const t = e.target.value.trim();
                      if (t && t !== g.title) void run(() => api.updateSkillGroup(g.id, t), "سرگروه ویرایش شد");
                    }}
                  />
                  <span className="text-xs text-stone-400">{(g.skills || []).length} مهارت</span>
                  <Button variant="danger" onClick={() => {
                    if (!window.confirm(`سرگروه «${g.title}» و همه زیرمهارت‌های آن حذف شود؟`)) return;
                    void run(() => api.deleteSkillGroup(g.id), "سرگروه حذف شد");
                  }}>حذف</Button>
                </div>
                <div className="mt-2 flex flex-wrap gap-1.5">
                  {(g.skills || []).map((s) => (
                    <span key={s.id} className="inline-flex items-center gap-1 rounded-full border border-stone-200 bg-white px-2 py-0.5 text-xs">
                      <input
                        className="w-24 border-0 bg-transparent text-xs outline-none"
                        defaultValue={s.title}
                        onBlur={(e) => {
                          const t = e.target.value.trim();
                          if (t && t !== s.title) void run(() => api.updateCatalogSkill(s.id, { title: t }), "ویرایش شد");
                        }}
                      />
                      {s.status === "inactive" && <span className="text-stone-400">غیرفعال</span>}
                      <button
                        type="button"
                        className="text-rose-600"
                        onClick={() => {
                          if (!window.confirm(`زیرمهارت «${s.title}» حذف شود؟`)) return;
                          void run(() => api.deleteCatalogSkill(s.id), "حذف شد");
                        }}
                      >×</button>
                    </span>
                  ))}
                </div>
                <div className="mt-2 flex gap-2">
                  <input className={inputClass} placeholder="زیرمهارت جدید" value={newSkill[g.id] || ""} onChange={(e) => setNewSkill({ ...newSkill, [g.id]: e.target.value })} />
                  <Button variant="ghost" onClick={() => run(() => api.createCatalogSkill(g.id, newSkill[g.id] || "").then(() => setNewSkill({ ...newSkill, [g.id]: "" })), "مهارت افزوده شد")}>افزودن</Button>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}
