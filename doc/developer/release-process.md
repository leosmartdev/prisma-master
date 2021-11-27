# How To: Create a PRISMA C2 Release

The release process is currently fairly manual, and this guide will go through all the steps we perform to create a new release.

The release process is the same for official releases, release candidates, and even alpha releases. 

## Preparing for the release

There a couple of things we must do before creating the official release tag in git and building the release. These steps involve ensuring the version number is correct and updating the changelog with any changes.

### Verify the version number

As of 1.7.5 the version number for the `orolia/prisma` project is retrieved from the `VERSION.txt` file at the root of the repo. During release builds tied to a tag, the version is retreived directly from the git tag with the leading `v` removed.
During the build process, there is a script `prisma/make/go-version` that reads that file or the tag and pulls the version number out during build so all the daemons report the correct versions with `--version`.

For the `orolia/prisma-ui` project, all version numbers are stored in the `package.json` files.

See: [update-version-numbers.md]()

Open the `prisma/client/package.json` and verify the `version` property matches the version you are about to release. This chart describes our versioning strategy and when to apply the alpha, rc tags.

| Release Type      | Version String Format | Build from Branch   | Comments                     |
|-------------------|-----------------------|---------------------|------------------------------|
| Alpha             |  X.X.X-alpha          | develop             | Builds on commits to develop |
| Release Candidate |  X.Y.Z-rcX            | release/X.Y         | Builds from release branch   |
| Stable Release    |  X.Y.Z                | master              | Builds from commit to master |

The flow is usually:

1. Developers are making merge requests to develop during the development periods. When merged into
develop a build is kicked off. This is the alpha channel, and all versions numbers in develop should
have the `-alpha` tag and builds will always include a data and git hash as well to differentiate
builds from one another.
2. When we are nearing a release and freezing active development to move to testing, we will create
a `release-X.Y` branch from `develop` (or a released tag if we are creating another minor release of
an existing release). When branching, we must at that point update the `-alpha` tag to `-rc1`.
Builds on this branch are not automatically created, they are manually created when we decide to cut
a  new release candidate. We will follow the release process in this document to tag, build, then
update the version on the branch with the next rc number for example `-rc2`. Every RC should only
ever have 1 build artifact. Creating another build of a different tag means you must increment the
rc build number.
3. When testing is complete and the release is to be official, we will modify the version number
one last time to drop the `-rcX` and create the official `X.Y.Z` release tag. Then we will merge
the release branch into master as the official release (or just tag if its a minor version of a past
release) and then create the final official build.

### Update the Changelog

Before tagging the repo, we must make one last commit with the changelogs updated for this release.
There are two `CHANGELOG.md` files that must be updated (may be more in the future as we add
packages to `prisma-ui/packages`):

1. `prisma/CHANGELOG.md` which holds the main changes for eveything update in the repository except changes to `prisma/client/**/*`
2. `prisma-ui/CHANGELOG.md` for all changes to the client.

!!! TODO

    In the future we may add other changelogs to `prisma-ui/packages/prisma-*/CHANGELOG.md`.

The easiest way to update the changelogs are to use the list of Pull Requests to see what code
has changed. You can grab the list of `MERGED` Pull Requests and loop through each one translating
the request into a line or couple lines in the changelog describing the changes. You should also
add a link next to that line that references the Pull Request so we can track the code associated
with those changes.

This means, the Pull Requests need to be formatted will as much information as possible about the
change to make creating the CHANGELOG as easy as possible. That includes referencing the actual
tickets whenever possible. By doing this, it means our Changelog references the PR wich then
references the ticket creating an audit log of sorts. The audience for our changelogs are developers
and SCRUM masters in this case, but these CHANGELOGs will be used to create
management/user/integrator changelogs for external customers.

#### Format of the changelog.

The CHANGELOGs follow the same format. Versions are always listed top -> bottom based on release date.
They are all in markdown format with versions set as the Header 2, or `##`

```
# CHANGELOG NAME

## <MOST RECENT VERSION NUMBER>

## <OLDER VERSIONS>

## <EVEN OLDER VERSION>

...
```

Under each version number is a summary of the release if applicable (mostly this is for official
releases to describe the general changes or theme to the release) followed by a bulleted list of
each change that necessary to log.

The bulleted list follows this order of types of change. The list should always be ordered top ->
bottom the same:

1. `[NEW]` New features that have been added in this release
2. `[CHG]` Changes to existing features that occured in this release.
3. `[FIX]` Bug fixes that have been addressed
4. `[RM]`  Features that have been removed in this release.

Each line always follows this format:

```
* [CHG] This feature was changed to do something new [PR #4](https://github.com/link/to/pull-request)
```

You can omit the link ONLY if there is no Pull Request or ticket associated with that change, but
ideally, every line will link to a Pull Request.

### Commit Changes

After editing the changelogs, commit the changes in git and push directly to the release branch.

```bash
  git add .
  git commit -m "Updated Changelogs for release"
  git push origin release/X.X
```

## Create the Release

The release consists of two parts: the git tag and the release build.

### Tag the release

Before we create the artifacts, we should first create the git tag for the release. This is only
done for official releases and release candidates. Alpha versions automatically built will not have
tags associated with them.

The format of the tag is always: `vX.Y.Z` or `vX.Y.Z-rcX`;

The tagging process is just a normal git tag, but it's preferred to use github to tag because 
we can add the changelog right when creating the tag. 

#### GitHub Method

Head to [orolia/prisma/releases](https://github.com/orolia/prisma/releases) and [orolia/prisma-ui/releases](https://github.com/orolia/prisma-ui/releases) and click the `Draft a New Release` at the top. 

Enter in the tag version in the correct format, eg, `v1.7.6`. For the title, it should always be first
the version number without the v, then, then if its a pre-release, one of the following: `Alpha X`, 
`Release Candidate X`, or some other descriptive text for a one off pre release like 
`Pre Release for Army Demo - Build 1`. So for a stable release: `1.7.6`, rc: `1.7.6 Release Candidate 3`.

Then, be sure to fill in the description with the changelog for this release. It should be the exact
changelog as commited to the CHANGELOG.md for this release. 

Be sure before saving to check the `This is a pre-release` checkbox if this is not a stable release.

Hit Publish Release to publish. This will create the release in GitHub and the tag in the repo. 


#### Git Tag Method

Be sure to do this for both `orolia/prisma` and `orolia/prisma-ui`.


``` bash
git checkout release/1.7
git tag v1.7.1-rc6 -m "1.7.1 Release Candidate 6"
git push origin v1.7.1-rc6
```

### Build

When the tag is pushed to the remote repo, AWS CodeBuild will automatically build the release and installers. 
Once complete, the `s3://prisma-c2/release/<VERSION>` folder will contain the server installer, windows installer, and 
the debian installers for the backend. 

!!! note 
    As of 1.7.6, the installer package is also build in the `release-prisma` job.
    
!!! note
    As of 1.7.5, when you tag a release in either `prisma` or `prisma-ui` repos, the `release-prisma` and the `release-prisma-electron` build jobs will automatically be triggered.

## Artifacts

When the builds are successful, the artifacts will be sent to the `prisma-c2` bucket in S3.

Currently, they will be in the `releases/<VERSION>/` folder, where `VERSION` is the version you set in the Environment variable when starting the build.

!!! info
    The default path is just `s3://prisma-c2/releases/<VERSION>` so we don't have to change the build configuration between each major version. This path variable is set in the CodeBuild configuration as `Path` in the artifacts section. (Note this is in the AWS UI NOT in the buildspec file).

!!! warning
    If you forget to set the Environment variable for VERSION, it will in the `releases/UNKNOWN_VERSION` directory. You should re-run the build again with the right version.


The table below specifies the artifacts, their folder location inside prisma-c2 bucket, and which build project creates them.

| Artifact Name                                        | S3 Path in s3://prisma-c2                  | Build Project           |
|------------------------------------------------------|--------------------------------------------|-------------------------|
| `tms_<version>-<date>-git<hash>_xenial_amd64.deb`    | `/release/<version>/dist/`                 | release-prisma-tms      |
| `tms-deb_<version>-<date>-git<hash>_xenial_amd64.deb`| `/release/<version>/dist/`                 | release-prisma-tms      |
| `tms-mcc_<version>-<date>-git<hash>_xenial_amd64.deb`| `/release/<version>/dist/`                 | release-prisma-tms      |
| `tms-dev_<version>-<date>-git<hash>_xenial_amd64.deb`| `/release/<version>/dist/`                 | release-prisma-tms      |
| `PRISMASetup-<version>.exe`                          | `/release/<version>/dist/installer/windows`| release-prisma-electron |
| `PRISMA-macOS-x64-<version>-<date>-git<hash>.tar`    | `/release/<version>/dist/package/macOS`    | release-prisma-electron |
| `PRISMA_Server_Install-<version>`                    | `/release/<version>`                       | package-prisma-installer|


### Update Github Release

Once the builds are complete, upload the built artifacts to the release in github, [orolia/prisma/releases](https://github.com/orolia/prisma/releases) and [orolia/prisma-ui/releases](https://github.com/orolia/prisma-ui/releases).

## Update the Release Site

There is a separate git repo that contains the release site that needs to be updated, then the site deployed to AWS S3 so that the release is published with links to the artifacts.

The repo is in gitlab as `orolia/prisma-release-documentation`. There are pages for each type of release. Pick the page for the release you are creating.

* **stable**: Official stable releases. m
* **prerelease**: Release Candidates
* **nightly**: Automated builds of the develop branch.
* **past**: List of past releases.

### Official Stable Releases

First open `docs/stable.md` and copy the contents starting with `## Latest Stable` and removing the two links for the badges (and the badge table) and paste into `docs/past.md`. Change the `## Latest Stable` to `## X.X.X - <DATE>` replacing X.X.X with the version and <DATE> with the release date. Save the file.

Then, change the version and release date in the top section of `docs/stable.md`.

```markdown
## Latest Stable

The latest PRISMA C2 stable release is 1.7.1 released on November 1, 2018.
```

Then at the bottom of the file are the links to all the artifacts. Changes the links to reference the correct artifacts built above.

For each artifact, take the S3 path as shown in the artifacts table above and add it to the end of `https://releases.mcmurdo.io`. So for example, for the file `S3://prisma-c2/release/1.7.1/PRISMA_Server_Install-1.7.1` the URL would be `https://releases.mcmurdo.io/release/1.7.1/PRISMA_Server_Install-1.7.1`.


### Release Candidates

For Pre-Releases (eg release candidate builds), we need to change the `docs/prerelease.md` file. This file only contains the latest release candidate, we don't keep a history of pre releases.

The process for updating the file is exactly the same of stable above, just ensure all versions have the correct version number and release candidate number as well.

For iterative release candidates, you only need
to change the `-rcX` references throughout the document.

When creating a new release candidate, just copy the contents of the `docs/stable.md` file and then changes all the versions to be the release candidate version you are releasing. The only additional change is to the badge, you will need to last part of the badge links from `branch=master` to `branch=release/X.X` replacing X.X with the version in the branch name.

!!! note
    If changing the branch does not properly show the badge status, you can copy the bade link by logging into CodeBuild then copying the badge URL from the release build page for the tms and client builds then change the branch to the release branch you are building.


### Build and Deploy Release Site

Commit the changes to the repo and push to master. CodeBuild will build and deploy the site automatically. 

## Extra Steps for Stable Releases

After creating the release and the build, there are a few extra steps that need to be done for stable releases.

### Merge the Release Branch

When the release is complete, we need to merge the release branch into the `develop` branch and the `master` branch, provided this release is not a point release for a version older than the newest release. In those cases, you may skip this section or cherry pick into develop relevant code fixes.

We will first do the merge to `master`. `master` HEAD should always be (or HEAD^1 if there was a merge and not a fast-forward in the last release) the same commit as the tag of the latest release. To do this, create a branch off the `release/X.X` branch titled `release/X.X-merge-master` then create a pull request. The Pull Request is to give a final sanity check that the merge is clean and let other developers ensure versions numbers etc... aren't overridden.

Once we merge into `master` we also need to merge into develop to get all new code changes and bug fixes into the next version as well. Do this by first creating a new branch of the release branch, then merge develop into that release branch and create a pull request. This allows the team a final chance to view the changes and verify the that the merge isn't overriding new changes in develop or causing other issues.

Once the PR is closed and the branch just created deleted, it is now safe to delete the official release branch if no work is immediately planned for a new point release.

!!! note "Remember"
    You can always re-create the release branch from the stable release tag, so deleting the branch is not irreversible. It's always a good idea to keep the branches clean and removed unused branches, especially when you have tags to re-create them.

### Copy to Sharepoint

Now that we have the official stable release, we need to copy the installers, debians, configurations, and links to any documentation or a copy of the install guide, over to the Sharepoint site so that GSS and management can access the release.

There are two sharepoint sites we place the installers into, one for GSS and one for Sales and Marketing.

  * [GSS Sharepoint Site](https://oroliagroup.sharepoint.com/sites/mcmurdogroup/DC/GSS/GSS%20Customer%20Files/Forms/AllItems.aspx?slrid=5b1b9e9e%2D400c%2D7000%2D1954%2D4296022539d8&RootFolder=%2Fsites%2Fmcmurdogroup%2FDC%2FGSS%2FGSS%20Customer%20Files%2FPRISMA&FolderCTID=0x01200046376510E6953748AACF1CC3554FCB8E)
  * [Sales & Marketing Sharepoint Site](https://oroliagroup.sharepoint.com/sites/mcmurdogroup/mcmurdord/PC2/Shared%20Documents/Forms/AllItems.aspx?slrid=5b1b9e9e%2Da04d%2D7000%2Dec22%2D032b49ae6c0b&FolderCTID=0x01200075FF57070D05E84C8F6C256BE1BB9BA0&id=%2Fsites%2Fmcmurdogroup%2Fmcmurdord%2FPC2%2FShared%20Documents%2FSales%20%26%20Marketing%2FSales%20Demo%2Freleases)

For both, create a folder with the version number as the name like `1.7.1` then inside that folder, add the windows installer and macOS tarball package.
Also place a configuration to access the demo and fleet server in each. [Demo Server Configuration](https://s3.amazonaws.com/prisma-c2/client+configurations/Demo/production.json)

#### GSS Sharepoint

For GSS, we also need upload the PRISMA Installer, Installation Guide PDF, and the build debian packages for TMS server. The installer and debian packages are the artifacts produced above. For the installation guide, use the pdf generated by the `make docs` command. The pdf will be located in `build/documentation_site/installation/installation/installation.pdf`.

### Upgrade Demo Sites

With the stable release we also need to update the demo sites, `demo.mcmurdo.io` and `fleet.mcmurdo.io` so that management, sales, and gss have access to the latest hosted version for training, review, demos, etc...

### Announce the Release

With all the links copied, demo servers updated, it's now time to announce to the company that the latest stable release is now available. Be sure to include the links to the sharepoint site for getting the code, download instructions, installation instructions, and configurations needed to get anyone up and running with a client on their box connected to our demo servers.
