import type { ReactNode } from "react";
import styles from "./PublicAuthShell.module.css";

type Field = "mustard" | "turquoise" | "navy" | "accent";

/* CSS-module class lookups type as string | undefined; fall back to "". */
const FIELD_CLASS: Record<Field, string> = {
  mustard: styles.fieldMustard ?? "",
  turquoise: styles.fieldTurquoise ?? "",
  navy: styles.fieldNavy ?? "",
  accent: styles.fieldAccent ?? "",
};

/*
 * Visual shell for the public auth routes (/register, /login,
 * /jury/register, /jury/login): the same editorial world as the homepage —
 * display typography on the canvas, one large artwork field beside the
 * content. These routes are still functional placeholders until the auth
 * stage: the note states that plainly instead of faking a form.
 */
export function PublicAuthShell({
  eyebrow,
  title,
  lead,
  field,
  art,
}: {
  eyebrow: string;
  title: string;
  lead: string;
  field: Field;
  art: ReactNode;
}) {
  return (
    <section className={styles.page} aria-labelledby="auth-heading">
      <div className={`container-wide ${styles.inner}`}>
        <div className={styles.content}>
          <p className={styles.eyebrow}>{eyebrow}</p>
          <h1 id="auth-heading" className={styles.title}>
            {title}
          </h1>
          <p className={styles.lead}>{lead}</p>
          <p className={styles.note}>
            Рабочая форма появится на следующем этапе миграции — страница уже занимает своё место в
            структуре сайта.
          </p>
        </div>
        <div className={`${styles.field} ${FIELD_CLASS[field]}`}>{art}</div>
      </div>
    </section>
  );
}
