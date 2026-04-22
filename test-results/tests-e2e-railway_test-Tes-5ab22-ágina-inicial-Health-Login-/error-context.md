# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests\e2e\railway_test.spec.js >> Testes E2E - Recrutamento IA >> Acesso à página inicial (Health/Login)
- Location: tests\e2e\railway_test.spec.js:8:7

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 200
Received: 502
```

# Page snapshot

```yaml
- main [ref=e2]:
  - generic [ref=e3]:
    - generic [ref=e4]:
      - img [ref=e5]
      - heading "Application failed to respond" [level=1] [ref=e8]
    - generic [ref=e9]:
      - paragraph [ref=e10]: This error appears to be caused by the application.
      - paragraph [ref=e11]:
        - text: If this is your project, check out your
        - link "deploy logs" [ref=e12] [cursor=pointer]:
          - /url: https://docs.railway.com/guides/logs
        - text: to see what went wrong. Refer to our
        - link "docs on Fixing Common Errors" [ref=e13] [cursor=pointer]:
          - /url: https://docs.railway.com/guides/fixing-common-errors
        - text: for help, or reach out over our
        - link "Help Station" [ref=e14] [cursor=pointer]:
          - /url: https://station.railway.com
        - text: .
      - paragraph [ref=e15]: If you are a visitor, please contact the application owner or try again later.
      - paragraph [ref=e17]:
        - text: "Request ID:"
        - text: zSAoQ4ONQzSiP8WZ9I3ezw
      - link "Go to Railway" [ref=e19] [cursor=pointer]:
        - /url: https://railway.com
```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | // Substitua pela URL gerada do Railway
  4  | const BASE_URL = 'https://app-recrutamento-ia-production.up.railway.app';
  5  | 
  6  | test.describe('Testes E2E - Recrutamento IA', () => {
  7  | 
  8  |   test('Acesso à página inicial (Health/Login)', async ({ page }) => {
  9  |     console.log(`Acessando ${BASE_URL}/...`);
  10 |     const response = await page.goto(BASE_URL);
> 11 |     expect(response?.status()).toBe(200);
     |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  12 | 
  13 |     // Verifica se a página de login foi renderizada
  14 |     const title = await page.title();
  15 |     console.log(`Título da página: ${title}`);
  16 |     expect(title).toContain('Login');
  17 |     
  18 |     // Verifica se o formulário existe
  19 |     const loginForm = page.locator('form');
  20 |     await expect(loginForm).toBeVisible();
  21 |     
  22 |     console.log('✔ A página de Login está carregando corretamente no Railway!');
  23 |   });
  24 | 
  25 |   test('Acesso ao endpoint de Health do servidor', async ({ request }) => {
  26 |     console.log(`Acessando ${BASE_URL}/health...`);
  27 |     const response = await request.get(`${BASE_URL}/health`);
  28 |     expect(response.ok()).toBeTruthy();
  29 |     
  30 |     const body = await response.text();
  31 |     console.log(`Health Check Body: ${body}`);
  32 |     expect(body).toContain('OK');
  33 |     
  34 |     console.log('✔ O Health Check do Servidor está funcionando!');
  35 |   });
  36 | 
  37 | });
```