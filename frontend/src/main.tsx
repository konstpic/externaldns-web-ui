import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Providers } from "@/components/providers";
import { AppRoutes } from "@/App";
import { syncSidebarDataset } from "@/lib/sidebar-store";
import "@/styles/index.css";

syncSidebarDataset();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <Providers>
        <div className="theme-ambient relative z-0 flex min-h-dvh flex-col">
          <AppRoutes />
        </div>
      </Providers>
    </BrowserRouter>
  </StrictMode>
);
