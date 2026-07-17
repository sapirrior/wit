---
name: wit-skill
description: Assists in generating, inspecting, and formatting valid, high-fidelity XML snapshot files for the wit CLI tool. Activate this skill when writing, validating, or editing a wit snapshot XML file.
---

# wit-skill — Generating and Formatting Valid wit XML Files

This skill guides AI models in reading and writing valid XML archives compatible with `wit`'s `grab` (rebuild) command. 

## 1. The XML Schema

Every `wit` snapshot file must be valid XML wrapped in a `<wit>` root element.

### Root element: `<wit>`
Attributes:
- `version`: Always `"2.0"`.
- `created`: UTC timestamp in ISO 8601 format (e.g., `2026-07-17T11:06:45Z`).
- `root`: The base name of the original root directory.
- `file_count`: Total number of entries (files, symlinks, empties).
- `message`: (Optional) A custom commit/snapshot message.

Example:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<wit version="2.0" created="2026-07-17T11:06:45Z" root="my_project" file_count="4" message="My first snapshot">
  <!-- Children tags sit here -->
</wit>
```

### Children elements (Four exhaustive types):
1. **Regular Text File (`<file>`)**:
   Contains raw text data inside CDATA.
   ```xml
   <file path="src/main.c" sha1="a0f3d9b8..." mode="0644" size="182"><![CDATA[#include <stdio.h>
int main() {
  printf("Hello!");
}]]></file>
   ```
2. **Binary File (`<binary>`)**:
   Contains base64 encoded data inside CDATA. Must include `encoding="base64"`.
   ```xml
   <binary path="assets/logo.png" sha1="e23fa0b..." mode="0644" size="12" encoding="base64"><![CDATA[aGVsbG8gd29ybGQ=]]></binary>
   ```
3. **Symbolic Link (`<symlink>`)**:
   Self-closing tag specifying target path.
   ```xml
   <symlink path="lib/current" target="lib/v1.0" mode="0777"/>
   ```
4. **Empty File (`<empty>`)**:
   Self-closing tag indicating a zero-byte file.
   ```xml
   <empty path="src/.gitkeep" mode="0644"/>
   ```

---

## 2. Formatting Rules & Constraints

### ⚠️ CRITICAL: Whitespace around CDATA
Do **NOT** write formatting whitespace (newlines or indentation spaces) between `<file>` / `<binary>` tag boundaries and their child CDATA blocks. 
*   **CORRECT**: `<file ...><![CDATA[content]]></file>`
*   **INCORRECT**: 
    ```xml
    <!-- INCORRECT: The newline and indentations will be read as file content -->
    <file ...>
      <![CDATA[content]]>
    </file>
    ```

### CDATA Escaping Rule
If a file's content contains the literal XML CDATA closer sequence `]]>`, split it across two CDATA blocks:
```
]]>  becomes  ]]]]><![CDATA[>
```

### Integrity Attributes
- **`sha1`**: The lowercase 40-character SHA-1 hex digest of the **raw, post-decoded** file content.
- **`size`**: The exact byte count of the **raw, post-decoded** file content.
- **`mode`**: Octal permission string (e.g. `"0644"`, `"0755"`).
- **`path`**: POSIX-style relative path with forward slashes (e.g. `src/main.go`).

---

## 3. Generating wit XML Outline Files (Directory Layout Index)

A **wit XML Outline** is a lightweight version of a `wit` snapshot. It preserves the exact file metadata structure but omits the CDATA contents. This allows an AI to understand or propose a complete workspace layout without transferring large file payloads.

### Outline Generation Schema
When generating an outline, write empty, self-closing `<file>` and `<binary>` tags. Symmetrically, omit the inner CDATA block and content.

Example Outline:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<wit version="2.0" created="2026-07-17T11:16:20Z" root="my_project" file_count="4" message="Workspace Outline Map">
  <!-- Files represented as self-closing tags containing only metadata -->
  <file path="src/main.go" sha1="4f8f411ba188ec8f67efdc8b6ff9a177958f6141" mode="0644" size="1054"/>
  <binary path="assets/logo.png" sha1="6a71e27a6f2b4ff95b9cdbb99a5e8484e55e51a8" mode="0644" size="14560"/>
  <symlink path="lib/current" target="lib/v1.0" mode="0777"/>
  <empty path="src/.gitkeep" mode="0644"/>
</wit>
```

### When to use Outlines:
1. **Initial Workspace Proposals**: When proposing a new project structure, the AI should first present an outline to let the user review the file layout before generating all target files.
2. **Context Summarization**: Use outlines when transmitting structural layout context of large codebases to save LLM context window tokens.
