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
  const role: TokenRole = home.startsWith("/admin") ? "admin" : "volunteer";

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
        setUnread((d.pending_task_requests || 0) + (d.open_tickets || 0) + (d.pending_certificates || 0) + (d.pending_skill_proposals || 0) + (d.pending_volunteers || 0) + (d.resubmitted_documents || 0));
      }).catch(() => undefined);
    }
  }, [home, role, router]);

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-20 border-b border-mahak-100/80 bg-white/85 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
          <Link href={home} className="flex items-center gap-2">
            <MahakLogo className="h-9 w-auto" />
            <div>
              <div className="text-sm font-bold text-ink-900">سامانه داوطلبان</div>
              <div className="text-[11px] text-stone-500">حمایت از کودکان مبتلا به سرطان</div>
            </div>
          </Link>
          <nav className="hidden items-center gap-1 md:flex">
            {links.map(([href, label]) => (
              <Link
                key={href}
                href={href}
                className={`relative rounded-lg px-3 py-1.5 text-sm ${path === href ? "bg-mahak-50 text-mahak-700" : "text-stone-600 hover:bg-stone-50"}`}
              >
                {label}
                {role === "admin" && href === "/admin/inbox" && unread > 0 && (
                  <span className="absolute -top-1 -end-1 min-w-[1.1rem] rounded-full bg-rose-600 px-1 text-center text-[10px] font-bold text-white">
                    {unread > 99 ? "۹۹+" : unread}
                  </span>
                )}
              </Link>
            ))}
          </nav>
          <div className="flex items-center gap-3 text-sm">
            {role === "volunteer" && (
              <Link href="/volunteer/notifications" className="relative text-stone-600">
                اعلان‌ها
                {unread > 0 && (
                  <span className="absolute -top-2 -end-3 min-w-[1.1rem] rounded-full bg-rose-600 px-1 text-center text-[10px] font-bold text-white">
                    {unread > 99 ? "۹۹+" : unread}
                  </span>
                )}
              </Link>
            )}
            <span className="hidden text-stone-600 sm:inline">{volunteer?.full_name || volunteer?.first_name || user?.email}</span>
            <button
              className="text-mahak-700"
              onClick={() => {
                clearToken(role);
                router.push(role === "admin" ? "/login?as=admin" : "/login?as=volunteer");
              }}
            >
              خروج
            </button>
          </div>
        </div>
        <nav className="flex gap-1 overflow-x-auto px-3 pb-2 md:hidden">
          {links.map(([href, label]) => (
            <Link
              key={href}
              href={href}
              className={`relative whitespace-nowrap rounded-lg px-3 py-1 text-xs ${path === href ? "bg-mahak-50 text-mahak-700" : "text-stone-600"}`}
            >
              {label}
              {role === "admin" && href === "/admin/inbox" && unread > 0 && (
                <span className="ms-1 rounded-full bg-rose-600 px-1 text-[10px] font-bold text-white">{unread > 99 ? "۹۹+" : unread}</span>
              )}
            </Link>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6">{children}</main>
    </div>
  );
}
