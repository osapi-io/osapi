/** The scheme of each reference to a secret: secret://<vault>/<key>. */
export const secretScheme = "secret://";

/** The reference of key in vault: secret://<vault>/<key>. */
export function secretRef(vault: string, key: string): string {
  return `${secretScheme}${vault}/${key}`;
}

/** Whether v is a reference, or the start of one that a person types. */
export function isSecretRef(v: string): boolean {
  return (
    v.startsWith(secretScheme) || (v.length > 0 && secretScheme.startsWith(v))
  );
}
