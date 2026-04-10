import { useState, useRef, useCallback } from "react";
import { Link, useLocation } from "react-router-dom";
import {
  Activity,
  Settings,
  ShieldCheck,
  LogOut,
  FileText,
  Briefcase,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/cn";
import { useAuth } from "@/lib/auth";
import { ROLES } from "@/lib/permissions";
import { Badge } from "@/components/ui/badge";
import { Dropdown } from "@/components/ui/dropdown";
import { useOutsideClick } from "@/hooks/use-outside-click";

const navItems = [
  { to: "/", label: "Dashboard", icon: Activity },
  { to: "/configure", label: "Configure", icon: Settings },
];

const adminLinks = [
  { to: "/admin/audit", label: "Audit Log", icon: FileText },
  { to: "/admin/jobs", label: "Jobs", icon: Briefcase },
  { to: "/admin/roles", label: "Roles", icon: Shield },
];

function roleBadgeVariant(name: string) {
  switch (name) {
    case "admin":
      return "error" as const;
    case "operator":
      return "running" as const;
    default:
      return "muted" as const;
  }
}

export function Navbar() {
  const location = useLocation();
  const { role, roleOverride, setRoleOverride, clearToken, can, tokenRoles } =
    useAuth();
  const isTokenAdmin = tokenRoles.includes("admin");
  const showAdmin = can("audit:read");
  const [adminOpen, setAdminOpen] = useState(false);
  const adminRef = useRef<HTMLDivElement>(null);

  const closeAdmin = useCallback(() => setAdminOpen(false), []);
  useOutsideClick(adminRef, closeAdmin, adminOpen);

  const isAdminPath = location.pathname.startsWith("/admin");

  return (
    <nav className="sticky top-0 z-50 flex h-16 items-center border-b-2 border-accent bg-card px-6 backdrop-blur-xl">
      <Link to="/" className="mr-8 flex items-center gap-2">
        <img src="/logo.png" alt="OSAPI" className="h-8 w-auto" />
      </Link>
      <div className="flex gap-1">
        {navItems.map((item) => {
          const active = location.pathname === item.to;
          return (
            <Link
              key={item.to}
              to={item.to}
              className={cn(
                "flex items-center gap-2 rounded-md px-4 py-2 text-base font-medium transition-colors",
                active ? "text-primary" : "text-text hover:text-primary",
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.label}
            </Link>
          );
        })}

        {/* Admin dropdown — visible based on effective role */}
        {showAdmin && (
          <div className="relative" ref={adminRef}>
            <button
              type="button"
              onClick={() => setAdminOpen(!adminOpen)}
              className={cn(
                "flex items-center gap-2 rounded-md px-4 py-2 text-base font-medium transition-colors",
                isAdminPath ? "text-primary" : "text-text hover:text-primary",
              )}
            >
              <ShieldCheck className="h-5 w-5" />
              Admin
            </button>
            {adminOpen && (
              <div className="absolute left-0 top-full mt-1 w-44 rounded-md border border-border/60 bg-card py-1 shadow-xl">
                {adminLinks.map((link) => {
                  const active = location.pathname === link.to;
                  return (
                    <Link
                      key={link.to}
                      to={link.to}
                      onClick={() => setAdminOpen(false)}
                      className={cn(
                        "flex items-center gap-2 px-4 py-2 text-sm transition-colors",
                        active
                          ? "bg-primary/10 text-primary"
                          : "text-text hover:bg-white/5 hover:text-primary",
                      )}
                    >
                      <link.icon className="h-4 w-4" />
                      {link.label}
                    </Link>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Role selector */}
      <div className="ml-auto flex items-center gap-3">
        <Badge variant={roleBadgeVariant(role.name)}>
          <Shield className="h-3 w-3" />
          {role.label}
        </Badge>
        {isTokenAdmin && (
          <Dropdown
            value={roleOverride ?? ""}
            onChange={(v) => setRoleOverride(v || null)}
            placeholder="Auto (from token)"
            options={[
              { value: "", label: "Auto (from token)" },
              ...ROLES.map((r) => ({ value: r.name, label: r.label })),
            ]}
            className="w-40"
          />
        )}
        <button
          onClick={clearToken}
          className="rounded-md p-1.5 text-text-muted transition-colors hover:bg-white/5 hover:text-status-error"
          title="Sign out"
        >
          <LogOut className="h-4 w-4" />
        </button>
      </div>
    </nav>
  );
}
