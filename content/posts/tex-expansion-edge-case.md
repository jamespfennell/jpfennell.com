---
title: "A time-traveling edge case in TeX"
date: 2021-07-08T20:05:40-04:00
---

Suppose you have the following raw TeX (not LaTeX) file `first.tex`:
```
\def\month{\relax The month is May.}
\input second.tex \month
```
This defines a custom macro, inputs a second file, and then prints the output of the custom macro.
Suppose that `second.tex` contains:
```
June has begun.
\def\month{\relax The month is now June.}
```
What's the output of `pdftex first.tex`? It's what you'd expect:
```
June has begun. The month is now June.
```
The custom macro `\month` is first defined, then the second file is processed during which the macro is redefined, 
so when `\month` is expanded it has the second definition.

But what if we change `first.tex` by removing the space after `\month`?
```
\def\month{\relax The month is May.}
\input second.tex\month
%                ^ space removed
```
Seems like an innocuous change. Let's run `pdftex first.tex`:
```
June has begun. The month is May.
```
The result is different!
And we've travelled back in time!

This example illustrates a funny little edge case in TeX processing.
Here's what's happening.
When processing the `\input` command, TeX scans the following characters to obtain the name of the file to input.
It keeps scanning until it encounters either a space or a command like `\relax` that cannot be expanded.
If it encounters an expandable command like a macro, however, it needs to expand the command.
This is to support cases where the filename is stored in a macro:
```
\def\filename{a_file.tex}
\input\filename
```
In this case the `\input` command needs to expand `\filename` to obtain `a_file.tex`.

In the original version of `first.tex` above, the processing stops when the space after
 the string `second.tex` is
    encountered.
In the edited version of `first.tex`, the input scanning continues until `\month` because there is no space to stop it.
Because `\month` is expandable, TeX expands the command in case the replacement text contains
    more characters for the file name.
After this expansion the line looks like this:
```
\input second.tex\relax The month is now May.
%                ^ processing continues here
```
The next command is `\relax` which is not expandable, so the input command takes `second.tex` as the filename
and carries on.
However expansion is irreversible! So when processing returns to `first.tex`, 
it is the previously expanded text of `\month` that is used.





