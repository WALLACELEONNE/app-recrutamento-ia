# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests\e2e\railway_test.spec.js >> Testes E2E Completos - Recrutamento IA >> Fluxo de Login e Acesso ao Dashboard
- Location: tests\e2e\railway_test.spec.js:8:7

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected substring: "/dashboard"
Received string:    "https://app-recrutamento-ia-production.up.railway.app/login?"
```

# Page snapshot

```yaml
- generic [ref=e2]:
  - generic [ref=e3]:
    - heading "Nova Voice" [level=1] [ref=e4]
    - paragraph [ref=e5]: Painel de Administração e RH
  - generic [ref=e6]:
    - generic [ref=e7]:
      - generic [ref=e8]: E-mail corporativo
      - textbox "E-mail corporativo" [ref=e10]
    - generic [ref=e11]:
      - generic [ref=e12]: Senha
      - textbox "Senha" [ref=e14]
    - generic [ref=e15]:
      - generic [ref=e16]:
        - checkbox "Lembrar-me" [ref=e17]
        - generic [ref=e18]: Lembrar-me
      - link "Esqueceu a senha?" [ref=e20] [cursor=pointer]:
        - /url: "#"
    - button "Entrar no Painel" [ref=e22] [cursor=pointer]:
      - generic [ref=e23]: Entrar no Painel
```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | // URL de Produção do Railway
  4  | const BASE_URL = 'https://app-recrutamento-ia-production.up.railway.app';
  5  | 
  6  | test.describe('Testes E2E Completos - Recrutamento IA', () => {
  7  | 
  8  |   test('Fluxo de Login e Acesso ao Dashboard', async ({ page }) => {
  9  |     page.on('pageerror', error => console.log('ERRO NA PÁGINA:', error));
  10 |     page.on('console', msg => console.log('CONSOLE:', msg.text()));
  11 | 
  12 |     // 1. Acessa a página de login
  13 |     console.log(`Acessando ${BASE_URL}/login...`);
  14 |     await page.goto(`${BASE_URL}/login`, { waitUntil: 'networkidle' });
  15 |     
  16 |     // Verifica se o título da página está correto
  17 |     await expect(page).toHaveTitle(/Login/);
  18 | 
  19 |     // 2. Preenche o formulário de login
  20 |     console.log('Preenchendo credenciais de admin...');
  21 |     await page.fill('input[type="email"]', 'admin@acme.com');
  22 |     await page.fill('input[type="password"]', 'password123');
  23 | 
  24 |     // 3. Submete o formulário
  25 |     console.log('Realizando login...');
  26 |     await Promise.all([
  27 |       page.waitForNavigation(), // Aguarda o redirecionamento
  28 |       page.click('button[type="submit"]')
  29 |     ]);
  30 | 
  31 |     // 4. Verifica se redirecionou para o dashboard
  32 |     console.log('Verificando redirecionamento para o Dashboard...');
> 33 |     expect(page.url()).toContain('/dashboard');
     |                        ^ Error: expect(received).toContain(expected) // indexOf
  34 | 
  35 |     // Verifica se elementos chave do Dashboard carregaram
  36 |     await expect(page.locator('text=Visão Geral')).toBeVisible();
  37 |     await expect(page.locator('text=Entrevistas Recentes')).toBeVisible();
  38 |     await expect(page.locator('text=Desempenho por Skill')).toBeVisible();
  39 | 
  40 |     console.log('✔ Fluxo de Login testado com sucesso no Railway!');
  41 |   });
  42 | 
  43 |   test('Healthcheck do Servidor', async ({ request }) => {
  44 |     const response = await request.get(`${BASE_URL}/health`);
  45 |     expect(response.ok()).toBeTruthy();
  46 |     const body = await response.text();
  47 |     expect(body).toContain('OK');
  48 |     console.log('✔ O Health Check do Servidor está funcionando!');
  49 |   });
  50 | 
  51 | });
```