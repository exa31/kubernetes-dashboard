---
name: clean-code-refactor
description: Universal code quality, refactoring patterns, and code smell elimination across all programming languages. Use when restructuring complex logic, breaking down large functions, applying guard clauses/early returns, eliminating duplicate code, improving naming, and enforcing SOLID/KISS/DRY principles without breaking existing behavior.
---

# Clean Code & Universal Refactoring Standards

This skill provides universal heuristics and patterns for transforming messy, brittle code into clean, readable, and maintainable systems without altering external behavior (*Zero Behavioral Regression*).

---

## 1. Prime Directive: Pure Refactoring

> [!IMPORTANT]
> **Refactoring means restructuring existing computer code without changing its external behavior.**
> Never combine functional behavior changes (adding new features or changing business logic) with refactoring in the same pass.

---

## 2. Top Code Smells & Tactical Cures

### Smell 1: Pyramid of Doom (Deep Nesting)
- **Symptom**: Code indented 4+ levels deep with nested `if-else` blocks.
- **Cure**: **Guard Clauses & Early Return**. Check failure or boundary conditions upfront and return immediately.

```ts
// ❌ Smelly: Deep nesting
function processUpload(file) {
  if (file) {
    if (file.size <= MAX_SIZE) {
      if (isValidType(file.type)) {
        return upload(file);
      } else {
        throw new Error('Invalid type');
      }
    } else {
      throw new Error('File too large');
    }
  }
}

// ✅ Clean: Guard clauses
function processUpload(file) {
  if (!file) return null;
  if (file.size > MAX_SIZE) throw new Error('File too large');
  if (!isValidType(file.type)) throw new Error('Invalid type');

  return upload(file);
}
```

---

### Smell 2: God Functions (>40–50 Lines)
- **Symptom**: A single function that fetches data, parses strings, handles business rules, formats HTML, and writes logs.
- **Cure**: **Extract Method & Single Responsibility Principle (SRP)**. Every function should do one thing at one level of abstraction.

---

### Smell 3: Magic Numbers & Hardcoded Strings
- **Symptom**: `if (status === 3)` or `setTimeout(cb, 86400000)`.
- **Cure**: Replace with descriptive named constants or typed Enums:
  ```ts
  const ONE_DAY_IN_MS = 24 * 60 * 60 * 1000;
  export const UserStatus = {
    PENDING: 1,
    ACTIVE: 2,
    ARCHIVED: 3,
  } as const;
  ```

---

### Smell 4: Boolean Flag Arguments
- **Symptom**: `createUser(name, email, true, false, true)`.
- **Cure**: Use an options object or explicit distinct functions:
  ```ts
  createUser({ name, email, isAdmin: true, sendWelcomeEmail: false });
  ```

---

### Smell 5: Copy-Paste Code Duplication (WET Code)
- **Symptom**: The same 10-line calculation or formatting logic copied in 3 different components/files.
- **Cure**: Extract into a shared, pure utility function. Pure functions (deterministic, no side-effects) are trivially easy to test.

---

## 3. The 4-Step Safe Refactoring Protocol

1. **Safety Net**: Confirm existing tests pass or establish baseline input/output expectations.
2. **Small Increments**: Make one micro-refactor at a time (e.g. rename a variable or extract one helper).
3. **Verify Cleanly**: Run static analysis / tests after each micro-step.
4. **Final Review**: Ensure no accidental changes in public interfaces or return values.
