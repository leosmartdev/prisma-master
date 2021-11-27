# Update Version Numbers

When we create a new release branch from `develop` we need to update the version numbers in develop
to reflect the next version.

We need to change the following files to update the versions in our two projects.

## `orolia/prisma`:
  * `VERSION.txt`

The `VERSION.txt` is a text file that just contains the single version string with no spaces:

```
1.8.0-alpha
```

## `orolia/prisma-ui`
  * `package.json`
  * `packages/prisma-electron/package.json`
  * `packages/prisma-map/package.json`
  * `packages/prisma-ui/package.json`

For the `package.json` files, there are two places the version number may need to be changed. For each one, there will be a `version` property at the top that will need to be updated. Optionally, some `package.json` will also have an `@prisma/*` property in the `dependencies` section that will need to be updated.

!!! note
    Right now, only `prisma-electron/package.json` has these `@prisma/*` references but others may get them in the future.

```json
{
  "name": "@prisma/electron",
  ...
  "version": "1.8.0-alpha",
  ...
  "dependencies": {
    ...
    "@prisma/map": "1.8.0-alpha",
    "@prisma/ui": "1.8.0-alpha",
    ...
  }
}
```

## Version Number Schemes

`develop` branch should _always_ end in `-alpha`. Only nightly builds are built from `develop`

When a release branch is created, the numbers will be need to be updated to reflect the release. Majority of the time, it means changing the `-alpha` to `-rc0` for the first release candidate.

On each release candidate, after creating the release, increment the release candidate number on that branch to the next `-rcX`.

When ready for the official release, change the versions as part of the changelog updates and drop the `-rcX` suffix so the version is only `X.X.X`. Commit and build.
