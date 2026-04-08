import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ErrorBanner } from "@/components/ui/error-banner";
import { FormField } from "@/components/ui/form-field";
import { Text } from "@/components/ui/text";
import { InfoBox } from "@/components/ui/info-box";
import { useAuth } from "@/lib/auth";
import { Key, LogIn } from "lucide-react";

export function SignIn() {
  const { setToken } = useAuth();
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = value.trim();
    if (!trimmed) return;

    // Basic JWT structure check (three dot-separated base64 parts)
    const parts = trimmed.split(".");
    if (parts.length !== 3) {
      setError(
        "Invalid token format. Expected a JWT (three dot-separated parts).",
      );
      return;
    }

    // Try to decode and check for roles
    try {
      const payload = JSON.parse(
        atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")),
      );
      if (!payload.roles || !Array.isArray(payload.roles)) {
        setError(
          "Token is missing roles claim. Generate one with: osapi token generate",
        );
        return;
      }
    } catch {
      setError("Failed to decode token payload.");
      return;
    }

    setError(null);
    setToken(trimmed);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md px-6">
        <div className="mb-8 text-center">
          <img
            src="/logo.png"
            alt="OSAPI"
            className="mx-auto mb-4 h-12 w-auto"
          />
          <h1 className="text-2xl font-bold text-white">Sign in to OSAPI</h1>
          <p className="mt-2 text-sm text-text-muted">
            Paste a JWT token to authenticate
          </p>
        </div>

        <Card>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <FormField id="token" label="Bearer Token">
                <div className="relative">
                  <Key className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
                  <input
                    id="token"
                    type="password"
                    value={value}
                    onChange={(e) => {
                      setValue(e.target.value);
                      setError(null);
                    }}
                    placeholder="eyJhbGciOiJIUzI1NiIs..."
                    autoFocus
                    className="h-10 w-full rounded-md border border-border bg-background pl-10 pr-3 text-sm text-text outline-none placeholder:text-text-muted/50 focus:border-primary/40 focus:ring-1 focus:ring-primary/20"
                  />
                </div>
              </FormField>

              {error && <ErrorBanner message={error} size="sm" />}

              <Button
                type="submit"
                variant="primary"
                size="lg"
                disabled={!value.trim()}
                className="w-full"
              >
                <LogIn className="h-4 w-4" />
                Sign In
              </Button>
            </form>

            <InfoBox className="mt-4">
              <Text variant="muted" as="p">
                Generate a token with the OSAPI CLI:
              </Text>
              <Text variant="mono-primary" as="code" className="mt-1 block">
                osapi token generate
              </Text>
            </InfoBox>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
