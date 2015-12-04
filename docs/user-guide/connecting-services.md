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
resources:
- name: guestbook-service
  type: Service
  properties:
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        app: guestbook-service
      name: guestbook-service
      namespace: default
    spec:
      ports:
      - name: guestbook
        port: 80
      selector:
        app: guestbook
      type: LoadBalancer
- name: guestbook-rc
  type: ReplicationController
  properties:
    apiVersion: v1
    kind: ReplicationController
    metadata:
      labels:
        app: guestbook-rc
      name: guestbook-rc
      namespace: default
    spec:
      replicas: 3
      selector:
        app: guestbook
      template:
        metadata:
          labels:
            app: guestbook
        spec:
          containers:
          - image: gcr.io/google_containers/example-guestbook-php-redis:v3
            name: guestbook
            ports:
            - containerPort: 80
              name: guestbook
```

For this entire example you'll want to save this and the following pieces of
configuration in the same file `guestbook.yaml`.

## Configure redis

The redis we're going to configure for storage consists of a master and slaves.
Both of these will need a `Service` and a `ReplicationController`.

### Redis master

Add the resources for the redis master to the guestbook configuration:

```
- name: redis-master
  type: Service
  properties:
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        app: redis-master
      name: redis-master
      namespace: default
    spec:
      ports:
      - name: master
        port: 6379
        targetPort: 6379
      selector:
        app: redis-master
- name: redis-master-rc
  type: ReplicationController
  properties:
    apiVersion: v1
    kind: ReplicationController
    metadata:
      labels:
        app: redis-master-rc
      name: redis-master-rc
      namespace: default
    spec:
      replicas: 1
      selector:
        app: redis-master
      template:
        metadata:
          labels:
            app: redis-master
        spec:
          containers:
          - image: redis
            name: master
            ports:
            - containerPort: 6379
              name: master
```

### Redis slaves

And then add the redis slaves as well:

```
- name: redis-slave
  type: Service
  properties:
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        app: redis-slave
      name: redis-slave
      namespace: default
    spec:
      ports:
      - name: worker
        port: 6379
      selector:
        app: redis-slave
- name: redis-slave-rc
  type: ReplicationController
  properties:
    apiVersion: v1
    kind: ReplicationController
    metadata:
      labels:
        app: redis-slave-rc
      name: redis-slave-rc
      namespace: default
    spec:
      replicas: 2
      selector:
        app: redis-slave
      template:
        metadata:
          labels:
            app: redis-slave
        spec:
          containers:
          - env:
            - name: GET_HOSTS_FROM
              value: env
            - name: REDIS_MASTER_SERVICE_HOST
              value: redis-master
            image: kubernetes/redis-slave:v2
            name: worker
            ports:
            - containerPort: 6379
              name: worker
```

You are now ready to deploy your working guestbook:

```
dm --name guestbook deploy guestbook.yaml
```

Now when you view your resources, you should see three `ReplicationControllers`
and three `Services`:

```
kubectl get rc,services
```

Once the `frontend-service` has an external IP, you can connect to it in your
browser and play!

## Deleting things

Again, you'll want to clean up when you're done with your resources:

```
dm delete guestbook
```

## Connecting services

Both the guestbook and the redis slaves need to know how to talk to the redis
master. Since the master is a kubernetes service, containers which consume it
can either hard-code the name of the service in application code, or pass it in
through an environment variable.

In this example, the guestbook has chosen to hard-code the name of the service
to `redis-master`, while the redis slave has chosen to pass the service name in
through the `REDIS_MASTER_SERVICE_HOST` environment variable.

## Example configuration

Please be sure to check out
[guestbook.yaml](https://github.com/kubernetes/deployment-manager/blob/master/examples/user-guide/guestbook/guestbook.yaml)
for the full example from this walk-through.

## Next steps

You might notice your configuration can get quite long, and you have a lot of
the same data repeated throughout, like ports and service names.

Next we'll take a look at [using templates](using-templates.md) to solve some
of these problems by parameterizing your configuration and making re-usable
components out of it.

