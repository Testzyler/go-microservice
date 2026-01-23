import { test, expect } from '@playwright/test';

test.describe('auth via gateway', () => {
  test('register and login', async ({ request, baseURL }) => {
    const email = `user+${Date.now()}@example.com`;
    const password = 'S3cretPass!';

    const register = await request.post(`${baseURL}/api/v1/auth/register`, {
      data: { email, password },
    });
    expect(register.status(), 'register status').toBe(201);
    const regBody = await register.json();
    expect(regBody.access_token, 'register token').toBeTruthy();

    const login = await request.post(`${baseURL}/api/v1/auth/login`, {
      data: { email, password },
    });
    expect(login.status(), 'login status').toBe(200);
    const loginBody = await login.json();
    expect(loginBody.access_token, 'login token').toBeTruthy();
    expect(loginBody.user_id, 'user id').toBe((regBody as any).user_id);
  });

  test('get current user with /me', async ({ request, baseURL }) => {
    const email = `user+${Date.now()}@example.com`;
    const password = 'S3cretPass!';

    // Register a new user
    const register = await request.post(`${baseURL}/api/v1/auth/register`, {
      data: { email, password },
    });
    expect(register.status()).toBe(201);
    const regBody = await register.json();
    const accessToken = regBody.access_token;
    const userId = regBody.user_id;

    // Call /me endpoint with token
    const me = await request.get(`${baseURL}/api/v1/auth/me`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
    expect(me.status(), '/me status').toBe(200);
    const meBody = await me.json();
    expect(meBody.user_id, '/me user_id').toBe(userId);
    expect(meBody.roles, '/me roles').toEqual(expect.arrayContaining(['user']));
    expect(meBody.permissions, '/me permissions').toBeTruthy();
  });

  test('/me without token returns 401', async ({ request, baseURL }) => {
    const me = await request.get(`${baseURL}/api/v1/auth/me`);
    expect(me.status()).toBe(401);
  });
});
