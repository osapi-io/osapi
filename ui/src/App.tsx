import { lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { PageLayout } from "./components/layout/page-layout";
import { AuthProvider } from "./lib/auth";
import { SessionProvider } from "./lib/session";
import { CatalogProvider } from "./lib/catalog-provider";
import { useSession } from "@/hooks/use-session";
import { CommandRegistryProvider } from "./lib/command-registry";
import { Dashboard } from "./pages/dashboard";
import { Spinner } from "./components/ui/spinner";
import { PlanProvider } from "./lib/plan-provider";
import { Roles } from "./pages/roles";
import { Audit } from "./pages/audit";
import { Traces } from "./pages/traces";
import { Jobs } from "./pages/jobs";
import { Admission } from "./pages/admission";
import { SignIn } from "./pages/sign-in";

// The pages of the behaviors and the plans load their code on the first
// visit; the dashboard does not need it.
const Behaviors = lazy(() =>
  import("./pages/behaviors").then((m) => ({ default: m.Behaviors })),
);
const BehaviorPage = lazy(() =>
  import("./pages/behavior").then((m) => ({ default: m.BehaviorPage })),
);
const Plan = lazy(() =>
  import("./pages/plan").then((m) => ({ default: m.Plan })),
);

function AuthenticatedApp() {
  const { isAuthenticated } = useSession();

  // The sign-in page needs no catalog. After sign-in, CatalogProvider loads
  // it once, and each page under it reads it.
  if (!isAuthenticated) {
    return <SignIn />;
  }

  return (
    <CatalogProvider>
      <AuthProvider>
        <AuthenticatedPages />
      </AuthProvider>
    </CatalogProvider>
  );
}

function AuthenticatedPages() {
  return (
    <PlanProvider>
      <PageLayout>
        <Suspense fallback={<Spinner />}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/behaviors" element={<Behaviors />} />
            <Route path="/behaviors/:name" element={<BehaviorPage />} />
            <Route
              path="/plans"
              element={<Navigate to="/plans/new" replace />}
            />
            <Route path="/plans/new" element={<Plan />} />
            <Route
              path="/roles"
              element={<Navigate to="/admin/roles" replace />}
            />
            <Route path="/admin/audit" element={<Audit />} />
            <Route path="/admin/jobs" element={<Jobs />} />
            <Route path="/admin/jobs/:id" element={<Jobs />} />
            <Route path="/admin/traces" element={<Traces />} />
            <Route path="/admin/traces/:id" element={<Traces />} />
            <Route path="/admin/roles" element={<Roles />} />
            <Route path="/admin/admission" element={<Admission />} />
          </Routes>
        </Suspense>
      </PageLayout>
    </PlanProvider>
  );
}

export default function App() {
  return (
    <SessionProvider>
      <BrowserRouter>
        <CommandRegistryProvider>
          <AuthenticatedApp />
        </CommandRegistryProvider>
      </BrowserRouter>
    </SessionProvider>
  );
}
