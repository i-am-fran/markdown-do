# Upgrading MarkdownDO

## Version 1.1.0

### What's New in 1.1.0

This release includes several new features and improvements:

**New Features:**
- **Ctrl+N keyboard shortcut** - Press `Ctrl+N` from anywhere in the TUI to quickly create a new task
- **Bulk task completion** - New `-cm` flag to complete multiple tasks at once: `mdd -cm 1 2 3`
- **Cleaner CLI output** - When adding tasks via CLI, output now shows just the task text without checkbox/number

**UI Improvements:**
- Menu reorganized with "Add note" moved lower for better workflow
- Removed redundant "Back" button from task list (use ESC instead)
- Removed redundant "Quit" button from main menu (use ESC ESC instead)

**Bug Fixes:**
- Various improvements to code quality and maintainability

---

## How to Upgrade from 1.0.0 to 1.1.0

### Option 1: Using `make install` (Recommended)

If you installed from source, this is the simplest method:

```bash
cd /path/to/markdown-do
git pull origin main
make install
```

This will:
1. Pull the latest code
2. Build the new version
3. Install it to your `$GOPATH/bin` (or `~/go/bin`) directory
4. Automatically replace the old version

**Verify the installation:**
```bash
mdd -v
# Should output: markdown-do v1.1.0
```

### Option 2: Using `go install`

If you installed using `go install`, simply run:

```bash
go install github.com/i-am-fran/markdowndo/cmd/mdd@latest
```

This will automatically fetch and install the latest version, replacing the old one.

**Verify the installation:**
```bash
mdd -v
# Should output: markdown-do v1.1.0
```

### Option 3: Manual Installation

If you have the binary in a custom location:

```bash
cd /path/to/markdown-do
git pull origin main
make build
```

Then copy the new binary to replace your old one:
```bash
# Find where your old mdd is installed
which mdd

# Copy the new binary there (example, adjust path as needed)
cp ./build/mdd /usr/local/bin/mdd
# or
cp ./build/mdd ~/bin/mdd
# or wherever your old binary was located
```

**Verify the installation:**
```bash
mdd -v
# Should output: markdown-do v1.1.0
```

### Option 4: Download Pre-built Binary

Download the latest binary for your platform from the [releases page](https://github.com/i-am-fran/markdown-do/releases) and replace your old binary.

---

## Troubleshooting

### Still seeing old version after upgrade?

1. **Check which binary is being used:**
   ```bash
   which mdd
   ```

2. **Check if you have multiple installations:**
   ```bash
   # On Linux/macOS
   find ~ -name "mdd" -type f 2>/dev/null
   ```

3. **Make sure your PATH includes the install location:**
   ```bash
   echo $PATH
   ```

4. **Try using the full path:**
   ```bash
   ~/go/bin/mdd -v
   ```

### Old binary still running?

If you have the TUI open, you'll need to close it and reopen it to use the new version. The CLI commands will use the new version immediately.

### Permission denied when copying?

You may need to use `sudo` if copying to a system directory:
```bash
sudo cp ./build/mdd /usr/local/bin/mdd
```

---

## Rollback (if needed)

If you encounter issues with 1.1.0 and need to rollback:

```bash
cd /path/to/markdown-do
git checkout v1.0.0  # or the specific version you want
make install
```

Or reinstall the old version:
```bash
go install github.com/i-am-fran/markdowndo/cmd/mdd@v1.0.0
```

---

## Configuration

Your configuration file (`~/.config/markdowndo/config.json`) and TODO files are **not affected** by the upgrade. All your tasks and settings will remain intact.

---

## Need Help?

If you encounter any issues during the upgrade, please [open an issue](https://github.com/i-am-fran/markdown-do/issues) on GitHub.
