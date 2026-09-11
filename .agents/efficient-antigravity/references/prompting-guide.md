# High-Efficiency Prompting & Collaboration Guide

How you prompt Antigravity directly dictates the speed and accuracy of the output. Following these prompting patterns will cut iteration cycles by more than half.

---

## 1. The Anatomy of a High-Velocity Prompt

A high-velocity prompt contains three key components:
1. **Target**: Where the work should happen (file name, component, line number, or symbol).
2. **Context / Evidence**: What is happening vs what should happen (error message, expected behavior, input/output).
3. **Constraint**: Any boundaries (e.g. "don't change the database schema", "keep backward compatibility", "use existing Tailwind/UI tokens").

### Comparison Example:

#### ❌ Low-Efficiency Prompt:
> *"Ini pas di-submit error tolong benerin"*
> 
> *Result*: The agent has to search across the entire project, run builds, guess what was submitted, and spend dozens of tool calls just discovering the error.

#### ✅ High-Efficiency Prompt:
> *"Di page `/dashboard/journeys/new.vue`, pas klik Save muncul error 500 dari `/api/journeys`. Ini potongan lognya: `null value in column title violates not-null constraint`. Tolong sesuaikan payload di frontend agar validasi form jalan sebelum submit."*
> 
> *Result*: The agent immediately inspects `new.vue` and the API handler, applies the fix in 1 turn, and verifies it cleanly.

---

## 2. Leverage Built-in Context Features

1. **Active File Tracking**:
   Antigravity automatically sees your open files and cursor position. If you keep the failing file open in your IDE tab, you can simply say:
   > *"Perbaiki fungsi upload di file yang lagi saya buka ini."*
2. **File Mentioning**:
   Use standard file paths or `@filename` notation so the agent skips exploratory searches:
   > *"Cek logic auth di `server/utils/auth.ts`."*

---

## 3. When and How to Use Slash Commands

| Command | When to Use | What it Does |
| :--- | :--- | :--- |
| **`/goal`** | Multi-step refactors, full feature builds, or bug hunting across multiple files | Directs Antigravity to work autonomously without pausing for trivial updates until the objective is fully satisfied. |
| **`/grill-me`** | Before coding a major feature, new architecture, or ambiguous requirement | Antigravity asks targeted architectural questions to eliminate ambiguity *before* writing code. |
| **`/learn`** | Right after solving a project-specific trick, fixing an obscure bug, or setting a convention | Stores the lesson permanently in Antigravity memory so it never makes that mistake again in this repo. |
| **`/schedule`** | Setting reminders or recurring background tasks | Sets a one-time timer or recurring cron job without blocking execution. |

---

## 4. Efficient Iteration Patterns

- **Don't wait for full rebuilds manually**: Let Antigravity execute background tasks. You can continue asking questions or reading while tasks run.
- **Provide terminal snippets directly**: If a command failed in your external terminal, copy the 5-10 lines of the actual error message and paste it. That saves the agent from re-running slow diagnostic builds.
- **Use Indonesian or English naturally**: Antigravity understands technical Indonesian and English equally well. Focus on clarity of technical terms (e.g., *endpoint, payload, props, query, migration*).
