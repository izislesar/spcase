import styles from "./DataState.module.css";

/** Polite loading indicator for data sections. */
export function LoadingState({ label }: { label: string }) {
  return (
    <p className={styles.loading} role="status">
      <span className={styles.loadingMark} aria-hidden="true" />
      {label}
    </p>
  );
}

/** Inline failure notice; the section around it stays fully rendered. */
export function ErrorNotice({ message }: { message: string }) {
  return (
    <p className={styles.error} role="alert">
      {message}
    </p>
  );
}
