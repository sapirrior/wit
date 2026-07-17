# wit — AI-Native High-Fidelity File Context Maker

`wit` is a lightweight, zero-dependency, streaming-native utility built in Go. It enables developers and AI models to package directory trees into human-readable XML context maker files and reconstruct them byte-for-byte on demand.

Unlike standard code prompt makers, `wit` is designed as a context maker for **symmetrical round-trips**. You can snap a directory, feed the single XML context file to an AI, receive the modified XML back, and rebuild the exact directory structure automatically.

---

## Key Features

- ⚡ **AI-Native & Symmetrical**: Perfect for exchanging complete codebases with AI assistants in a single message.
- 📦 **100% Standalone**: Compiled as a static binary with CGO disabled. Zero dependencies.
- 📥 **Streaming XML Engine**: Memory-efficient parsing of huge files via `xml.Decoder` and custom CDATA streaming writer.
- 🛡️ **Path Traversal Protection**: Rebuilding verifies relative target directories to prevent directory traversal attacks.
- 🔑 **SHA-1 Integrity**: Every file is validated using a SHA-1 checksum and size verification during rebuild.
- 🙈 **Smart .gitignore Support**: Automatically walks subdirectories and respects nested `.gitignore` files recursively.

---

## Installation

### From Source
Requires Go 1.20+:
```bash
make
```
This builds a statically linked binary named `wit` in the root folder.

---

## Usage

### 1. Snapshot a Directory (Snap)
Captures a directory tree into an XML archive.
```bash
# Symmetrical snap (defaults to snap if the target is a directory)
wit <folder> [-m "optional message"] [-o <output_file>]

# Or explicitly
wit snap <folder> [-m "optional message"] [-o <output_file>]
```
- By default, it writes to `witFile.xml`.
- If `witFile.xml` already exists, it auto-increments the output name to `witFile-n.xml`.

### 2. Rebuild a Workspace (Grab)
Reconstructs the directory tree from the snapshot XML.
```bash
wit grab <archive> [-o <destination_dir>]
```

### 3. View Snapshot Metadata (Meta)
Prints the metadata and the full file tree structure stored inside the archive.
```bash
wit meta <archive>
```

### 4. View Commit Message (Msg)
Displays the snapshot commit message saved in the XML header.
```bash
wit msg <archive>
```

---

## XML Archive Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<wit version="2.0" created="2026-07-17T11:06:45Z" root="my_project" file_count="4" message="My snapshot message">
  <file path="src/main.go" sha1="4f8f411..." mode="0644" size="51"><![CDATA[package main ...]]></file>
  <binary path="assets/image.png" sha1="6a71e2..." mode="0644" size="12" encoding="base64"><![CDATA[aGVsbG8gd29ybGQ=]]></binary>
  <symlink path="lib/current" target="lib/v1.0" mode="0777"/>
  <empty path="src/.gitkeep" mode="0644"/>
</wit>
```

---

## Instructing AI to Generate Valid wit XML Context Files

When pair-programming or generating workspaces with an AI, you can copy-paste or link the [wit-skill](wit-skill/SKILL.md) instructions. Instruct the AI with the following system rules:

1. **Output formatting**: Never wrap the XML blocks in additional markdown backticks if outputting directly to a `.xml` file.
2. **Whitespace**: Make sure the AI places no line breaks or indentation spaces around the CDATA block edges. It must write:
   `<file path="relative/path" ...><![CDATA[content]]></file>`
3. **Escaping**: Make sure it splits the illegal sequence `]]>` using `]]]]><![CDATA[>`.
4. **Integrity values**: The AI must compute valid `sha1` and `size` attributes for the files it modifies to ensure the `grab` client builds them successfully.

> [!IMPORTANT]
> Any deviation from these formatting constraints (such as incorrect SHA-1 digests, mismatched sizes, or extra formatting whitespace around CDATA blocks) will trigger a **fatal validation error** and immediately abort the workspace rebuild.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
Copyright © 2026 Nolan Stark.
