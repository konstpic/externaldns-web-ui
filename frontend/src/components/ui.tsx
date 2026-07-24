import { type ReactNode } from "react";
import { motion } from "framer-motion";

export function FadeIn({ children, delay = 0 }: { children: ReactNode; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay, ease: [0.25, 0.46, 0.45, 0.94] }}
    >
      {children}
    </motion.div>
  );
}

export function StatCard({
  label,
  value,
  hint,
  icon,
}: {
  label: string;
  value: string | number;
  hint?: string;
  icon: ReactNode;
}) {
  return (
    <div className="stat-card">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wider text-subtle">{label}</p>
          <p className="mt-2 text-3xl font-bold text-fg">{value}</p>
          {hint && <p className="mt-1 text-xs text-muted">{hint}</p>}
        </div>
        <div className="rounded-xl bg-primary/10 p-2.5 text-primary">{icon}</div>
      </div>
    </div>
  );
}

export function PageHeader({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="mb-8">
      <h1 className="text-3xl font-bold tracking-tight text-fg sm:text-4xl">{title}</h1>
      {subtitle && <p className="mt-2 max-w-2xl text-muted">{subtitle}</p>}
    </div>
  );
}

export function EmptyState({ message }: { message: string }) {
  return (
    <div className="glass px-6 py-14 text-center">
      <p className="text-muted">{message}</p>
    </div>
  );
}

export function LoadingState() {
  return (
    <div className="flex justify-center py-16">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
    </div>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="glass border-danger/30 px-6 py-8 text-center">
      <p className="text-danger">{message}</p>
    </div>
  );
}
