const API_URL = import.meta.env.VITE_API_URL || "";

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
    throw new Error(text || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export async function register(email: string, password: string) {
  return request<{ token: string }>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function login(email: string, password: string) {
  return request<{ token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function createLink(token: string, url: string) {
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
