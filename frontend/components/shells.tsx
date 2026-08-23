"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ReactNode, useEffect, useState } from "react";
import { api, clearToken, getToken, TokenRole, User, Volunteer } from "@/lib/api";
import { MahakLogo } from "@/components/mahak-logo";

export function VolunteerShell({ children }: { children: ReactNode }) {
  return (
    <Shell
      home="/volunteer"
      links={[
        ["/volunteer", "داشبورد"],
        ["/volunteer/profile", "پروفایل و مدارک"],
        ["/volunteer/tasks", "فعالیت‌ها"],
        ["/volunteer/work", "کارهای من"],
        ["/volunteer/missions", "ماموریت‌ها"],
        ["/volunteer/certificates", "گواهی‌ها"],
        ["/volunteer/tickets", "پشتیبانی"],
      ]}
    >
      {children}
    </Shell>
  );
}

export function AdminShell({ children }: { children: ReactNode }) {
  return (
    <Shell
      home="/admin"
      links={[
        ["/admin", "داشبورد"],
        ["/admin/inbox", "درخواست‌ها"],
        ["/admin/volunteers", "داوطلبان"],
        ["/admin/skills", "مهارت‌ها"],
        ["/admin/tasks", "فعالیت‌ها"],
        ["/admin/assignments", "حضور و امتیاز"],
        ["/admin/tickets", "تیکت‌ها"],
        ["/admin/certificates", "گواهی‌ها"],
        ["/admin/missions", "ماموریت‌ها"],
        ["/admin/reports", "گزارش‌ها"],
      ]}
    >
      {children}
    </Shell>
  );
}

function Shell({
  children,
  links,
  home,
}: {
  children: ReactNode;
  links: [string, string][];
  home: string;
}) {
  const path = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [volunteer, setVolunteer] = useState<Volunteer | null>(null);
  const [unread, setUnread] = useState(0);
  const [menuOpen, setMenuOpen] = useState(false);
  const role: TokenRole = home.startsWith("/admin") ? "admin" : "volunteer";

  useEffect(() => {
    setMenuOpen(false);
  }, [path]);

  useEffect(() => {
    if (!getToken(role)) {
      router.replace(role === "admin" ? "/login?as=admin" : "/login?as=volunteer");
      return;
    }
    api
      .me()
      .then((m) => {
        setUser(m.user);
        setVolunteer(m.volunteer || null);
        if (role === "admin" && m.user.role === "volunteer") router.replace("/volunteer");
        if (role === "volunteer" && m.user.role !== "volunteer") router.replace("/admin");
      })
      .catch(() => {
        clearToken(role);
        router.replace(role === "admin" ? "/login?as=admin" : "/login?as=volunteer");
      });
    if (role === "volunteer") {
      api.notifications().then((n) => setUnread((n || []).filter((x) => !x.read).length)).catch(() => undefined);
    } else {
      api.dashboard().then((d) => {
        setUnread(
          (d.pending_task_requests || 0) +
            (d.pending_deliveries || 0) +
            (d.open_tickets || 0) +
            (d.pending_certificates || 0) +
            (d.pending_skill_proposals || 0) +
            (d.pending_volunteers || 0) +
            (d.resubmitted_documents || 0),
        );
      }).catch(() => undefined);
    }
  }, [home, role, router]);

  function navLink(href: string, label: string) {
    const active = path === href;
    return (
      <Link
        key={href}
        href={href}
        onClick={() => setMenuOpen(false)}
        className={`relative flex items-center justify-between rounded-xl px-3 py-2 text-sm ${active ? "bg-mahak-50 font-bold text-mahak-800" : "text-stone-600 hover:bg-stone-50"}`}
      >
        <span>{label}</span>
        {role === "admin" && href === "/admin/inbox" && unread > 0 && (
          <span className="min-w-[1.1rem] rounded-full bg-rose-600 px-1.5 text-center text-[10px] font-bold text-white">
            {unread > 99 ? "۹۹+" : unread}
          </span>
        )}
      </Link>
    );
  }

  const sidebar = (
    <div className="flex h-full flex-col">
      <Link href={home} className="flex items-center gap-2 px-4 py-4" onClick={() => setMenuOpen(false)}>
        <MahakLogo className="h-9 w-auto" />
        <div>
          <div className="text-sm font-bold text-ink-900">سامانه داوطلبان</div>
          <div className="text-[11px] text-stone-500">حمایت از کودکان مبتلا به سرطان</div>
        </div>
      </Link>
      <nav className="flex-1 space-y-1 overflow-y-auto px-3 pb-4">
        {links.map(([href, label]) => navLink(href, label))}
        {role === "volunteer" && (
          <Link
            href="/volunteer/notifications"
            onClick={() => setMenuOpen(false)}
            className={`relative flex items-center justify-between rounded-xl px-3 py-2 text-sm ${path === "/volunteer/notifications" ? "bg-mahak-50 font-bold text-mahak-800" : "text-stone-600 hover:bg-stone-50"}`}
          >
            <span>اعلان‌ها</span>
            {unread > 0 && (
              <span className="min-w-[1.1rem] rounded-full bg-rose-600 px-1.5 text-center text-[10px] font-bold text-white">
                {unread > 99 ? "۹۹+" : unread}
              </span>
            )}
          </Link>
        )}
      </nav>
      <div className="border-t border-stone-100 px-4 py-3 text-sm">
        <div className="truncate text-stone-600">{volunteer?.full_name || volunteer?.first_name || user?.email}</div>
        <button
          className="mt-2 text-mahak-700"
          onClick={() => {
            clearToken(role);
            router.push(role === "admin" ? "/login?as=admin" : "/login?as=volunteer");
          }}
        >
          خروج
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen md:flex">
      <aside className="sticky top-0 hidden h-screen w-60 shrink-0 border-e border-mahak-100/80 bg-white/90 backdrop-blur md:block">
        {sidebar}
      </aside>
      {menuOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <button type="button" className="absolute inset-0 bg-black/40" aria-label="بستن منو" onClick={() => setMenuOpen(false)} />
          <aside className="absolute start-0 top-0 h-full w-64 bg-white shadow-xl">{sidebar}</aside>
        </div>
      )}
      <div className="min-w-0 flex-1">
        <header className="sticky top-0 z-20 flex items-center justify-between gap-3 border-b border-mahak-100/80 bg-white/85 px-4 py-3 backdrop-blur md:hidden">
          <button
            type="button"
            className="rounded-xl border border-stone-200 px-3 py-1.5 text-sm"
            onClick={() => setMenuOpen(true)}
          >
            منو
          </button>
          <Link href={home} className="flex items-center gap-2">
            <MahakLogo className="h-8 w-auto" />
            <span className="text-sm font-bold">سامانه داوطلبان</span>
          </Link>
          <button
            className="text-sm text-mahak-700"
            onClick={() => {
              clearToken(role);
              router.push(role === "admin" ? "/login?as=admin" : "/login?as=volunteer");
            }}
          >
            خروج
          </button>
        </header>
        <main className="mx-auto max-w-6xl px-4 py-6">{children}</main>
      </div>
    </div>
  );
}
