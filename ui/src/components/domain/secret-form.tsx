import { useState, type FormEvent } from "react";
import { KeyRound, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface SecretFormProps {
  /** The key to write, such as a key of the table; empty for a new key. */
  initialKey?: string;
  /** Writes the value of the key, and says whether the write succeeded. */
  onWrite: (key: string, value: string) => Promise<boolean>;
}

/**
 * SecretForm writes the value of a key of a vault. The value field is a
 * password field, and it is empty again after each write: the page never
 * holds a value longer than one write, and the API never gives one back.
 */
export function SecretForm({ initialKey = "", onWrite }: SecretFormProps) {
  const [key, setKey] = useState(initialKey);
  const [value, setValue] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    const ok = await onWrite(key, value);
    setBusy(false);
    setValue("");
    if (ok) setKey("");
  };

  return (
    <form onSubmit={submit} className="flex flex-wrap items-end gap-3">
      <div className="min-w-48 flex-1">
        <Input
          id="secret-key"
          label="Key"
          icon={KeyRound}
          placeholder="registry-token"
          value={key}
          onChange={(e) => setKey(e.target.value)}
        />
      </div>
      <div className="min-w-48 flex-1">
        <Input
          id="secret-value"
          label="Value"
          type="password"
          autoComplete="new-password"
          placeholder="The value, which no page shows"
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
      </div>
      <Button type="submit" size="sm" disabled={busy || !key || !value}>
        {busy && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
        Write
      </Button>
    </form>
  );
}
