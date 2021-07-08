---
title: "TeX and global variables"
date: 2021-05-10T20:05:40-04:00
draft: true
---

Suppose you have the following raw TeX (not LaTeX) file `first.tex`:

```
\def\mycommand{\relax First version of my command.}
\input second.tex \mycommand
```
This defines a custom macro, inputs a second file, and then prints the output of the custom macro.
Suppose that `second.tex` contains:
```
\def\mycommand{\relax Second version of my command.}
Command has been redefined.
```
What's the output of `pdftex first.tex`? It's what you'd expect:
```
Command has been redefined. Second version of my command.
```
The custom macro is first defined, then the second file is processed during which the macro is redefined, 
so when `\mycommand` is expanded it has the second definition.

But what if we change `first.tex` by removing the space before `\mycommand`?
```
\def\mycommand{\relax First version of my command.}
\input second.tex\mycommand
```
Seems like an innocuous change. Let's run `pdftex first.tex`:
```
Command has been redefined. First version of my command.
```
The result is different!

This example illustrates an interesting edge case in TeX processing.
Here's what's happening.
When processing the `\input` command, TeX scans the following characters to get the filename.
It will keep scanning until it encounters a space, or an unexpandable command like `\relax`.
If it encounters an expandable command like a macro, however, it needs to expand the command.
This is to support cases where the filename is in a macro:
```
\def\filename{second.tex}
\input\filename
```
In the edited version of `first.tex`, the input scanning runs until `\mycommand`.
The macro is expanded because it may return additional characters that help identify the filename.
However the expansion begins with `\relax`, so no characters are used.
At this point the expansion was essentially an irreversible mistake.



