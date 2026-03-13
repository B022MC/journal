import { spawn } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";
import process from "node:process";

const host = "127.0.0.1";
const port = 3100;
const baseUrl = `http://${host}:${port}`;
const expectedSearchReleaseEngine =
  process.env.JOURNAL_SEARCH_RELEASE_ENGINE?.trim() ||
  process.env.NEXT_PUBLIC_JOURNAL_SEARCH_RELEASE_ENGINE?.trim() ||
  "fulltext";

const server = spawn(
  process.platform === "win32" ? "npx.cmd" : "npx",
  ["next", "start", "--hostname", host, "--port", String(port)],
  {
    cwd: new URL("..", import.meta.url),
    env: { ...process.env, NODE_ENV: "production" },
    stdio: ["ignore", "pipe", "pipe"],
  },
);

let stdout = "";
let stderr = "";
server.stdout.on("data", (chunk) => {
  stdout += chunk.toString();
});
server.stderr.on("data", (chunk) => {
  stderr += chunk.toString();
});

async function waitForReady() {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    if (server.exitCode !== null) {
      throw new Error(`next start exited early with code ${server.exitCode}\n${stdout}\n${stderr}`);
    }
    try {
      const response = await fetch(`${baseUrl}/`);
      if (response.ok) {
        return;
      }
    } catch {}
    await delay(500);
  }
  throw new Error(`timed out waiting for ${baseUrl}\n${stdout}\n${stderr}`);
}

async function assertPage(path, expectations = []) {
  const response = await fetch(`${baseUrl}${path}`);
  if (!response.ok) {
    throw new Error(`expected ${path} to return 200, got ${response.status}`);
  }
  if (expectations.length === 0) {
    return;
  }
  const body = await response.text();
  for (const text of expectations) {
    if (!body.includes(text)) {
      throw new Error(`expected ${path} to contain ${JSON.stringify(text)}\n${body.slice(0, 800)}`);
    }
  }
}

function stopServer() {
  if (server.exitCode === null) {
    server.kill("SIGTERM");
  }
}

process.on("exit", stopServer);
process.on("SIGINT", () => {
  stopServer();
  process.exitCode = 1;
});
process.on("SIGTERM", () => {
  stopServer();
  process.exitCode = 1;
});

try {
  await waitForReady();
  await assertPage("/", ["S.H.I.T Journal"]);
  await assertPage("/papers", ["Release default engine:", expectedSearchReleaseEngine]);
  await assertPage("/papers?query=%E4%BA%BA%E5%B7%A5%E6%99%BA%E8%83%BD%E8%AE%BA%E6%96%87&sort=relevance&page=2&engine=auto", [
    "Release default engine:",
    expectedSearchReleaseEngine,
  ]);
  await assertPage("/papers?query=%E4%BA%BA%E5%B7%A5%E6%99%BA%E8%83%BD%E8%AE%BA%E6%96%87&sort=relevance&engine=hybrid&shadow_compare=true");
} finally {
  stopServer();
  await delay(500);
}
