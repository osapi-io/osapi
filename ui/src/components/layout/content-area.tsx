import type { ReactNode } from "react";

interface ContentAreaProps {
  children: ReactNode;
  className?: string;
}

export function ContentArea({ children, className }: ContentAreaProps) {
  return (
    <main
      className={`mx-auto max-w-7xl px-6 py-8 bg-background/60 ${className ?? ""}`}
    >
      {children}
    </main>
  );
}
