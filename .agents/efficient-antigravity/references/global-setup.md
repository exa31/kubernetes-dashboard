# Global Installation Guide for Antigravity

By default, this skill is installed in the workspace repository under:
`.agents/skills/efficient-antigravity/`

If you want this skill to be available across **all** your repositories and projects automatically, you can mirror it to your user-wide Antigravity configuration directory.

---

## Global Directory Path

The global configuration root for Antigravity on Windows is located at:
```text
C:\Users\<Your-Username>\.gemini\config\skills\
```
Or via PowerShell environment variable:
```text
$env:USERPROFILE\.gemini\config\skills\
```

---

## 1-Command Global Setup (PowerShell)

Open PowerShell and run the following command from the root of `portofolio-v2`:

```powershell
# Create global skills directory if it doesn't exist
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.gemini\config\skills"

# Copy the efficient-antigravity skill to global config
Copy-Item -Recurse -Force ".agents\skills\efficient-antigravity" "$env:USERPROFILE\.gemini\config\skills\"
```

---

## Verification

To verify that the skill has been copied globally:

```powershell
Test-Path "$env:USERPROFILE\.gemini\config\skills\efficient-antigravity\SKILL.md"
```

If it returns `True`, Antigravity will now automatically recognize this skill across all your workspaces and projects!
