# EpicEnv

An epic environment manager

## Quick Start

### Install EpicEnv

```
go install github.com/epicenv@latest
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

### Add personal environment variables

For something like database or AWS credentials, you'll want to use (and enforce) using local credentials.

```
epicenv set KEY VALUE -e myenv -p
```

This will mark the env var as personal, preventing it from being committed to git.

If someone sources the environment in the future without setting their own personal value, they will see a warning in the console notifying them that they are missing part of the environment.

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

**Note that if they do not pull, or do not re-source, then the values are accessible by them. This is not a replacement for rotating secrets!**

### Source the environment

```
source .epicenv/myenv/activate
```

Your local shell will decrypt and load the variables into the environment!

You can also run this command to update the local environment when changes are pulled from GitHub.

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

EpicEnv fixes this.

EpicEnv creates encrypted environment variables for you local environment, sharing what can be, and replacing developer-specific values where required.

Environment variables are encrypted using your `github.com/{username}.keys` keys, and you "invite" your collaborators to the environment.

## Safety

### Preventing personal variables from being added globally

If you attempt to `epicenv set` on a variable that is marked as personal, that set will update the personal variable instead of adding to the global variables to prevent personal values from being leaked via git.

A warning will be thrown when this occurs.

To make a personal variable shared, first `rm` the personal variable, then set it again as shared. Vice-versa for making a shared variable personal.