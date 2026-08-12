import { motion, type Transition, useReducedMotion } from "motion/react";
import { useId, useState } from "react";
import { PublicStatus } from "../../components/public/PublicStatus";
import { QUIET_EASE } from "../../lib/motion";
import { errorMessage, useFaq } from "./api";
import styles from "./FaqPreview.module.css";

/*
 * FAQ — quiet: heading, questions, answers, separators. The section's only
 * motion is the accordion open/close (state communication). Accordion
 * semantics are unchanged: one open item at a time,
 * aria-expanded/aria-controls wiring intact, and collapsed regions stay in
 * the DOM hidden visually and from AT. The answer reveals through a Motion
 * height animation (~300ms, near-instant under reduced motion), so
 * neighboring items move smoothly instead of jumping.
 */
export function FaqPreview() {
  const faq = useFaq();
  const reduced = useReducedMotion();
  const idPrefix = useId();
  const [openId, setOpenId] = useState<number | null>(null);

  const answerTransition: Transition = reduced
    ? { duration: 0.01 }
    : { duration: 0.3, ease: QUIET_EASE };

  return (
    <section className={styles.faq} aria-labelledby="faq-heading">
      <div className={`container-wide ${styles.faqInner}`}>
        <header className={styles.sectionHeader}>
          <h2 id="faq-heading" className={styles.sectionTitle}>
            Частые вопросы
          </h2>
          <p className={styles.sectionIntro}>
            Если ответа здесь нет, организаторы помогут уточнить детали до начала чемпионата.
          </p>
        </header>
        {faq.isPending && <PublicStatus kind="loading" title="Загружаем вопросы…" />}
        {faq.isError && (
          <PublicStatus
            kind="error"
            title="Не удалось загрузить ответы"
            detail={errorMessage(faq.error)}
            onRetry={() => faq.refetch()}
          />
        )}
        {faq.isSuccess &&
          (faq.data.faq.length > 0 ? (
            <div className={styles.faqList}>
              {faq.data.faq.map((item) => {
                const open = openId === item.id;
                const buttonId = `${idPrefix}-q-${item.id}`;
                const regionId = `${idPrefix}-a-${item.id}`;
                return (
                  <div className={styles.faqItem} key={item.id}>
                    <h3 className={styles.faqQuestion}>
                      <button
                        type="button"
                        id={buttonId}
                        className={styles.faqToggle}
                        aria-expanded={open}
                        aria-controls={regionId}
                        onClick={() => setOpenId(open ? null : item.id)}
                      >
                        {item.question}
                        <svg
                          className={
                            open ? `${styles.faqIcon} ${styles.faqIconOpen}` : styles.faqIcon
                          }
                          width="16"
                          height="16"
                          viewBox="0 0 16 16"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2"
                          aria-hidden="true"
                          focusable="false"
                        >
                          <path d="M8 1v14M1 8h14" />
                        </svg>
                      </button>
                    </h3>
                    <motion.section
                      id={regionId}
                      aria-labelledby={buttonId}
                      aria-hidden={!open}
                      className={styles.faqAnswer}
                      initial={false}
                      animate={{
                        height: open ? "auto" : 0,
                        opacity: open ? 1 : 0,
                      }}
                      transition={answerTransition}
                    >
                      <p className={styles.faqAnswerText}>{item.answer}</p>
                    </motion.section>
                  </div>
                );
              })}
            </div>
          ) : (
            <PublicStatus
              kind="empty"
              title="Ответы появятся позже"
              detail="Организаторы соберут частые вопросы ближе к старту чемпионата."
            />
          ))}
      </div>
    </section>
  );
}
