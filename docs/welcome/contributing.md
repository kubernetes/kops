## Getting Involved and Contributing

Are you interested in contributing to kops? We, the maintainers and community,
would love your suggestions, contributions, and help! We have a quick-start
guide on [adding a feature](../development/adding_a_feature.md). Also, the
maintainers can be contacted at any time to learn more about how to get
involved.

In the interest of getting more newer folks involved with kops, we are starting to
tag issues with `good-starter-issue`. These are typically issues that have
smaller scope but are good ways to start to get acquainted with the codebase.

We also encourage ALL active community participants to act as if they are
maintainers, even if you don't have "official" write permissions. This is a
community effort, we are here to serve the Kubernetes community. If you have an
active interest and you want to get involved, you have real power! Don't assume
that the only people who can get things done around here are the "maintainers".

We also would love to add more "official" maintainers, so show us what you can
do!

What this means:

__Issues__

- Help read and triage issues, assist when possible.
- Point out issues that are duplicates, out of date, etc.
  - Even if you don't have tagging permissions, make a note and tag maintainers (`/close`,`/dupe #127`).

__Pull Requests__

- Read and review the code. Leave comments, questions, and critiques (`/lgtm` ).
- Download, compile, and run the code and make sure the tests pass (make test).
  - Also verify that the new feature seems sane, follows best architectural patterns, and includes tests.

This repository uses the Kubernetes bots.  See a full list of the commands [here](
https://go.k8s.io/bot-commands).


## Office Hours

Kops maintainers set aside one hour every other week for **public** office hours. This time is used to gather with community members interested in kops. This session is open to both developers and users.

For more information, checkout the [office hours page.](office_hours.md)

### Other Ways to Communicate with the Contributors

Please check in with us in the [#kops-users](https://kubernetes.slack.com/messages/kops-users/) or [#kops-dev](https://kubernetes.slack.com/messages/kops-dev/) channel. Often-times, a well crafted question or potential bug report in slack will catch the attention of the right folks and help quickly get the ship righted.

## GitHub Issues


### Bugs

If you think you have found a bug please follow the instructions below.

- Please spend a small amount of time giving due diligence to the issue tracker. Your issue might be a duplicate.
- Set `-v 10` command line option and save the log output. Please paste this into your issue.
- Note the version of kops you are running (from `kops version`), and the command line options you are using.
- Open a [new issue](https://github.com/kubernetes/kops/issues/new).
- Remember users might be searching for your issue in the future, so please give it a meaningful title to helps others.
- Feel free to reach out to the kops community on [kubernetes slack](https://github.com/kubernetes/community/blob/master/communication.md#social-media).


### Features

We also use the issue tracker to track features. If you have an idea for a feature, or think you can help kops become even more awesome follow the steps below.

- Open a [new issue](https://github.com/kubernetes/kops/issues/new).
- Remember users might be searching for your issue in the future, so please give it a meaningful title to helps others.
- Clearly define the use case, using concrete examples. EG: I type `this` and kops does `that`.
- Some of our larger features will require some design. If you would like to include a technical design for your feature please include it in the issue.
- After the new feature is well understood, and the design agreed upon we can start coding the feature. We would love for you to code it. So please open up a **WIP** *(work in progress)* pull request, and happy coding.
