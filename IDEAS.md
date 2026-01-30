# Future Enhancement Ideas

Ideas for enhancing MarkdownDO using the Charmbracelet ecosystem.

## Visual & Animation Enhancements

### Animated Progress Dashboard
Use **Harmonica** + **Progress** bubbles to create a satisfying "task completion" dashboard:
- Spring-animated progress bar that bounces when tasks are completed
- Show daily/weekly completion percentage with smooth transitions
- "Streak" counter with celebratory animation when you complete all tasks

## New Components & Views

### Kanban Board View
Use **Table** bubble to create a kanban-style view:
```
┌─────────────┬─────────────┬─────────────┐
│   INBOX     │  IN PROGRESS│    DONE     │
├─────────────┼─────────────┼─────────────┤
│ □ Task 1    │ □ Task 3    │ ✓ Task 5    │
│ □ Task 2    │ □ Task 4    │ ✓ Task 6    │
└─────────────┴─────────────┴─────────────┘
```
Navigate horizontally between columns, vertically within them.

### Pomodoro Timer Integration
Use **Timer** + **Stopwatch** + **Spinner** bubbles:
- Start a Pomodoro session for any task
- Visual countdown with customizable work/break intervals
- Task auto-marks as "in progress" during timer
- Satisfying completion animation when timer ends

### File Picker for Multi-Project
Use **File Picker** bubble to:
- Browse and switch between TODO.md files across projects
- Visual tree of your workspace's TODO files
- Quick jump to any project's tasks

## Better Forms & Input

### Huh-Powered Task Creation
Replace the basic text input with **Huh** forms:
```
┌─ New Task ─────────────────────────────────┐
│                                            │
│  Task: ___________________________         │
│                                            │
│  Section: ◉ Inbox  ○ Features  ○ Bugs     │
│                                            │
│  Priority: [ Low ▾ ]                       │
│                                            │
│  Due date: [ None ▾ ]                      │
│                                            │
│         [ Cancel ]  [ Create Task ]        │
└────────────────────────────────────────────┘
```
- Multi-step wizard for complex task creation
- Validation feedback inline
- Themed to match your color scheme

### Quick Capture with Tags
Inline tag parsing with visual pills:
```
Task: Fix login bug @bugs #urgent !tomorrow
       └──────────────┴────────┴──────────┘
        Section: Bugs  Tag: urgent  Due: tomorrow
```

## Theming & Personality

### Theme Gallery
Integrate Huh's built-in themes + custom ones:
- **Dracula** - dark purple aesthetic
- **Catppuccin** - pastel warmth
- **Nord** - cool arctic blues
- **Tokyo Night** - neon cyberpunk
- Live preview when selecting themes

### Celebration Animations
When you complete all tasks in a section:
- Confetti-style ASCII art burst
- Satisfying "ding" sound (optional)
- Streak counter increments with bounce animation

## Power Features

### Split-Pane View
Using **Viewport** bubbles:
```
┌─────────────────────┬─────────────────────┐
│  TASKS              │  TASK DETAIL        │
│                     │                     │
│  □ Fix bug #123     │  ## Fix bug #123    │
│ >□ Add dark mode  <─┤  **Priority:** High │
│  ✓ Update docs      │  **Section:** Bugs  │
│                     │  **Notes:**         │
│                     │  The login form...  │
└─────────────────────┴─────────────────────┘
```

### Inbox Zero Mode
Special view that shows only uncategorized tasks, encouraging you to:
- Process each task one-by-one
- Move to a section or complete
- Celebrate when inbox is empty
