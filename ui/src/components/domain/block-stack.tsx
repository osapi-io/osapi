import type { ReactNode } from "react";

interface BlockStackProps {
  children: ReactNode[];
}

export function BlockStack({ children }: BlockStackProps) {
  return (
    <div className="space-y-0">
      {children.map((child, i) => (
        <div key={i}>
          {i > 0 && (
            <div className="flex justify-center py-1">
              <div className="h-4 w-px border-l-2 border-dashed border-accent/30" />
            </div>
          )}
          {child}
        </div>
      ))}
    </div>
  );
}
