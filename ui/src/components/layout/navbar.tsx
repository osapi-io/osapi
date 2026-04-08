import { Link, useLocation } from "react-router-dom";
import { Activity, Settings, Shield, LogOut } from "lucide-react";
import { cn } from "@/lib/cn";
import { useAuth } from "@/lib/auth";
import { ROLES } from "@/lib/permissions";
import { Badge } from "@/components/ui/badge";
import { Dropdown } from "@/components/ui/dropdown";

const navItems = [
  { to: "/", label: "Dashboard", icon: Activity },
  { to: "/configure", label: "Configure", icon: Settings },
  { to: "/roles", label: "Roles", icon: Shield },
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
  const { role, roleOverride, setRoleOverride, clearToken } = useAuth();

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
      </div>

      {/* Role selector */}
      <div className="ml-auto flex items-center gap-3">
        <Badge variant={roleBadgeVariant(role.name)}>
          <Shield className="h-3 w-3" />
          {role.label}
        </Badge>
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
