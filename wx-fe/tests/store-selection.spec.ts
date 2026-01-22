import { test, expect } from '@playwright/test';

test('store selection updates current store display', async ({ page }) => {
  await page.goto('/#/pages/stores/index');

  const targetStore = '隽也YUYE茶馆 · 城南店';
  await page.locator('[data-testid="store-select-btn"][data-store-name="隽也YUYE茶馆 · 城南店"]').click();
  await expect(page).toHaveURL(/#\/pages\/menu\/index/);

  const stored = await page.evaluate(() => localStorage.getItem('current_store_name'));
  let normalized = stored || '';
  if (normalized.trim().startsWith('{')) {
    try {
      const parsed = JSON.parse(normalized);
      if (parsed && typeof parsed.data === 'string') normalized = parsed.data;
    } catch (_) {}
  }
  expect(normalized).toBe(targetStore);

  await page.goto('/#/pages/home/index');
  await expect(page.locator('.header-store-name')).toHaveText(targetStore);
});

test('store selection navigates to menu page', async ({ page }) => {
  await page.goto('/#/pages/stores/index');
  await page.locator('[data-testid="store-select-btn"][data-store-name="隽也YUYE茶馆 · 城南店"]').click();
  await expect(page).toHaveURL(/#\/pages\/menu\/index/);
  await expect(page.getByText('商家推荐')).toBeVisible();
});
