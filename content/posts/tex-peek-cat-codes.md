---
title: "Another token peeking side effect in TeX"
date: 2021-09-24T20:05:40-04:00
---

When processing a stream of tokens in a language parser,
    we generally assume that peeking at the next token does not have any side effects. 
In [a previous post](../tex-expansion-edge-case) I described how this is not the case in TeX. 

In that post we had a token whose expansion rules were changed 
    after the token was peeked at but before it was fully consumed. 
In TeX, the peek operation has the side effect of expanding tokens it encounters, 
    and TeX never undoes this expansion. 
In our case the token was expanded using the rules in effect at the time of the peek 
    rather than the time of consumption, giving a somewhat counterintuitive result.

Here’s another example of a side effect that can occur during TeX token peeking: cat code assignment.

When TeX is reading an input file, 
    each character is assigned one of [16 category codes](https://en.wikibooks.org/wiki/TeX/catcode)
     (or cat codes). 
The cat code of a character determines its semantic function in the TeX language. 
Characters like `{` have cat code 1 and start a new group;
    `^` has cat code 7 and denotes a subscript; 
    `A` has catcode 11 and represents a regular character to be typeset. 
The cat code assignment rules can be dynamically changed in TeX source using the `\catcode` primitive.

What happens if we peek at a character, 
    change the cat code rules for the character, 
    and then consume the character afterwards? 
Will the character’s cat code be determined by the old cat code rules,
    or the new cat code rules? 
We can write a TeX snippet to answer this question in plain English:

    \catcode`\T=13
    \def The new rules{The old rules}
    \catcode`\T=11The new rules.

Running this script in TeX gives:

    The old rules.

So again, peeking has an irreversible side effect, in this case assigning a cat code to a character.

## How the snippet works

The first line of the snippet ``\catcode`\T=13`` sets the
    cat code of `T` to be 13 which means _active character_.
When `T` is an active character, 
    it can be used in many places where a control sequence is usually required.
For example, we can define macros that are expanded when the active character `T` appears in the input.

In the second line we define such a macro using `\def`. 
The target of the macro is the active character `T`.
The following characters `he new rules` are regular letter tokens, 
    and form the the macro’s prefix. 
When the macro is expanded, 
    it will first trim `he new rules` from the input stream 
    (or throw an error if these characters don’t appear).
It will then push `The old rules` to the front of the input stream.

The upshot is that when TeX encounters the text `The new rules`
     what happens depends on whether `T` is an active character or a regular character.
If it is regular, `The new rules` is printed verbatim. 
If it is active, the macro is invoked and `The new rules` is replaced by `The old rules`.

Finally in the last line we change the cat code of `T` back to 11 (regular letter).
Here’s where the peeking happens.
After the equals sign, TeX needs to parse in a number for the cat code.
It first reads the two digits `11`.
It then peeks at the next token to see if it is also a digit.
If the next token was `2`, say, the number would be 112.
At the point of peeking, `T` is assigned the active character
    cat code because the `\catcode` primitive hasn’t finished yet.
`T` is not a digit so the command runs with input 11, 
    the cat code rules are updated, and then `T` is processed.

At this point, one of two things can happen:

- If the cat code assignment is not updated with the new rules,
    `T` remains an active character and `The new rules` is replaced by `The old rules`.

- If the cat code assignment is updated with the new rules, 
    `T` will be changed to a letter character and `The new rules` will be returned verbatim.

In both cases, the snippet will correctly describe what happens.
As we saw, it is the first case that applies.
