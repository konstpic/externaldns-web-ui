import { Routes, Route } from "react-router-dom";
import { AppShell } from "@/components/app-shell";
import { AdminRoute, ProtectedRoute } from "@/components/providers";
import { DashboardPage } from "@/pages/dashboard";
import { RecordsPage } from "@/pages/records";
import { SourcesPage } from "@/pages/sources";
import { ControllerPage } from "@/pages/controller";
import { LogsPage } from "@/pages/logs";
import { AuthPage } from "@/pages/auth";
import { AuthCallbackPage } from "@/pages/auth-callback";
import { AdminSettingsPage } from "@/pages/admin/settings";

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/auth" element={<AuthPage />} />
      <Route path="/auth/callback" element={<AuthCallbackPage />} />
      <Route
        path="/*"
        element={
          <ProtectedRoute>
            <AppShell>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/records" element={<RecordsPage />} />
                <Route path="/sources" element={<SourcesPage />} />
                <Route path="/controller" element={<ControllerPage />} />
                <Route path="/logs" element={<LogsPage />} />
                <Route
                  path="/admin/settings"
                  element={
                    <AdminRoute>
                      <AdminSettingsPage />
                    </AdminRoute>
                  }
                />
              </Routes>
            </AppShell>
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}
