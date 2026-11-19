import { getAuthToken } from "@/lib/auth-token";

const API_BASE_URL = import.meta.env.OSAPI_API_URL || "";

export const apiFetch = async <T>(
  url: string,
  options?: RequestInit,
): Promise<T> => {
  const headers: Record<string, string> = {};

  // Don't set Content-Type for FormData — browser sets it with boundary
  const isFormData = options?.body instanceof FormData;
  if (!isFormData) {
    headers["Content-Type"] = "application/json";
  }

  const token = getAuthToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      ...headers,
      ...(options?.headers as Record<string, string>),
    },
  });

  return {
    data: await body(res),
    status: res.status,
    headers: res.headers,
  } as T;
};

// body is a response's JSON, or undefined when it has none, as a 204 does.
// A body that is not JSON, such as a proxy's error page, comes back as an
// error message, so a caller shows it rather than a parse error.
async function body(res: Response): Promise<unknown> {
  const text = await res.text();
  if (text === "") return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return { error: text };
  }
}
