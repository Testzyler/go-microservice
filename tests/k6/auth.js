import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const PASSWORD = 'S3cretPass!';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<400'],
  },
};

export default function () {
  const email = `k6+${__VU}-${Date.now()}@example.com`;

  const register = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({ email, password: PASSWORD }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(register, {
    'register status 201': (r) => r.status === 201,
  });

  const login = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({ email, password: PASSWORD }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(login, {
    'login status 200': (r) => r.status === 200,
    'login token present': (r) => !!(r.json('access_token')),
  });

  sleep(1);
}
