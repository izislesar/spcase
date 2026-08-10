import { useId, useState } from "react";
import { BubbleScene } from "../../components/graphics/scenes/BubbleScene";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, useFaq } from "./api";
import styles from "./FaqPreview.module.css";

/*
 * FAQ — the quiet decompression after the louder sections: hairline rows
 * and one large conversation scene beside them. Accordion semantics are
 * unchanged: one open item at a time, collapsed regions stay in the DOM
 * hidden visually and from AT.
 */
export function FaqPreview() {
  const faq = useFaq();
  const idPrefix = useId();
  const [openId, setOpenId] = useState<number | null>(null);

  return (
    <section className={styles.faq} aria-labelledby="faq-heading">
      <div className={`container ${styles.faqInner}`}>
        <header className={styles.sectionHeader}>
          <p className={styles.eyebrow}>03 · Вопросы</p>
          <h2 id="faq-heading" className={styles.sectionTitle}>
            Коротко о главном
          </h2>
          <p className={styles.sectionIntro}>
            Если ответа здесь нет, организаторы помогут уточнить детали до начала чемпионата.
          </p>
          {/* One large quiet scene under the header; the rows stay calm. */}
          <BubbleScene className={styles.faqArt} />
        </header>
        {faq.isPending && <LoadingState label="Загружаем вопросы…" />}
        {faq.isError && <ErrorNotice message={errorMessage(faq.error)} />}
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
                    <section
                      id={regionId}
                      aria-labelledby={buttonId}
                      aria-hidden={!open}
                      className={
                        open ? `${styles.faqAnswer} ${styles.faqAnswerOpen}` : styles.faqAnswer
                      }
                    >
                      <p className={styles.faqAnswerText}>{item.answer}</p>
                    </section>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className={styles.sectionIntro}>Ответы появятся позже.</p>
          ))}
      </div>
    </section>
  );
}
