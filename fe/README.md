# Vue 3 SaaS Dashboard Template

A production-ready, modern SaaS dashboard template built with Vue 3 and Vite.

## 🚀 Tech Stack

- **Framework**: [Vue 3](https://vuejs.org/) (Composition API & `<script setup>`)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **UI Components**: [PrimeVue](https://primevue.org/) (Unstyled mode / Aura Theme)
- **State Management**: [Pinia](https://pinia.vuejs.org/)
- **Routing**: [Vue Router](https://router.vuejs.org/)
- **Icons**: [PrimeIcons](https://primefaces.org/primeicons/)

## 📂 Modular Architecture

This template uses a domain-driven, feature-sliced architecture designed for scalability.

```text
src/
├── api/             # API clients and endpoints grouped by domain (e.g., users.ts)
├── assets/          # Static assets (images, global CSS)
├── components/      # Global shared UI components
│   └── layout/      # Layout-specific components (Sidebar, TopNav)
├── composables/     # Reusable Vue composition functions (e.g., useDarkMode)
├── features/        # Domain-specific components grouped by feature (e.g., users/)
├── layouts/         # Page layout wrappers (DashboardLayout, AuthLayout)
├── pages/           # Route-level Vue components (Views)
├── router/          # Vue Router configuration
├── stores/          # Pinia state management stores
├── types/           # TypeScript interfaces and types
└── utils/           # Helper functions and formatters
```

> **Note on Barrel Exports**: Every major directory contains an `index.ts` (Barrel Export). This allows you to import multiple items from a directory cleanly, e.g., `import { formatCurrency } from '@/utils'`.

## 🛠️ Getting Started

### 1. Install Dependencies
```bash
npm install
```

### 2. Run Development Server
```bash
npm run dev
```

### 3. Build for Production
```bash
npm run build
```

## 🎨 Design System

The application uses a custom set of CSS variables (`src/style.css`) that are fully integrated with Tailwind CSS v4. It features a premium, clean aesthetic inspired by modern SaaS applications (Linear, Vercel, Stripe), complete with dark mode support.
