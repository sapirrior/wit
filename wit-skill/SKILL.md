---
name: wit-skill
description: Assists in generating, inspecting, and formatting valid, high-fidelity XML snapshot files for the wit CLI tool. Activate this skill when writing, validating, or editing a wit snapshot XML file.
---

# wit-skill — Technical Specification & AI Generation Rules

Authoritative Repository: [github.com/sapirrior/wit](https://github.com/sapirrior/wit)  
Raw Reference Docs: [raw.githubusercontent.com/sapirrior/wit/main/README.md](https://raw.githubusercontent.com/sapirrior/wit/main/README.md)

This skill contains the rigorous technical specifications, schema rules, and parser boundaries required to generate valid, rebuilt-compatible `wit` XML archives.

---

## 1. Symmetrical XML Schema Definition

All archives must strictly conform to the following XML structure. Deviations in tag names, attributes, or casing will trigger parser validation errors.

### Root Node: `<wit>`
The root tag wrapping the file database.
- **Attributes**:
  - `version` (string): Must be exactly `"2.0"`.
  - `created` (string): UTC timestamp formatted in ISO-8601/RFC3339 (`YYYY-MM-DDTHH:MM:SSZ`).
  - `root` (string): POSIX directory name of the target root.
  - `file_count` (int): Total count of child tags.
  - `message` (string, optional): XML-escaped custom description message.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<wit version="2.0" created="2026-07-17T11:34:17Z" root="app" file_count="4" message="Technical Commit">
  <!-- Children tags -->
</wit>
```

### Children Nodes:
1. **Regular Text File (`<file>`)**:
   Contains file content in a CDATA wrapper.
   - **Attributes**:
     - `path` (string): Relative destination path using POSIX forward slashes (`/`).
     - `sha1` (string): 40-character lowercase hex SHA-1 digest of the raw contents.
     - `mode` (string): Octal string representing file permissions (e.g. `"0644"`, `"0755"`).
     - `size` (int): Total byte count of the uncompressed, post-decoded content.
   - **Casing**: `<file path="..." sha1="..." mode="..." size="...">`
2. **Binary File (`<binary>`)**:
   Contains base64 encoded data inside CDATA.
   - **Attributes**: Same as `<file>` plus:
     - `encoding` (string): Must be exactly `"base64"`.
   - **Casing**: `<binary path="..." sha1="..." mode="..." size="..." encoding="base64">`
3. **Symbolic Link (`<symlink>`)**:
   A self-closing tag for symbolic links.
   - **Attributes**:
     - `path` (string): Destination path.
     - `target` (string): Path the link points to.
     - `mode` (string): Octal representation (usually `"0777"`).
   - **Casing**: `<symlink path="..." target="..." mode="0777"/>`
4. **Empty File (`<empty>`)**:
   A self-closing tag for zero-byte files.
   - **Attributes**:
     - `path` (string): Destination path.
     - `mode` (string): Octal representation.
   - **Casing**: `<empty path="..." mode="0644"/>`

---

## 2. Formatting Constraints (Strict Parser Requirements)

### ⚠️ Whitespace Sensitive CDATA
Go's `xml.Decoder` captures all character data inside a node. Do **not** format XML layout with newlines or indents around CDATA blocks.
- **Valid (100% correct)**:
  ```xml
  <file path="main.go" sha1="d39a3..." mode="0644" size="52"><![CDATA[package main
func main() {}]]></file>
  ```
- **Invalid (will corrupt files with extra spaces/newlines)**:
  ```xml
  <file path="main.go" sha1="d39a3..." mode="0644" size="52">
    <![CDATA[package main
func main() {}]]>
  </file>
  ```

### CDATA Escaping Rule
If a file's content contains the sequence `]]>`, you must escape it by splitting it across multiple CDATA blocks:
- Sequence: `]]>`
- Transformed: `]]]]><![CDATA[>`

---

## 3. Generating wit XML Outlines (Metadata Indexing)

To present a proposed workspace outline without transferring file bodies, generate self-closing elements omitting CDATA contents:
```xml
<file path="src/main.go" sha1="4f8f411ba188ec8f67efdc8b6ff9a177958f6141" mode="0644" size="1054"/>
```
The rebuilding client validates size and hash on files, so these outlines are strictly for context map exchange and cannot be rebuilt by the client directly until content is populated.
