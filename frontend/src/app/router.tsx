import { createBrowserRouter, Navigate } from "react-router";
import { AdminPage } from "../routes/admin/AdminPage";
import { LoginPage } from "../routes/auth/LoginPage";
import { RegisterPage } from "../routes/auth/RegisterPage";
import { DashboardPage } from "../routes/dashboard/DashboardPage";
import { HomePage } from "../routes/home/HomePage";
import { JuryLoginPage } from "../routes/jury/JuryLoginPage";
import { JuryRegisterPage } from "../routes/jury/JuryRegisterPage";
import { JuryTeamsPage } from "../routes/jury/JuryTeamsPage";
import { NoTeamPage } from "../routes/no-team/NoTeamPage";
import { NotFoundPage } from "../routes/not-found/NotFoundPage";
import { SchedulePage } from "../routes/schedule/SchedulePage";
import { AppShell } from "./shell/AppShell";

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/schedule", element: <SchedulePage /> },
      { path: "/no-team", element: <NoTeamPage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/dashboard", element: <DashboardPage /> },
      { path: "/jury", element: <Navigate to="/jury/teams" replace /> },
      { path: "/jury/login", element: <JuryLoginPage /> },
      { path: "/jury/register", element: <JuryRegisterPage /> },
      { path: "/jury/teams", element: <JuryTeamsPage /> },
      { path: "/admin", element: <AdminPage /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
