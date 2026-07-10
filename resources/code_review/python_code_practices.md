# Python Code Review Guidelines

This document provides comprehensive code review guidelines based on PEP 8 (Style Guide for Python Code). These guidelines ensure code readability, maintainability, and consistency across Python projects.

## General Approach and Objectives

The primary goal of these guidelines is to improve code readability and consistency. As noted in PEP 8, "code is read much more often than it is written," making readability the paramount concern. These guidelines support several key objectives:

**Consistency Hierarchy**: Maintain consistency at multiple levels - with this style guide first, then within a project, then within a module or function. Consistency within a single module or function is most critical.

**Readability Over Rigidity**: While these guidelines provide structure, they should enhance rather than diminish code readability. When following a guideline would make code less readable, it's appropriate to deviate with good judgment.

**Maintainability**: Clear formatting, logical organization, and consistent patterns make code easier to maintain, debug, and extend over time.

**Team Collaboration**: Consistent style reduces cognitive overhead when switching between different parts of a codebase, enabling more effective team collaboration.

**Backwards Compatibility**: Don't break existing functionality or backwards compatibility solely to comply with style guidelines. When in doubt, prioritize working code over perfect formatting.

## Practices to Follow

| Practice | Explanation and Rationale |
|----------|---------------------------|
| Use 4 spaces per indentation level | Four spaces provide clear visual hierarchy without excessive horizontal space consumption. This is Python's standard and ensures consistency across all Python projects and tools. |
| Limit lines to 79 characters | Short lines enable side-by-side file viewing, improve readability, and work well with code review tools. Use 72 characters for comments and docstrings to account for indentation in documentation tools. |
| Use implicit line continuation | Prefer parentheses, brackets, and braces for line continuation over backslashes. This creates cleaner, more readable multi-line expressions: `result = (long_expression + another_expression)` |
| Break before binary operators | Place operators at the beginning of continuation lines, not the end. This follows mathematical convention and makes it easier to scan operators: `total = (income + taxable_interest + qualified_dividends)` |
| Use proper hanging indents | When using hanging indents, ensure no arguments appear on the first line and add extra indentation to distinguish continuation lines from the following suite of code. |
| Group imports logically | Organize imports in three groups separated by blank lines: (1) standard library, (2) related third party, (3) local application/library imports. This creates predictable import organization. |
| Place imports on separate lines | Write `import os` and `import sys` on separate lines, not `import os, sys`. This improves readability and makes version control diffs cleaner when imports change. |
| Use absolute imports | Absolute imports are more readable and better behaved: `import mypkg.sibling` rather than `from . import sibling`. They provide clear module paths and work better with import tools. |
| Surround top-level definitions with blank lines | Use two blank lines before and after function and class definitions at module level. Use one blank line around method definitions within classes. This creates clear visual separation. |
| Be consistent with string quotes | Choose either single or double quotes and stick with it project-wide. When a string contains quote characters, use the opposite type to avoid backslashes and improve readability. |
| Use double quotes for docstrings | Always use double quotes for triple-quoted strings (docstrings) to maintain consistency with PEP 257 docstring conventions. |
| Place single spaces around binary operators | Use one space before and after assignment (=), comparison (==, <, >, !=), and Boolean operators (and, or, not). This improves readability without excessive spacing. |
| Follow operator precedence in spacing | When mixing operators with different priorities, consider adding space around lower-priority operators: `c = (a+b) * (x+y)`. Use judgment and maintain consistency. |
| Use UTF-8 encoding | Always use UTF-8 for Python files. No encoding declaration is needed for UTF-8 in Python 3, keeping files clean and ensuring universal compatibility. |
| Apply proper function annotation spacing | Use normal colon spacing rules and spaces around the `->` arrow: `def func(x: int) -> str:`. When combining annotations with default values, use spaces around `=`. |

## Things to Avoid

| Anti-Pattern | Explanation and Why to Avoid |
|--------------|-------------------------------|
| Mixing tabs and spaces | Python disallows mixing tabs and spaces for indentation. This creates invisible formatting errors and inconsistent indentation that can cause IndentationError or unexpected behavior. |
| Exceeding line length limits | Lines longer than 79 characters (72 for comments) are harder to read and don't work well with side-by-side viewing or various development tools. Long lines force horizontal scrolling and reduce readability. |
| Breaking after binary operators | The old style of breaking after operators (`total = income +\n    taxable_interest +`) separates operators from their operands, making code harder to scan and understand. |
| Using wildcard imports | `from module import *` pollutes the namespace, makes it unclear which names are available, and confuses both readers and automated tools. It can lead to unexpected behavior and name conflicts. |
| Trailing whitespace | Trailing whitespace is invisible, can cause version control noise, and serves no purpose. Many editors highlight it as an error, making code appear messy and unprofessional. |
| Excessive whitespace in expressions | Avoid spaces inside parentheses, brackets, or braces (`spam( ham[ 1 ], { eggs: 2 } )`), before commas/semicolons/colons, or before function call parentheses (`func ()`). This creates visual clutter without benefit. |
| Aligning assignments for readability | Don't add extra spaces to align assignment operators vertically (`x             = 1`). This creates maintenance overhead when variable names change and provides no real readability benefit. |
| Inconsistent spacing around operators | Avoid mixing spacing styles within the same codebase. If you use spaces around `=` in one place, use them consistently throughout. Inconsistency creates cognitive overhead for readers. |
| Using backslashes for line continuation | When parentheses, brackets, or braces can provide implicit line continuation, avoid backslashes (`\`). Implicit continuation is cleaner and less error-prone than explicit backslash continuation. |
| Putting multiple imports on one line | `import sys, os` should be split into separate lines. Combined imports make it harder to see version control changes, organize imports, and understand dependencies at a glance. |
| Ignoring spacing rules for slicing | In slice notation, treat the colon like a binary operator with equal spacing on both sides (`ham[1:3]`), unless a parameter is omitted (`ham[1:]`). Inconsistent slice spacing hurts readability. |
| Spacing around keyword argument equals | Don't use spaces around `=` when used for keyword arguments (`func(real=5, imag=0)`) or default parameter values in function definitions, unless combining with type annotations. |