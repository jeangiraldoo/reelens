# reelens

/ˈriː.lɛnz/

_reelens_ tracks package changes and releases so you don't have to do it
yourself.

## Elevator Pitch

If you are like me, you like to be up to date with certain packages or tools
because that new shiny feature looks cool! So you end up checking repositories
or your package manager(s) from time to time.

But manually checking whether a new release for a given package is out is
tedious, so is manually fetching the release notes or changelog file. This
problem only gets worse if you perform this process for more than one package.

Why not writing a script?

Honestly, that's enough in some cases. Then one day you want to check a package
hosted in a different platform, back to updating the script. Oops, that package
doesn't have a `CHANGELOG.md`, it instead places the release notes in a tag,
back to updating the script. Personally, I'd just stop caring about tracking new
releases after having to update the script a second time!

So, if you have better things to do than manually keeping up with all your
packages, or maintaining a (likely) ever changing script, Reelens might become
your handy companion.

## Table of contents

- [Features](#features)
- [Configuration](#Configuration)

## Features

- Track the latest release of any package from the command line (soon)
- List every tracked package with its latest known version in a table (soon)
- Read a package's changelog without leaving your terminal (soon)
- Support for many package providers such as GitHub (soon)

## Configuration

reelens reads its configuration from `~/.config/reelens/config.yaml`:

```yaml
packages:
  gum:
    provider:
      type: github
      repo: neovim/neovim
    version: release # "release" for the latest GitHub release, or "tag" for the latest stable semantic version tag
```
