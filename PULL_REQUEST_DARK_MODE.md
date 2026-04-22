# Pull Request: Implementação do Dark Mode Nativo (Tailwind + Alpine.js)

## 📝 Descrição
Este PR implementa o suporte nativo a temas Claro/Escuro (Light/Dark mode) em todo o frontend do projeto, abrangendo o painel de administração (RH) e a área do candidato, conforme os requisitos estabelecidos de UI/UX.

As alterações incluem:
- Configuração do Tailwind CSS para utilizar a estratégia `darkMode: 'class'`.
- Implementação de script bloqueante no `<head>` de todas as páginas para avaliar o tema ativo no `localStorage` ou, na ausência deste, herdar a preferência do sistema operacional (`window.matchMedia`). Isso garante a eliminação do efeito FOUC (Flash of Unstyled Content).
- Injeção de lógica interativa via `Alpine.store('theme')` gerenciando a troca dinâmica através de um botão `toggle` posicionado na Topbar do Dashboard e no Header do Candidato.
- Atualização semântica das classes CSS nos templates Templ (Login, Dashboard, e Interview), aplicando variações do Tailwind (ex: `dark:bg-slate-900`, `dark:text-slate-50`, `dark:border-slate-700`) em tipografias, formulários, cards, ícones e alertas.
- Adição da classe `transition-colors duration-200` ao `<body>` para garantir que a mudança de cores ocorra de forma suave para o usuário final.

## 🔗 Link para Preview no Ambiente
**Railway:** [app-recrutamento-ia-production.up.railway.app](https://app-recrutamento-ia-production.up.railway.app)
*(O preview será atualizado assim que o CI/CD fizer o merge com a `main`)*

## ✅ Checklist de Validação
- [x] Variável de tema persistida via `localStorage`.
- [x] O tema padrão respeita a preferência do Sistema Operacional do usuário (`prefers-color-scheme: dark`).
- [x] Nenhum Flash de conteúdo sem estilo (FOUC) ocorre durante o carregamento inicial da página.
- [x] O botão de *toggle* alterna os ícones corretamente (Sol/Lua) e atualiza a interface instantaneamente.
- [x] O **Dashboard** adapta todas as tabelas, modais, painéis laterais e tipografias.
- [x] A tela de **Login** adapta as bordas dos inputs, cores do botão e alertas de erro.
- [x] A **Sala de Entrevista (Candidato)** adapta o avatar SVG da IA, os textos e os estados dos botões (mute/unmute).
- [x] Testes visuais executados manualmente (Light/Dark) nas três interfaces primárias.

## 📸 Evidências / Screenshots
> *(Como trata-se de um PR via CLI, por favor execute a branch localmente ou acesse o preview da Railway para as validações visuais completas)*

**Testes executados:**
1. Inicialização do sistema em aba anônima (Chrome) e validação de que o SO em modo Dark forçou o site a renderizar Dark nativamente.
2. Troca do botão Toggle forçou `localStorage.setItem('color-theme', 'light')` e alterou a árvore do DOM `classList.remove('dark')`.
3. Navegação entre `/login` -> `/dashboard` -> `/interview` manteve o estado do tema perfeitamente lido no carregamento de cada rota HTTP graças à interceptação no `<head>`.