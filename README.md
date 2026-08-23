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
- [Configuration](#configuration)

## Features

- Track the latest release of any package from the command line
- List every tracked package with its latest known version
- Read a package's changelog without leaving your terminal (soon)
- Support for many package providers such as GitHub

## Configuration

**reelens** reads its configuration from `~/.config/reelens/config.yaml`:

There is no default configuration, so everything must be explicitely set.

```yaml
packages:
  # See the `packages` section
```

### Packages

Packages are defined in the `packages` key, which is expected at the root-level
of the configuration file.

A specific package is defined as a key under `packages`; the key is the package
name, and package options are defined under it.

#### Common options

The following options are available for all packages:

| Name       | Expected type    | Description                                     |
| ---------- | ---------------- | ----------------------------------------------- |
| `provider` | `string`         | Provider used to retrieve package data          |
| `version`  | `release \| tag` | Determines how the package version is obtained. |

##### Providers

###### github

| Name   | Expected type | Description       |
| ------ | ------------- | ----------------- |
| repoID | string        | The repository ID |

```yaml
packages:
  neovim:
    provider:
      type: github
      repoID: neovim/neovim
    version: release # "release" for the latest GitHub release, or "tag" for the latest stable semantic version tag
```

##### `version`

The `version` option determines where the package version is obtained from. It
supports two strategies: `release` and `tag`.

###### `release`

Uses the **name of the latest release** published by the provider as the package
version.

This is the **recommended and most efficient** option when releases are
available.

###### `tag`

Uses the latest stable tag as the package version.

This is useful for packages that do not publish releases and instead use tags to
mark stable versions. It should generally only be used when `release` is not
suitable, as determining the latest stable tag may require inspecting a large
number of tags. For repositories with many tags, this can be inefficient and may
cause the provider's API rate limit to be reached.
