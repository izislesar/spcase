import { useQuery } from "@tanstack/react-query";
import { z } from "zod";
import { apiGet } from "../../lib/api/client";
import { ApiError } from "../../lib/api/errors";

/*
 * Public data for the home page. Schemas parse only what the page needs and
 * ignore unknown extra fields (Zod strips them by default). Response shapes
 * follow docs/contracts/http-api.md §3.
 */

const infoSchema = z.object({
  registration_deadline: z.string(),
  submission_deadline: z.string().optional(),
  is_registration_open: z.boolean().optional(),
  is_submission_open: z.boolean().optional(),
});

const scheduleSchema = z.object({
  events: z.array(
    z.object({
      id: z.number(),
      title: z.string(),
      start_time: z.string(),
      description: z.string(),
    }),
  ),
});

const faqSchema = z.object({
  faq: z.array(
    z.object({
      id: z.number(),
      question: z.string(),
      answer: z.string(),
    }),
  ),
});

export type PublicInfo = z.infer<typeof infoSchema>;
export type ScheduleEvent = z.infer<typeof scheduleSchema>["events"][number];
export type FaqItem = z.infer<typeof faqSchema>["faq"][number];

export function usePublicInfo() {
  return useQuery({
    queryKey: ["public", "info"],
    queryFn: async () => infoSchema.parse(await apiGet("/info")),
  });
}

export function useSchedule() {
  return useQuery({
    queryKey: ["public", "schedule"],
    queryFn: async () => scheduleSchema.parse(await apiGet("/schedule")),
  });
}

export function useFaq() {
  return useQuery({
    queryKey: ["public", "faq"],
    queryFn: async () => faqSchema.parse(await apiGet("/faq")),
  });
}

const dateTimeFormatter = new Intl.DateTimeFormat("ru-RU", {
  dateStyle: "long",
  timeStyle: "short",
});

/** Long-date/short-time in the visitor's timezone; «—» for missing/invalid. */
export function formatEventTime(value: string): string {
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? "—" : dateTimeFormatter.format(time);
}

/** The datetime attribute needs a parseable value; undefined otherwise. */
export function eventDateTimeAttr(value: string): string | undefined {
  return Number.isNaN(new Date(value).getTime()) ? undefined : value;
}

/** Server message from the error envelope; a generic line as fallback. */
export function errorMessage(error: unknown): string {
  return error instanceof ApiError ? error.message : "Не удалось загрузить данные.";
}
