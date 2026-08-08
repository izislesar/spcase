import { z } from "zod";

const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
  }),
});

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export async function toApiError(response: Response): Promise<ApiError> {
  try {
    const body: unknown = await response.json();
    const parsed = errorEnvelopeSchema.safeParse(body);
    if (parsed.success) {
      return new ApiError(response.status, parsed.data.error.code, parsed.data.error.message);
    }
  } catch {
    // Тело ответа не является JSON — обрабатывается ниже.
  }
  return new ApiError(response.status, "invalid_response", "Сервер вернул некорректный ответ");
}

export function networkError(): ApiError {
  return new ApiError(0, "network_error", "Нет соединения с сервером");
}
