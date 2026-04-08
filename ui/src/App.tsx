import { BrowserRouter, Routes, Route } from "react-router-dom";
import { PageLayout } from "./components/layout/page-layout";
import { AuthProvider, useAuth } from "./lib/auth";
import { CommandRegistryProvider } from "./lib/command-registry";
import { Dashboard } from "./pages/dashboard";
import { Configure } from "./pages/configure";
import { Roles } from "./pages/roles";
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
        <Route path="/roles" element={<Roles />} />
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
