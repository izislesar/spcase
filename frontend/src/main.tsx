import { MotionConfig } from "motion/react";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { AppProviders } from "./app/providers";
import { router } from "./app/router";
import "./styles/tokens.css";
import "./styles/base.css";
import "./styles/view-transitions.css";

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Не найден корневой элемент #root");
}

createRoot(rootElement).render(
  <StrictMode>
    <MotionConfig reducedMotion="user">
      <AppProviders>
        <RouterProvider router={router} />
      </AppProviders>
    </MotionConfig>
  </StrictMode>,
);
