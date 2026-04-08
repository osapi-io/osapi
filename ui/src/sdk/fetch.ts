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

  const data = await res.json();

  return { data, status: res.status, headers: res.headers } as T;
};
