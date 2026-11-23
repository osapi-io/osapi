import { useEffect, useRef, type ReactNode } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { NetworkMapBackground } from "./network-map-background";
import { Navbar } from "./navbar";
import { CommandBar } from "@/components/ui/command-bar";
import { useCommands } from "@/hooks/use-commands";

interface PageLayoutProps {
  children: ReactNode;
}

export function PageLayout({ children }: PageLayoutProps) {
  const navigate = useNavigate();
  const { pathname } = useLocation();

  // Global navigation commands — hide the one for the current page
  useCommands(
    [
      ...(pathname !== "/"
        ? [
            {
              id: "nav:dash",
              name: "dash",
              description: "Go to Dashboard",
              category: "navigate",
              action: () => navigate("/"),
            },
          ]
        : []),
      ...(pathname !== "/behaviors"
        ? [
            {
              id: "nav:behaviors",
              name: "behaviors",
              description: "Go to Behaviors",
              category: "navigate",
              action: () => navigate("/behaviors"),
            },
          ]
        : []),
      ...(pathname !== "/plans/new"
        ? [
            {
              id: "nav:plans",
              name: "plans",
              description: "Go to Plans",
              category: "navigate",
              action: () => navigate("/plans/new"),
            },
          ]
        : []),
      ...(pathname !== "/admin/audit"
        ? [
            {
              id: "nav:audit",
              name: "admin audit",
              description: "Go to Audit Log",
              category: "admin",
              action: () => navigate("/admin/audit"),
            },
          ]
        : []),
      ...(pathname !== "/admin/jobs"
        ? [
            {
              id: "nav:jobs",
              name: "admin jobs",
              description: "Go to Jobs",
              category: "admin",
              action: () => navigate("/admin/jobs"),
            },
          ]
        : []),
      ...(pathname !== "/admin/roles"
        ? [
            {
              id: "nav:roles",
              name: "admin roles",
              description: "Go to Roles",
              category: "admin",
              action: () => navigate("/admin/roles"),
            },
          ]
        : []),
      ...(pathname !== "/admin/admission"
        ? [
            {
              id: "nav:admission",
              name: "admin admission",
              description: "Go to Admission",
              category: "admin",
              action: () => navigate("/admin/admission"),
            },
          ]
        : []),
    ],
    [pathname, navigate],
  );

  // Section headings take the logo's colors in page order, as the CLI's
  // nextHeadingColor does, so two neighbors never share a color. Data loads
  // after the first render, so the order is set again when the page changes.
  const root = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = root.current;
    if (!el) return;
    const colors = 5;
    const assign = () => {
      el.querySelectorAll<HTMLElement>("[data-section-heading]").forEach(
        (h, i) => {
          const c = String((i % colors) + 1);
          if (h.dataset.headingColor !== c) h.dataset.headingColor = c;
        },
      );
    };
    assign();
    const observer = new MutationObserver(assign);
    observer.observe(el, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);

  return (
    <div ref={root} className="relative min-h-screen">
      <NetworkMapBackground />
      <div className="relative z-10">
        <Navbar />
        {children}
      </div>
      <CommandBar />
    </div>
  );
}
