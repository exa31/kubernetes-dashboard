---
name: secret-and-config-safety
description: Universal secrets management, environment variable safety, configuration hygiene, and sensitive data protection. Use when managing .env files, setting up config schemas, auditing repositories for leaked credentials, configuring .gitignore, or handling secrets across dev, staging, and production environments.
---

# Secrets Management & Configuration Safety

This skill establishes strict guidelines for managing sensitive environment variables, preventing credential leaks, and ensuring zero-downtime configuration hygiene.

---

## 1. The Three Inviolable Laws of Secrets

1. **NEVER Commit Real Secrets**:
   Files containing genuine API keys, passwords, private keys, or tokens must never be tracked in Git.
2. **ALWAYS Synchronize `.env.example`**:
   Whenever a new environment variable is introduced, immediately add a placeholder entry to `.env.example` with clear documentation or dummy values.
3. **FAIL-FAST at Startup**:
   Validate that required configuration values exist upon application boot. If a mandatory variable is missing, throw an explicit, descriptive error immediately rather than failing silently halfway through execution.

---

## 2. Fail-Fast Configuration Pattern

Validate environment configurations early using a schema validator (e.g. Zod) or an initialization assertion:

```ts
import { z } from 'zod';

const envSchema = z.object({
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32, 'JWT_SECRET must be at least 32 characters'),
  PORT: z.coerce.number().default(3000),
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
});

// Throws detailed error at server bootstrap if any variable is missing or malformed
export const env = envSchema.parse(process.env);
```

---

## 3. Log Redaction & Masking Rules

Sensitive data must NEVER leak into console logs, error messages, or APM monitoring:

### Strings to Mask:
- Authorization headers (`Bearer eyJ...` -> `Bearer eyJ...[REDACTED]`)
- Database connection strings (`postgres://user:password@host...` -> `postgres://user:***@host...`)
- Credit card numbers, CVVs, passwords, private keys
- Webhook secret signatures

### Masking Utility Heuristic:
```ts
export function maskSecret(secret?: string): string {
  if (!secret) return '';
  if (secret.length <= 8) return '********';
  return `${secret.slice(0, 3)}...${secret.slice(-3)}`;
}
```

---

## 4. Git Hygiene & `.gitignore` Baseline

Ensure every repository's `.gitignore` contains the minimum baseline:

```gitignore
# Environment files
.env
.env.*
!.env.example

# Credentials & Keys
*.pem
*.key
*.p12
*.keystore
id_rsa*

# Service Account files
*service-account*.json
*credentials*.json

# OS & Editor files
.DS_Store
Thumbs.db
```
