---
title: "Non-reproducible builds in Python"
date: 2022-06-22T22:05:40-04:00
---


For the last four or five years I’ve worked on-and-off on a 
[sizable enough Python side project](https://github.com/jamespfennell/transiter).
By far the most frustrating aspect of maintaining it has been
    that if I leave it alone for a few months and then come back to add some small feature, 
    I cannot in general rebuild the project. 
Something simple like executing `pytest` to run all the unit tests just does not work. 

Instead, I have to first spend a non-trivial amount of time debugging error messages
 like the one that appeared when I tried to build the project earlier this week:

```
ImportError: cannot import name 'soft_unicode' from 'markupsafe'
```
When I first saw this error, the identifiers `soft_unicode` and `markupsafe` were totally unknown to me.

My side project uses Docker to make the build environment mostly consistent,
    and I use pip for managing dependencies. 
In the past I also spent a significant amount of time ensuring the CI and the local development environments are the same.
This means that when I triggered a CI re-run for the commit that passed CI a few months ago,
    it also failed (!) and the error message was identical.

You can Google the error and find many 
[Github issues](https://github.com/aws/aws-sam-cli/issues/3661) and
[Stack Overflow questions](https://stackoverflow.com/questions/72191560/importerror-cannot-import-name-soft-unicode-from-markupsafe)
about it.
The easiest fix is to manually pin the MarkdownSafe package (which I had never heard of) to
    the magic version number 0.23.

But this is not a full fix - we’ve only just started!
When this dependency is pinned the next error appears:
```
ImportError: cannot import name 'json' from 'itsdangerous' 
```
This error looks largely the same. 
It, too, has its collection of [Github issues](https://github.com/MTG/mtg-arousal-valence-annotator/issues/5)
and
[Stack Overflow questions](https://stackoverflow.com/questions/71189819/python-docker-importerror-cannot-import-name-json-from-itsdangerous).
The easiest fix here is to pin `itsdangerous==2.0.1`.

With this done the next error appears:
```
ImportError: cannot import name 'BaseResponse' from 'werkzeug.wrappers' 
```
[Github issue](https://github.com/mjmeijer/donald_2021/issues/4); dependency to pin: `werkzeug==2.0.3`.


All of these specific errors have the same root cause.
If you have a Python project whose dependencies are being managed with pip,
    your build is not, in general, deterministic.
This is the case even if everyone involved (you, developers of your dependencies) 
    is following pip best practices. 
Your build changes over time, and can randomly start to fail at any point.

The problem is well known and comes from three pieces:

1. Library authors often do not pin their dependencies exactly, 
    and instead specify a liberal range.
    I thought, but could not verify, that some Python documentation explicitly recommended this.
    As an example, the library Jinja at version 2.11.3 
    [specifies that MarkdownSafe should have version greater than or equal to 0.23](https://github.com/pallets/jinja/blob/cf215390d4a4d6f0a4de27e2687eed176878f13d/setup.py#L53).

1. For something like my side project which is not a library, 
    pinning dependencies exactly is recommend. 
    But only the direct dependencies are pinned, 
    and pip seems to lack a built-in mechanism for easily pinning transitive dependencies.

1. When choosing which version of a package to download, pip always chooses the *newest* version
    that satisfies all of the dependency constraints.

When building my project, pip eventually needs to fetch MarkupSafe because it used by Jinja 2,
    a direct dependency of my project.
The constraint from Jinja is for a version greater than or equal to 0.23;
    in particular, there is no upper bound.
So pip always uses the latest version.
At some point when a 
[backward incompatible change to Markupsafe is committed](https://github.com/pallets/markupsafe/pull/261/commits/7856c3d945a969bc94a19989dda61c3d50ac2adb),
    my project no longer builds.

I had been aware of this issue for a while, 
    but what was interesting to me in my recent debugging was seeing just how much wasted time this causes.
My debugging involved fixing 3 largely identical errors.
But each separate error had spawned its own collection of Github issues,
    each of which needed triaging, investigation and downstream fixes,
    as well as Stack Overflow questions that needed to be answered.

My understanding is that the default package managers for other languages like Rust don’t suffer from this problem.
Russ Cox has a really nice [series of blog posts](https://research.swtch.com/vgo) about dependency management,
    and in particular [how to choose package versions to get reproducible builds](https://research.swtch.com/vgo-mvs).
 Roughly speaking, in item 3 above,
    instead of picking the _newest_ version that satisfies the version constraints,
    pick the _oldest_ version. 
As new versions are released, this choice doesn't change and the build will be the same
    (at least from the perspective of dependencies).

Russ Cox was writing as part of his work on developing a dependency resolver for Go,
    which also has a better out-of-the-box reproducible builds story.
For Python, I think with some manual toil when updating dependencies, it is
    [possible to get deterministic builds](https://pip.pypa.io/en/stable/topics/repeatable-installs/).
However, having grown accustomed to toolchains where this is all handled automatically,
    the solution I have converged on is, sadly,
    is just not to use Python for large projects I intend to maintain for many years.
None of the work in this area -- be it managing dependency versions explicitly,
    or debugging non-reproducible builds -- is simply any fun.
And for hobby projects, that's really the main point!
