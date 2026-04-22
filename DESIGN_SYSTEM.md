# 🎨 Design System & UI/UX Guidelines: Plataforma de Recrutamento IA

Este documento estabelece a identidade visual, padrões de interação e diretrizes de acessibilidade para a Plataforma de Recrutamento com IA por Voz. Rejeitamos designs genéricos: este sistema foi pensado especificamente para transmitir **Confiança (B2B/RH)** e **Inovação Tranquila (Candidatos)**.

---

## 1. Identidade Visual e Atmosfera

A plataforma possui duas "faces" distintas, mas conectadas pela mesma linguagem visual:
- **Painel do RH (Admin/B2B):** Focado em produtividade, dados densos, leitura de relatórios. Tema *Light Mode* predominante.
- **Sala de Entrevista (Candidato/B2C):** Focado em imersão, redução de ansiedade e foco na voz da IA. Tema *Dark Mode* imersivo.

---

## 2. Paleta de Cores (Tailwind CSS customizado)

A paleta de cores foi escolhida para garantir contraste (WCAG AA e AAA) e transmitir tecnologia e segurança.

### 🔵 Cores Primárias (Confiança & Ação)
- **Primary Base (`indigo-600`):** `#4F46E5` (Botões principais, links ativos)
- **Primary Hover (`indigo-700`):** `#4338CA` (Interação de hover)
- **Primary Active (`indigo-800`):** `#3730A3` (Clique/Pressed)
- **Primary Light (`indigo-50`):** `#EEF2FF` (Fundos de seleção, highlights sutis)

### 🟢 Cores Secundárias / Acento (IA & Voz)
- **Accent Base (`teal-500`):** `#14B8A6` (Ondas sonoras, status "Gravando/Ouvindo")
- **Accent Glow (`teal-400`):** `#2DD4BF` (Efeitos de brilho para quando a IA estiver "Pensando")

### ⚪ Tons Neutros (Superfícies & Textos)
- **Background RH (`slate-50`):** `#F8FAFC`
- **Surface RH (`white`):** `#FFFFFF` (Cards, Modais)
- **Background Candidato (`slate-900`):** `#0F172A`
- **Surface Candidato (`slate-800`):** `#1E293B`
- **Texto Principal (`slate-900` / `slate-50`):** `#0F172A` (Light) / `#F8FAFC` (Dark)
- **Texto Secundário (`slate-500` / `slate-400`):** `#64748B` (Light) / `#94A3B8` (Dark)

### 🔴 Cores Semânticas (Status)
- **Success (`emerald-600`):** `#059669` (Aprovado, Concluído)
- **Warning (`amber-500`):** `#F59E0B` (Atenção, Faltam X minutos)
- **Error/Destructive (`rose-600`):** `#E11D48` (Erro de mic, Desconectar)

---

## 3. Tipografia

A tipografia foca em máxima legibilidade em telas densas de dados (RH) e clareza absoluta nas instruções (Candidato).

- **Fonte Primária (Headings & UI):** `Inter` (Sans-serif)
  - *Pesos:* 400 (Regular), 500 (Medium), 600 (SemiBold), 700 (Bold).
  - *Tracking:* `-0.02em` em Títulos (H1, H2) para um visual mais coeso e moderno.
- **Fonte Secundária (Dados e Transcrições):** `JetBrains Mono` ou `Fira Code` (Apenas para exibir os logs de JSONB, métricas técnicas e timestamps da transcrição).

### Escala Tipográfica (Mobile-First)
- `text-xs` (0.75rem) - Labels secundários, timestamps.
- `text-sm` (0.875rem) - UI secundária, botões pequenos.
- `text-base` (1rem) - Corpo de texto padrão.
- `text-lg` (1.125rem) - Corpo de texto em destaque (instruções do candidato).
- `text-xl` (1.25rem) - H4 (Títulos de cards).
- `text-2xl` (1.5rem) - H3 (Títulos de seções).
- `text-3xl` (1.875rem) - H2 (Títulos de página mobile).
- `text-4xl` (2.25rem) - H1 (Títulos principais desktop).

---

## 4. Grid e Layout de Interface

O layout utiliza uma estrutura fluida de CSS Grid/Flexbox de 12 colunas, aderindo aos breakpoints nativos do Tailwind.

### Breakpoints
- `sm`: `640px` (Tablets retrato)
- `md`: `768px` (Tablets paisagem)
- `lg`: `1024px` (Laptops - Transição do Menu Mobile para Sidebar fixo)
- `xl`: `1280px` (Monitores Desktop padrão)
- `2xl`: `1536px` (Telas Ultra-Wide, limitador de conteúdo `max-w-7xl`)

### Estruturas
1. **Dashboard RH (Admin):**
   - *Sidebar:* Fixo na esquerda (256px de largura) em telas `>= lg`. Oculto sob botão "Hamburger" no mobile.
   - *Header:* Barra superior (64px de altura) com Breadcrumbs e perfil do usuário.
   - *Main Content:* `max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8`.
2. **Player de Entrevista (Candidato):**
   - *Layout:* 100vh / 100vw, centralizado (Flexbox `items-center justify-center`).
   - Sem distrações (sem sidebar, sem navegação complexa). Apenas o Visualizador de Áudio, status da conexão e botões de controle (Mutar, Encerrar).

---

## 5. Componentes Personalizados e Estados de Interação

Todos os componentes interativos possuem 4 estados obrigatórios: **Default**, **Hover**, **Active (Focus/Pressed)** e **Disabled**.

### 5.1. Botões (Buttons)
- **Primary Button (Ação Principal - ex: "Iniciar Entrevista")**
  - *Default:* `bg-indigo-600 text-white shadow-sm rounded-lg px-4 py-2 font-medium transition-all duration-200`
  - *Hover:* `hover:bg-indigo-700 hover:shadow-md transform hover:-translate-y-0.5`
  - *Active:* `active:bg-indigo-800 active:scale-95`
  - *Focus (Teclado):* `focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2`
  - *Disabled:* `disabled:bg-indigo-300 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none opacity-60`

- **Secondary Button (Ações Alternativas)**
  - *Default:* `bg-white text-slate-700 border border-slate-300 rounded-lg px-4 py-2 font-medium transition-colors`
  - *Hover:* `hover:bg-slate-50 hover:text-slate-900`
  - *Focus:* `focus:ring-2 focus:ring-slate-200 focus:ring-offset-2`

### 5.2. Visualizador de Voz da IA (Ondas Sonoras)
- Elemento central da tela do candidato.
- **Estado "Listening" (Candidato falando):** Ondas sutis na cor `teal-500` com opacidade de 40%. Pulsação baseada no volume do microfone local.
- **Estado "Processing" (IA processando):** Animação de *loading/spinner* pulsante linear, com *glow* `teal-400`.
- **Estado "Speaking" (IA falando):** Ondas sonoras intensas e dinâmicas em `indigo-500` e `teal-400`, sincronizadas com os pacotes RTP de áudio (LiveKit).

### 5.3. Cards (Cartões de Vagas / Relatórios)
- *Default:* `bg-white rounded-xl shadow-sm border border-slate-100 p-6`
- *Hover (Se clicável):* `hover:shadow-md hover:border-indigo-100 transition-shadow duration-200 cursor-pointer`

---

## 6. Diretrizes de Acessibilidade (WCAG 2.1 AA/AAA)

O design não é apenas visual, ele exige implementação técnica de acessibilidade (A11y) no Go Templ:

1. **Contraste de Cores:**
   - O texto primário (`slate-900`) contra fundos brancos/cinzas excede a razão de `4.5:1` (WCAG AA).
   - O texto branco contra botões `indigo-600` possui contraste seguro de `5.1:1`.
2. **Navegação por Teclado:**
   - Todos os botões, links e inputs de formulário devem utilizar as classes `focus:ring-2` para garantir que o usuário que navega por "Tab" veja claramente onde está o foco. Rejeitamos o uso de `outline: none` sem um fallback de ring visual.
3. **Leitores de Tela (Screen Readers):**
   - O visualizador de áudio da IA (que é puramente visual/Canvas) deve conter `aria-hidden="true"`.
   - Ícones SVG devem conter `<span class="sr-only">Descrição da ação</span>`.
   - O estado da entrevista (Ouvindo, Falando, Processando) deve utilizar uma região `aria-live="polite"` para anunciar mudanças de estado aos leitores de tela.
4. **Semântica HTML:**
   - Uso estrito de `<nav>`, `<main>`, `<aside>`, `<section>`, `<article>`, `<header>` e `<footer>` nos templates Go (`.templ`).

---

## 7. Transições e Micro-interações (Alpine.js)

- **Modais e Dropdowns:** Entrada suave.
  - *Enter:* `transition ease-out duration-200`, `From: opacity-0 scale-95`, `To: opacity-100 scale-100`
  - *Leave:* `transition ease-in duration-150`, `From: opacity-100 scale-100`, `To: opacity-0 scale-95`
- **Feedback de Microfone:** O microfone pedirá permissão via browser. Um ícone de microfone no player terá a cor `rose-500` (mutado) transitando para `emerald-500` com um leve `pulse` (desmutado).
