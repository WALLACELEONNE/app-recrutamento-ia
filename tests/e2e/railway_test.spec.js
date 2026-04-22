import { test, expect } from '@playwright/test';

// URL de Produção do Railway
const BASE_URL = 'https://app-recrutamento-ia-production.up.railway.app';

test.describe('Testes E2E Completos - Recrutamento IA', () => {

  test('Fluxo de Login e Acesso ao Dashboard', async ({ page }) => {
    page.on('pageerror', error => console.log('ERRO NA PÁGINA:', error));
    page.on('console', msg => console.log('CONSOLE:', msg.text()));

    // 1. Acessa a página de login
    console.log(`Acessando ${BASE_URL}/login...`);
    await page.goto(`${BASE_URL}/login`, { waitUntil: 'networkidle' });
    
    // Verifica se o título da página está correto
    await expect(page).toHaveTitle(/Login/);

    // 2. Preenche o formulário de login
    console.log('Preenchendo credenciais de admin...');
    await page.fill('input[type="email"]', 'admin@acme.com');
    await page.fill('input[type="password"]', 'password123');

    // 3. Submete o formulário
    console.log('Realizando login...');
    await Promise.all([
      page.waitForNavigation(), // Aguarda o redirecionamento
      page.click('button[type="submit"]')
    ]);

    // 4. Verifica se redirecionou para o dashboard
    console.log('Verificando redirecionamento para o Dashboard...');
    expect(page.url()).toContain('/dashboard');

    // Verifica se elementos chave do Dashboard carregaram
    await expect(page.locator('text=Visão Geral')).toBeVisible();
    await expect(page.locator('text=Entrevistas Recentes')).toBeVisible();

    console.log('✔ Fluxo de Login testado com sucesso no Railway!');
  });

  test('Healthcheck do Servidor', async ({ request }) => {
    const response = await request.get(`${BASE_URL}/health`);
    expect(response.ok()).toBeTruthy();
    const body = await response.text();
    expect(body).toContain('OK');
    console.log('✔ O Health Check do Servidor está funcionando!');
  });

});