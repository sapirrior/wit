# wit CLI Technical Command Specification

Authoritative Repository: [github.com/sapirrior/wit](https://github.com/sapirrior/wit)  
Issue Tracker: [github.com/sapirrior/wit/issues](https://github.com/sapirrior/wit/issues)

This document contains the low-level CLI command specifications, exit codes, and flag definitions for the `wit` tool.

---

## Command Routing & Verbs

The `wit` entrypoint routes commands based on the first argument (`args[1]`).

### 1. `wit snap <folder>` (Directory Packaging)
Explicitly packages a folder structure into an XML snapshot file.
- **Syntax**: `wit snap <folder> [options]`
- **Fallback**: Running `wit <folder> [options]` also executes a snapshot if the target argument resolves to a valid directory.
- **Parameters**:
  - `-m` (string): Snapshot message, written to the `<wit message="...">` attribute.
  - `-o` (string): Custom destination path. Automatically appends `.wit.xml`. Defaults to `file.wit.xml`.
  - `-c`: Compresses the snapshot output into `<name>.wit.xml.zip` containing the `.wit.xml`.
  - `--exclude <pattern>`: Excludes files matching the glob pattern (can be specified multiple times).
  - `--max-size <size>`: Ignores files exceeding the size limit (supports units like B, KB, MB, GB).
- **Exit Codes**:
  - `0`: Success.
  - `1`: Target not found, not a directory, size limit parsing error, or file write failure.

### 2. `wit grab <archive>` (Workspace Reconstruction)
Rebuilds a directory tree from an XML archive. Automatically detects and decompresses `.zip` archives.
- **Syntax**: `wit grab <archive> [-o <destination_dir>]`
- **Parameters**:
  - `-o` (string): Rebuild output directory. Defaults to the `root` attribute specified in the XML.
- **Safety Features**:
  - Rebuilding will exit with `1` if the archive contains paths starting with `..` or absolute paths (Path Traversal Protection).
  - Verifies computed SHA-1 and file size against the XML tags before writing.
- **Exit Codes**:
  - `0`: Success (all files rebuilt, verified, and permissions applied).
  - `1`: Size/SHA-1 mismatches, formatting exceptions, zip decompression errors, or target file creation failures.

### 3. `wit verify <archive>` (Integrity Verification)
Validates all SHA-1 hashes and sizes in the archive without rebuilding any files. Supports `.zip` archives.
- **Syntax**: `wit verify <archive>`
- **Exit Codes**:
  - `0`: Success (all files matching).
  - `1`: Integrity failure (hash or size mismatches, file reading errors).

### 4. `wit diff <archiveA> <archiveB>` (Snapshot Differencing)
Shows a unified diff (using Myers diff algorithm) of file modifications, additions, and removals between two wit snapshots. Supports `.zip` archives.
- **Implementation**: Runs in a two-pass stream. Pass 1 reads only the tag attributes and skips CDATA blocks using `decoder.Skip()`. Pass 2 streams the file content for modified files only, reducing RAM footprint.
- **Smart Binary Diff**: Detects registered image formats (PNG, JPEG, GIF) via the standard `image` library to print dimension changes, falling back to size comparison for other binaries.
- **Syntax**: `wit diff <archiveA> <archiveB>`
- **Exit Codes**:
  - `0`: Success.
  - `1`: File not found or invalid XML.

### 5. `wit list <archive>` (File List Printer)
Lists all files in the archive by showing their `identity` hash and relative path (separated by space). Use `-l` for detailed list view (shows type, size, mode, and full identity). Supports `.zip` archives.
- **Syntax**: `wit list <archive> [-l]`
- **Exit Codes**:
  - `0`: Success.
  - `1`: File not found or invalid XML.

### 6. `wit glance <archive> <identity>` (File Content Inspection)
Inspects a specific file inside the archive by its identity SHA-1 hash. It displays its metadata (path, type, size, mode, identity) and prints its text content. Supports prefix matching (minimum 4 characters) and `.zip` archives.
- **Syntax**: `wit glance <archive> <identity>` or `wit glance <archive> -i <identity>`
- **Exit Codes**:
  - `0`: Success.
  - `1`: Archive not found, invalid XML, or identity hash not found.

### 7. `wit patch <archive> [dest_folder]` (Partial Workspace Application)
Applies only the changed or new files from the archive into the target folder, leaving identical files untouched. Supports `.zip` archives.
- **Syntax**: `wit patch <archive> [dest_folder]`
- **Exit Codes**:
  - `0`: Success.
  - `1`: File not found or write errors.

### 8. `wit msg <archive>` (Message Extraction)
Reads and displays the snapshot message embedded inside the archive's header.
- **Syntax**: `wit msg <archive>`
- **Exit Codes**:
  - `0`: Success.
  - `1`: File not found or invalid XML.

### 9. `wit meta <archive>` (Metadata & Structure Viewer)
Prints metadata info and lists all files packed within the archive.
- **Syntax**: `wit meta <archive>`
- **Output format**:
  ```
  Archive Metadata for: my_archive.xml

  Root Directory: app
  Format Version: 3.0
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
