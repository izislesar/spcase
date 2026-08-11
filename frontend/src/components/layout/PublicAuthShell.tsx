import type { ReactNode } from "react";
import styles from "./PublicAuthShell.module.css";

/*
 * Shell for the public auth routes (/register, /login, /jury/register,
 * /jury/login): one dark canvas, a clear title, concise context, secondary
 * navigation. These routes are still functional placeholders until the auth
 * stage: the note states that plainly instead of faking a form. No art
 * panel, no motion — auth surfaces are especially quiet.
 */
export function PublicAuthShell({
  context,
  title,
  lead,
  secondary,
}: {
  /** Factual context label, e.g. the audience («Участникам», «Жюри»). */
  context: string;
  title: string;
  lead: string;
  /** Secondary navigation (e.g. the login↔registration cross-link). */
  secondary?: ReactNode;
}) {
  return (
    <section className={styles.page} aria-labelledby="auth-heading">
      <div className={`container ${styles.inner}`}>
        <p className={styles.context}>{context}</p>
        <h1 id="auth-heading" className={styles.title}>
          {title}
        </h1>
        <p className={styles.lead}>{lead}</p>
        <p className={styles.note}>
          Рабочая форма появится на следующем этапе миграции — страница уже занимает своё место в
          структуре сайта.
        </p>
        {secondary && <div className={styles.secondary}>{secondary}</div>}
      </div>
    </section>
  );
}
