import type { ReactNode } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { NetworkMapBackground } from "./network-map-background";
import { Navbar } from "./navbar";
import { CommandBar } from "@/components/ui/command-bar";
import { useCommands } from "@/lib/command-registry";

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
      ...(pathname !== "/configure"
        ? [
            {
              id: "nav:config",
              name: "config",
              description: "Go to Configure",
              category: "navigate",
              action: () => navigate("/configure"),
            },
          ]
        : []),
      ...(pathname !== "/roles"
        ? [
            {
              id: "nav:roles",
              name: "roles",
              description: "Go to Roles",
              category: "navigate",
              action: () => navigate("/roles"),
            },
          ]
        : []),
    ],
    [pathname, navigate],
  );

  return (
    <div className="relative min-h-screen">
      <NetworkMapBackground />
      <div className="relative z-10">
        <Navbar />
        {children}
      </div>
      <CommandBar />
    </div>
  );
}
