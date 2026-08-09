import { Link, type LinkProps } from "react-router";
import styles from "./ActionLinks.module.css";

/** Primary action: a flat accent block — editorial, not a pill. */
export function ButtonLink({ className, ...props }: LinkProps) {
  return <Link className={`${styles.button} ${className ?? ""}`} {...props} />;
}

interface ArrowLinkProps extends LinkProps {
  /** Arrow glyph shown after the label; decorative, hidden from AT. */
  arrow?: string;
}

/** Secondary action: quiet text link with an underline and a trailing arrow. */
export function ArrowLink({ className, arrow = "→", children, ...props }: ArrowLinkProps) {
  return (
    <Link className={`${styles.arrowLink} ${className ?? ""}`} {...props}>
      {children}
      <span className={styles.arrow} aria-hidden="true">
        {arrow}
      </span>
    </Link>
  );
}
