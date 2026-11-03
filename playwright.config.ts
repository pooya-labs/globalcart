import { defineConfig, devices } from "@playwright/test";

const basePath = process.env.GITHUB_ACTIONS === "true" ? "/globalcart" : "";
const baseURL = "http://127.0.0.1:4173" + (basePath ? basePath + "/" : "/");

export default defineConfig({
  testDir: "./e2e",
  testMatch: "**/*.spec.ts",
  fullyParallel: false,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 30_000,
  expect: { timeout: 7_000 },
  reporter: [
    ["list"],
    ["html", { outputFolder: "playwright-report", open: "never" }]
  ],
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    actionTimeout: 10_000
  },
  projects: [
    { name: "chromium-desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "chromium-mobile", use: { ...devices["Pixel 7"] } },
    {
      name: "chromium-tablet",
      testMatch: ["**/accessibility.spec.ts", "**/performance.spec.ts"],
      use: { ...devices["iPad (gen 7)"], browserName: "chromium" }
    },
    {
      name: "firefox-accessibility",
      testMatch: "**/accessibility.spec.ts",
      use: { browserName: "firefox", viewport: { width: 1280, height: 900 } }
    },
    {
      name: "webkit-accessibility",
      testMatch: "**/accessibility.spec.ts",
      use: { browserName: "webkit", viewport: { width: 1280, height: 900 } }
    }
  ],
  webServer: {
    command: "node scripts/serve-static.mjs",
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 15_000,
    env: {
      PORT: "4173",
      E2E_BASE_PATH: basePath
    }
  }
});
