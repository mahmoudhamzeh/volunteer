"use client";

import { ReactNode } from "react";
import { statusClass, STATUS_LABEL } from "@/lib/labels";

export function Badge({ status }: { status: string }) {
  return (
    <span className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${statusClass(status)}`}>
      {STATUS_LABEL[status] || status}
    </span>
  );
}

export function Card({ children, className = "" }: { children: ReactNode; className?: string }) {
  return <div className={`rounded-3xl border border-white/70 bg-white/90 shadow-card ${className}`}>{children}</div>;
}

export function Button({
  children,
  onClick,
  type = "button",
  variant = "primary",
  disabled,
}: {
  children: ReactNode;
  onClick?: () => void;
  type?: "button" | "submit";
  variant?: "primary" | "ghost" | "danger" | "outline";
  disabled?: boolean;
}) {
  const styles = {
    primary: "bg-mahak-500 text-white hover:bg-mahak-600",
    ghost: "bg-stone-100 text-ink-800 hover:bg-stone-200",
    danger: "bg-rose-600 text-white hover:bg-rose-700",
    outline: "border border-mahak-200 text-mahak-700 hover:bg-mahak-50",
  }[variant];
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      className={`rounded-2xl px-4 py-2.5 text-sm font-medium transition disabled:opacity-50 ${styles}`}
    >
      {children}
    </button>
  );
}

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block space-y-1.5">
      <span className="text-sm text-stone-600">{label}</span>
      {children}
    </label>
  );
}

export const inputClass =
  "w-full rounded-2xl border border-stone-200 bg-white px-3 py-2.5 text-sm outline-none ring-mahak-400 focus:ring-2";

export function Modal({
  open,
  title,
  children,
  onClose,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  onClose?: () => void;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div
        className="w-full max-w-md rounded-3xl bg-white p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <h2 id="modal-title" className="text-lg font-black text-ink-900">{title}</h2>
        <div className="mt-3">{children}</div>
      </div>
    </div>
  );
}
