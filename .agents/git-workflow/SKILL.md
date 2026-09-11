---
name: git-workflow
description: Standardized Git version control workflow, conventional commits, branch hygiene, secret leak prevention, atomic staging, and PR documentation. Use when committing code, writing commit messages, staging changes, creating branches, reviewing Git status, or opening Pull Requests.
---

# Git Workflow & Version Control Standard

This skill enforces industry-standard Git hygiene across all repositories regardless of technology stack or framework. It ensures clean commit histories, prevents accidental secret leaks, and standardizes collaboration.

---

## 1. Conventional Commits Specification

All commit messages MUST follow the [Conventional Commits v1.0.0](https://www.conventionalcommits.org/) format:

```text
<type>(<optional scope>): <short imperative description>

[optional longer body explaining WHY, not just WHAT]

[optional footer(s): e.g., Closes #123, BREAKING CHANGE: ...]
```

### Allowed Types:
| Type | Purpose | Example |
| :--- | :--- | :--- |
| **`feat`** | A new user-facing or system feature | `feat(auth): add google oauth2 login flow` |
| **`fix`** | A bug fix | `fix(journey): handle null thumbnail when uploading` |
| **`refactor`** | Code change that neither fixes a bug nor adds a feature | `refactor(db): extract pool connection into shared service` |
| **`perf`** | Code change that improves performance | `perf(images): convert uploaded assets to webp with sharp` |
| **`chore`** | Maintenance, package upgrades, auxiliary config | `chore(deps): update pinia to v3.0.4` |
| **`docs`** | Documentation only changes | `docs(readme): add docker setup instructions` |
| **`test`** | Adding or correcting tests | `test(worker): add unit test for job scraper parsing` |
| **`style`** | Formatting, white-space, semicolon fixes (no code change) | `style(css): reorder utility classes for journey card` |
| **`ci`** | CI/CD pipeline and GitHub Actions modifications | `ci(github): add automated typecheck step on PR` |

### Rules for the Summary Line:
1. Use lowercase for the type and description.
2. Use the imperative, present tense ("add", not "added" or "adds").
3. Do not capitalize the first letter.
4. Do not put a period (`.`) at the end.
5. Keep it under 72 characters.

---

## 2. Atomic Staging Protocol

> [!IMPORTANT]
> **One Commit = One Logical Unit of Work.**
> Never lump unrelated changes (e.g. a refactor + a bugfix + a config update) into a single commit.

### Pre-Staging Workflow:
1. Run `git status` to observe modified, untracked, and deleted files.
2. Stage specifically by file or hunk (`git add <file>`), **never** run blind `git add .` if untracked temporary or secret files are present.
3. Review staged diff before committing:
   ```powershell
   git diff --staged --stat
   ```

---

## 3. Pre-Commit Secret Scanner Checklist

Before executing any commit, verify that none of the following files or patterns are staged:

- [ ] `.env`, `.env.local`, `.env.production`
- [ ] `*.pem`, `*.key`, `*.p12`, `*.keystore`
- [ ] Service account JSON keys (`*service-account*.json`, `*credentials*.json`)
- [ ] Hardcoded secret strings (look for strings matching `sk-proj-`, `AIzaSy`, `ghp_`, `postgres://user:pass@`)
- [ ] Build artifacts (`node_modules/`, `dist/`, `.output/`, `.nuxt/`, `build/`)

If a secret is accidentally staged, immediately unstage:
```powershell
git restore --staged <file>
```

---

## 4. Pull Request (PR) Documentation Template

When preparing a Pull Request description, structure it as follows:

```markdown
## Summary of Changes
- Concise bullet points of what was implemented or resolved.

## Motivation & Context
- Why was this change necessary? (Link issue/ticket if applicable).

## Verification & Testing
- [x] Static typecheck / lint passed cleanly.
- [x] Verified locally in browser / terminal.
- [x] No regressions observed in existing features.

## Checklist
- [x] No secrets, temporary files, or debug logs committed.
- [x] Documentation updated (if applicable).
```
