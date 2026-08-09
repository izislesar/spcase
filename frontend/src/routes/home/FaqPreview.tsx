import { useId, useState } from "react";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, useFaq } from "./api";
import styles from "./home.module.css";

export function FaqPreview() {
  const faq = useFaq();
  const idPrefix = useId();
  const [openId, setOpenId] = useState<number | null>(null);

  return (
    <section className={styles.faq} aria-labelledby="faq-heading">
      <div className="container">
        <header className={styles.sectionHeader}>
          <p className={styles.eyebrow}>03 · Вопросы</p>
          <h2 id="faq-heading" className={styles.sectionTitle}>
            Коротко о главном
          </h2>
          <p className={styles.sectionIntro}>
            Если ответа здесь нет, организаторы помогут уточнить детали до начала чемпионата.
          </p>
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
                    {/* One open item at a time; collapsed regions stay in the
                        DOM but are hidden visually and from AT. The labelled
                        section carries the implicit region role. */}
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
