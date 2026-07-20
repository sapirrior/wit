# ✨ wit — AI-Native File Context Maker

[![Build & Release](https://github.com/sapirrior/wit/actions/workflows/build.yml/badge.svg)](https://github.com/sapirrior/wit/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sapirrior/wit)](https://golang.org)
[![License](https://img.shields.io/github/license/sapirrior/wit)](LICENSE)

Struggling to feed codebase context to AI models? Tired of manually copying and pasting changes back and forth?

Meet **`wit`**—the ultimate high-fidelity context bridge between you and your AI coding assistant. Symmetrically capture your entire directory in a single human-readable XML file, and restore modifications byte-for-byte in seconds. 🚀

---

## 🌟 Why `wit`?

- Symmetrically package directories to exchange codebases with AI assistants in a single message.
- Rebuild modifications made by AI assistants securely and automatically.
- 100% standalone single-binary utility with zero dependencies.
- Respects nested `.gitignore` files to keep archives clean.
- Auto-checks file integrity via SHA-1 hashes and sizes.
- Built-in path traversal safeguards to protect your filesystem.

---

## 🚀 Quick Start (60 Seconds)

### 1. Automated Installer (Recommended)
You can install `wit` interactively by executing:
```bash
curl -fsSLO https://raw.githubusercontent.com/sapirrior/wit/main/install.sh && sh install.sh && rm install.sh
```

### 2. Manual Download
Download the pre-compiled static binary matching your operating system directly from our **[Releases Page](https://github.com/sapirrior/wit/releases)**:

- 🐧 **Linux**: [wit-linux-amd64](https://github.com/sapirrior/wit/releases) | [wit-linux-arm64](https://github.com/sapirrior/wit/releases)
- 🤖 **Android (Termux)**: [wit-android-arm64](https://github.com/sapirrior/wit/releases)
- 🍎 **macOS**: [wit-darwin-arm64 (Apple Silicon)](https://github.com/sapirrior/wit/releases) | [wit-darwin-amd64 (Intel)](https://github.com/sapirrior/wit/releases)
- 🪟 **Windows**: [wit-windows-amd64.exe](https://github.com/sapirrior/wit/releases) | [wit-windows-arm64.exe](https://github.com/sapirrior/wit/releases)

On UNIX systems, make it executable and move it to your system PATH:
```bash
chmod +x wit-linux-amd64
mv wit-linux-amd64 /usr/local/bin/wit
```

### 2. Capture a Directory (Snap)
```bash
# Capture the 'src' directory into 'witFile.xml'
wit src -m "My project context"
```

### 3. Rebuild the Workspace (Grab)
```bash
# Rebuild the project structure from the XML file
wit grab witFile.xml -o my_restored_project
```

---

## 🛠️ Commands Guide

### 📦 `wit [folder] [options]`
> **Alias:** `wit snap`

Walks through your directory tree, respects Gitignore rules, and packs all contents into a clean XML file.
- Custom naming (`-o <name>`) automatically enforces the `.wit.xml` extension.
- **Compression flag (`-c`):** Compresses the output XML into a `.wit.xml.zip` file containing the `.wit.xml` file.
- **Exclusion flag (`--exclude <pattern>`):** Allows specifying glob patterns to skip (e.g. `*.log`, `node_modules/`). Can be specified multiple times.
- **Max Size flag (`--max-size <size>`):** Skip files exceeding the size limit (e.g. `1MB`, `500KB`, `100B`).
- **Example:** `wit my-app/ -m "Adding oauth support" -o my-app --exclude "*.log" --max-size 1MB -c`

### 🏗️ `wit grab [archive] [-o dest_folder]`
Reconstructs the workspace directory structure from the archive. Automatically detects and decompresses `.zip` archives on the fly.
- **Example:** `wit grab my-app.wit.xml.zip -o rebuilt-app`

### ✅ `wit verify [archive]`
Validates all SHA-1 hashes and sizes in the archive without rebuilding files. Supports `.zip` archives on the fly.
- **Example:** `wit verify my-app.wit.xml.zip`

### 🔀 `wit diff [archiveA] [archiveB]`
Shows a unified diff (using Myers diff algorithm) of file changes, additions, removals, and content differences between two wit snapshots. Supports `.zip` archives.
- **Memory Optimized**: Runs a two-pass streaming parse, holding only the metadata index in memory and streaming CDATA contents solely for modified files to maintain a minimal memory footprint.
- **Smart Binary Diffing**: Automatically identifies image files (PNG, JPEG, GIF) and prints dimension changes (e.g. `image/png (800x600) -> image/png (1024x768)`) alongside file size comparisons for other binaries.
- **Example:** `wit diff before.wit.xml after.wit.xml`

### 📃 `wit list [archive] [-l]`
Lists all file paths stored in the archive, one per line. Use `-l` for detailed list view. Supports `.zip` archives.
- **Example:** `wit list my-app.wit.xml.zip -l`

### 🩹 `wit patch [archive] [dest_folder]`
Applies only the *changed or new* files from the archive into the target folder, keeping existing matching files untouched. Supports `.zip` archives.
- **Example:** `wit patch after.wit.xml.zip ./my-project`

### 🗺️ `wit meta [archive]`
Inspects the archive metadata (creation time, original root folder name, file count, and message).
- **Example:** `wit meta my-app.wit.xml`

### 💬 `wit msg [archive]`
Displays the custom commit/snapshot message embedded inside the XML.
- **Example:** `wit msg my-app.wit.xml`

---

## 🤖 Guide: Pair Programming with AI

When working with an AI assistant (like ChatGPT, Gemini, or Claude), attach your generated `wit` XML file and feed it the following instructions:

> **System Rules for AI Assistants:**
> 1. You are modifying a project packaged via `wit`. Return the full updated XML archive matching the layout.
> 2. **No formatting whitespace**: Write the `<file>` and `<binary>` boundaries flush with their inner CDATA blocks:
>    `<file path="src/main.go" sha1="xxx" mode="0644" size="123"><![CDATA[content]]></file>`
> 3. **Integrity Validation**: Compute the correct SHA-1 hash (lowercase hex) and size attributes for any file you add or modify.
> 
> *⚠️ Note: Any formatting deviations or mismatching hashes will fail validation checks and abort the rebuild.*


> [!IMPORTANT]
> 
> I built wit because I wanted to give AI my entire codebase without uploading dozens of files or paying for larger context features. It turned into a portable workspace snapshot format that both developers and AI can understand.

---

## 📄 License
This project is open-source under the [MIT License](LICENSE).  
Copyright © 2026 Nolan Stark.
