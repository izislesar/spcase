import { expect, test } from "@playwright/test";

/*
 * Home page smoke coverage. Runs against the Vite dev server without the Go
 * backend: public data sections then render their error/empty states, while
 * the static composition (headings, navigation, graphics) must stay intact.
 */

test("домашняя страница показывает ключевые разделы", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "СПК кейс-чемпионат", level: 1 })).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Три этапа. Одна сильная работа.", level: 2 }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { name: "Расписание", level: 2 })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Коротко о главном", level: 2 })).toBeVisible();
  await expect(page.getByRole("link", { name: "Подать заявку" }).first()).toBeVisible();
});

test("навигация доступна на всех вьюпортах", async ({ page, isMobile }) => {
  await page.goto("/");
  if (isMobile) {
    await expect(page.getByRole("button", { name: "Меню" })).toBeVisible();
  } else {
    const nav = page.getByRole("navigation", { name: "Основная навигация" });
    await expect(nav.getByRole("link", { name: "Главная" })).toBeVisible();
    await expect(nav.getByRole("link", { name: "Расписание" })).toBeVisible();
    await expect(nav.getByRole("link", { name: "Вход" })).toBeVisible();
  }
});

test("skip-ссылка переносит фокус к основному содержимому", async ({ page }) => {
  await page.goto("/");
  await page.keyboard.press("Tab");
  const skipLink = page.getByRole("link", { name: "К содержимому" });
  await expect(skipLink).toBeFocused();
  await expect(skipLink).toBeVisible();
  await page.keyboard.press("Enter");
  await expect(page.locator("#main-content")).toBeFocused();
});

test("публичные состояния загрузки имеют status-семантику", async ({ page }) => {
  await page.goto("/");
  // Публичные секции начинают с загрузки: PublicStatus с role="status"
  // присутствует в DOM независимо от доступности бэкенда.
  await expect(page.getByRole("status").first()).toBeVisible();
});

test("редуцированная анимация не ломает страницу", async ({ page }) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "СПК кейс-чемпионат", level: 1 })).toBeVisible();
  await expect(page.getByRole("link", { name: "Подать заявку" }).first()).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Три этапа. Одна сильная работа.", level: 2 }),
  ).toBeVisible();
  await page.goto("/schedule");
  await expect(page.getByRole("heading", { name: "Расписание", level: 1 })).toBeVisible();
});

test.describe("мобильное меню", () => {
  test.skip(({ isMobile }) => !isMobile, "только для мобильного вьюпорта");

  test("открывается, ведёт по ссылке и закрывается", async ({ page }) => {
    await page.goto("/");
    const toggle = page.getByRole("button", { name: "Меню" });
    await toggle.tap();
    const dialog = page.getByRole("dialog", { name: "Меню" });
    await expect(dialog).toBeVisible();
    await dialog.getByRole("link", { name: "Расписание" }).tap();
    await expect(page).toHaveURL(/\/schedule$/);
    await expect(dialog).not.toBeVisible();
    await expect(page.getByRole("heading", { name: "Расписание", level: 1 })).toBeVisible();
  });

  test("закрывается по Escape и возвращает фокус на кнопку", async ({ page }) => {
    await page.goto("/");
    const toggle = page.getByRole("button", { name: "Меню" });
    await toggle.tap();
    await expect(page.getByRole("dialog", { name: "Меню" })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog", { name: "Меню" })).not.toBeVisible();
    await expect(toggle).toBeFocused();
  });
});

test.describe("вьюпорт 320px", () => {
  test.use({ viewport: { width: 320, height: 568 } });

  test("нет горизонтального переполнения", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("heading", { name: "СПК кейс-чемпионат", level: 1 })).toBeVisible();
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow).toBeLessThanOrEqual(0);
  });
});
