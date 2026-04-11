import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { PageLayout } from "./components/layout/page-layout";
import { AuthProvider } from "./lib/auth";
import { useAuth } from "@/hooks/use-auth";
import { CommandRegistryProvider } from "./lib/command-registry";
import { Dashboard } from "./pages/dashboard";
import { Configure } from "./pages/configure";
import { Roles } from "./pages/roles";
import { Audit } from "./pages/audit";
import { Jobs } from "./pages/jobs";
import { SignIn } from "./pages/sign-in";

function AuthenticatedApp() {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <SignIn />;
  }

  return (
    <PageLayout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/configure" element={<Configure />} />
        <Route path="/roles" element={<Navigate to="/admin/roles" replace />} />
        <Route path="/admin/audit" element={<Audit />} />
        <Route path="/admin/jobs" element={<Jobs />} />
        <Route path="/admin/roles" element={<Roles />} />
      </Routes>
    </PageLayout>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <CommandRegistryProvider>
          <AuthenticatedApp />
        </CommandRegistryProvider>
      </BrowserRouter>
    </AuthProvider>
  );
}
