"use client";

export function TabBar({
  items,
  active,
  onChange,
  numbered,
}: {
  items: { id: string; label: string }[];
  active: string;
  onChange: (id: string) => void;
  numbered?: boolean;
}) {
  return (
    <nav className="overflow-x-auto rounded-2xl border border-stone-100 bg-white p-1.5 shadow-card">
      <div className="flex min-w-max gap-1">
        {items.map((item, i) => {
          const on = active === item.id;
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onChange(item.id)}
              className={`flex min-w-fit items-center gap-2 rounded-xl px-3.5 py-2.5 text-sm transition ${
                on ? "bg-mahak-500 font-bold text-white shadow-sm" : "text-stone-600 hover:bg-stone-50"
              }`}
            >
              {numbered && (
                <span className={`grid h-6 w-6 place-items-center rounded-full text-xs font-bold ${on ? "bg-white/20 text-white" : "bg-stone-100 text-stone-500"}`}>
                  {i + 1}
                </span>
              )}
              {item.label}
            </button>
          );
        })}
      </div>
    </nav>
  );
}
