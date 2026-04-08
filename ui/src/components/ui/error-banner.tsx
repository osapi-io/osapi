import { cn } from "@/lib/cn";
import { AlertCircle } from "lucide-react";

interface ErrorBannerProps {
  message: string;
  size?: "sm" | "md";
  className?: string;
}

export function ErrorBanner({
  message,
  size = "md",
  className,
}: ErrorBannerProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-md border border-status-error/20 bg-status-error/5",
        size === "sm" ? "items-start px-3 py-2" : "p-4",
        className,
      )}
    >
      <AlertCircle
        className={cn(
          "shrink-0 text-status-error",
          size === "sm" ? "mt-0.5 h-3.5 w-3.5" : "h-5 w-5",
        )}
      />
      <p
        className={cn(
          "text-status-error",
          size === "sm" ? "text-xs" : "text-sm",
        )}
      >
        {message}
      </p>
    </div>
  );
}
