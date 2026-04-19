import { defineConfig } from "@playwright/test";
import { fileURLToPath } from "node:url";
import path from "path";

const currentDir = path.dirname(fileURLToPath(import.meta.url));
const backendDir = path.resolve(currentDir, "..", "backend");

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false,
  workers: 1,
  timeout: 120_000,
  expect: {
    timeout: 15_000,
  },
  globalSetup: "./tests/e2e/global-setup.ts",
  use: {
    baseURL: "http://127.0.0.1:4173",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  webServer: [
    {
      command: "go run ./cmd/api",
      cwd: backendDir,
      env: {
        ...process.env,
        HTTP_PORT: "18080",
        FRONTEND_ORIGIN: "http://127.0.0.1:4173",
      },
      url: "http://127.0.0.1:18080/healthz",
      reuseExistingServer: false,
      timeout: 120_000,
    },
    {
      command: "npm run dev -- --host 127.0.0.1 --port 4173",
      cwd: currentDir,
      env: {
        ...process.env,
        VITE_API_BASE_URL: "http://127.0.0.1:18080/api/v1",
      },
      url: "http://127.0.0.1:4173",
      reuseExistingServer: false,
      timeout: 120_000,
    },
  ],
});
