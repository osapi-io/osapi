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
  ShieldAlert,
  GitBranch,
} from "lucide-react";
import { cn } from "@/lib/cn";
import { IconButton } from "@/components/ui/icon-button";
import { useAuth } from "@/hooks/use-auth";
import { ROLES, type Permission } from "@/lib/permissions";
import { Badge } from "@/components/ui/badge";
import { Dropdown } from "@/components/ui/dropdown";
import { useOutsideClick } from "@/hooks/use-outside-click";

const navItems = [
  { to: "/", label: "Dashboard", icon: Activity },
  { to: "/configure", label: "Configure", icon: Settings },
];

// Each link shows to a user with its permission. A link without one shows to
// a user who can read the audit log, as the whole menu did before.
const adminLinks: {
  to: string;
  label: string;
  icon: typeof FileText;
  perm?: Permission;
}[] = [
  { to: "/admin/audit", label: "Audit Log", icon: FileText },
  { to: "/admin/admission", label: "Admission", icon: ShieldAlert },
  { to: "/admin/jobs", label: "Jobs", icon: Briefcase },
  { to: "/admin/roles", label: "Roles", icon: Shield },
  { to: "/admin/traces", label: "Traces", icon: GitBranch, perm: "trace:read" },
];

export function Navbar() {
  const location = useLocation();
  const { role, roleOverride, setRoleOverride, clearToken, can, tokenRoles } =
    useAuth();
  const isTokenAdmin = tokenRoles.includes("admin");
  const visibleLinks = adminLinks.filter((link) =>
    can(link.perm ?? "audit:read"),
  );
  const showAdmin = visibleLinks.length > 0;
  const [adminOpen, setAdminOpen] = useState(false);
  const adminRef = useRef<HTMLDivElement>(null);

  const closeAdmin = useCallback(() => setAdminOpen(false), []);
  useOutsideClick(adminRef, closeAdmin, adminOpen);

  const isAdminPath = location.pathname.startsWith("/admin");

  return (
    <nav className="sticky top-0 z-50 flex h-16 items-center border-b border-border bg-card px-6 backdrop-blur-xl">
      <Link to="/" className="mr-8 flex items-center gap-2">
        <img src="/logo.svg" alt="OSAPI" className="h-8 w-auto" />
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
                active ? "text-text" : "text-text-muted hover:text-text",
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
                isAdminPath ? "text-text" : "text-text-muted hover:text-text",
              )}
            >
              <ShieldCheck className="h-5 w-5" />
              Admin
            </button>
            {adminOpen && (
              <div className="absolute left-0 top-full mt-1 w-44 rounded-md border border-border/60 bg-card py-1 shadow-xl">
                {visibleLinks.map((link) => {
                  const active = location.pathname.startsWith(link.to);
                  return (
                    <Link
                      key={link.to}
                      to={link.to}
                      onClick={() => setAdminOpen(false)}
                      className={cn(
                        "flex items-center gap-2 px-4 py-2 text-sm transition-colors",
                        active
                          ? "bg-white/10 text-text"
                          : "text-text-muted hover:bg-white/5 hover:text-text",
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
        <Badge variant="muted">
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
        <IconButton
          icon={LogOut}
          variant="danger"
          size="md"
          onClick={clearToken}
          title="Sign out"
        />
      </div>
    </nav>
  );
}
