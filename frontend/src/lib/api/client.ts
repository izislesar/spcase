import { networkError, toApiError } from "./errors";

const BASE_PATH = "/api/v1";

type JsonBody = unknown;

interface RequestOptions {
  body?: JsonBody;
  signal?: AbortSignal;
}

async function request<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers({ Accept: "application/json" });
  let body: string | undefined;
  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
    body = JSON.stringify(options.body);
  }

  let response: Response;
  try {
    response = await fetch(`${BASE_PATH}${path}`, {
      method,
      headers,
      body,
      credentials: "same-origin",
      signal: options.signal,
    });
  } catch {
    throw networkError();
  }

  if (!response.ok) {
    throw await toApiError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export function apiGet<T>(path: string, options?: RequestOptions): Promise<T> {
  return request<T>("GET", path, options);
}

export function apiPost<T>(path: string, body?: JsonBody, options?: RequestOptions): Promise<T> {
  return request<T>("POST", path, { ...options, body });
}

export function apiPatch<T>(path: string, body?: JsonBody, options?: RequestOptions): Promise<T> {
  return request<T>("PATCH", path, { ...options, body });
}

export function apiDelete<T>(path: string, options?: RequestOptions): Promise<T> {
  return request<T>("DELETE", path, options);
}
