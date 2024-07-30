# EpicEnv

An epic environment manager

## Quick Start

### Install EpicEnv

```
go install github.com/epicenv@latest
```

### Initialize EpicEnv

```
epicenv init
```

### Add environment variables

You can add individual variables with:

```
epicenv set [KEY] [VALUE]
```

_Pro-tip: put a space before typing the command to prevent it from being added to your shell history (thus preventing leaks)_

You can also import an existing `.env` file with:

```
epicenv import [PATH]
```

### Invite colleagues

**2) Invite your colleagues via their Github usernames**

_They must have at least 1 key added to their account, which can be checked at `github.com/{username}.keys`_

**3) You commit and push the encrypted environment to your git repo**

Now everyone can share updated env vars!

**4) You source the env **

## Motivation

Local environment management is pretty hacky when you have multiple people working on a production system.

There are great managements systems like kubeseal for kubernetes to keep encrypted secrets right in git, but similar tools for local environment management is lacking.

Considering each developer has (or at least should have) their own specific environment values like database credentials, team members can quickly get left behind in terms of keeping up to date with the required environment variable needed to run the project locally.

This spirals out of control until eventually everyone is fighting over the staging deployment as their dev environment, and nobody has run the service locally in 8 months.

EpicEnv fixes this.

EpicEnv creates encrypted environment variables for you local environment, sharing what can be, and replacing developer-specific values where required.

Environment variables are encrypted using your `github.com/{username}.keys` keys, and you "invite" your collaborators to the environment.

## The Audit Log

Within the `.epicenv` folder, there is an `auditlog.txt`.

This file contains logs of every operation performed on the environment, so you can follow the history of invitations, variable changes, etc.

These are the same logs written to the console.# EpicEnv
