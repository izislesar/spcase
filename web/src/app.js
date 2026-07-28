import Alpine from "alpinejs";

const API_BASE = "/api/v1";
const LOCK_CODES = new Set([
  "MUTATIONS_LOCKED",
  "DEADLINE_PASSED",
  "EVALUATIONS_LOCKED"
]);

class APIError extends Error {
  constructor(status, code, message) {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.code = code;
  }
}

function authRoute() {
  return window.location.pathname.startsWith("/jury") ? "/jury/login" : "/login";
}

function onAuthPage() {
  return ["/login", "/register", "/jury/login", "/jury/register"].includes(window.location.pathname);
}

function redirectForRole(role) {
  const routes = {
    USER: "/dashboard",
    JURY: "/jury/teams",
    ADMIN: "/admin"
  };
  window.location.assign(routes[role] || "/login");
}

function nestedError(payload, fallbackStatus) {
  const error = payload && typeof payload === "object" ? payload.error : null;
  if (
    error &&
    typeof error === "object" &&
    typeof error.code === "string" &&
    typeof error.message === "string"
  ) {
    return new APIError(fallbackStatus, error.code, error.message);
  }
  return new APIError(fallbackStatus, "UNEXPECTED_RESPONSE", "Сервер вернул некорректный ответ.");
}

async function apiFetch(path, options = {}) {
  const {
    redirectOnUnauthorized = true,
    body,
    headers = {},
    ...requestOptions
  } = options;
  const requestHeaders = { Accept: "application/json", ...headers };
  const init = {
    ...requestOptions,
    credentials: "include",
    headers: requestHeaders
  };
  if (body !== undefined) {
    requestHeaders["Content-Type"] = "application/json";
    init.body = JSON.stringify(body);
  }

  let response;
  try {
    response = await fetch(API_BASE + path, init);
  } catch {
    throw new APIError(0, "NETWORK_ERROR", "Не удалось связаться с сервером. Проверьте подключение.");
  }

  let payload = null;
  if (response.status !== 204) {
    try {
      payload = await response.json();
    } catch {
      if (!response.ok) {
        throw new APIError(response.status, "UNEXPECTED_RESPONSE", "Сервер вернул некорректный ответ.");
      }
    }
  }

  if (!response.ok) {
    const error = nestedError(payload, response.status);
    if (response.status === 401 && redirectOnUnauthorized && !onAuthPage()) {
      window.location.assign(authRoute());
    }
    if (response.status === 403 && LOCK_CODES.has(error.code)) {
      window.dispatchEvent(new CustomEvent("spcase:lock", { detail: error }));
    }
    throw error;
  }
  return payload;
}

function formatDate(value) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "long",
    timeStyle: "short"
  }).format(date);
}

function notify(message, type = "info") {
  Alpine.store("toast").show(message, type);
}

function errorMessage(error) {
  if (error instanceof APIError) return error.message;
  return "Произошла непредвиденная ошибка.";
}

async function logout() {
  try {
    await apiFetch("/auth/logout", { method: "POST" });
  } finally {
    window.location.assign("/");
  }
}

document.addEventListener("alpine:init", () => {
  Alpine.store("toast", {
    items: [],
    nextID: 1,
    show(message, type = "info") {
      const id = this.nextID++;
      this.items.push({ id, message, type });
      window.setTimeout(() => this.remove(id), 5000);
    },
    remove(id) {
      this.items = this.items.filter((toast) => toast.id !== id);
    }
  });

  Alpine.data("homePage", () => ({
    info: {},
    faq: [],
    schedule: [],
    countdown: "Загрузка…",
    openFAQ: null,
    loading: true,
    showTeamChoice: false,
    teamModal: false,
    teamName: "",
    inviteCode: "",
    busy: false,
    timer: null,
    formatDate,
    async init() {
      try {
        const [info, faq, schedule] = await Promise.all([
          apiFetch("/info", { redirectOnUnauthorized: false }),
          apiFetch("/faq", { redirectOnUnauthorized: false }),
          apiFetch("/schedule", { redirectOnUnauthorized: false })
        ]);
        this.info = info;
        this.faq = faq.faq;
        this.schedule = schedule.events;
        this.updateCountdown();
        this.timer = window.setInterval(() => this.updateCountdown(), 1000);
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.loading = false;
      }
      try {
        const profile = await apiFetch("/user/me", { redirectOnUnauthorized: false });
        this.showTeamChoice = profile.role === "USER" && profile.team_status === "NO_TEAM";
      } catch (error) {
        if (!(error instanceof APIError) || ![401, 403].includes(error.status)) {
          notify(errorMessage(error), "error");
        }
      }
    },
    updateCountdown() {
      const deadline = new Date(this.info.registration_deadline).getTime();
      const distance = deadline - Date.now();
      if (!Number.isFinite(deadline) || distance <= 0) {
        this.countdown = "Регистрация завершена";
        if (this.timer) window.clearInterval(this.timer);
        return;
      }
      const days = Math.floor(distance / 86400000);
      const hours = Math.floor((distance % 86400000) / 3600000);
      const minutes = Math.floor((distance % 3600000) / 60000);
      const seconds = Math.floor((distance % 60000) / 1000);
      this.countdown = `${days}д ${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
    },
    async createTeam() {
      if (!this.teamName) return;
      this.busy = true;
      try {
        await apiFetch("/team/create", { method: "POST", body: { name: this.teamName } });
        window.location.assign("/dashboard");
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.busy = false;
      }
    },
    async joinTeam() {
      if (this.inviteCode.length !== 8) return;
      this.busy = true;
      try {
        await apiFetch("/team/join", {
          method: "POST",
          body: { invite_code: this.inviteCode }
        });
        window.location.assign("/dashboard");
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.busy = false;
      }
    }
  }));

  Alpine.data("schedulePage", () => ({
    events: [],
    formatDate,
    async init() {
      try {
        const response = await apiFetch("/schedule", { redirectOnUnauthorized: false });
        this.events = response.events;
      } catch (error) {
        notify(errorMessage(error), "error");
      }
    }
  }));

  Alpine.data("noTeamPage", () => ({
    message: "Загружаем инструкцию…",
    telegramURL: "",
    async init() {
      try {
        const response = await apiFetch("/no-team", { redirectOnUnauthorized: false });
        this.message = response.message;
        this.telegramURL = response.telegram_chat_url;
      } catch (error) {
        notify(errorMessage(error), "error");
      }
    }
  }));

  Alpine.data("loginPage", () => ({
    email: "",
    password: "",
    errors: {},
    busy: false,
    async submit() {
      this.errors = {};
      if (!this.email || !this.email.includes("@")) this.errors.email = "Введите корректный email.";
      if (!this.password) this.errors.password = "Введите пароль.";
      if (Object.keys(this.errors).length > 0) return;

      this.busy = true;
      try {
        const response = await apiFetch("/auth/login", {
          method: "POST",
          redirectOnUnauthorized: false,
          body: { email: this.email, password: this.password }
        });
        notify("Вход выполнен.", "success");
        redirectForRole(response.role);
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.busy = false;
      }
    }
  }));

  Alpine.data("registerPage", () => ({
    form: {
      full_name: "",
      university: "",
      email: "",
      telegram: "",
      password: ""
    },
    formError: "",
    busy: false,
    async submit() {
      this.formError = "";
      if (Object.values(this.form).some((value) => !String(value).trim())) {
        this.formError = "Заполните все поля.";
        return;
      }
      if (!this.form.email.includes("@")) {
        this.formError = "Введите корректный email.";
        return;
      }
      if (this.form.password.length < 8) {
        this.formError = "Пароль должен содержать минимум 8 символов.";
        return;
      }
      this.busy = true;
      try {
        await apiFetch("/auth/register", {
          method: "POST",
          redirectOnUnauthorized: false,
          body: this.form
        });
        notify("Аккаунт создан.", "success");
        window.location.assign("/dashboard");
      } catch (error) {
        this.formError = errorMessage(error);
        notify(this.formError, "error");
      } finally {
        this.busy = false;
      }
    }
  }));

  Alpine.data("juryAuthPage", (mode) => ({
    form: {
      secret_key: "",
      full_name: "",
      email: "",
      password: ""
    },
    busy: false,
    async submit() {
      const isRegistration = mode === "register";
      const body = isRegistration
        ? this.form
        : { email: this.form.email, password: this.form.password };
      if (Object.values(body).some((value) => !String(value).trim())) {
        notify("Заполните все поля.", "error");
        return;
      }
      this.busy = true;
      try {
        await apiFetch(isRegistration ? "/jury/register" : "/jury/login", {
          method: "POST",
          redirectOnUnauthorized: false,
          body
        });
        notify(isRegistration ? "Аккаунт жюри создан." : "Вход выполнен.", "success");
        window.location.assign("/jury/teams");
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.busy = false;
      }
    }
  }));

  Alpine.data("dashboardPage", () => ({
    profile: {},
    team: { members: [] },
    info: {},
    loading: true,
    busy: false,
    teamName: "",
    inviteCode: "",
    copied: false,
    transferModal: false,
    disbandModal: false,
    newCaptainID: "",
    solutionURL: "",
    submissionError: "",
    editingSubmission: false,
    deadlinePassed: false,
    now: Date.now(),
    get isCaptain() {
      return this.profile.team_status === "CAPTAIN";
    },
    get nonCaptainMembers() {
      return (this.team.members || []).filter((member) => !member.is_captain);
    },
    get submissionLocked() {
      const deadline = new Date(this.info.submission_deadline).getTime();
      return this.deadlinePassed || (Number.isFinite(deadline) && this.now >= deadline);
    },
    async init() {
      window.addEventListener("spcase:lock", (event) => {
        if (event.detail.code === "MUTATIONS_LOCKED") this.team.mutations_locked = true;
        if (event.detail.code === "DEADLINE_PASSED") this.deadlinePassed = true;
      });
      window.setInterval(() => {
        this.now = Date.now();
      }, 30000);
      try {
        const [profile, info] = await Promise.all([
          apiFetch("/user/me"),
          apiFetch("/info", { redirectOnUnauthorized: false })
        ]);
        if (profile.role !== "USER") {
          redirectForRole(profile.role);
          return;
        }
        this.profile = profile;
        this.info = info;
        if (profile.team_status !== "NO_TEAM") await this.loadTeam();
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.loading = false;
      }
    },
    async loadTeam() {
      this.team = await apiFetch("/team/my");
      this.solutionURL = this.team.submission?.solution_url || "";
      this.editingSubmission = !this.team.submission;
    },
    badgeClass(status) {
      return {
        SEARCHING: "badge-searching",
        READY: "badge-ready",
        SUBMITTED: "badge-submitted"
      }[status] || "badge";
    },
    badgeLabel(status) {
      return {
        SEARCHING: "В поиске ⏳",
        READY: "Команда готова 🟢",
        SUBMITTED: "Решение сдано 🚀"
      }[status] || status;
    },
    async copyInvite() {
      try {
        await navigator.clipboard.writeText(this.team.invite_code);
        this.copied = true;
        window.setTimeout(() => {
          this.copied = false;
        }, 2000);
      } catch {
        notify("Не удалось скопировать код.", "error");
      }
    },
    async createTeam() {
      if (!this.teamName) return;
      await this.runBusy(async () => {
        await apiFetch("/team/create", { method: "POST", body: { name: this.teamName } });
        notify("Команда создана.", "success");
        window.location.reload();
      });
    },
    async joinTeam() {
      if (this.inviteCode.length !== 8) return;
      await this.runBusy(async () => {
        await apiFetch("/team/join", {
          method: "POST",
          body: { invite_code: this.inviteCode }
        });
        notify("Вы вступили в команду.", "success");
        window.location.reload();
      });
    },
    async leaveTeam() {
      const losesSubmission = Boolean(this.team.submission) && this.team.members.length === 2;
      const warning = losesSubmission
        ? "После выхода останется один участник, поэтому ссылка на решение будет удалена. Продолжить?"
        : "Выйти из команды?";
      if (!window.confirm(warning)) return;
      await this.runBusy(async () => {
        await apiFetch("/team/leave", { method: "POST" });
        notify("Вы вышли из команды.", "success");
        window.location.reload();
      });
    },
    async kickMember(member) {
      const losesSubmission = Boolean(this.team.submission) && this.team.members.length === 2;
      const warning = losesSubmission
        ? `Исключить ${member.full_name}? В команде останется один участник, ссылка на решение будет удалена.`
        : `Исключить ${member.full_name} из команды?`;
      if (!window.confirm(warning)) return;
      await this.runBusy(async () => {
        await apiFetch("/team/kick", {
          method: "POST",
          body: { user_id: member.id }
        });
        notify("Участник исключён.", "success");
        await this.loadTeam();
      });
    },
    async transferOwnership() {
      if (!this.newCaptainID) return;
      await this.runBusy(async () => {
        await apiFetch("/team/transfer-ownership", {
          method: "POST",
          body: { new_captain_id: this.newCaptainID }
        });
        notify("Права капитана переданы.", "success");
        this.transferModal = false;
        window.location.reload();
      });
    },
    async disbandTeam() {
      await this.runBusy(async () => {
        await apiFetch("/team/disband", { method: "DELETE" });
        notify("Команда расформирована.", "success");
        window.location.reload();
      });
    },
    async submitSolution() {
      this.submissionError = "";
      if (!/^https?:\/\/.+/i.test(this.solutionURL)) {
        this.submissionError = "Введите полный URL, начинающийся с http:// или https://.";
        return;
      }
      if (this.team.members.length < 2) {
        this.submissionError = "Для сдачи нужны минимум два участника.";
        return;
      }
      await this.runBusy(async () => {
        const submission = await apiFetch("/team/submit", {
          method: "POST",
          body: { solution_url: this.solutionURL }
        });
        this.team.submission = submission;
        this.team.status_badge = "SUBMITTED";
        this.editingSubmission = false;
        notify("Ссылка на решение сохранена.", "success");
      });
    },
    async runBusy(action) {
      this.busy = true;
      try {
        await action();
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.busy = false;
      }
    },
    logout
  }));

  Alpine.data("juryWorkspace", () => ({
    teams: [],
    scores: {},
    loading: true,
    hideEvaluated: false,
    evaluationsLocked: false,
    savingTeamID: "",
    criteria: [
      { id: 1, label: "Критерий 1" },
      { id: 2, label: "Критерий 2" },
      { id: 3, label: "Критерий 3" },
      { id: 4, label: "Критерий 4" },
      { id: 5, label: "Критерий 5" },
      { id: 6, label: "Критерий 6" }
    ],
    get visibleTeams() {
      return this.hideEvaluated
        ? this.teams.filter((team) => !team.is_evaluated_by_me)
        : this.teams;
    },
    async init() {
      window.addEventListener("spcase:lock", (event) => {
        if (event.detail.code === "EVALUATIONS_LOCKED") this.evaluationsLocked = true;
      });
      try {
        const [teamResponse, evaluationResponse] = await Promise.all([
          apiFetch("/jury/teams"),
          apiFetch("/jury/evaluations")
        ]);
        this.teams = teamResponse.teams;
        this.evaluationsLocked = teamResponse.evaluations_locked;
        for (const team of this.teams) {
          this.scores[team.team_id] = {};
          for (const criterion of this.criteria) {
            this.scores[team.team_id][criterion.id] = 0;
          }
        }
        for (const evaluation of evaluationResponse.evaluations) {
          if (this.scores[evaluation.team_id]) {
            this.scores[evaluation.team_id][evaluation.criterion_id] = evaluation.score;
          }
        }
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.loading = false;
      }
    },
    totalFor(teamID) {
      return Object.values(this.scores[teamID] || {}).reduce(
        (total, score) => total + (Number(score) || 0),
        0
      );
    },
    validScores(teamID) {
      const values = Object.values(this.scores[teamID] || {});
      return (
        values.length === 6 &&
        values.every((score) => Number.isInteger(Number(score)) && Number(score) >= 0 && Number(score) <= 10)
      );
    },
    async save(team) {
      if (!this.validScores(team.team_id)) {
        notify("Каждый из шести критериев должен быть целым числом от 0 до 10.", "error");
        return;
      }
      this.savingTeamID = team.team_id;
      try {
        const scores = this.criteria.map((criterion) => ({
          criterion_id: criterion.id,
          score: Number(this.scores[team.team_id][criterion.id])
        }));
        await apiFetch("/jury/evaluations", {
          method: "POST",
          body: { team_id: team.team_id, scores }
        });
        team.is_evaluated_by_me = true;
        notify("Все шесть оценок сохранены атомарно.", "success");
      } catch (error) {
        if (error instanceof APIError && error.code === "EVALUATIONS_LOCKED") {
          this.evaluationsLocked = true;
        }
        notify(errorMessage(error), "error");
      } finally {
        this.savingTeamID = "";
      }
    },
    logout
  }));

  Alpine.data("adminPage", () => ({
    stats: {},
    downloading: false,
    stateBusy: false,
    async init() {
      try {
        this.stats = await apiFetch("/admin/stats");
      } catch (error) {
        notify(errorMessage(error), "error");
      }
    },
    async downloadExcel() {
      this.downloading = true;
      try {
        const response = await fetch(API_BASE + "/admin/export/excel", {
          credentials: "include",
          headers: { Accept: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }
        });
        if (response.status === 401) {
          window.location.assign("/login");
          return;
        }
        if (!response.ok) {
          let payload = null;
          try {
            payload = await response.json();
          } catch {
            throw new APIError(response.status, "UNEXPECTED_RESPONSE", "Не удалось сформировать отчёт.");
          }
          throw nestedError(payload, response.status);
        }
        const blob = await response.blob();
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = "hackathon_results.xlsx";
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(url);
        notify("Excel-отчёт скачан.", "success");
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.downloading = false;
      }
    },
    async setEvaluationState(closed) {
      this.stateBusy = true;
      try {
        await apiFetch(`/admin/evaluations/${closed ? "close" : "open"}`, { method: "POST" });
        this.stats.evaluations_closed = closed;
        notify(closed ? "Оценивание закрыто." : "Оценивание открыто.", "success");
      } catch (error) {
        notify(errorMessage(error), "error");
      } finally {
        this.stateBusy = false;
      }
    },
    logout
  }));
});

window.Alpine = Alpine;
Alpine.start();
