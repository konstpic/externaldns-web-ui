import { NavLink } from "react-router-dom";
import {
  Activity,
  Globe,
  LayoutDashboard,
  LogOut,
  Menu,
  Network,
  Server,
  ShieldCheck,
  User,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useSidebarStore } from "@/lib/sidebar-store";
import { useAuthStore } from "@/lib/auth-store";

const nav = [
  { to: "/", label: "Обзор", icon: LayoutDashboard },
  { to: "/records", label: "DNS записи", icon: Globe },
  { to: "/sources", label: "Источники", icon: Network },
  { to: "/controller", label: "Controller", icon: Server },
  { to: "/logs", label: "Логи", icon: Activity },
];

export function AppSidebar() {
  const expanded = useSidebarStore((s) => s.expanded);
  const toggle = useSidebarStore((s) => s.toggle);
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  return (
    <>
      <header className="fixed inset-x-0 top-0 z-40 flex h-14 items-center justify-between border-b border-border/10 bg-[rgb(var(--bg))]/80 px-4 backdrop-blur-xl md:hidden">
        <div className="flex items-center gap-2">
          <Globe className="h-5 w-5 text-primary" />
          <span className="font-semibold text-fg">ExternalDNS</span>
        </div>
        <button type="button" className="btn-secondary !px-3 !py-2" onClick={toggle}>
          <Menu className="h-4 w-4" />
        </button>
      </header>

      <aside
        className={cn(
          "fixed z-50 flex flex-col border-r border-border/10 bg-[rgb(var(--bg))]/90 backdrop-blur-xl transition-all duration-300",
          "inset-x-0 bottom-0 top-14 md:top-0 md:left-0 md:h-dvh md:w-[4.75rem]",
          expanded ? "md:w-60" : "max-md:-translate-x-full md:translate-x-0"
        )}
      >
        <div className="hidden h-16 items-center gap-3 border-b border-border/10 px-4 md:flex">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/15 text-primary">
            <Globe className="h-5 w-5" />
          </div>
          {expanded && (
            <div>
              <div className="text-sm font-semibold text-fg">ExternalDNS</div>
              <div className="text-xs text-subtle">Web UI</div>
            </div>
          )}
        </div>

        <nav className="flex flex-1 flex-col gap-1 p-2">
          {nav.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === "/"}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary/15 text-primary"
                    : "text-muted hover:bg-surface/10 hover:text-fg-secondary"
                )
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span className={cn(!expanded && "md:sr-only")}>{label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="space-y-1 border-t border-border/10 p-2">
          {user?.is_admin && (
            <NavLink
              to="/admin"
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary/15 text-primary"
                    : "text-muted hover:bg-surface/10 hover:text-fg-secondary"
                )
              }
            >
              <ShieldCheck className="h-4 w-4 shrink-0" />
              <span className={cn(!expanded && "md:sr-only")}>Админ</span>
            </NavLink>
          )}
          {user && (
            <div className={cn("flex items-center gap-2 px-3 py-2", !expanded && "md:justify-center")}>
              <User className="h-4 w-4 shrink-0 text-subtle" />
              {expanded && (
                <div className="min-w-0 flex-1">
                  <p className="truncate text-xs font-medium text-fg-secondary">{user.display_name}</p>
                  <p className="truncate text-[11px] text-subtle">{user.email}</p>
                </div>
              )}
            </div>
          )}
          <button
            type="button"
            className="btn-secondary flex w-full items-center justify-center gap-2 !py-2"
            onClick={logout}
          >
            <LogOut className="h-4 w-4" />
            <span className={cn(!expanded && "md:sr-only")}>Выйти</span>
          </button>
          <button type="button" className="btn-secondary hidden w-full !py-2 md:block" onClick={toggle}>
            {expanded ? "Свернуть" : "≡"}
          </button>
        </div>
      </aside>

      {expanded && (
        <button
          type="button"
          className="fixed inset-0 z-40 bg-black/40 md:hidden"
          aria-label="Close menu"
          onClick={toggle}
        />
      )}
    </>
  );
}
