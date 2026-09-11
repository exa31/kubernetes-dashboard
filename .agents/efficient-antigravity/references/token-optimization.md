# Token & Context Optimization Guide

Context saturation is the primary cause of model degradation, hallucinations, slow response times, and high token costs. Follow these technical rules to keep conversations lean and responsive.

---

## 1. The Cost of Context Bloat

When context accumulates unnecessary code dumps or verbose logs:
- **Attention degradation**: The agent may miss subtle details located in earlier turns.
- **Latency increase**: Turn times scale with context length.
- **Token budget exhaustion**: Large files can consume 15,000+ tokens in a single tool call.

---

## 2. Reading Code Conservatively

### Anti-Pattern: Full-File Dumps
```json
// BAD: Reading all 800 lines of a file to inspect one method
{
  "AbsolutePath": "d:/Project/portofolio-v2/server/services/project.service.ts"
}
```

### Best Practice: Targeted Slices
```json
// GOOD: Pinpointing the method and reading only its bounds
{
  "AbsolutePath": "d:/Project/portofolio-v2/server/services/project.service.ts",
  "StartLine": 45,
  "EndLine": 85
}
```

### Workflow for Finding Code
1. Run `grep_search` with a specific `Query` and filter with `Includes` (e.g. `*.vue` or `*.ts`).
2. Identify the exact matching file and line number from the ripgrep result.
3. Call `view_file` specifying `StartLine` and `EndLine` with a small padding ($\approx 20-30$ lines above and below).

---

## 3. Terminal Output Throttling

Command output directly enters the conversation transcript. Verbose commands can dump thousands of lines.

### Best Practices:
1. **Piping & Slicing**:
   In PowerShell (`pwsh`), limit output lines when expecting heavy dumps:
   ```powershell
   npm run build | Select-Object -Last 40
   ```
2. **Filtering for Errors Only**:
   When testing compilation or type errors:
   ```powershell
   npx vue-tsc --noEmit | Select-Object -First 30
   ```
3. **Suppressing Progress Bars & Spammers**:
   Use flags like `--silent`, `--quiet`, or `--no-progress` when available.

---

## 4. Surgical Editing vs Full Rewrites

| Operation | Tool to Use | Why |
| :--- | :--- | :--- |
| Changing a function or few lines | `replace_file_content` | Minimal token consumption, clear diff generation |
| Modifying several separate sections in one file | `multi_replace_file_content` | Single-pass edit, preserves unchanged lines |
| Creating a new file from scratch | `write_to_file` | Appropriate for new modules only |
| Rewriting an existing 500-line file | **AVOID** | Wastes thousands of tokens and increases risk of introducing regression |

---

## 5. Cleaning Up Scratch Files

Temporary testing scripts should always be kept inside the designated scratch directory:
- `<appDataDir>/brain/<conversation-id>/scratch/`
- Never pollute workspace roots with temporary debug dumps or ad-hoc test files.
