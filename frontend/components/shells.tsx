"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ReactNode, useEffect, useState } from "react";
import { api, clearToken, getToken, User, Volunteer } from "@/lib/api";

export function VolunteerShell({ children }: { children: ReactNode }) {
  return (
    <Shell
      home="/volunteer"
      links={[
        ["/volunteer", "داشبورد"],
        ["/volunteer/profile", "پروفایل و مدارک"],
        ["/volunteer/tasks", "تسک‌ها"],
        ["/volunteer/work", "کارهای من"],
        ["/volunteer/missions", "ماموریت‌ها"],
        ["/volunteer/certificates", "گواهی‌ها"],
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
        ["/admin/volunteers", "داوطلبان"],
        ["/admin/tasks", "تسک‌ها"],
        ["/admin/assignments", "حضور و امتیاز"],
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

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    api
      .me()
      .then((m) => {
        setUser(m.user);
        setVolunteer(m.volunteer || null);
        if (home.startsWith("/admin") && m.user.role === "volunteer") router.replace("/volunteer");
        if (home.startsWith("/volunteer") && m.user.role !== "volunteer") router.replace("/admin");
      })
      .catch(() => {
        clearToken();
        router.replace("/login");
      });
  }, [home, router]);

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-20 border-b border-mahak-100/80 bg-white/85 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
          <Link href={home} className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-xl bg-mahak-500 text-sm font-black text-white">محک</span>
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
                className={`rounded-lg px-3 py-1.5 text-sm ${path === href ? "bg-mahak-50 text-mahak-700" : "text-stone-600 hover:bg-stone-50"}`}
              >
                {label}
              </Link>
            ))}
          </nav>
          <div className="flex items-center gap-3 text-sm">
            <span className="hidden text-stone-600 sm:inline">{volunteer?.full_name || user?.email}</span>
            <button
              className="text-mahak-700"
              onClick={() => {
                clearToken();
                router.push("/login");
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
              className={`whitespace-nowrap rounded-lg px-3 py-1 text-xs ${path === href ? "bg-mahak-50 text-mahak-700" : "text-stone-600"}`}
            >
              {label}
            </Link>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6">{children}</main>
    </div>
  );
}
