import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Input } from "@/components/ui/input";
import { Text } from "@/components/ui/text";
import { InfoBox } from "@/components/ui/info-box";
import { useAuth } from "@/hooks/use-auth";
import { checkTokenFormat } from "@/lib/token-format";
import { getHealthStatus } from "@/sdk/gen/health-check-api-health/health-check-api-health";
import { Key, LogIn } from "lucide-react";

export function SignIn() {
  const { setToken } = useAuth();
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);

  const [checking, setChecking] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = value.trim();
    if (!trimmed) return;

    const check = checkTokenFormat(trimmed);
    if (!check.ok) {
      setError(check.error);
      return;
    }

    // Ask the server once. Only a 401 means it refused the token; a 403
    // still means the signature and expiry are good.
    setChecking(true);
    try {
      const res = await getHealthStatus({
        headers: { Authorization: `Bearer ${trimmed}` },
      });
      if (res.status === 401) {
        const reason = (res.data as { error?: string } | undefined)?.error;
        setError(
          `The server refused this token${reason ? `: ${reason}` : "."} Check that it was signed with this controller's signing key.`,
        );
        return;
      }
    } catch (err) {
      setError(
        `Could not reach the API to check the token${err instanceof Error ? `: ${err.message}` : "."}`,
      );
      return;
    } finally {
      setChecking(false);
    }

    setError(null);
    setToken(trimmed);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md px-6">
        <div className="mb-8 text-center">
          <img
            src="/logo.svg"
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
              <Input
                id="token"
                label="Bearer Token"
                icon={Key}
                type="password"
                value={value}
                onChange={(e) => {
                  setValue(e.target.value);
                  setError(null);
                }}
                placeholder="eyJhbGciOiJIUzI1NiIs..."
                autoFocus
              />

              {error && <ErrorBanner message={error} size="sm" />}

              <Button
                type="submit"
                variant="primary"
                size="lg"
                disabled={!value.trim() || checking}
                className="w-full"
              >
                <LogIn className="h-4 w-4" />
                {checking ? "Checking..." : "Sign In"}
              </Button>
            </form>

            <InfoBox className="mt-4">
              <Text variant="muted" as="p">
                Generate a token with the OSAPI CLI:
              </Text>
              <Text variant="mono-primary" as="code" className="mt-1 block">
                osapi controller token generate
              </Text>
            </InfoBox>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
