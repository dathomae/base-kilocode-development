There are general practices that apply to any code written in any language. This file captures some of those. In
addition to these general practices, language specific practices may also apply.

General Practices:
1) The code should be well designed, with clear intent stated plainly in the comments and reflected in class, function, method, attribute and variable names.
2) Comments should concentrate on why things are being done in a certain way, while the code captures the how. Avoid comments that simply reiterate what the code is doing.
3) Complexity needs to pay for itself. Simpler is not always better, provided that additional complexity buys something that's worth the cost of the additional complexity.  Looking at it from the other direction, if there are two approaches that yield the same benefit the simpler one should always be chosen.
4) While some code may be written to support an upcoming story, never write code that *might* be used if there's no current need for it or that there's a clear need for in the future.  Even then, it's better to modify a future story, when possible, to include the implementation of such things (and then implement them when they are needed) than to implement code that isn't needed now unless that makes the implementation too complex. Ask the user when unsure of the best trade-off.
5) Code should almost always have tests implemented or updated when the code is implemented or updated.
6) Code should always follow the style guidelines, when available.

