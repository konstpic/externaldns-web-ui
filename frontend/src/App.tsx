import { Routes, Route, Navigate } from "react-router-dom";
import { AppShell } from "@/components/app-shell";
import { AdminRoute, ProtectedRoute } from "@/components/providers";
import { DashboardPage } from "@/pages/dashboard";
import { RecordsPage } from "@/pages/records";
import { SourcesPage } from "@/pages/sources";
import { ControllerPage } from "@/pages/controller";
import { LogsPage } from "@/pages/logs";
import { AuthPage } from "@/pages/auth";
import { AuthCallbackPage } from "@/pages/auth-callback";
import { AdminLayout } from "@/pages/admin/layout";
import { AdminOverviewPage } from "@/pages/admin/index";
import { AdminManagePage } from "@/pages/admin/manage";
import { AdminAuditPage } from "@/pages/admin/audit";
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
                  path="/admin/*"
                  element={
                    <AdminRoute>
                      <AdminLayout />
                    </AdminRoute>
                  }
                >
                  <Route index element={<AdminOverviewPage />} />
                  <Route path="manage" element={<AdminManagePage />} />
                  <Route path="audit" element={<AdminAuditPage />} />
                  <Route path="settings" element={<AdminSettingsPage />} />
                  <Route path="*" element={<Navigate to="/admin" replace />} />
                </Route>
              </Routes>
            </AppShell>
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}
