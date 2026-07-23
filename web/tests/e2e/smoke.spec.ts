import { expect, test } from '@playwright/test';

test('desktop demo table renders commercial table zones', async ({ page }) => {
  await page.goto('/?demo=table');
  await expect(page.getByText('斗地主').first()).toBeVisible();
  await expect(page.getByRole('button', { name: '出牌' })).toBeVisible();
  await expect(page.getByLabel('手牌')).toBeVisible();
  await expect(page.getByLabel('底牌')).toBeVisible();
});

test('bidding demo hides bottom cards and only shows bidding actions', async ({ page }) => {
  await page.goto('/?demo=bidding');
  await expect(page.getByLabel('底牌')).toContainText('底牌未揭');
  await expect(page.getByRole('button', { name: '叫地主' })).toBeVisible();
  await expect(page.getByRole('button', { name: '不叫' })).toBeVisible();
  await expect(page.getByRole('button', { name: '出牌' })).toHaveCount(0);
});

test('mobile demo table keeps controls and drawers visible', async ({ page }) => {
  await page.goto('/?demo=table');
  await expect(page.getByRole('button', { name: '提示' })).toBeVisible();
  await expect(page.getByRole('button', { name: '聊天' })).toBeVisible();
});
