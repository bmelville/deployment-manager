# User Guide

## Overview

Deployment Manager is a service in your Kubernetes cluster which provides
declarative configuration for your various Kubernetes resources. This means you
can define all of the resources that make up your application in one file, then
you can give it to the DM service to deploy all of the resources together.

Declarative configuration is great because it lets you know exactly what
resources make up your application or applications, and lets you check your
application configuration into your source control system right along-side your
application code.

## What is this?

This is a guide that will walk you through using DM to configure some example
applications and will show you how to use the powerful features of DM, like
building re-usable types and using pre-built types from our type registry.

If at any time you get bored and want to jump to the end, you should really
check out some real working
[example configs](https://github.com/kubernetes/deployment-manager/tree/master/examples)
and templates from our
[template registry](https://github.com/kubernetes/application-dm-templates).

## Before you begin

Before you begin, this guide needs you to have DM set up and running in a
kubernetes cluster somewhere, and you'll need to build our client to interact
with it.

You can find steps to get this all set up in our
[README](https://github.com/kubernetes/deployment-manager/blob/master/README.md)
under the "Installing Deployment Manager" and "Using Deployment Manager"
sections.

We also assume you're pretty familiar with kubernetes and its concepts. If you
need a refresher or are brand new to the game, you might want to check out the
[Kubernetes User Guide](https://github.com/kubernetes/kubernetes/tree/master/docs/user-guide).

## Get started

To get started let's jump in and [deploy something](deploy-something.md).
