import { useRef, useState } from "react";
import { ChevronDown, KeyRound } from "lucide-react";
import { cn } from "@/lib/cn";
import { FormField } from "@/components/ui/form-field";
import { PopoverItem, PopoverPanel } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { isSecretRef, secretScheme as scheme } from "@/lib/secret-ref";

interface SecretInputProps {
  id?: string;
  label?: string;
  value: string;
  onChange: (value: string) => void;
  /** The references that the picker offers, each with its vault and key. */
  refs: { ref: string; vault: string; key: string }[];
  /** Whether the picker shows: the role may list the vaults and keys. */
  picker: boolean;
  className?: string;
}

/**
 * SecretInput is a secret field. It takes a reference, which a picker
 * beside it inserts, as FactInput inserts a @fact. reference. A value that
 * is not a reference is a password field; a reference shows as text, as it
 * is not a secret.
 */
export function SecretInput({
  id,
  label,
  value,
  onChange,
  refs,
  picker,
  className,
}: SecretInputProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  useOutsideClick(ref, () => setOpen(false), open);

  // Only a whole scheme shows as text: a value that starts as "se" may be a
  // password, and stays hidden.
  const isRef = value.startsWith(scheme);
  const filter = value.startsWith(scheme) ? value : "";
  const shown = filter ? refs.filter((r) => r.ref.startsWith(filter)) : refs;

  const pick = (r: string) => {
    onChange(r);
    setOpen(false);
  };

  return (
    <FormField id={id} label={label} className={className}>
      <div className="relative" ref={ref}>
        <input
          ref={inputRef}
          id={id}
          type={isRef ? "text" : "password"}
          value={value}
          placeholder={picker ? "secret://<vault>/<key>" : undefined}
          onChange={(e) => {
            onChange(e.target.value);
            setOpen(picker && isSecretRef(e.target.value));
          }}
          onKeyDown={(e) => {
            if (e.key === "Escape") setOpen(false);
          }}
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          className={cn(
            "flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm text-text",
            picker && "pr-8",
            "placeholder:text-text-muted",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
            isRef && "text-accent-light",
          )}
        />
        {picker && (
          <button
            type="button"
            onClick={() => {
              setOpen(!open);
              inputRef.current?.focus();
            }}
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-text-muted hover:text-primary"
            title="Insert secret reference"
          >
            {open ? (
              <ChevronDown className="h-3.5 w-3.5" />
            ) : (
              <KeyRound className="h-3.5 w-3.5" />
            )}
          </button>
        )}
        {open && (
          <PopoverPanel onClose={() => setOpen(false)}>
            <div className="sticky top-0 border-b border-border/30 bg-card px-3 py-1.5">
              <p className="text-xs font-semibold text-text-muted">
                Secret references
              </p>
            </div>
            {shown.length === 0 ? (
              <p className="px-3 py-2 text-xs text-text-muted">
                No secret. Add one on the Secrets page.
              </p>
            ) : (
              shown.map((r, i) => (
                <div key={r.ref}>
                  {(i === 0 || shown[i - 1].vault !== r.vault) && (
                    <p className="px-3 pt-2 text-xs font-semibold uppercase tracking-wide text-text-muted">
                      {r.vault}
                    </p>
                  )}
                  <PopoverItem
                    selected={value === r.ref}
                    onClick={() => pick(r.ref)}
                  >
                    <span className="text-xs font-medium text-accent-light">
                      {r.ref}
                    </span>
                  </PopoverItem>
                </div>
              ))
            )}
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
