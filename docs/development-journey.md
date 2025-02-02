# My AI Pair Programming Journey: Building MedTrack

I wanted to document my experience building MedTrack using AI pair programming with Claude (specifically claude-3.5-sonnet) via the Cline extension in VSCode. What made this journey particularly interesting was that I never had to manually create or edit any files - everything was handled by Cline executing Claude's suggestions.

The main objective of this project was to explore whether it's possible to create a fully functional, potentially useful application solely through AI programming - including all coding, debugging, and testing. I wanted to see how far we could get with zero manual coding intervention.

## How I Approached Development

The most impressive aspect of this project was the complete hands-off development approach:

- **True AI Pair Programming**: 
  - Claude suggested the architecture and project structure
  - Cline automatically created all directories and files
  - Every code change was executed by Cline
  - Even running and testing the app was handled by Cline
  - I just had to review and guide the process

- **Just Having a Conversation**: Instead of crafting complex prompts or manually implementing suggestions, I simply talked to Claude like I would with any other developer. It was amazing to see how we could build a complex app through simple conversation while Cline handled all the actual implementation.

- **Planning Then Doing**: I made good use of Cline's Plan/Act modes:
  - Used Plan mode when I needed to think through requirements and architecture
  - Switched to Act mode for actual implementation and testing
  - This helped me stay focused while still being able to step back and plan when needed

## The Technical Side

I chose Go and the Fyne UI toolkit as the foundation, and the AI filled in the rest of the tech stack:
- Go programming language (my choice)
- Fyne UI toolkit for the desktop interface (my choice)
- SQLite for storing data locally (AI's suggestion)
- A clean MVC-like architecture (AI's suggestion)

What impressed me most was how Claude took these base choices and suggested the entire project structure, which Cline then created automatically - from setting up the basic directory structure to creating individual files with proper Go package organization. I never had to manually create a single directory or file.

## The Challenges I Faced

### The UI Layout Nightmare
This was probably the most frustrating part of the whole experience, and it really tested Claude's problem-solving capabilities:

- **The Problem**: The form layout would break when changes were made to completely unrelated parts of the app
- **What It Looked Like**: The forms would do this annoying double-render:
  1. First showing just the labels
  2. Then the complete form below it
- **The Impact**:
  - This happened 3-4 times during development
  - Each time, Claude would attempt multiple fixes
  - The AI would spend time refactoring code to find the right sequence
  - Burned through a lot of credits as Claude tried different approaches
  - Any time saved in the initial development was eaten up by these fixes
- **The Process**:
  - Each time it broke, Claude would start a new cycle of fixes
  - The AI would try different approaches and would repeat try the same approaches during every regression
  - Refactor the code until it worked
  - Then watch it break again with other changes
- **What I Learned**: While I could have fixed it faster manually, I wanted to see if Claude could solve it independently as part of the experiment in fully AI-driven development

### Was It Worth It?
Looking at the trade-offs in this experiment:
- **The Good**: 
  - Initial development was super fast and efficient
  - Never had to manually create or edit files
  - Complete automation of all implementation tasks
  - AI suggested and Cline executed everything
  - Proved that complex apps can be built entirely by AI
- **The Bad**: 
  - Spent too much time watching Claude attempt repetitive fixes
  - Used up more credits than necessary
  - Got really frustrating at times
  - The AI kept trying similar approaches to the same problems
- **Bottom Line**: The experiment showed that while fully AI-driven development is possible, it still has some limitations, particularly with complex UI state management

### What Worked Well

1. **Complete Automation**
   - AI suggested all architectural decisions
   - Cline handled all file and directory creation
   - Code changes were automatically implemented
   - Even running the app was automated

2. **Keeping Things Consistent**
   - Claude was great at maintaining patterns
   - Helped me stick to Go best practices
   - Set up good concurrent access patterns
   - Established solid error handling

3. **Quick Development**
   - Features came together quickly
   - AI handled all debugging
   - Code generation was fast
   - No manual file handling needed

4. **Documentation**
   - Generated clear docs automatically
   - Added good code comments
   - Created solid README files

5. **Impressive Debugging Capabilities**
   - One incident really showed the AI's problem-solving evolution
   - When the medication listing wasn't displaying properly, Claude first tried static code analysis
   - After that approach didn't yield results, it switched tactics
   - Started adding strategic log statements throughout the code
   - Used the running app's log output to diagnose the issue
   - Successfully identified and fixed the problem
   - Demonstrated ability to adapt debugging strategies when initial approaches fail

## What I Learned

1. **What AI (claude-3.5-sonnet) is Good At**
   - Great at suggesting and implementing architecture
   - Good at initial UI layout
   - Excellent at automating repetitive tasks
   - Struggles with UI state
   - Tends to repeat itself when stuck

2. **Development Process**
   - It's possible to achieve objectives with simple prompts
   - Plan/Act mode switching really helps
   - AI can handle testing and debugging
   - UI state needs special attention
   - Sometimes it's okay to step in manually
   - More sophisticated prompting might have helped avoid some issues

3. **Best Practices I Picked Up**
   - Let AI handle the repetitive tasks
   - Trust but verify automated changes
   - Let the AI attempt fixes independently
   - Know when to step in manually
   - Consider experimenting with different prompting strategies

## What Could Be Better

Things I'd improve next time:

1. **State Management**
   - Need better UI state handling
   - Better form layout preservation
   - Ways to prevent regressions
   - More efficient approaches to recurring issues

2. **Testing**
   - More unit tests
   - Automated UI testing
   - Tests specifically for layouts

3. **Code Structure**
   - More modular views
   - Better separation of concerns
   - Isolated UI state management

4. **AI Interaction**
   - Experiment with more sophisticated prompting
   - Better strategies for avoiding recurring issues
   - More structured approach to problem-solving

## Looking Back

This experiment in building MedTrack showed me the incredible potential of true AI pair programming. Using claude-3.5-sonnet and Cline, I was able to create a complete, functional application without writing a single line of code manually. The AI handled everything from architecture to debugging, and Cline executed all the changes.

While the automation was fantastic for initial development and feature implementation, the recurring UI issues highlighted where this approach still needs improvement. The experiment proved that while fully AI-driven development is possible, you need to be smart about when to let the AI work through problems and when to step in manually. While we achieved our goals with simple prompts, there's room to explore whether more sophisticated prompting strategies could help avoid some of the challenges we encountered. The future of development might look like this - but with better handling of tricky state management issues and more refined AI interaction strategies.
