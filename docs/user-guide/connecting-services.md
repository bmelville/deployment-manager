# Connecting Services Together

## Overview

We've looked at a configuration that creates a replicated service when we were
deploying [multiple resources](multiple-resources.md). Chances are your
configuration is going to have more than one of these things, and you'll need
to figure out how to connect the pieces.

## Getting a little persistent

A common example of what we want to take a look at is a website that needs to
store posts made to it. Let's see what it would take to deploy a guestbook
that uses redis to persist its data.

## Configure the guestbook

First we'll need the guestbook replicated service:

```
```

For this entire example you'll want to save this and the following pieces of
configuration in the same file `guestbook.yaml`.

## Configure redis

The redis we're going to configure for storage consists of a master and slaves.
Both of these will need a `Service` and a `ReplicationController`.

### Redis master

Add the resources for the redis master to the guestbook configuration:

```
```

### Redis slaves

And then add the redis slaves as well:

```
```

## Connecting services

Both the guestbook and the redis slaves need to know how to talk to the redis
master. Since the master is a kubernetes service, the guestbook can either
hard-code the name of the service in application code, or pass it in through an
environment variable.

This example chooses to hard-code the names within the application, but if

```
```

An unfortunate side-effect of this is that you now have two places where you're
defining the redis service name, once where it is configured, and once where it
is used. Don't worry about this, we'll look at how to fix that next.

## Next steps

You might notice your configuration can get quite long, and you have a lot of
the same data repeated throughout, like ports, and service names.

So next let's take a look at [using templates](using-templates.md) to
parameterize your configuration and make re-usable components out of it.

