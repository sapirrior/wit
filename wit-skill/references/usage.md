# wit CLI Usage Reference

This document lists the available commands and option patterns for the `wit` snapshot utility.

## Commands

### 1. Create a Snapshot (Snap)
Captures a directory tree into an XML archive file.
- **Usage**: `wit <folder> [-m "message"] [-o <output_file>]`
- **Fallback**: If no command is provided, `wit` defaults to snapping if the argument is a directory.
- **Example**: `wit src -m "Initial commit" -o backup.xml`

### 2. Rebuild from Snapshot (Grab)
Extracts and restores the directory structure from an XML archive.
- **Usage**: `wit grab <archive> [-o <output_dir>]`
- **Example**: `wit grab backup.xml -o restore_dir`

### 3. View Snapshot Message (Msg)
Displays the snapshot message saved inside the archive.
- **Usage**: `wit msg <archive>`
- **Example**: `wit msg backup.xml`

### 4. View Metadata and File List (Meta)
Displays detailed metadata (created at, version, original root directory) and lists all files contained in the archive.
- **Usage**: `wit meta <archive>`
- **Example**: `wit meta backup.xml`
