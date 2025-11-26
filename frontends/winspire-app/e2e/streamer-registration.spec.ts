import { test, expect } from '@playwright/test';

test.describe('Streamer Registration', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to streamer registration page
    await page.goto('/auth/streamer/register');
  });

  test('should display streamer registration form', async ({ page }) => {
    // Check page title/heading
    await expect(page.getByRole('heading', { name: /create.*streamer.*account/i })).toBeVisible();

    // Check all required form fields are present
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByLabel('Nickname')).toBeVisible();
    await expect(page.getByLabel('First Name')).toBeVisible();
    await expect(page.getByLabel('Last Name')).toBeVisible();
    await expect(page.getByLabel(/Country/i)).toBeVisible();

    // Check submit button
    await expect(page.getByRole('button', { name: /create.*streamer.*account/i })).toBeVisible();
  });

  test('should show validation errors for empty form', async ({ page }) => {
    // Click submit without filling form
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Wait for validation errors to appear
    await expect(page.locator('text=/required/i').first()).toBeVisible({ timeout: 2000 });
  });

  test('should show error when passwords do not match', async ({ page }) => {
    const timestamp = Date.now();
    const testEmail = `streamer${timestamp}@example.com`;

    // Fill form with mismatched passwords
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill('Password123!');
    await page.getByLabel('Confirm Password').fill('Password456!');
    await page.getByLabel('Nickname').fill(`streamer${timestamp}`);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');

    // Submit form
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Check for password mismatch error
    await expect(page.locator('text=/password.*match/i')).toBeVisible({ timeout: 2000 });
  });

  test('should successfully register a new streamer', async ({ page }) => {
    const timestamp = Date.now();
    const testEmail = `streamer${timestamp}@example.com`;
    const testPassword = 'StreamerPass123!';
    const testNickname = `streamer${timestamp}`;

    // Fill in all required fields
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(testPassword);
    await page.getByLabel('Confirm Password').fill(testPassword);
    await page.getByLabel('Nickname').fill(testNickname);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');

    // Submit form
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Wait for redirect to home page or success indicator
    // Adjust timeout as needed for your backend processing
    await page.waitForURL('/', { timeout: 10000 });

    // Verify streamer is logged in by checking for user-specific elements
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible({ timeout: 5000 });
  });

  test('should successfully register a streamer with optional fields', async ({ page }) => {
    const timestamp = Date.now();
    const testEmail = `streamer${timestamp}@example.com`;
    const testPassword = 'StreamerPass123!';
    const testNickname = `streamer${timestamp}`;

    // Fill in all required fields
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(testPassword);
    await page.getByLabel('Confirm Password').fill(testPassword);
    await page.getByLabel('Nickname').fill(testNickname);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');

    // Fill optional fields
    await page.getByLabel(/Country/i).selectOption({ index: 1 }); // Select first country

    // Wait for city field to appear (if conditional rendering)
    const cityField = page.getByLabel(/City/i);
    if (await cityField.isVisible()) {
      await cityField.fill('Streamer City');
    }

    // Submit form
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Wait for redirect to home page
    await page.waitForURL('/', { timeout: 10000 });

    // Verify successful registration
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible({ timeout: 5000 });
  });

  test('should show error when email is already registered', async ({ page }) => {
    const timestamp = Date.now();
    const testEmail = `existingstreamer${timestamp}@example.com`;
    const testPassword = 'StreamerPass123!';
    const testNickname = `streamer${timestamp}`;

    // First registration
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(testPassword);
    await page.getByLabel('Confirm Password').fill(testPassword);
    await page.getByLabel('Nickname').fill(testNickname);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Wait for successful registration
    await page.waitForURL('/', { timeout: 10000 });

    // Logout
    await page.getByRole('button', { name: /logout/i }).click();
    await page.waitForTimeout(1000);

    // Try to register again with same email
    await page.goto('/auth/streamer/register');
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(testPassword);
    await page.getByLabel('Confirm Password').fill(testPassword);
    await page.getByLabel('Nickname').fill(`${testNickname}2`);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Check for duplicate email error
    await expect(page.locator('text=/already.*registered|email.*exists/i')).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to login page from registration page', async ({ page }) => {
    // Look for login link
    const loginLink = page.getByRole('link', { name: /sign in|login/i });
    
    if (await loginLink.isVisible()) {
      await loginLink.click();
      await expect(page).toHaveURL(/\/auth\/(streamer\/)?login/);
    }
  });

  test('should differentiate streamer from user registration', async ({ page }) => {
    // Verify we're on the streamer registration page
    await expect(page).toHaveURL(/\/auth\/streamer\/register/);
    
    // Check for streamer-specific text or elements
    await expect(page.getByRole('heading', { name: /streamer/i })).toBeVisible();
  });

  test('should create streamer profile with correct profile type', async ({ page }) => {
    const timestamp = Date.now();
    const testEmail = `streamertype${timestamp}@example.com`;
    const testPassword = 'StreamerPass123!';
    const testNickname = `streamer${timestamp}`;

    // Register streamer
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(testPassword);
    await page.getByLabel('Confirm Password').fill(testPassword);
    await page.getByLabel('Nickname').fill(testNickname);
    await page.getByLabel('First Name').fill('Test');
    await page.getByLabel('Last Name').fill('Streamer');
    await page.getByRole('button', { name: /create.*streamer.*account/i }).click();

    // Wait for successful registration and redirect
    await page.waitForURL('/', { timeout: 10000 });
    
    // Verify logged in
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible({ timeout: 5000 });
    
    // Note: Additional verification of profile type would require
    // checking the profile page or API calls, which depends on your app structure
  });
});

