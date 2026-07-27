import { NavLink, Outlet } from "react-router-dom";
import { ClipboardList, LayoutDashboard, Settings, Wrench } from "lucide-react";
import { cn, pageShellClass } from "@/lib/utils";
import { PageHeader } from "@/components/ui";

const adminNav = [
  { to: "/admin", label: "Обзор", icon: LayoutDashboard, end: true },
  { to: "/admin/manage", label: "Управление DNS", icon: Wrench },
  { to: "/admin/audit", label: "Audit log", icon: ClipboardList },
  { to: "/admin/settings", label: "Настройки", icon: Settings },
];

export function AdminLayout() {
  return (
    <div className={cn(pageShellClass, "py-10")}>
      <PageHeader
        title="Администрирование"
        subtitle="Управление DNS записями, audit log и конфигурация"
      />
      <div className="mb-8 flex flex-wrap gap-2">
        {adminNav.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              cn(
                "inline-flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-primary/15 text-primary"
                  : "bg-surface/10 text-muted hover:text-fg-secondary"
              )
            }
          >
            <Icon className="h-4 w-4" />
            {label}
          </NavLink>
        ))}
      </div>
      <Outlet />
    </div>
  );
}
