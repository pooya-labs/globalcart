import { test, expect } from "@playwright/test";

// Portfolio evidence is captured from the production static export that the
// browser suite serves, not from a hand-built mock or a separate design file.
test("capture the Persian case study for portfolio evidence", async ({ page }, testInfo) => {
  await page.goto("./case-study/");
  const main = page.locator("main.case-study");
  await expect(main).toHaveAttribute("lang", "fa");
  await expect(main).toHaveAttribute("dir", "rtl");
  await expect(main.getByRole("heading", { level: 1 })).toContainText("یک بازار جهانی");
  await expect(main.getByRole("heading", { name: "مدل داده و روابط پایگاه‌داده" })).toBeVisible();
  await expect(main.getByRole("link", { name: /مشاهده نمودار کامل ERD/ })).toBeVisible();

  await page.evaluate(() => document.fonts.ready);
  await page.screenshot({
    path: testInfo.outputPath(`case-study-${testInfo.project.name}.png`),
    fullPage: true,
    animations: "disabled"
  });
});
