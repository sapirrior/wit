# wit CLI Technical Command Specification

Authoritative Repository: [github.com/sapirrior/wit](https://github.com/sapirrior/wit)  
Issue Tracker: [github.com/sapirrior/wit/issues](https://github.com/sapirrior/wit/issues)

This document contains the low-level CLI command specifications, exit codes, and flag definitions for the `wit` tool.

---

## Command Routing & Verbs

The `wit` entrypoint routes commands based on the first argument (`args[1]`).

### 1. `wit snap <folder>` (Directory Packaging)
Explicitly packages a folder structure into an XML snapshot file.
- **Syntax**: `wit snap <folder> [-m "message"] [-o <output_file>]`
- **Fallback**: Running `wit <folder>` also executes a snapshot if the target argument resolves to a valid directory.
- **Parameters**:
  - `-m` (string): Snapshot message, written to the `<wit message="...">` attribute.
  - `-o` (string): Custom destination path. Defaults to `witFile.xml` (or `witFile-n.xml` if conflicts arise).
- **Exit Codes**:
  - `0`: Success.
  - `1`: Target not found, not a directory, or file write failure.

### 2. `wit grab <archive>` (Workspace Reconstruction)
Rebuilds a directory tree from an XML archive.
- **Syntax**: `wit grab <archive> [-o <destination_dir>]`
- **Parameters**:
  - `-o` (string): Rebuild output directory. Defaults to the `root` attribute specified in the XML.
- **Safety Features**:
  - Rebuilding will exit with `1` if the archive contains paths starting with `..` or absolute paths (Path Traversal Protection).
  - Verifies computed SHA-1 and file size against the XML tags before writing.
- **Exit Codes**:
  - `0`: Success (all files rebuilt, verified, and permissions applied).
  - `1`: Size/SHA-1 mismatches, formatting exceptions, or target file creation failures.

### 3. `wit msg <archive>` (Message Extraction)
Reads and displays the snapshot message embedded inside the archive's header.
- **Syntax**: `wit msg <archive>`
- **Exit Codes**:
  - `0`: Success (displays message or notes no message exists).
  - `1`: File not found or invalid XML.

### 4. `wit meta <archive>` (Metadata & Structure Viewer)
Prints metadata info and lists all files packed within the archive.
- **Syntax**: `wit meta <archive>`
- **Output format**:
  ```
  Archive Metadata for: my_archive.xml

  Root Directory: app
  Format Version: 2.0.1
  Created At:     2026-07-17T11:34:17Z
  File Count:     10
  Message:        Technical commit

  Files:
    [text]    src/main.go  (size: 1054, mode: 0644)
    [binary]  assets/logo.png  (size: 14560, mode: 0644)
  ```
- **Exit Codes**:
  - `0`: Success.
  - `1`: File not found or invalid XML.
