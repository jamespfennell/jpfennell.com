---
title: "A surprising enum size optimization in the Rust compiler"
date: 2025-04-07T18:05:40-04:00
---


Enums are one of the most popular features in Rust.
An enum is type whose value is one of a specified set of variants.

```
/// Foo is either a 32-bit integer or character.
enum Foo {
    Int(u32),
    Char(char),
}
```

Values of type `Foo` are either
    integers (e.g. the variant `Foo::Int(3)` with a payload of `3`)
    or characters (e.g. the variant `Foo::Char('A')` with a payload of `'A'`).
If you think of structs as being the *and* combination of their fields,
    enums are the _or_ combination of their variants.

This post is about a surprising optimization that the Rust compiler
    performs on the memory representation of enum values
    in order to make their memory usage smaller
    (spoiler: it's not the niche optimization).
In general, keeping values small can result in faster programs because values get passed
    around in CPU registers and more values fit in single CPU cache lines.

Generally the size of an enum is the size of the largest payload,
    plus some extra bytes for storing a _tag_ that specifies which variant the value is.
For the type `Foo` above, both variant payloads take up 4 bytes,
    and we need an additional byte at least for the tag.
A requirement called "type alignment" (which I won't get into)
    requires that Rust actually use 4 bytes for the tag.
So the overall size of the type is 8 bytes:

```
assert_eq!(std::mem::size_of::<Foo>(), 8);
```

(All the code snippets can be run [here in the Rust Playground](https://play.rust-lang.org/?version=stable&mode=debug&edition=2024&gist=c5daeb2d75d83f9f5e1bdb9d46d77182).)

For the rest of this article it will be useful and interesting to actually see the memory
    representation of various enum values.
Here's a function that prints the raw bytes representation of any Rust value:

```
/// Print the memory representation of a value of a type T.
fn print_memory_representation<T: std::fmt::Debug>(t: T) {
    print!("type={} value={t:?}: ", std::any::type_name::<T>());
    let start = &t as *const _ as *const u8;
    for i in 0..std::mem::size_of::<T>() {
        print!("{:02x} ", unsafe {*start.offset(i as isize)});
    }
    println!();
}
```

(This function was adapted from this [10-year old Reddit post](https://www.reddit.com/r/rust/comments/2ngnmk/viewing_any_object_as_a_raw_memory_array).)

Let's run this for our enum `Foo`.

```
print_memory_representation(Foo::Int(5));
// type=Foo value=Int(5): 00 00 00 00 05 00 00 00
//                        |-- tag --| |- value -|

print_memory_representation(Foo::Char('A'));
// type=Foo value=Char('A'): 01 00 00 00 41 00 00 00 
//                           |-- tag --| |- value -|
```

The first thing to point out is that this computer's memory is _little endian_,
    so the lowest bytes come first.
In 32-bit hex the number 5 is `0x00000005`, but its little endian representation is `05 00 00 00`.

With that in mind, we see that the first 4 bytes are the tag.
The integer variant has been assigned tag 0,
and the character variant tag 1.
The second 4 bytes are then just the usual values of the payload.
Note that the lowercase letter "a" is 41 in ASCII,
    which is why its memory representation is `41 00 00 00`.

## The niche optimization

Aside from the general tags scheme,
    there is one well known enum size optimization called the niche optimization.
This optimization works for types where only one of the variants has a payload.
A good example is the built in option type:

```
enum Option<char> {
    None,
    Some(char),
}
```

Based on the tags analysis in the last section we might guess the enum size will be 8 bytes
    (the largest payload of 4 bytes plus 4 bytes for the tag).
But actually values of this type only use 4 bytes of memory in total:

```
assert_eq!(std::mem::size_of::<Option<char>>(), 4);
```

What's going on?
The Rust compiler knows that while `char` takes up 4 bytes of memory,
    not every value of those 4 bytes is a valid value of `char`.
Char only has about `2^21` valid values (one for each Unicode code point),
    whereas 4 bytes support `2^32` different values.
The compiler choses one of these invalid bit patterns as a _niche_.
It then represents the enum value without using tags.
It represents the `Some` variant identically to char.
It represents the `None` variant using the niche.

One interesting question is: what exact niche does Rust use?
Let's print the memory representations to see:

```
let a: char = 'A'
print_memory_representation(a);
// type=char value='A': 41 00 00 00 

print_memory_representation(Some(a));
// type=Option<char> value=Some('A'): 41 00 00 00 

let none: Option<char> = None;
print_memory_representation(none);
// type=Option<char> value=None: 00 00 11 00 
```

As we see, the memory representations of `'A'` and `Some('A')` are identical.
Rust represents `None` using the 32-bit number `0x00110000`.
A quick search reveals that this number is exactly one bigger than the largest
    valid Unicode code point.


## Beyond the niche optimization?

My understanding was that Rust doesn't perform any more optimizations,
    so I was pleasantly surprised recently when I found one.

The context is nested enums. Start with an inner enum

```
enum Inner {
    A(u32),
    B(u32),
}
```

If we look at the representation in memory,
    it's as we expect: 8 bytes, where the first 4 bytes store the tag
    and the last 4 bytes store the payload.

```
assert_eq!(std::mem::size_of::<Inner>(), 8);

print_memory_representation(Inner::A(2));
// type=Inner value=A(2): 00 00 00 00 02 00 00 00 
//                        |-- tag --| |- value -|

print_memory_representation(Inner::B(3));
// type=Inner value=B(3): 01 00 00 00 03 00 00 00 
//                        |-- tag --| |- value -|
```

Now add another enum that contains the `Inner` enum as a payload:

```
enum Outer {
    C(u32),
    D(Inner),
}
```

My guess was that the size of values of this type would be 12 bytes - 
    8 bytes for the largest payload `Inner`, plus 4 bytes for the tag.
But it's not - values take up only 8 bytes!

```
assert_eq!(std::mem::size_of::<Outer>(), 8);
```

What's going on here?

First let's check what values of type `Outer::C` look like in memory:

```
print_memory_representation(Outer::C(5));
// type=Outer value=C(5): 02 00 00 00 05 00 00 00
//                        |-- tag --| |- value -|
```

Already we see something weird happening:
    Rust has chosen to use the tag number 2 for `Outer::C`
    instead of starting from 0 like it did for `Inner::A`.
Next look at `Outer::D`:

```
print_memory_representation(Outer::D(Inner::A(2)));
// type=Outer value=D(A(2)): 00 00 00 00 02 00 00 00 
//                           |-- tag --| |- value -|

print_memory_representation(Outer::D(Inner::B(3)));
// type=Outer value=D(B(3)): 01 00 00 00 03 00 00 00 
//                           |-- tag --| |- value -|
```

The representation of the value `Outer::D(inner)`
    is identical to the representation of `inner`!



I guess the Rust compiler has put the following pieces together:

- The first 4 bytes of `Inner` form a tag whose
    values are just 0 or 1.
    In particular we have many more values we can store here, like in the niche optimization.

- The payload for every other variant of `Outer` is no larger than any of the payloads of `Inner`.
    In particular, if `Inner` values are of the form `<Inner tag><Inner payload>`,
    then the payload for every other variant of `Outer` fits inside `<Inner payload>`. 

We can thus represent values of `Outer` in the form `<Outer tag><Outer remainder>` where

- If the `<Outer tag>` matches any `<Inner tag>`, the value is `Outer::D` and the payload is the entire
    bit pattern `<Outer tag><Outer remainder>`.

- Otherwise, the value is another variant and the payload is in `<Outer remainder>`.
