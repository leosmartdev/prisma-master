# Pull Request and Code Review Guide

This document describes the workflow and expectations for a Pull Request (or Merge Request).

The Pull Request is a critical part of the engineering workflow. It provides a place to discuss code,
design, style, and correctness as a group, while also providing a gate to ensure code quality stays
high. This is not to say that all code reviewed will be perfect, it never is, but it allows for
discussions so that the team can ensure changes are of a reasonable quality and correctness.

The following guidelines will help with the process of writing and reviewing a pull request, and
ensure everyone is on the same page so the aren't too many issues that arise outside of discussion
on the code being merged.

## Reasons

The job of the pull request is to ensure that new code:

* Follows the style guidelines of the project
* Is technically correct and actually solves the problem at hand.
* Includes tests to ensure correctness and passes existing tests to verify no regressions.
* Code is maintainable
* Provides documentation for the code or changes to existing documentation.

The review also provides and opportunity for other engineers to discuss the code and have their
input heard as well. This can be related to the code itself, or the overal design and architecture.

## Approvals

All code will require an approval from a Code Owner for the parts of the application their code is
touching. There can be more than one code owner for a part of the project, and it's up to the team
to indentify those owners.

!!! note
    Although code owners are responsible ultimately for the merge requests being merged in, they are not the only ones who can or should review a merge request. Every developer can and should review merge requests and provide input.

!!! note
    In some cases there may be only one code owner for a particular part of the application. In these cases, it is expected that code owners will not merge their own code without a review from the team. Though they are solely responsible for the merge, they cannot be responsible for the review itself and should get input from the rest of the team before merging.

## For the Committer

When creating a pull request, be as descriptive as possble in the details. Include a link to the
ticket, but put as much information about how the code solves the ticket in the details. It's
important to provide as many details as possible so the reviewer can understand your code and how
to run the code or setup the test environment to verify it works properly. Also be sure to include
details on any architectural or design decisions made in the request. This is important for better
discussions over these types of design details.

Also, if the code makes any changes to APIs, you must include examples of the new API or provide a
link to POSTMAN examples and documentation if the API changes are part of the RESTful interface.
It's also import to address any backwards compatibility problems and ensure these changes do not
break systems during upgrades. If there is no current place to document the API, then the team should discuss either adding documentation for that API with examples or just provide the examples as part of the merge request.

Provide an itemized list of changes, new features, removed code, and fixed bugs. This follows the
exact same format as the changelog, but instead of including links to pull requests, add a link
to the tickets. The list should be ordered [NEW] at the top followed by [CHG], [FIX], [RM]. (See
[Format of the Changelog for information on formatting](../../developer/release-process/#format-of-the-changelog))

```
* [NEW] Added a function to convert lat/lng to decimal values [PC2-XX](http://link.to/ticket/PC2-XX)
* [CHG] `convertToDegreesSeconds()` added new round parameter [PC2-XY](http://link.to/ticket/PC2-XY)
```

There is a new template in GitLab to help assist you in creating the pull request and filling in
all the above information. Not every section is needed for every pull request, but do remember the
more information provided, the better and more timely the review will be. Your audience is the
reviewer and the more information you provide them the better they will be able to understand your
code and the less back and forth will need to happen.

!!! note
    These pull requests are how the CHANGELOGs are generated during release, so accurate information
    in the pull request is a great help to creating the release notes and ensuring we capture all
    code changes. The better the pull request, the better the release documentation. Same goes
    for release testing, its much easier to find the bug inducing code if the requests are well
    organized and descriptive.

So, when creating the pull request:

1. Communicate! More details the better the review
2. Read your code
    - Check the diff, does it look right?
    - Does the diff clearly show your changes?
    - Should your code be broken into smaller chunks or multiple pull requests?
    - Is everything styled correctly?
    - Did I include documentation or change the documentation to reflect the change?
    - Did I provide unit tests and where applicable, integration tests?
3. Discuss feedback
    - Respond to all code review feedback. Discuss any feedback you disagree with
4. Identify and tag the reviewer
    - Identify the code owner for the code you are committing against.
    - Tag them in the Approvals section of the pull request so they know their approval is requried.
5. Merge target branch
    - Ensure the branch you are requesting to merge into has been merged into your branc

Lastly, don't take anything personally. It's hard to convey positive or negative context in text
sometimes, but a code review is never attacking the developer. This is about getting extra eyes
on your code and having discussions. Everyone benefits from having extra eyes on their code. If you
don't agree with a comment say so! But be sure to provide reasoning for why you disagree, remember
communication is key, this is a discussion after all.

## For the Reviewer

The reviewers job is to understand the code in the request and verifying that it is correct and
solves the problem at hand, while also making sure it's stylistically consistent with the code base
and fits with overall architecture and maintainability in the future. You are just as responsible
for the code as the committer.

When reviewing, be sure to ask questions when your unsure of something. Ensure you understand the
code and what it's doing, if you don't ask!

Verify the correctness of the code. Does it solve the problem at hand? Are edge cases accounted for?
Are the there tests included to show the code works, and especially if its a bug, a test case
that catches the bug then proves its fixed. Don't just verify correctness by looking at the code,
you must also BUILD AND RUN IT! Check out the branch and make sure it works. Run the tests. The
committer should have provided enough information to replicate the bug, or setup an environment for
you to run the code easily. If they haven't ask for it.

Check for stylistic consistency. Hopefully most of these issues are solved by linters or formatters
(like prettier for the frontend) but you should verify it adheres to the style guidelines set by the
team. Consistency is key here.

Evaluate any design or architecture tradeoffs. Discuss these with the committer. It's important to
asses changes in design or architecture, doesnt mean the code is wrong or should be changed, but
its important to understand the _why_ behind a change and evaluate it. When in doubt, prefer
code readability and maintainability. It's better to optimize for developers first then address
performance issues when problems actually arise in performance testing.

Verify backwards compatibility or a deprecation path when APIs change. Especially when related
to protobuf API changes or serialization/deserialization. If the committer hasn't addressed this and
you think there might be issues. Ask!!

So when reviewing a pull request:

1. Understand the code
2. Check for code quality
    - Enforce sylistic consistency in the application
    - Ensure code maintainability
    - Prioritize code readability over pre-mature optimization.
    - Check code organization
3. Review design and architecutre
    - Evaluate tradeoffs
    - Discuss changes to design or architecture. When in doubt, bring in others on the team.
4. Verify correctness
    - Check for test coverage, ensure edge cases are well tested.
    - Build and run the code. Verify it works in your own environment as well.
    - If its a bug, ensure there are tests that fail when the bug is present and pass with the code changes
5. Verify backwards compatibility and/or deprecation paths
    - Check for issues with serialization.
    - Check for issues that may arise when upgrade old code.
6. Check for security holes or issues.
7. Be clear in your comments, when in doubt phrase as a question.

-------

Inspiration for this document provided by:

  * [Code Review Guidelines (Yelp)](https://engineeringblog.yelp.com/2017/11/code-review-guidelines.html)
  * [A GitHub Pull Request Template](https://embeddedartistry.com/blog/2017/8/4/a-github-pull-request-template-for-your-projects)
  * Other posts and developers I (John Conway) forgot to grab links from or got from slack conversations.
