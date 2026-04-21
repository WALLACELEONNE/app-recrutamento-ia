# Guia de Contribuição

Bem-vindo ao guia de contribuição da plataforma de Recrutamento com IA! Este documento define as regras e o fluxo de trabalho Git (Git Flow) que utilizamos.

## 🌿 Estratégia de Branching (Git Flow Simplificado)

Utilizamos um modelo estruturado para proteger o código de produção e garantir estabilidade.

- `main`: Branch de produção. **Nunca comite diretamente aqui.** Todo código aqui é implantável.
- `develop`: Branch de integração e testes. Todo o desenvolvimento ativo converge para cá.
- `feature/*`: Branches para novas funcionalidades (ex: `feature/stt-integration`). Nascem de `develop`.
- `bugfix/*` ou `fix/*`: Correção de bugs não críticos que estão em `develop`.
- `hotfix/*`: Correções emergenciais em produção. Nascem de `main` e voltam para `main` e `develop`.

## 🔄 Fluxo de Trabalho (Workflow)

1. **Clone e Setup**:
   ```bash
   git clone git@github.com:WALLACELEONNE/app-recrutamento-ia.git
   cd app-recrutamento-ia
   git checkout develop
   ```

2. **Crie sua branch de trabalho**:
   ```bash
   git checkout -b feature/minha-nova-feature
   ```

3. **Desenvolva usando Docker**:
   Use o Docker Compose para subir o ambiente local isolado (Postgres, Redis, LiveKit, Server, Worker):
   ```bash
   cp .env.example .env.development
   docker-compose up --build -d
   ```

4. **Commits Semânticos**:
   Exigimos o padrão [Conventional Commits](https://www.conventionalcommits.org/). O prefixo diz o que o commit faz:
   - `feat:` Nova funcionalidade
   - `fix:` Correção de bug
   - `docs:` Alterações em documentação
   - `test:` Adicionando ou refatorando testes
   - `refactor:` Refatoração de código
   - `chore:` Tarefas de manutenção, dependências, build
   
   *Exemplo*: `git commit -m "feat: add user authentication via JWT"`

5. **Testes e Linting**:
   Antes de subir o código, garanta que os testes e a cobertura (>80%) passem:
   ```bash
   go test -v -cover ./...
   ```

6. **Abra um Pull Request (PR)**:
   - Faça o push da sua branch: `git push origin feature/minha-nova-feature`
   - Abra o PR no GitHub apontando para a branch `develop`.
   - O template de PR será carregado automaticamente. Preencha o checklist.
   - Aguarde o pipeline de CI/CD (GitHub Actions) passar.
   - Solicite Review.

## 🛡️ Proteções do Repositório (GitHub)

As seguintes regras estão ativas nas configurações do repositório:
- **`main` e `develop` são bloqueadas** para push direto.
- Requer aprovação de Pull Request antes de merge.
- Requer aprovação do status check (Pipeline do GitHub Actions `CI/CD Pipeline`).
