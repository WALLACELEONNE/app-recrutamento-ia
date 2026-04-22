import { test, expect } from '@playwright/test';

// Substitua pela URL gerada do Railway
const BASE_URL = 'https://server-production-73243.up.railway.app';

test.describe('Testes E2E - Recrutamento IA', () => {

  test('Acesso à página inicial (Health/Login)', async ({ page }) => {
    console.log(`Acessando ${BASE_URL}/...`);
    const response = await page.goto(BASE_URL);
    expect(response?.status()).toBe(200);

    // Verifica se a página de login foi renderizada
    const title = await page.title();
    console.log(`Título da página: ${title}`);
    expect(title).toContain('Login - Recrutamento IA');
    
    // Verifica se o formulário existe
    const loginForm = page.locator('form');
    await expect(loginForm).toBeVisible();
    
    console.log('✔ A página de Login está carregando corretamente no Railway!');
  });

  test('Acesso ao endpoint de Health do servidor', async ({ request }) => {
    console.log(`Acessando ${BASE_URL}/health...`);
    const response = await request.get(`${BASE_URL}/health`);
    expect(response.ok()).toBeTruthy();
    
    const body = await response.text();
    console.log(`Health Check Body: ${body}`);
    expect(body).toContain('OK');
    
    console.log('✔ O Health Check do Servidor está funcionando!');
  });

});