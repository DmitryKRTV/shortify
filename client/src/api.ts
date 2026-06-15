import { hasProfanity as checkProfanity } from "./profanity";

const API_URL = import.meta.env.VITE_API_URL || "";

const MIN_PASSWORD_LENGTH = 6;

export const MSG_PROFANITY = "Фу как некультурно";
export const MSG_INVALID_URL = "Enter a valid URL, e.g. https://example.com";
export const MSG_PASSWORD_SHORT =
  "А что ещё у тебя такое же короткое как этот пароль?";

export function hasProfanity(value: string): boolean {
  return checkProfanity(value);
}
const API_ERROR_MESSAGES: Record<string, string> = {
  "invalid email format": "Enter a valid email address.",
  "invalid url": MSG_INVALID_URL,
  [MSG_PASSWORD_SHORT]: MSG_PASSWORD_SHORT,
  [MSG_PROFANITY]: MSG_PROFANITY,
  "invalid input": "Invalid email or password.",
  "invalid credentials": "Invalid email or password.",
  "already exists": "This email is already registered."
};

function assertNoProfanity(email: string, password: string) {
  if (checkProfanity(email) || checkProfanity(password)) {
    throw new Error(MSG_PROFANITY);
  }
}

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

function assertValidEmail(email: string) {
  if (!isValidEmail(email)) {
    throw new Error("Enter a valid email address.");
  }
}

export function isValidUrl(value: string): boolean {
  try {
    const parsed = new URL(value.trim());
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

function formatApiError(text: string): string {
  try {
    const data = JSON.parse(text) as { error?: string };
    if (data.error && API_ERROR_MESSAGES[data.error]) {
      return API_ERROR_MESSAGES[data.error];
    }
    if (data.error) {
      return data.error;
    }
  } catch {
    // plain text response
  }

  return text || "Unknown error";
}

export type Link = {
  id: string;
  original_url: string;
  short_code: string;
  short_url: string;
  click_count: number;
  created_at: string;
};

export type RecentClick = {
  created_at: string;
};

export type LinkStats = {
  total_clicks: number;
  recent: RecentClick[];
};

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string
): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {})
    }
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(formatApiError(text) || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export async function register(email: string, password: string) {
  assertNoProfanity(email, password);
  assertValidEmail(email);

  if (password.length < MIN_PASSWORD_LENGTH) {
    throw new Error(MSG_PASSWORD_SHORT);
  }

  return request<{ token: string }>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function login(email: string, password: string) {
  assertNoProfanity(email, password);
  assertValidEmail(email);

  return request<{ token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function createLink(token: string, url: string) {
  if (checkProfanity(url)) {
    throw new Error(MSG_PROFANITY);
  }
  if (!isValidUrl(url)) {
    throw new Error(MSG_INVALID_URL);
  }

  return request<Link>(
    "/api/v1/links",
    {
      method: "POST",
      body: JSON.stringify({ url })
    },
    token
  );
}

export async function listLinks(token: string) {
  return request<Link[]>("/api/v1/links", {}, token);
}

export async function deleteLink(token: string, id: string) {
  return request<void>(
    `/api/v1/links/${id}`,
    {
      method: "DELETE"
    },
    token
  );
}

export async function getStats(token: string, id: string) {
  return request<LinkStats>(`/api/v1/links/${id}/stats`, {}, token);
}

export function getClickCount(link: Link) {
  return link.click_count;
}

/** Public short URL for display, copy, and open. */
export function resolveShortUrl(link: Link): string {
  const envBase = import.meta.env.VITE_SHORT_BASE_URL?.replace(/\/$/, "");
  if (envBase) {
    return `${envBase}/${link.short_code}`;
  }

  if (typeof window !== "undefined") {
    return `${window.location.origin}/${link.short_code}`;
  }

  return link.short_url;
}
