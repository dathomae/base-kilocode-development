# Go Synthesis Target - Table of Contents

This directory contains design documentation for the Go code synthesis target, which generates Go code for calculating plan data and dimensions.

## Files

**[synthesis.md](synthesis.md)**
: Design documentation for synthesizing Go code from CIL representations

**[synthesis_design_decisions.md](synthesis_design_decisions.md)**
: Design decisions and rationale for the Go synthesis target

**[eros/toc.md](eros/toc.md)**
: Epic Restraining Orders (EROs) for the Go synthesis target.

## Overview

The Go Synthesis Target is the Go code that is generated to calculate a
particular set of metrics, dimensions, types, forcast lenghts and so on specified by the output specification passed to the CPL compiler.

To avoid confusion, the Go Synthesis Target is concerned with how the Go code
is synthesized, not compiling or running the code or tracking changes. Those will the responsibilities of other parts of the Chronocone system.
