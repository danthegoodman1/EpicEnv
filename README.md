# EpicEnv

An epic environment manager to fix local environment variable management among git collaborators.

It's like python virtual envs + overlayfs, but for environment variables.

All of your environments encrypted and managed in git, with basic permissions via an invitation system.

_Currently only targets macOS and Linux, but it appears to work on windows?_

<!-- TOC -->
- [EpicEnv](#epicenv)
  - [Quick Start](#quick-start)
    - [Install EpicEnv](#install-epicenv)
    - [Initialize EpicEnv](#initialize-epicenv)
    - [Overlay Environments](#overlay-environments)
    - [Set shared environment variables](#set-shared-environment-variables)
    - [Add personal environment variables](#add-personal-environment-variables)
    - [Invite collaborators](#invite-collaborators)
    - [Add headless keys](#add-headless-keys)
    - [Source the environment](#source-the-environment)
    - [Deactivate the environment](#deactivate-the-environment)
    - [Commit the `.epicenv` directory](#commit-the-epicenv-directory)
    - [Remove variables](#remove-variables)
  - [Motivation](#motivation)
  - [Safety](#safety)
    - [Encryption](#encryption)
    - [Preventing personal variables from being added globally](#preventing-personal-variables-from-being-added-globally)
    - [Rotating keys](#rotating-keys)
  - [Developing](#developing)
<!-- TOC -->

## Quick Start

### Install EpicEnv

```
go install github.com/danthegoodman1/epicenv@latest
```

### Initialize EpicEnv

Within your project directory, run:

```
epicenv init GITHUB_USERNAME [-e ENVIRONMENT]
```

For example:
```
epicenv init danthegoodman1
# Creates the default "local" environment

epicenv init danthegoodman1 -e staging
# Creates a "staging" environment
```

Your GitHub username is required as the first argument to fetch your public SSH keys. The environment name can be specified with the `-e` flag and defaults to "local" if not provided.

You can use different environments to link your local environment to different infrastructure, such as staging and production.

This will create a `.epicenv` directory, and add `.epicenv/*/personal` to your `.gitignore`.

### Overlay Environments

Overlay environments let you create environments that inherit from a base environment and override specific variables. This is useful for scenarios like having a `testing` environment that shares most settings with `local` but uses a different S3 bucket.

```
epicenv init -e testing --overlay local
```

Overlays can be stacked arbitrarily deep:

```
epicenv init -e agent-testing --overlay testing
```

This creates a chain: `local` → `testing` → `agent-testing`

When you load `agent-testing`, variables are resolved by stacking each layer:
1. Load `local` secrets
2. Apply `testing` overrides
3. Apply `agent-testing` overrides

**Example:**
```bash
# Set base values in local
epicenv set DB_HOST localhost -e local
epicenv set S3_BUCKET s3://local-bucket -e local
epicenv set LOG_LEVEL debug -e local

# Override S3 in testing
epicenv set S3_BUCKET s3://test-bucket -e testing

# Override LOG_LEVEL in agent-testing
epicenv set LOG_LEVEL info -e agent-testing

# Loading agent-testing gives:
# DB_HOST=localhost (from local)
# S3_BUCKET=s3://test-bucket (from testing)
# LOG_LEVEL=info (from agent-testing)
```

**Key behaviors:**
- Overlays inherit encryption keys and users from their root environment
- Inviting users to an overlay will add them to the root environment (with a warning)
- Personal secrets are also stacked, with each layer able to add or override
- Removing a variable from an overlay only removes it from that layer; if it exists in an underlay, it will still be visible

### Set shared environment variables

You can set individual variables with:

```
epicenv set KEY [VALUE] -e myenv
```

If you omit the value, then it will ask for the value via hidden stdin:

```
epicenv set MYVAR -e myenv

MYVAR> 🔑
```
You can also import an existing `.env` file with:

```
epicenv import PATH
```

If a line ends with `#personal` like:

```ini
shared_thing="this val is shared"
personal_thing="this val is personal" #personal
```

then it will automatically be added as a personal variable. This is very convenient if you have an existing `.env` file to import that has many mixed shared and personal env vars. EpicEnv will log when it imports a personal value.

Imports will overwrite existing values, using the rules for personal flag collisions mentioned below.

### Add personal environment variables

For something like database or AWS credentials, you'll want to use (and enforce) using local credentials.

```
epicenv set KEY VALUE -e myenv -p
```

This will mark the env var as personal, preventing it from being committed to git.

If someone sources the environment in the future without setting their own personal value, they will see a warning in the console notifying them that they are missing part of the environment.

If they attempt to write a personal env var as a shared env var (omitting `-p`), EpicEnv will recognize this and assume they meant personal, while throwing a warning (see safety section for more).

If they attempt to write a shared env var as personal, it will reject the operation and ask the user to `rm` the variable to change it to a personal variable.

### Invite collaborators

Collaborators are invited via their GitHub usernames.  They must have at least 1 RSA key added to their account, which can be checked at `github.com/{username}.keys`.

```
epicenv invite danthegoodman1
```

and you can uninvite them with

```
epicenv uninvite danthegoodman1
```

Which will re-encrypt all values, removing their access.

**Note you still need to rotate your secrets if someone leaves your team!**

### Add headless keys

You can also add public keys directly from files (not associated with GitHub users) using the `--path` option:

```
epicenv invite keyname --path /path/to/public_key.pub
```

These "headless keys" can be managed just like GitHub users:

```
epicenv uninvite keyname
```

This is useful for CI/CD systems, service accounts, or other automated systems that need access to environment variables.

### Source the environment

```
source .epicenv/myenv/activate
```

Your local shell will decrypt and load the variables into the environment!

You can also run this command to update the local environment when changes are pulled from GitHub.

### Deactivate the environment

You can deactivate the environment, which will return environment variables to their previous state (previous value or unset).

```
epic-deactivate
```

If you are already in an environment, deactivate will automatically be run when you switch or refresh.

### Commit the `.epicenv` directory

```
git add .epicenv/*
git commit -m "add epicenv"
```

### Remove variables

You can remove global and personal variables with:

```
epicenv rm KEY -e myenv
```


## Motivation

Local environment management is pretty hacky when you have multiple people working on a production system.

There are great managements systems like kubeseal for kubernetes to keep encrypted secrets right in git, but similar tools for local environment management is lacking.

Considering each developer has (or at least should have) their own specific environment values like database credentials, team members can quickly get left behind in terms of keeping up to date with the required environment variable needed to run the project locally.

This spirals out of control until eventually everyone is fighting over the staging deployment as their dev environment, and nobody has run the service locally in 8 months.

EpicEnv aims to fix this.

EpicEnv stores your environments encrypted in git, and decrypts them when you activate the environment. You can share variables that can be, and replace developer-specific ("personal") variables where required (e.g. DB or AWS credentials).

Environment variables are encrypted using RSA keys found in your `github.com/{username}.keys`, and you "invite" your collaborators to the environment.

Everything in git is encrypted and nobody has to manage local `.env` files or prevent them from being committed.

This is also great for streamers, as they never have to worry about accidentally opening a `.env` file and spilling their production secrets to viewers >.<

In fact the ability to be able to stream [Tangia](https://www.tangia.co) development was the inception of this idea, as I (used to) happen to have very sensitive environment variables decrypted in my local `.env` file 😬

## Safety

### Encryption

Variables are encrypted with AES GCM mode, the symmetric key is encrypted with the RSA keys from each collaborator.

### Preventing personal variables from being added globally

If you attempt to `epicenv set` on a variable that is marked as personal, that set will update the personal variable instead of adding to the global variables to prevent personal values from being leaked via git.

A warning will be thrown when this occurs.

To make a personal variable shared, first `rm` the personal variable, then set it again as shared. Vice-versa for making a shared variable personal.

### Rotating keys

We explicitly DO NOT change the symmetric key for decryption of environment variables when you uninvite a collaborator to FORCE YOU TO ROTATE KEYS IF SOMEONE LEAVES YOUR TEAM!!!!!!!!!!

## Developing

Need to:

```
git tag -a v#.#.# -m "some patch notes"
git push origin v#.#.#
```
