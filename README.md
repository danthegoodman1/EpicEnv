# EpicEnv

An epic environment manager to fix local environment variable management among git collaborators.

It's like python virtual envs, but for environment variables.

All of your environments encrypted and managed in git, with basic permissions via an invitation system.

_Currently only supports macOS and Linux_

<!-- TOC -->
* [EpicEnv](#epicenv)
  * [Quick Start](#quick-start)
    * [Install EpicEnv](#install-epicenv)
    * [Initialize EpicEnv](#initialize-epicenv)
    * [Add shared environment variables](#add-shared-environment-variables)
    * [Add personal environment variables](#add-personal-environment-variables)
    * [Invite collaborators](#invite-collaborators)
    * [Source the environment](#source-the-environment)
    * [Deactivate the environment](#deactivate-the-environment)
    * [Commit the `.epicenv` directory](#commit-the-epicenv-directory)
    * [Remove variables](#remove-variables)
  * [Motivation](#motivation)
  * [Safety](#safety)
    * [Encryption](#encryption)
    * [Preventing personal variables from being added globally](#preventing-personal-variables-from-being-added-globally)
    * [Rotating keys](#rotating-keys)
<!-- TOC -->

## Quick Start

### Install EpicEnv

```
go install github.com/danthegoodman1/epicenv@latest
```

### Initialize EpicEnv

Within your project directory, run:

```
epicenv init
```

This will walk you through creating an EpicEnv `environment`. You can use different environments to link your local environment to different infrastructure, such as staging and production.

This will create a `.epicenv` directory, and add `.epicenv/*/personal` to your `.gitignore`.

### Add shared environment variables

You can add individual variables with:

```
epicenv set KEY VALUE -e myenv
```

You can run this again to replace values as well.

_Pro-tip: put a space before typing the command to prevent it from being added to your shell history (thus preventing leaks). You can also use `-i` to have the command prompt for the input from stdin_

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

### Source the environment

```
source .epicenv/myenv/activate
```

Your local shell will decrypt and load the variables into the environment!

You can also run this command to update the local environment when changes are pulled from GitHub.

### Deactivate the environment

You can deactivate the environment, which will return environment variables to their previous state (previous value or unset).

```
deactivate
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
