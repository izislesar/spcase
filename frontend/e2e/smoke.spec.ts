import { expect, test } from "@playwright/test";

const routes: { path: string; heading: string }[] = [
  { path: "/", heading: "Главная" },
  { path: "/schedule", heading: "Расписание" },
  { path: "/no-team", heading: "Нет команды" },
  { path: "/login", heading: "Вход" },
  { path: "/register", heading: "Регистрация" },
  { path: "/dashboard", heading: "Личный кабинет" },
  { path: "/jury/login", heading: "Вход для жюри" },
  { path: "/jury/register", heading: "Регистрация жюри" },
  { path: "/jury/teams", heading: "Команды" },
  { path: "/admin", heading: "Администрирование" },
];

for (const route of routes) {
  test(`загружает ${route.path}`, async ({ page }) => {
    await page.goto(route.path);
    await expect(page.getByRole("heading", { name: route.heading, level: 1 })).toBeVisible();
  });
}

test("/jury перенаправляет на /jury/teams", async ({ page }) => {
  await page.goto("/jury");
  await expect(page).toHaveURL(/\/jury\/teams$/);
  await expect(page.getByRole("heading", { name: "Команды", level: 1 })).toBeVisible();
});

test("неизвестный путь показывает страницу 404", async ({ page }) => {
  await page.goto("/such-page-does-not-exist");
  await expect(page.getByRole("heading", { name: "Страница не найдена", level: 1 })).toBeVisible();
});
