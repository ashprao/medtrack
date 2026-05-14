# MedTrack: An AI Pair Programming Experiment

This document preserves the history of how MedTrack was originally created — as an experiment in AI-driven development. The application has since grown into a maintained project, but this origin story is worth keeping for context.

---

## The Experiment

This project represents a unique experiment in AI pair programming, where MedTrack was built without writing a single line of code manually. Every aspect of development was handled through natural language interaction with Claude (claude-3.5-sonnet) via the Cline extension in VSCode.

### The Goal

Could we create a fully functional, production-quality application solely through AI programming? Go and Fyne UI were chosen as the foundation, and the AI handled everything else — from suggesting SQLite for storage to implementing a clean MVC architecture.

### Key Findings

🌟 **What Worked Well**

1. **Complete Automation**
   - AI suggested architectural patterns
   - Cline handled all file operations
   - Automated testing and debugging
   - Generated comprehensive documentation

2. **Sophisticated Debugging**
   - When the medication listing broke, Claude showed impressive adaptability
   - Evolved from static code analysis to strategic logging
   - Successfully diagnosed and fixed complex issues

⚠️ **Critical Challenges**

The most significant insight came from UI state management issues:
- Forms would break 3–4 times during development
- Each regression triggered multiple AI fix attempts
- While initial development was lightning fast…
- Repeated fixes consumed significant credits
- Time savings were partially offset by regression cycles

This revealed a crucial trade-off: AI's rapid initial development can be offset by recurring issues and credit consumption from repeated fix attempts.

### Development Journey

For a deep dive into the AI pair programming experience, including detailed insights into the challenges, successes, and lessons learned, see the [Development Journey](development-journey.md) documentation.

---

## Looking Forward

While this experiment proved AI-only development is possible, it highlighted both potential and limitations. Success requires smart decisions about when to let AI iterate versus stepping in manually. The future of development might look like this — but with better handling of UI state and more refined AI interaction strategies.
