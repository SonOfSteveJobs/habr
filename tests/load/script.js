import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Counter } from "k6/metrics";
import encoding from "k6/encoding";

// ─── CONFIG ─────────────────────────────────────────────────────────────────────

const cfg = JSON.parse(open("./config.json"));

const BASE_URL = cfg.baseUrl;
const RPS = cfg.rps;
const DURATION = cfg.duration;
const READ_RATIO = cfg.readRatio;
const SETUP_USERS = cfg.setupUsers;
const MAX_VUS = cfg.maxVUs;
const TH = cfg.thresholds;

const READ_RPS = Math.max(1, Math.round((RPS * READ_RATIO) / 100));
const WRITE_RPS = Math.max(1, RPS - READ_RPS);
const AUTH_RPS = 10;

const PRE_VUS = Math.ceil(MAX_VUS * 0.5);

// ─── CUSTOM METRICS ─────────────────────────────────────────────────────────────

const articleListDuration = new Trend("article_list_duration", true);
const articleGetDuration = new Trend("article_get_duration", true);
const articleCreateDuration = new Trend("article_create_duration", true);
const articleUpdateDuration = new Trend("article_update_duration", true);
const articleDeleteDuration = new Trend("article_delete_duration", true);

const authRegisterDuration = new Trend("auth_register_duration", true);
const authLoginDuration = new Trend("auth_login_duration", true);
const authVerifyDuration = new Trend("auth_verify_duration", true);
const authRefreshDuration = new Trend("auth_refresh_duration", true);
const authLogoutDuration = new Trend("auth_logout_duration", true);

const authFlowErrors = new Counter("auth_flow_errors");
const writeFlowErrors = new Counter("write_flow_errors");

// ─── SCENARIOS ──────────────────────────────────────────────────────────────────

export const options = {
  scenarios: {
    reads: {
      executor: "constant-arrival-rate",
      rate: READ_RPS,
      timeUnit: "1s",
      duration: DURATION,
      preAllocatedVUs: PRE_VUS,
      maxVUs: MAX_VUS,
      exec: "readScenario",
    },
    writes: {
      executor: "constant-arrival-rate",
      rate: WRITE_RPS,
      timeUnit: "1s",
      duration: DURATION,
      preAllocatedVUs: PRE_VUS,
      maxVUs: MAX_VUS,
      exec: "writeScenario",
    },
    auth_flow: {
      executor: "constant-arrival-rate",
      rate: AUTH_RPS,
      timeUnit: "1s",
      duration: DURATION,
      preAllocatedVUs: 20,
      maxVUs: 50,
      exec: "authScenario",
    },
  },
  thresholds: {
    http_req_duration: [
      `p(95)<${TH.httpReqDurationP95}`,
      `p(99)<${TH.httpReqDurationP99}`,
    ],
    http_req_failed: [`rate<${TH.httpReqFailedRate}`],
    article_list_duration: [`p(95)<${TH.endpointP95}`],
    article_get_duration: [`p(95)<${TH.endpointP95}`],
    article_create_duration: [`p(95)<${TH.endpointP95}`],
    article_update_duration: [`p(95)<${TH.endpointP95}`],
    article_delete_duration: [`p(95)<${TH.endpointP95}`],
    auth_register_duration: [`p(95)<${TH.endpointP95}`],
    auth_login_duration: [`p(95)<${TH.endpointP95}`],
    auth_verify_duration: [`p(95)<${TH.endpointP95}`],
    auth_refresh_duration: [`p(95)<${TH.endpointP95}`],
    auth_logout_duration: [`p(95)<${TH.endpointP95}`],
  },
};

// ─── HELPERS ────────────────────────────────────────────────────────────────────

function jsonHeaders(token) {
  const h = { "Content-Type": "application/json" };
  if (token) h["Authorization"] = `Bearer ${token}`;
  return h;
}

function decodeJwtPayload(token) {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  const payload = encoding.b64decode(parts[1], "rawurl", "s");
  return JSON.parse(payload);
}

function getUserId(accessToken) {
  const claims = decodeJwtPayload(accessToken);
  return claims ? claims.sub : null;
}

function isTokenExpiringSoon(accessToken, bufferSec) {
  const claims = decodeJwtPayload(accessToken);
  if (!claims || !claims.exp) return true;
  return claims.exp - Math.floor(Date.now() / 1000) < (bufferSec || 60);
}

function refreshIfNeeded(user) {
  if (!isTokenExpiringSoon(user.access_token, 60)) return;

  const res = http.post(
    `${BASE_URL}/api/v1/auth/refresh`,
    JSON.stringify({
      user_id: user.user_id,
      refresh_token: user.refresh_token,
    }),
    { headers: jsonHeaders(), tags: { name: "auth_refresh" } }
  );

  if (res.status === 200) {
    const body = res.json();
    user.access_token = body.access_token;
    user.refresh_token = body.refresh_token;
  }
}

function randomString(len) {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  let s = "";
  for (let i = 0; i < len; i++) {
    s += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return s;
}

// ─── SETUP ──────────────────────────────────────────────────────────────────────

export function setup() {
  const users = [];

  for (let i = 0; i < SETUP_USERS; i++) {
    const email = `loadtest_${randomString(8)}_${i}@test.local`;
    const password = `Pass${randomString(10)}`;

    const regRes = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ email, password }),
      { headers: jsonHeaders(), tags: { name: "setup_register" } }
    );

    if (regRes.status !== 201) {
      console.warn(`setup: register failed for ${email}: ${regRes.status}`);
      continue;
    }

    const loginRes = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email, password }),
      { headers: jsonHeaders(), tags: { name: "setup_login" } }
    );

    if (loginRes.status !== 200) {
      console.warn(`setup: login failed for ${email}: ${loginRes.status}`);
      continue;
    }

    const tokens = loginRes.json();
    const userId = getUserId(tokens.access_token);

    users.push({
      email,
      password,
      user_id: userId,
      access_token: tokens.access_token,
      refresh_token: tokens.refresh_token,
    });
  }

  if (users.length === 0) {
    throw new Error("setup: no users created, cannot run test");
  }

  console.log(`setup: created ${users.length} users`);

  // Seed articles for read scenarios
  const seedArticleIds = [];
  const seedUser = users[0];

  for (let i = 0; i < 10; i++) {
    const res = http.post(
      `${BASE_URL}/api/v1/articles`,
      JSON.stringify({
        title: `Seed Article ${i + 1}`,
        content: `This is seed article content number ${i + 1}. ${randomString(200)}`,
      }),
      {
        headers: jsonHeaders(seedUser.access_token),
        tags: { name: "setup_create_article" },
      }
    );

    if (res.status === 201) {
      seedArticleIds.push(res.json().id);
    } else {
      console.warn(`setup: seed article ${i} failed: ${res.status}`);
    }
  }

  console.log(`setup: created ${seedArticleIds.length} seed articles`);

  return { users, seedArticleIds };
}

// ─── READ SCENARIO ──────────────────────────────────────────────────────────────

export function readScenario(data) {
  // List articles
  const listRes = http.get(`${BASE_URL}/api/v1/articles?limit=20`, {
    headers: jsonHeaders(),
    tags: { name: "article_list" },
  });
  articleListDuration.add(listRes.timings.duration);
  check(listRes, { "list articles 200": (r) => r.status === 200 });

  // Get article by id
  if (data.seedArticleIds.length > 0) {
    const articleId =
      data.seedArticleIds[
        Math.floor(Math.random() * data.seedArticleIds.length)
      ];
    const getRes = http.get(`${BASE_URL}/api/v1/articles/${articleId}`, {
      headers: jsonHeaders(),
      tags: { name: "article_get" },
    });
    articleGetDuration.add(getRes.timings.duration);
    check(getRes, { "get article 200": (r) => r.status === 200 });
  }
}

// ─── WRITE SCENARIO ─────────────────────────────────────────────────────────────

export function writeScenario(data) {
  const user = data.users[__VU % data.users.length];
  refreshIfNeeded(user);

  // Create
  const createRes = http.post(
    `${BASE_URL}/api/v1/articles`,
    JSON.stringify({
      title: `Load Test Article ${randomString(6)}`,
      content: `Generated content for load testing. ${randomString(100)}`,
    }),
    {
      headers: jsonHeaders(user.access_token),
      tags: { name: "article_create" },
    }
  );
  articleCreateDuration.add(createRes.timings.duration);

  const created = check(createRes, {
    "create article 201": (r) => r.status === 201,
  });

  if (!created) {
    writeFlowErrors.add(1);
    return;
  }

  const articleId = createRes.json().id;

  // Update
  const updateRes = http.patch(
    `${BASE_URL}/api/v1/articles/${articleId}`,
    JSON.stringify({
      title: `Updated ${randomString(6)}`,
      content: `Updated content. ${randomString(100)}`,
    }),
    {
      headers: jsonHeaders(user.access_token),
      tags: { name: "article_update" },
    }
  );
  articleUpdateDuration.add(updateRes.timings.duration);
  check(updateRes, { "update article 200": (r) => r.status === 200 });

  // Delete (self-cleaning)
  const deleteRes = http.del(
    `${BASE_URL}/api/v1/articles/${articleId}`,
    null,
    {
      headers: jsonHeaders(user.access_token),
      tags: { name: "article_delete" },
    }
  );
  articleDeleteDuration.add(deleteRes.timings.duration);
  check(deleteRes, { "delete article 200": (r) => r.status === 200 });
}

// ─── AUTH SCENARIO ──────────────────────────────────────────────────────────────

export function authScenario() {
  const email = `authflow_${randomString(10)}_${Date.now()}@test.local`;
  const password = `Pass${randomString(10)}`;

  // Register
  const regRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({ email, password }),
    { headers: jsonHeaders(), tags: { name: "auth_register" } }
  );
  authRegisterDuration.add(regRes.timings.duration);
  const registered = check(regRes, {
    "register 201": (r) => r.status === 201,
  });

  if (!registered) {
    authFlowErrors.add(1);
    return;
  }

  // Login
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: jsonHeaders(), tags: { name: "auth_login" } }
  );
  authLoginDuration.add(loginRes.timings.duration);
  const loggedIn = check(loginRes, { "login 200": (r) => r.status === 200 });

  if (!loggedIn) {
    authFlowErrors.add(1);
    return;
  }

  const tokens = loginRes.json();
  const userId = getUserId(tokens.access_token);

  // Verify email — always 400 (code delivered via Kafka, not available in k6)
  const verifyRes = http.post(
    `${BASE_URL}/api/v1/auth/verify-email`,
    JSON.stringify({ user_id: userId, code: "000000" }),
    { headers: jsonHeaders(), tags: { name: "auth_verify" } }
  );
  authVerifyDuration.add(verifyRes.timings.duration);
  check(verifyRes, {
    "verify-email 400 (expected)": (r) => r.status === 400,
  });

  // Refresh
  const refreshRes = http.post(
    `${BASE_URL}/api/v1/auth/refresh`,
    JSON.stringify({
      user_id: userId,
      refresh_token: tokens.refresh_token,
    }),
    { headers: jsonHeaders(), tags: { name: "auth_refresh" } }
  );
  authRefreshDuration.add(refreshRes.timings.duration);
  check(refreshRes, { "refresh 200": (r) => r.status === 200 });

  const newAccessToken =
    refreshRes.status === 200
      ? refreshRes.json().access_token
      : tokens.access_token;

  // Logout
  const logoutRes = http.post(`${BASE_URL}/api/v1/auth/logout`, null, {
    headers: jsonHeaders(newAccessToken),
    tags: { name: "auth_logout" },
  });
  authLogoutDuration.add(logoutRes.timings.duration);
  check(logoutRes, { "logout 200": (r) => r.status === 200 });
}
