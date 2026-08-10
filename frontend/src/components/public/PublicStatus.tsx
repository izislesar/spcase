import styles from "./PublicStatus.module.css";

/*
 * Editorial status block for public data sections: loading, error and
 * empty states of the public API. Replaces the pale generic boxes with the
 * system's own grammar — a hairline rule, a small accent marker, a clear
 * human-facing heading and one concise explanation. Retry is offered only
 * where the caller already has a safe retry (TanStack Query `refetch`).
 * ARIA: loading/empty are polite live regions; error uses alert semantics.
 */
export function PublicStatus({
  kind,
  title,
  detail,
  onRetry,
  retryLabel = "Попробовать ещё раз",
}: {
  kind: "loading" | "error" | "empty";
  title: string;
  detail?: string;
  onRetry?: () => void;
  retryLabel?: string;
}) {
  return (
    <div className={styles.status} role={kind === "error" ? "alert" : "status"}>
      <span
        className={`${styles.marker} ${kind === "loading" ? (styles.markerLoading ?? "") : ""}`}
        aria-hidden="true"
      />
      <p className={styles.title}>{title}</p>
      {detail && <p className={styles.detail}>{detail}</p>}
      {kind === "error" && onRetry && (
        <button type="button" className={styles.retry} onClick={onRetry}>
          {retryLabel}
        </button>
      )}
    </div>
  );
}
