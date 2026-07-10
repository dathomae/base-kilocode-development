There are certain standard "end of task activities" that should be
performed ONLY AT THE EXPLICIT DIRECTION OF THE USER. NEVER perform
these without direction, the user will want to inspect the state of the
project before proceeding.

1. Execute the tests. If any are failing for any reason (even if the
   failure is due to some cause that is unrelated to the task that was just
   completed), STOP and ask the user what to do next.
2. If all of the tests pass, use the command 'git status' to determine
   which files have been added, deleted or modified.
3. Update any plan or story files that need to be updated. If working on a story
   update the status of the story in the toc.md file.
4. If working on a story that has been completed (all tasks marked "Completed" or "Done"):

- Move the story file from memory-bank/stories/ to memory-bank/stories/finished-stories/
- Update toc.md to move the story entry from the "Active Stories" section to the "Finished Stories" section
- Update the story file link to use the finished-stories/ subdirectory path

5. Using that information plus information on the task just completed,
   compose a brief commit message, review it with the user and then
   perform the command 'git commit -m <commit message>'
