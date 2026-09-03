# AGENTS.md — Vue Template Instructions for AI Coding Agents

This document is read automatically by opencode, Claude Code, Cursor, Antigravity, and other AI tools. Follow all guidelines below **without exception** when writing or modifying code in this Vue template.

---

## 1. Stack Overview

- **Framework**: Vue 3 (`vue`) with Composition API & `<script setup lang="ts">`
- **Build Tool**: Vite (`@vitejs/plugin-vue`)
- **Language**: TypeScript (Strict mode, `vue-tsc`, `@/` path alias mapped to `src/`)
- **Styling**: Tailwind CSS v4 (`@tailwindcss/vite`, `tailwindcss`) + custom CSS variables (`src/style.css`)
- **UI Components**: PrimeVue (`primevue`, `@primevue/themes/aura`) + `primeicons`
- **State Management**: Pinia (`pinia`) using Setup Store syntax
- **Routing**: Vue Router (`vue-router`)
- **Logging**: Structured logger via `logger` from `@/utils` (Consola)

---

## 2. Mandatory Commands & Quality Gate

- `npm run build` — `vue-tsc -b && vite build` — **MUST pass cleanly** before finishing any task.
- `npm run lint` — `eslint . --fix` — **MUST pass with ZERO errors**.
- `npm run format` — `prettier --write src/` — Formats source code according to repo configuration.
- `npm run dev` — Launches Vite development server.

---

## 3. Architecture & Modular Layout

This project follows a domain-driven, feature-sliced modular architecture:

```text
src/
├── api/             # API client & domain endpoint functions (e.g., users.ts, client.ts)
├── assets/          # Static assets & global CSS
├── components/      # Global shared Vue UI components
│   └── layout/      # Shared layout components (Sidebar, TopNav)
├── composables/     # Reusable Vue composable functions (e.g., useDarkMode)
├── features/        # Feature modules containing domain components (e.g., features/users/)
├── layouts/         # Page layout wrappers (DashboardLayout, AuthLayout)
├── pages/           # Route-level Vue components / Views
├── router/          # Vue Router configuration
├── stores/          # Pinia global state management stores
├── types/           # TypeScript interfaces, types & DTOs
└── utils/           # Shared helper functions, formatters, and logger
```

### Key Architectural Rules:
- **Barrel Exports**: Every major directory (`api`, `components`, `composables`, `features`, `layouts`, `pages`, `stores`, `types`, `utils`) MUST contain an `index.ts` file exporting its public interface.
- **Import Paths**: Always use path alias `@/` for imports from `src/` (e.g., `import { logger, formatCurrency } from '@/utils'`, `import { useAuthStore } from '@/stores'`).
- **Feature Isolation**: Place domain-specific components inside `src/features/<feature>/`. Do not bloat `src/components/` with feature-specific code.

---

## 4. Code & Style Conventions

### Single File Component (SFC) Conventions
- **Block Order**: Strict block order enforced by ESLint:
  1. `<script setup lang="ts">`
  2. `<template>`
  3. `<style scoped>`
- **Template Component Casing**: Use PascalCase for component tags in template (`<Sidebar />`, `<UserCard />`).
- **Self-Closing Tags**: Use self-closing syntax for HTML elements and components without slot content (`<input />`, `<button />`, `<UserAvatar />`).

### TypeScript & Imports
- **Type-only imports**: Use explicit type imports: `import type { User } from '@/types'` (`@typescript-eslint/consistent-type-imports`).
- **Strict Typing**: Avoid `any`. Define proper interfaces/types in `src/types/`.
- **Import Sorting**: `eslint-plugin-simple-import-sort` is enforced. Run `npm run lint` to fix automatically.
- **Unused Imports**: Unused imports/variables are disallowed unless prefixed with `_`.

### State Management (Pinia)
- Store files live in `src/stores/`.
- Use **Setup Store** syntax (`defineStore('id', () => { ref, computed, function })`) for consistency across all stores.

### Logging & Side Effects
- **DO NOT** use `console.log()`. Use the app logger: `import { logger } from '@/utils'; logger.info(...)`.

---

## 5. Definition of Done

Before marking any task complete:
1. [ ] `npm run lint` passes without errors.
2. [ ] `npm run build` compiles without `vue-tsc` or Vite errors.
3. [ ] All imports are clean, sorted, and use path alias `@/`.
4. [ ] No `console.log` statements remain in code (use `logger` from `@/utils`).
5. [ ] Barrel export `index.ts` is updated if new composables, modules, or components are added.
