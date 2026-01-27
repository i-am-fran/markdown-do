#!/usr/bin/env node

// src/cli/commands.ts
import { spawn } from "child_process";
import pc from "picocolors";

// src/config/settings.ts
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "fs";
import { homedir } from "os";
import { join } from "path";

// src/config/defaults.ts
var DEFAULT_SETTINGS = {
  theme: "default",
  fullscreen: false,
  showCompleted: true,
  editor: "system"
};
var EDITOR_OPTIONS = [
  { value: "system", label: "System default", hint: "Uses $EDITOR or $VISUAL" },
  { value: "vim", label: "Vim", hint: "Open in Vim" },
  { value: "nano", label: "Nano", hint: "Open in Nano" },
  {
    value: "default-app",
    label: "Default app",
    hint: "Open with system default (Preview, etc.)"
  }
];

// src/config/settings.ts
var CONFIG_DIR = join(homedir(), ".config", "markdowndo");
var CONFIG_FILE = join(CONFIG_DIR, "config.json");
var cachedSettings = null;
function ensureConfigDir() {
  if (!existsSync(CONFIG_DIR)) {
    mkdirSync(CONFIG_DIR, { recursive: true });
  }
}
function getSettings() {
  if (cachedSettings) return cachedSettings;
  try {
    if (existsSync(CONFIG_FILE)) {
      const content = readFileSync(CONFIG_FILE, "utf-8");
      cachedSettings = { ...DEFAULT_SETTINGS, ...JSON.parse(content) };
    } else {
      cachedSettings = { ...DEFAULT_SETTINGS };
    }
  } catch {
    cachedSettings = { ...DEFAULT_SETTINGS };
  }
  return cachedSettings;
}
async function saveSettings(settings) {
  ensureConfigDir();
  writeFileSync(CONFIG_FILE, JSON.stringify(settings, null, 2));
  cachedSettings = settings;
}
async function updateSettings(partial) {
  const current = getSettings();
  const updated = { ...current, ...partial };
  await saveSettings(updated);
}
async function resetSettings() {
  await saveSettings({ ...DEFAULT_SETTINGS });
}

// src/core/finder.ts
import { existsSync as existsSync2 } from "fs";
import { dirname, join as join2 } from "path";
import fg from "fast-glob";
var TODO_PATTERNS = ["**/TODO*.md", "**/todo*.md"];
var DEFAULT_TODO_FILE = "TODO.md";
async function findTodoFiles(cwd = process.cwd(), recursive = false) {
  const pattern = recursive ? TODO_PATTERNS : ["TODO*.md", "todo*.md"];
  const files = await fg(pattern, {
    cwd,
    absolute: true,
    caseSensitiveMatch: false,
    ignore: ["**/node_modules/**", "**/.git/**"]
  });
  return files.map((path) => ({
    path,
    relativePath: path.replace(`${cwd}/`, "")
  }));
}
async function findDefaultTodoFile(cwd = process.cwd()) {
  const defaultPath = join2(cwd, DEFAULT_TODO_FILE);
  if (existsSync2(defaultPath)) {
    return defaultPath;
  }
  const files = await findTodoFiles(cwd, false);
  if (files.length > 0) {
    return files[0].path;
  }
  return defaultPath;
}
async function findTodoFilesInSubdirs(cwd = process.cwd()) {
  const files = await findTodoFiles(cwd, true);
  const byDir = /* @__PURE__ */ new Map();
  for (const file of files) {
    const dir = dirname(file.relativePath);
    const dirKey = dir === "." ? "(root)" : dir;
    const existing = byDir.get(dirKey);
    if (existing) {
      existing.push(file);
    } else {
      byDir.set(dirKey, [file]);
    }
  }
  return byDir;
}

// src/core/todo.ts
import { existsSync as existsSync3 } from "fs";
import { readFile, writeFile } from "fs/promises";

// src/core/task.ts
function createTask(id, text3, status, lineNumber, section = null) {
  return { id, text: text3, status, lineNumber, section };
}
function parseHeaderLine(line) {
  const match = line.match(/^##\s+(.+)$/);
  return match ? match[1].trim() : null;
}
function formatTask(task) {
  const checkboxMap = {
    pending: "[ ]",
    "in-progress": "[/]",
    completed: "[x]"
  };
  return `- ${checkboxMap[task.status]} ${task.text}`;
}
function parseTaskLine(line, lineNumber, id, section = null) {
  const match = line.match(/^(\s*)-\s*\[([ xX/])\]\s*(.*)$/);
  if (!match) return null;
  const marker = match[2].toLowerCase();
  let status = "pending";
  if (marker === "x") {
    status = "completed";
  } else if (marker === "/") {
    status = "in-progress";
  }
  const text3 = match[3].trim();
  return createTask(id, text3, status, lineNumber, section);
}
var SECTION_ALIASES = {
  ff: "Features",
  bb: "Bugs",
  ii: "Ideas",
  ww: "Warnings"
};
function parseTaskInput(input) {
  const match = input.match(/(?:^|\s)@(\w+)\s*$/);
  if (!match) {
    return { text: input.trim(), sectionTag: null };
  }
  const rawTag = match[1];
  const sectionTag = SECTION_ALIASES[rawTag.toLowerCase()] || rawTag;
  const text3 = input.slice(0, match.index).trim();
  return { text: text3, sectionTag };
}

// src/core/todo.ts
var TodoFile = class _TodoFile {
  constructor(filePath, content) {
    this.filePath = filePath;
    if (content !== void 0) {
      this.parse(content);
    }
  }
  lines = [];
  tasks = [];
  sections = [];
  static async load(filePath) {
    if (!existsSync3(filePath)) {
      return new _TodoFile(filePath, "");
    }
    const content = await readFile(filePath, "utf-8");
    return new _TodoFile(filePath, content);
  }
  static async create(filePath) {
    const todoFile = new _TodoFile(filePath, "# TODO\n\n");
    await todoFile.save();
    return todoFile;
  }
  parse(content) {
    this.lines = content.split("\n");
    this.tasks = [];
    this.sections = [];
    let taskId = 1;
    let currentSection = null;
    for (let i = 0; i < this.lines.length; i++) {
      const line = this.lines[i];
      const sectionName = parseHeaderLine(line);
      if (sectionName) {
        currentSection = sectionName;
        this.sections.push({ name: sectionName, lineNumber: i });
        continue;
      }
      const task = parseTaskLine(line, i, taskId, currentSection);
      if (task) {
        this.tasks.push(task);
        taskId++;
      }
    }
  }
  getTasks() {
    return [...this.tasks];
  }
  getTask(id) {
    return this.tasks.find((t) => t.id === id);
  }
  getPendingTasks() {
    return this.tasks.filter((t) => t.status === "pending");
  }
  getInProgressTasks() {
    return this.tasks.filter((t) => t.status === "in-progress");
  }
  getCompletedTasks() {
    return this.tasks.filter((t) => t.status === "completed");
  }
  getSections() {
    return [...this.sections];
  }
  getTasksBySection(sectionName) {
    return this.tasks.filter((t) => t.section === sectionName);
  }
  getTasksGroupedBySection() {
    const groups = /* @__PURE__ */ new Map();
    groups.set(null, []);
    for (const section of this.sections) {
      groups.set(section.name, []);
    }
    for (const task of this.tasks) {
      const sectionTasks = groups.get(task.section);
      if (sectionTasks) {
        sectionTasks.push(task);
      }
    }
    for (const [key, tasks] of groups) {
      if (tasks.length === 0) {
        groups.delete(key);
      }
    }
    return groups;
  }
  addTask(text3) {
    const parsed = parseTaskInput(text3);
    const taskText = parsed.text;
    const newTaskLine = `- [ ] ${taskText}`;
    let lineNumber;
    if (parsed.sectionTag) {
      lineNumber = this.findOrCreateSection(parsed.sectionTag);
    } else {
      lineNumber = this.findInsertPosition();
    }
    this.lines.splice(lineNumber, 0, newTaskLine);
    this.parse(this.lines.join("\n"));
    const task = this.tasks.find((t) => t.text === taskText);
    if (!task) {
      throw new Error("Failed to add task");
    }
    return task;
  }
  findOrCreateSection(sectionTag) {
    const existingSection = this.sections.find(
      (s) => s.name.toLowerCase() === sectionTag.toLowerCase()
    );
    if (existingSection) {
      const sectionTasks = this.tasks.filter(
        (t) => t.section?.toLowerCase() === sectionTag.toLowerCase()
      );
      if (sectionTasks.length > 0) {
        const lastTask = sectionTasks[sectionTasks.length - 1];
        return lastTask.lineNumber + 1;
      }
      return existingSection.lineNumber + 1;
    }
    const sectionHeader = `## ${sectionTag}`;
    let insertPos = this.lines.length;
    while (insertPos > 0 && this.lines[insertPos - 1].trim() === "") {
      insertPos--;
    }
    if (insertPos > 0) {
      this.lines.splice(insertPos, 0, "", sectionHeader);
      return insertPos + 2;
    }
    this.lines.splice(insertPos, 0, sectionHeader);
    return insertPos + 1;
  }
  findInsertPosition() {
    for (let i = 0; i < this.lines.length; i++) {
      const line = this.lines[i];
      if (line.match(/^#\s+/) && !line.startsWith("##")) {
        let insertPos = i + 1;
        while (insertPos < this.lines.length && this.lines[insertPos].trim() === "") {
          insertPos++;
        }
        return insertPos;
      }
    }
    return 0;
  }
  updateTask(id, text3) {
    const task = this.getTask(id);
    if (!task) return false;
    const checkboxMap = {
      pending: " ",
      "in-progress": "/",
      completed: "x"
    };
    const updatedLine = `- [${checkboxMap[task.status]}] ${text3}`;
    this.lines[task.lineNumber] = updatedLine;
    this.parse(this.lines.join("\n"));
    return true;
  }
  toggleTask(id) {
    const task = this.getTask(id);
    if (!task) return false;
    const nextStatus = {
      pending: "in-progress",
      "in-progress": "completed",
      completed: "pending"
    };
    const newStatus = nextStatus[task.status];
    const updatedLine = formatTask({ ...task, status: newStatus });
    this.lines[task.lineNumber] = updatedLine;
    this.parse(this.lines.join("\n"));
    return true;
  }
  setTaskStatus(id, status) {
    const task = this.getTask(id);
    if (!task) return false;
    const updatedLine = formatTask({ ...task, status });
    this.lines[task.lineNumber] = updatedLine;
    this.parse(this.lines.join("\n"));
    return true;
  }
  addNote(text3) {
    const noteLine = `- ${text3}`;
    const lineNumber = this.findOrCreateSection("Notes");
    this.lines.splice(lineNumber, 0, noteLine);
    this.parse(this.lines.join("\n"));
  }
  getSectionNames() {
    return this.sections.map((s) => s.name);
  }
  moveTask(taskId, targetSection) {
    const task = this.getTask(taskId);
    if (!task) return false;
    this.lines.splice(task.lineNumber, 1);
    this.parse(this.lines.join("\n"));
    let insertLine;
    if (targetSection === null) {
      insertLine = this.findInsertPosition();
    } else {
      insertLine = this.findOrCreateSection(targetSection);
    }
    this.lines.splice(insertLine, 0, formatTask(task));
    this.parse(this.lines.join("\n"));
    return true;
  }
  deleteTask(id) {
    const task = this.getTask(id);
    if (!task) return false;
    this.lines.splice(task.lineNumber, 1);
    this.parse(this.lines.join("\n"));
    return true;
  }
  deleteCompletedTasks() {
    const completed = this.getCompletedTasks();
    if (completed.length === 0) return 0;
    const sortedByLine = [...completed].sort(
      (a, b) => b.lineNumber - a.lineNumber
    );
    for (const task of sortedByLine) {
      this.lines.splice(task.lineNumber, 1);
    }
    this.parse(this.lines.join("\n"));
    return completed.length;
  }
  findTasks(keyword) {
    const lower = keyword.toLowerCase();
    return this.tasks.filter((t) => t.text.toLowerCase().includes(lower));
  }
  lint() {
    const issues = [];
    let fixedCount = 0;
    let consecutiveBlankStart = null;
    for (let i = 0; i < this.lines.length; i++) {
      if (this.lines[i].trim() === "") {
        if (consecutiveBlankStart === null) {
          consecutiveBlankStart = i;
        }
      } else {
        if (consecutiveBlankStart !== null) {
          const blankCount = i - consecutiveBlankStart;
          if (blankCount > 1) {
            const extraLines = blankCount - 1;
            this.lines.splice(consecutiveBlankStart + 1, extraLines);
            issues.push({
              line: consecutiveBlankStart + 2,
              issue: `${extraLines} extra blank line${extraLines > 1 ? "s" : ""}`,
              fixed: true
            });
            fixedCount++;
            i -= extraLines;
          }
        }
        consecutiveBlankStart = null;
      }
    }
    if (consecutiveBlankStart !== null) {
      const blankCount = this.lines.length - consecutiveBlankStart;
      if (blankCount > 1) {
        const extraLines = blankCount - 1;
        this.lines.splice(consecutiveBlankStart + 1, extraLines);
        issues.push({
          line: consecutiveBlankStart + 2,
          issue: `${extraLines} extra trailing blank line${extraLines > 1 ? "s" : ""}`,
          fixed: true
        });
        fixedCount++;
      }
    }
    for (let i = 0; i < this.lines.length; i++) {
      const line = this.lines[i];
      const original = line;
      if (line.trim() === "" || line.startsWith("#")) continue;
      const taskPattern = /^(\s*)-\s*\[([^\]]*)\]\s*(.*)$/;
      const match = line.match(taskPattern);
      if (match) {
        const indent = match[1];
        let checkbox = match[2];
        const text3 = match[3];
        let fixed = false;
        if (checkbox === "") {
          checkbox = " ";
          fixed = true;
          issues.push({
            line: i + 1,
            issue: "Empty checkbox",
            fixed: true
          });
        }
        if (checkbox === "X") {
          checkbox = "x";
          fixed = true;
          issues.push({
            line: i + 1,
            issue: "Uppercase X in checkbox",
            fixed: true
          });
        }
        if (checkbox !== " " && checkbox !== "x" && checkbox !== "/") {
          if (checkbox.trim() === "" || checkbox.trim().toLowerCase() === "x") {
            checkbox = checkbox.trim() === "" ? " " : "x";
            fixed = true;
            issues.push({
              line: i + 1,
              issue: "Malformed checkbox content",
              fixed: true
            });
          }
        }
        if (fixed) {
          this.lines[i] = `${indent}- [${checkbox}] ${text3.trim()}`;
          fixedCount++;
        }
      }
    }
    if (fixedCount > 0) {
      this.parse(this.lines.join("\n"));
    }
    return { issues, fixedCount };
  }
  serialize() {
    return this.lines.join("\n");
  }
  reorderTasks() {
    if (this.tasks.length <= 1) return;
    const tasksBySection = this.getTasksGroupedBySection();
    const sortByStatus = (tasks) => {
      const pending = tasks.filter((t) => t.status === "pending");
      const inProgress = tasks.filter((t) => t.status === "in-progress");
      const completed = tasks.filter((t) => t.status === "completed");
      return [...pending, ...inProgress, ...completed];
    };
    const sortedTasks = [];
    const noSection = tasksBySection.get(null) || [];
    sortedTasks.push(...sortByStatus(noSection));
    for (const section of this.sections) {
      const sectionTasks = tasksBySection.get(section.name) || [];
      sortedTasks.push(...sortByStatus(sectionTasks));
    }
    const taskLineNumbers = this.tasks.map((t) => t.lineNumber).sort((a, b) => a - b);
    for (let i = 0; i < sortedTasks.length; i++) {
      const lineNum = taskLineNumbers[i];
      this.lines[lineNum] = formatTask(sortedTasks[i]);
    }
    this.parse(this.lines.join("\n"));
  }
  async save() {
    this.reorderTasks();
    let content = this.serialize();
    if (!content.endsWith("\n")) {
      content += "\n";
    }
    await writeFile(this.filePath, content, "utf-8");
  }
};

// src/cli/commands.ts
function parseArgs(args) {
  const flags = /* @__PURE__ */ new Set();
  const positional = [];
  let value = null;
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg.startsWith("-")) {
      const flag = arg.replace(/^-+/, "");
      flags.add(flag);
      if (i + 1 < args.length && !args[i + 1].startsWith("-")) {
        if (flag === "c" || flag === "d" || flag === "e" || flag === "f" || flag === "fs" || flag === "n" || flag === "t") {
          value = args[i + 1];
          i++;
        }
      }
    } else {
      positional.push(arg);
    }
  }
  const text3 = positional.length > 0 ? positional.join(" ") : null;
  return { text: text3, value, flags };
}
function formatTaskLine(task, showFile, displayId) {
  let checkbox;
  let text3;
  switch (task.status) {
    case "completed":
      checkbox = pc.green("[x]");
      text3 = pc.strikethrough(pc.dim(task.text));
      break;
    case "in-progress":
      checkbox = pc.yellow("[/]");
      text3 = pc.yellow(task.text);
      break;
    default:
      checkbox = pc.dim("[ ]");
      text3 = task.text;
  }
  const id = pc.dim(`${displayId ?? task.id}.`);
  const file = showFile ? pc.dim(` (${showFile})`) : "";
  return `  ${id} ${checkbox} ${text3}${file}`;
}
async function listTasks(recursive = false) {
  const cwd = process.cwd();
  if (recursive) {
    const filesByDir = await findTodoFilesInSubdirs(cwd);
    if (filesByDir.size === 0) {
      console.log(pc.dim("No TODO files found"));
      return;
    }
    let globalTaskId = 1;
    for (const [dir, files] of filesByDir) {
      console.log(pc.bold(pc.cyan(dir)));
      for (const file of files) {
        const todoFile = await TodoFile.load(file.path);
        const tasks = todoFile.getTasks();
        if (tasks.length === 0) {
          console.log(pc.dim("  (no tasks)"));
        } else {
          for (const task of tasks) {
            console.log(formatTaskLine(task, void 0, globalTaskId));
            globalTaskId++;
          }
        }
      }
      console.log();
    }
  } else {
    const filePath = await findDefaultTodoFile(cwd);
    const todoFile = await TodoFile.load(filePath);
    const tasksBySection = todoFile.getTasksGroupedBySection();
    if (tasksBySection.size === 0) {
      console.log(pc.dim("No tasks found"));
      return;
    }
    for (const [section, tasks] of tasksBySection) {
      if (section === null) {
        console.log(pc.bold(pc.cyan("Tasks:")));
      } else {
        console.log();
        console.log(pc.bold(pc.magenta(`## ${section}`)));
      }
      for (const task of tasks) {
        console.log(formatTaskLine(task));
      }
    }
  }
}
async function addTask(text3) {
  const filePath = await findDefaultTodoFile();
  let todoFile = await TodoFile.load(filePath);
  if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === "") {
    todoFile = await TodoFile.create(filePath);
  }
  const task = todoFile.addTask(text3);
  await todoFile.save();
  console.log(pc.green("Added:"), formatTaskLine(task));
}
async function deleteTask(idStr) {
  const id = Number.parseInt(idStr, 10);
  if (Number.isNaN(id)) {
    console.error(pc.red("Invalid task ID"));
    process.exit(1);
  }
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  const task = todoFile.getTask(id);
  if (!task) {
    console.error(pc.red(`Task ${id} not found`));
    process.exit(1);
  }
  todoFile.deleteTask(id);
  await todoFile.save();
  console.log(pc.yellow("Deleted:"), task.text);
}
async function editTask(idStr, newText) {
  const id = Number.parseInt(idStr, 10);
  if (Number.isNaN(id)) {
    console.error(pc.red("Invalid task ID"));
    process.exit(1);
  }
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  if (!todoFile.updateTask(id, newText)) {
    console.error(pc.red(`Task ${id} not found`));
    process.exit(1);
  }
  await todoFile.save();
  const task = todoFile.getTask(id);
  if (task) {
    console.log(pc.green("Updated:"), formatTaskLine(task));
  }
}
async function toggleTask(idStr) {
  const id = Number.parseInt(idStr, 10);
  if (Number.isNaN(id)) {
    console.error(pc.red("Invalid task ID"));
    process.exit(1);
  }
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  if (!todoFile.toggleTask(id)) {
    console.error(pc.red(`Task ${id} not found`));
    process.exit(1);
  }
  await todoFile.save();
  const task = todoFile.getTask(id);
  if (task) {
    const actionMap = {
      pending: "Reopened",
      "in-progress": "Started",
      completed: "Completed"
    };
    console.log(pc.green(`${actionMap[task.status]}:`), formatTaskLine(task));
  }
}
async function completeTask(idStr) {
  const id = Number.parseInt(idStr, 10);
  if (Number.isNaN(id)) {
    console.error(pc.red("Invalid task ID"));
    process.exit(1);
  }
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  if (!todoFile.setTaskStatus(id, "completed")) {
    console.error(pc.red(`Task ${id} not found`));
    process.exit(1);
  }
  await todoFile.save();
  const task = todoFile.getTask(id);
  if (task) {
    console.log(pc.green("Completed:"), formatTaskLine(task));
  }
}
async function addNote(text3) {
  const filePath = await findDefaultTodoFile();
  let todoFile = await TodoFile.load(filePath);
  if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === "") {
    todoFile = await TodoFile.create(filePath);
  }
  todoFile.addNote(text3);
  await todoFile.save();
  console.log(pc.green("Added note:"), text3);
}
async function findTasksByKeyword(keyword, recursive = false) {
  const cwd = process.cwd();
  const files = await findTodoFiles(cwd, recursive);
  if (files.length === 0) {
    console.log(pc.dim("No TODO files found"));
    return;
  }
  let found = false;
  console.log(pc.bold(pc.cyan(`Searching for: "${keyword}"`)));
  console.log();
  for (const file of files) {
    const todoFile = await TodoFile.load(file.path);
    const matches = todoFile.findTasks(keyword);
    if (matches.length > 0) {
      found = true;
      if (recursive) {
        console.log(pc.dim(file.relativePath));
      }
      for (const task of matches) {
        console.log(formatTaskLine(task));
      }
    }
  }
  if (!found) {
    console.log(pc.dim("No matching tasks found"));
  }
}
async function openInEditor() {
  const filePath = await findDefaultTodoFile();
  const settings = getSettings();
  let command;
  let args;
  switch (settings.editor) {
    case "vim":
      command = "vim";
      args = [filePath];
      break;
    case "nano":
      command = "nano";
      args = [filePath];
      break;
    case "default-app":
      command = "open";
      args = [filePath];
      break;
    default:
      command = process.env.EDITOR || process.env.VISUAL || "vim";
      args = [filePath];
      break;
  }
  console.log(pc.dim(`Opening ${filePath}...`));
  const child = spawn(command, args, {
    stdio: "inherit"
  });
  child.on("error", (err) => {
    console.error(pc.red(`Failed to open editor: ${err.message}`));
    process.exit(1);
  });
}
async function deleteCompletedTasks() {
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  const count = todoFile.deleteCompletedTasks();
  if (count === 0) {
    console.log(pc.dim("No completed tasks to delete"));
    return;
  }
  await todoFile.save();
  console.log(
    pc.green(`Deleted ${count} completed task${count === 1 ? "" : "s"}`)
  );
}
async function lintFile() {
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  console.log(pc.dim(`Linting ${filePath}...`));
  console.log();
  const tasks = todoFile.getTasks();
  const pendingCount = tasks.filter((t) => t.status === "pending").length;
  const inProgressCount = tasks.filter((t) => t.status === "in-progress").length;
  const completedCount = tasks.filter((t) => t.status === "completed").length;
  const result = todoFile.lint();
  if (result.issues.length === 0) {
    console.log(pc.green("\u2713 No issues found"));
    console.log();
    console.log(
      pc.dim(
        `Checked ${tasks.length} task${tasks.length === 1 ? "" : "s"} (${pendingCount} pending, ${inProgressCount} in progress, ${completedCount} completed)`
      )
    );
    return;
  }
  console.log(pc.bold(pc.cyan("Lint results:")));
  console.log();
  for (const issue of result.issues) {
    const status = issue.fixed ? pc.green("fixed") : pc.yellow("found");
    console.log(`  Line ${issue.line}: ${issue.issue} [${status}]`);
  }
  console.log();
  if (result.fixedCount > 0) {
    await todoFile.save();
    console.log(
      pc.green(
        `\u2713 Fixed ${result.fixedCount} issue${result.fixedCount === 1 ? "" : "s"}`
      )
    );
  }
  console.log(
    pc.dim(
      `Checked ${tasks.length} task${tasks.length === 1 ? "" : "s"} (${pendingCount} pending, ${inProgressCount} in progress, ${completedCount} completed)`
    )
  );
}
function showHelp() {
  console.log(`
${pc.bold(pc.cyan("mdd"))} - MarkdownDO: Manage TODO.md files

${pc.bold("Usage:")}
  mdd                    Open interactive TUI
  mdd <task text>        Add a new task (quotes optional)
  mdd -l                 List tasks
  mdd -ls                List tasks recursively
  mdd -t <id>            Toggle task status
  mdd -c <id>            Complete task by ID
  mdd -e <id> <text>     Edit task text
  mdd -d <id>            Delete task by ID
  mdd -dc                Delete all completed tasks
  mdd -f <keyword>       Find tasks by keyword
  mdd -fs <keyword>      Find tasks recursively
  mdd -n <text>          Add a note to ## Notes section
  mdd -o                 Open TODO file in editor
  mdd -lint              Lint and fix TODO file formatting
  mdd -h, --help         Show this help

${pc.bold("Examples:")}
  mdd Buy groceries      Add new task (no quotes needed)
  mdd -l                 Show all tasks
  mdd -t 2               Toggle task #2
  mdd -c 1               Complete task #1
  mdd -e 1 Fix the bug   Edit task #1
  mdd -d 3               Delete task #3
  mdd -f bug             Find tasks containing "bug"
`);
}

// src/tui/app.ts
import * as p3 from "@clack/prompts";
import pc5 from "picocolors";

// src/tui/hints.ts
var MENU_FOOTER = "\u2191\u2193 navigate \u2022 enter select \u2022 esc esc quit";
var LIST_FOOTER = "\u2191\u2193 navigate \u2022 enter select \u2022 c complete \u2022 d delete \u2022 e edit \u2022 m move \u2022 esc back";
function getTaskFooter(taskId) {
  return `CLI: mdd -t ${taskId} (toggle) \u2022 mdd -e ${taskId} (edit) \u2022 mdd -d ${taskId} (delete)`;
}

// src/tui/select.ts
import { SelectPrompt, isCancel } from "@clack/core";
import pc2 from "picocolors";
var S_BAR = "\u2502";
var S_BAR_END = "\u2514";
var S_RADIO_ACTIVE = "\u25CF";
var S_RADIO_INACTIVE = "\u25CB";
var S_STEP_ACTIVE = "\u25C6";
var S_STEP_CANCEL = "\u25A0";
var S_STEP_SUBMIT = "\u25C7";
function symbol(state) {
  switch (state) {
    case "initial":
    case "active":
      return pc2.cyan(S_STEP_ACTIVE);
    case "cancel":
      return pc2.red(S_STEP_CANCEL);
    case "submit":
      return pc2.green(S_STEP_SUBMIT);
    default:
      return pc2.cyan(S_STEP_ACTIVE);
  }
}
function renderOption(option, state) {
  const label = option.label ?? String(option.value);
  switch (state) {
    case "disabled":
      return `${pc2.dim(S_RADIO_INACTIVE)} ${pc2.dim(label)}${option.hint ? ` ${pc2.dim(`(${option.hint ?? "disabled"})`)}` : ""}`;
    case "selected":
      return pc2.dim(label);
    case "active":
      return `${pc2.green(S_RADIO_ACTIVE)} ${label}${option.hint ? ` ${pc2.dim(`(${option.hint})`)}` : ""}`;
    case "cancelled":
      return pc2.strikethrough(pc2.dim(label));
    default:
      return `${pc2.dim(S_RADIO_INACTIVE)} ${label}`;
  }
}
async function selectWithFooter(config) {
  return new Promise((resolve) => {
    const prompt = new SelectPrompt({
      options: config.options,
      initialValue: config.initialValue,
      render() {
        const title = `${symbol(this.state)}  ${config.message}`;
        const bar = pc2.cyan(S_BAR);
        const barGray = pc2.gray(S_BAR);
        const barEnd = pc2.cyan(S_BAR_END);
        switch (this.state) {
          case "submit": {
            const selected = renderOption(
              config.options[this.cursor],
              "selected"
            );
            return `${title}
${barGray}  ${selected}`;
          }
          case "cancel": {
            const cancelled = renderOption(
              config.options[this.cursor],
              "cancelled"
            );
            return `${title}
${barGray}  ${cancelled}
${barGray}`;
          }
          default: {
            const opts = config.options.map(
              (opt, i) => renderOption(
                opt,
                opt.disabled ? "disabled" : i === this.cursor ? "active" : "inactive"
              )
            ).join(`
${bar}  `);
            const footer = config.footer ? `
${barEnd}  ${pc2.dim(config.footer)}` : `
${barEnd}`;
            return `${title}
${bar}  ${opts}${footer}
`;
          }
        }
      }
    });
    if (config.hotkeys) {
      prompt.on("key", (key) => {
        if (key && config.hotkeys && key in config.hotkeys) {
          const currentOption = config.options[prompt.cursor];
          if (currentOption && !currentOption.disabled) {
            const promptInternal = prompt;
            promptInternal.state = "submit";
            promptInternal.close();
            resolve({
              action: config.hotkeys[key],
              value: currentOption.value
            });
          }
        }
      });
    }
    prompt.prompt().then((result) => {
      resolve(result);
    });
  });
}
function isHotkeyResult(value) {
  return typeof value === "object" && value !== null && "action" in value && "value" in value;
}

// src/tui/settings.ts
import * as p from "@clack/prompts";
import pc3 from "picocolors";
async function showSettingsMenu() {
  while (true) {
    const settings = getSettings();
    console.log();
    const editorLabel = EDITOR_OPTIONS.find((e) => e.value === settings.editor)?.label || "System default";
    const action = await selectWithFooter({
      message: "Settings",
      options: [
        {
          value: "editor",
          label: `Editor: ${pc3.cyan(editorLabel)}`
        },
        {
          value: "fullscreen",
          label: `Fullscreen mode: ${settings.fullscreen ? pc3.green("on") : pc3.dim("off")}`
        },
        {
          value: "showCompleted",
          label: `Show completed: ${settings.showCompleted ? pc3.green("on") : pc3.dim("off")}`
        },
        {
          value: "theme",
          label: `Theme ${pc3.dim("(coming soon)")}`,
          disabled: true
        },
        {
          value: "reset",
          label: pc3.yellow("Reset to defaults")
        },
        {
          value: "back",
          label: pc3.dim("Back")
        }
      ],
      footer: LIST_FOOTER
    });
    if (isCancel(action) || action === "back") break;
    switch (action) {
      case "editor":
        await selectEditor();
        break;
      case "fullscreen":
        await updateSettings({ fullscreen: !settings.fullscreen });
        p.log.success(
          `Fullscreen mode ${!settings.fullscreen ? "enabled" : "disabled"}`
        );
        break;
      case "showCompleted":
        await updateSettings({ showCompleted: !settings.showCompleted });
        p.log.success(
          `Show completed ${!settings.showCompleted ? "enabled" : "disabled"}`
        );
        break;
      case "reset": {
        const shouldReset = await p.confirm({
          message: "Reset all settings to defaults?"
        });
        if (shouldReset === true) {
          await resetSettings();
          p.log.success("Settings reset");
        }
        break;
      }
    }
  }
}
async function selectEditor() {
  const settings = getSettings();
  const editor = await selectWithFooter({
    message: "Select editor:",
    options: EDITOR_OPTIONS.map((opt) => ({
      value: opt.value,
      label: opt.value === settings.editor ? `${opt.label} ${pc3.green("(current)")}` : opt.label,
      hint: opt.hint
    })),
    footer: LIST_FOOTER
  });
  if (!isCancel(editor)) {
    await updateSettings({ editor });
    p.log.success(
      `Editor set to ${EDITOR_OPTIONS.find((e) => e.value === editor)?.label}`
    );
  }
}

// src/tui/taskList.ts
import * as p2 from "@clack/prompts";
import pc4 from "picocolors";
var TASK_HOTKEYS = {
  c: "complete",
  d: "delete",
  e: "edit",
  m: "move"
};
async function showTaskList(todoFile) {
  const settings = getSettings();
  while (true) {
    const tasksBySection = todoFile.getTasksGroupedBySection();
    if (tasksBySection.size === 0) {
      p2.log.warn("No tasks found");
      return null;
    }
    const options = [];
    let hasVisibleTasks = false;
    for (const [section, tasks] of tasksBySection) {
      const visibleTasks = settings.showCompleted ? tasks : tasks.filter((t) => t.status !== "completed");
      if (visibleTasks.length === 0) continue;
      hasVisibleTasks = true;
      if (section !== null) {
        options.push({
          value: `section:${section}`,
          label: pc4.bold(pc4.magenta(`\u2500\u2500 ${section} \u2500\u2500`)),
          disabled: true
        });
      }
      for (const task of visibleTasks) {
        options.push({
          value: task.id,
          label: formatTaskOption(task)
        });
      }
    }
    if (!hasVisibleTasks) {
      p2.log.warn(
        settings.showCompleted ? "No tasks found" : "No pending tasks (completed tasks hidden)"
      );
      return null;
    }
    console.log();
    const result = await selectWithFooter({
      message: "Select a task:",
      options: [
        ...options,
        { value: "add", label: pc4.green("+ Add new task") },
        { value: -1, label: pc4.dim("Back") }
      ],
      footer: LIST_FOOTER,
      hotkeys: TASK_HOTKEYS
    });
    if (isHotkeyResult(result)) {
      const taskId = result.value;
      const handled = await handleHotkeyAction(todoFile, taskId, result.action);
      if (handled) continue;
      return taskId;
    }
    if (typeof result === "string" && result.startsWith("section:")) {
      continue;
    }
    if (result === "add") {
      return "add";
    }
    return result;
  }
}
async function handleHotkeyAction(todoFile, taskId, action) {
  const task = todoFile.getTask(taskId);
  if (!task) return false;
  switch (action) {
    case "complete": {
      const prevStatus = task.status;
      todoFile.toggleTask(taskId);
      await todoFile.save();
      const messageMap = {
        pending: "Task started",
        "in-progress": "Task completed",
        completed: "Task reopened"
      };
      p2.log.success(messageMap[prevStatus]);
      return true;
    }
    case "delete": {
      const shouldDelete = await p2.confirm({
        message: `Delete "${task.text}"?`
      });
      if (shouldDelete === true) {
        todoFile.deleteTask(taskId);
        await todoFile.save();
        p2.log.success("Task deleted");
      }
      return true;
    }
    case "edit": {
      const newText = await p2.text({
        message: "Enter new text:",
        initialValue: task.text,
        validate: (value) => {
          if (!value.trim()) return "Task cannot be empty";
        }
      });
      if (!p2.isCancel(newText)) {
        todoFile.updateTask(taskId, newText);
        await todoFile.save();
        p2.log.success("Task updated");
      }
      return true;
    }
    case "move": {
      await showMoveToSectionMenu(todoFile, taskId, task);
      return true;
    }
  }
  return false;
}
function formatTaskOption(task) {
  const id = pc4.dim(`${task.id}.`);
  let checkbox;
  let text3;
  switch (task.status) {
    case "completed":
      checkbox = pc4.green("[x]");
      text3 = pc4.strikethrough(pc4.dim(task.text));
      break;
    case "in-progress":
      checkbox = pc4.yellow("[/]");
      text3 = pc4.yellow(task.text);
      break;
    default:
      checkbox = pc4.dim("[ ]");
      text3 = task.text;
  }
  return `${id} ${checkbox} ${text3}`;
}
async function showTaskMenu(todoFile, taskId) {
  if (taskId === -1) return;
  const task = todoFile.getTask(taskId);
  if (!task) return;
  console.log();
  const toggleLabelMap = {
    pending: "Start progress",
    "in-progress": "Mark as complete",
    completed: "Reopen task"
  };
  const action = await selectWithFooter({
    message: `Task: ${task.text}`,
    options: [
      {
        value: "toggle",
        label: toggleLabelMap[task.status]
      },
      { value: "edit", label: "Edit text" },
      { value: "move", label: "Move to section" },
      { value: "delete", label: pc4.red("Delete") },
      { value: "back", label: pc4.dim("Back") }
    ],
    footer: getTaskFooter(taskId)
  });
  if (isCancel(action) || action === "back") return;
  switch (action) {
    case "toggle": {
      const prevStatus = task.status;
      todoFile.toggleTask(taskId);
      await todoFile.save();
      const messageMap = {
        pending: "Task started",
        "in-progress": "Task completed",
        completed: "Task reopened"
      };
      p2.log.success(messageMap[prevStatus]);
      break;
    }
    case "edit": {
      const newText = await p2.text({
        message: "Enter new text:",
        initialValue: task.text,
        validate: (value) => {
          if (!value.trim()) return "Task cannot be empty";
        }
      });
      if (!p2.isCancel(newText)) {
        todoFile.updateTask(taskId, newText);
        await todoFile.save();
        p2.log.success("Task updated");
      }
      break;
    }
    case "move": {
      await showMoveToSectionMenu(todoFile, taskId, task);
      break;
    }
    case "delete": {
      const shouldDelete = await p2.confirm({
        message: "Are you sure you want to delete this task?"
      });
      if (shouldDelete === true) {
        todoFile.deleteTask(taskId);
        await todoFile.save();
        p2.log.success("Task deleted");
      }
      break;
    }
  }
}
async function showMoveToSectionMenu(todoFile, taskId, task) {
  const sections = todoFile.getSectionNames();
  const options = [];
  if (task.section !== null) {
    options.push({ value: null, label: "Inbox (no section)" });
  }
  for (const section of sections) {
    if (section !== task.section) {
      options.push({ value: section, label: section });
    }
  }
  options.push({ value: "__new__", label: pc4.green("+ Create new section") });
  options.push({ value: "__back__", label: pc4.dim("Back") });
  console.log();
  const selected = await selectWithFooter({
    message: "Move to section:",
    options,
    footer: LIST_FOOTER
  });
  if (isCancel(selected) || selected === "__back__") return;
  let targetSection;
  if (selected === "__new__") {
    const newSection = await p2.text({
      message: "Enter section name:",
      validate: (value) => {
        if (!value.trim()) return "Section name cannot be empty";
      }
    });
    if (p2.isCancel(newSection)) return;
    targetSection = newSection.trim();
  } else {
    targetSection = selected;
  }
  todoFile.moveTask(taskId, targetSection);
  await todoFile.save();
  const targetName = targetSection ?? "Inbox";
  p2.log.success(`Moved to ${targetName}`);
}

// src/tui/app.ts
function clearScreen() {
  process.stdout.write("\x1B[2J\x1B[0f");
}
function enterAlternateScreen() {
  process.stdout.write("\x1B[?1049h");
  process.stdout.write("\x1B[2J\x1B[0f");
}
function exitAlternateScreen() {
  process.stdout.write("\x1B[?1049l");
}
function showTitle() {
  console.log(pc5.cyan("\u250C") + "  " + pc5.bgCyan(pc5.black(" MarkdownDO ")));
}
var DOUBLE_ESC_WINDOW_MS = 500;
var lastEscTime = 0;
async function runTui() {
  const settings = getSettings();
  let useAlternateScreen = settings.fullscreen;
  if (useAlternateScreen) {
    enterAlternateScreen();
    const cleanup = () => {
      exitAlternateScreen();
      process.exit(0);
    };
    process.on("SIGINT", cleanup);
    process.on("SIGTERM", cleanup);
  }
  console.log();
  showTitle();
  while (true) {
    const action = await showMainMenu();
    if (action === "quit") {
      if (useAlternateScreen) {
        exitAlternateScreen();
      }
      p3.outro(pc5.dim("Goodbye!"));
      break;
    }
    if (isCancel(action)) {
      const now = Date.now();
      if (now - lastEscTime < DOUBLE_ESC_WINDOW_MS) {
        if (useAlternateScreen) {
          exitAlternateScreen();
        }
        p3.outro(pc5.dim("Goodbye!"));
        break;
      }
      lastEscTime = now;
      p3.log.info("Press ESC again to quit");
      continue;
    }
    lastEscTime = 0;
    console.log();
    await handleAction(action);
    const currentSettings = getSettings();
    if (currentSettings.fullscreen && !useAlternateScreen) {
      enterAlternateScreen();
      useAlternateScreen = true;
    } else if (!currentSettings.fullscreen && useAlternateScreen) {
      exitAlternateScreen();
      useAlternateScreen = false;
    }
    if (useAlternateScreen) {
      clearScreen();
      showTitle();
    }
  }
}
async function showMainMenu() {
  const filePath = await findDefaultTodoFile();
  const todoFile = await TodoFile.load(filePath);
  const tasks = todoFile.getTasks();
  const pending = tasks.filter((t) => !t.completed).length;
  const taskSummary = tasks.length > 0 ? pc5.dim(` (${pending} pending, ${tasks.length} total)`) : pc5.dim(" (no tasks)");
  return selectWithFooter({
    message: "What would you like to do?",
    options: [
      { value: "list", label: `List tasks${taskSummary}` },
      { value: "add", label: "Add new task" },
      { value: "open", label: "Open in editor" },
      { value: "subfolders", label: "List subfolders" },
      { value: "settings", label: "Settings" },
      { value: "quit", label: pc5.dim("Quit") }
    ],
    footer: MENU_FOOTER
  });
}
async function handleAction(action) {
  switch (action) {
    case "list":
      await handleListTasks();
      break;
    case "add":
      await handleAddTask();
      break;
    case "open":
      await openInEditor();
      break;
    case "subfolders":
      await handleSubfolders();
      break;
    case "settings":
      await showSettingsMenu();
      break;
  }
}
async function handleListTasks() {
  const filePath = await findDefaultTodoFile();
  let todoFile = await TodoFile.load(filePath);
  while (true) {
    const selectedTask = await showTaskList(todoFile);
    if (selectedTask === null || isCancel(selectedTask) || selectedTask === -1) {
      break;
    }
    if (selectedTask === "add") {
      await handleQuickAddTask(todoFile);
      todoFile = await TodoFile.load(filePath);
      continue;
    }
    await showTaskMenu(todoFile, selectedTask);
    todoFile = await TodoFile.load(filePath);
  }
}
async function handleQuickAddTask(todoFile) {
  const text3 = await p3.text({
    message: "Enter task description:",
    placeholder: 'Buy groceries or "Fix bug @Features" to add to section'
  });
  if (p3.isCancel(text3)) return;
  const trimmed = typeof text3 === "string" ? text3.trim() : "";
  if (!trimmed) {
    p3.log.info("Empty input - cancelled");
    return;
  }
  const task = todoFile.addTask(trimmed);
  await todoFile.save();
  p3.log.success(`Added: ${task.text}`);
}
async function handleAddTask() {
  const filePath = await findDefaultTodoFile();
  let todoFile = await TodoFile.load(filePath);
  if (todoFile.getTasks().length === 0 && todoFile.serialize().trim() === "") {
    todoFile = await TodoFile.create(filePath);
  }
  while (true) {
    const text3 = await p3.text({
      message: "Enter task description (empty to finish):",
      placeholder: 'Buy groceries or "Fix bug @Features" to add to section'
    });
    if (p3.isCancel(text3)) break;
    const trimmed = typeof text3 === "string" ? text3.trim() : "";
    if (!trimmed) {
      p3.log.info("Empty input - returning to menu");
      break;
    }
    const task = todoFile.addTask(trimmed);
    await todoFile.save();
    p3.log.success(`Added: ${task.text}`);
    console.log();
  }
}
async function handleSubfolders() {
  const filesByDir = await findTodoFilesInSubdirs();
  if (filesByDir.size === 0) {
    p3.log.warn("No TODO files found in subfolders");
    return;
  }
  const options = Array.from(filesByDir.entries()).map(([dir, files]) => ({
    value: files[0].path,
    label: `${dir} (${files.length} file${files.length > 1 ? "s" : ""})`
  }));
  const selected = await selectWithFooter({
    message: "Select a folder:",
    options: [...options, { value: "back", label: pc5.dim("Back") }],
    footer: LIST_FOOTER
  });
  if (isCancel(selected) || selected === "back") return;
  const filePath = selected;
  let todoFile = await TodoFile.load(filePath);
  while (true) {
    const selectedTask = await showTaskList(todoFile);
    if (selectedTask === null || isCancel(selectedTask) || selectedTask === -1) {
      break;
    }
    if (selectedTask === "add") {
      await handleQuickAddTask(todoFile);
      todoFile = await TodoFile.load(filePath);
      continue;
    }
    await showTaskMenu(todoFile, selectedTask);
    todoFile = await TodoFile.load(filePath);
  }
}

// src/index.ts
async function main() {
  const args = process.argv.slice(2);
  const { text: text3, value, flags } = parseArgs(args);
  if (flags.has("h") || flags.has("help")) {
    showHelp();
    return;
  }
  if (flags.has("l")) {
    await listTasks(false);
    return;
  }
  if (flags.has("ls")) {
    await listTasks(true);
    return;
  }
  if (flags.has("d") && value) {
    await deleteTask(value);
    return;
  }
  if (flags.has("dc")) {
    await deleteCompletedTasks();
    return;
  }
  if (flags.has("lint")) {
    await lintFile();
    return;
  }
  if (flags.has("e") && value && text3) {
    await editTask(value, text3);
    return;
  }
  if (flags.has("t") && value) {
    await toggleTask(value);
    return;
  }
  if (flags.has("c") && value) {
    await completeTask(value);
    return;
  }
  if (flags.has("n") && value) {
    const noteText = text3 ? `${value} ${text3}` : value;
    await addNote(noteText);
    return;
  }
  if (flags.has("f") && value) {
    await findTasksByKeyword(value, false);
    return;
  }
  if (flags.has("fs") && value) {
    await findTasksByKeyword(value, true);
    return;
  }
  if (flags.has("o")) {
    await openInEditor();
    return;
  }
  if (text3 && !flags.size) {
    await addTask(text3);
    return;
  }
  await runTui();
}
main().catch((err) => {
  console.error("Error:", err.message);
  process.exit(1);
});
//# sourceMappingURL=index.js.map