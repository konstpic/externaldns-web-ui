import { Routes, Route } from "react-router-dom";
import { AppShell } from "@/components/app-shell";
import { DashboardPage } from "@/pages/dashboard";
import { RecordsPage } from "@/pages/records";
import { SourcesPage } from "@/pages/sources";
import { ControllerPage } from "@/pages/controller";
import { LogsPage } from "@/pages/logs";

export function AppRoutes() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/records" element={<RecordsPage />} />
        <Route path="/sources" element={<SourcesPage />} />
        <Route path="/controller" element={<ControllerPage />} />
        <Route path="/logs" element={<LogsPage />} />
      </Routes>
    </AppShell>
  );
}
