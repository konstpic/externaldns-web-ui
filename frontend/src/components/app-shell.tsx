import { type ReactNode } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { cn } from "@/lib/utils";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <>
      <AppSidebar />
      <div className={cn("relative flex min-h-dvh flex-1 flex-col app-content-offset")}>
        <main className="relative z-0 flex-1">{children}</main>
      </div>
    </>
  );
}
