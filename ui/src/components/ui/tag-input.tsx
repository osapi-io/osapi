import { useState, type KeyboardEvent } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/cn";
import { FormField } from "@/components/ui/form-field";

interface TagInputProps {
  id?: string;
  label?: string;
  value: readonly string[];
  onChange: (value: string[]) => void;
  placeholder?: string;
  className?: string;
}

/**
 * TagInput takes a list of words, such as DNS servers: Enter, a comma or a
 * space adds the word that is typed, a paste of several words adds each,
 * and Backspace in an empty field removes the last. Each word is a tag with
 * its own remove button.
 */
export function TagInput({
  id,
  label,
  value,
  onChange,
  placeholder,
  className,
}: TagInputProps) {
  const [draft, setDraft] = useState("");

  const add = (text: string) => {
    const words = text
      .split(/[\s,]+/)
      .map((w) => w.trim())
      .filter((w) => w !== "" && !value.includes(w));
    if (words.length > 0) onChange([...value, ...words]);
    setDraft("");
  };

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === "," || e.key === " ") {
      if (draft.trim() !== "") {
        e.preventDefault();
        add(draft);
      } else if (e.key !== "Enter") {
        e.preventDefault();
      }
    } else if (e.key === "Backspace" && draft === "" && value.length > 0) {
      onChange(value.slice(0, -1));
    }
  };

  return (
    <FormField id={id} label={label} className={className}>
      <div
        className={cn(
          "flex min-h-9 w-full flex-wrap items-center gap-1.5 rounded-md border border-border bg-background px-2 py-1",
          "focus-within:border-primary/40 focus-within:ring-1 focus-within:ring-primary/20",
        )}
      >
        {value.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 rounded border border-border bg-white/[0.04] px-1.5 py-0.5 font-mono text-xs text-text"
          >
            {tag}
            <button
              type="button"
              onClick={() => onChange(value.filter((t) => t !== tag))}
              aria-label={`Remove ${tag}`}
              className="text-text-muted transition-colors hover:text-text"
            >
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
        <input
          id={id}
          value={draft}
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          placeholder={value.length === 0 ? placeholder : undefined}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={onKeyDown}
          onBlur={() => draft.trim() !== "" && add(draft)}
          onPaste={(e) => {
            const text = e.clipboardData.getData("text");
            if (/[\s,]/.test(text.trim())) {
              e.preventDefault();
              add(text);
            }
          }}
          className="h-7 min-w-24 flex-1 bg-transparent px-1 text-sm text-text placeholder:text-text-muted focus:outline-none"
        />
      </div>
    </FormField>
  );
}
