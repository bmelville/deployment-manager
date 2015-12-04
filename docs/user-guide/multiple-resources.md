# Deploying Mulitple Resources

## Overview

Of course, the whole point of DM is to allow you to specify all the resources
that make up your application deployment, and likely you'll have quite a few
in your configurations.

If you just came from [deploying something](deploy-something.md), you know how
to deploy a single resource (we deployed a `Pod`), so let's take a look at the
logical next step in that.

## A replicated service

Nginx gets really powerful when you can scale it out, and of course you'll want
it exposed to the world to get all that beautiful traffic. To do that, we can
deploy a `ReplicationController` and a `Service` together in the same config:

```
resources:
- name: nginx-rc
  type: ReplicationController
  properties:
    apiVersion: v1
    kind: ReplicationController
    metadata:
      name: nginx-controller
      spec:
        replicas: 2
        selector:
          app: nginx
        template:
          metadata:
            labels:
              app: nginx
          spec:
            containers:
            - name: nginx
              image: nginx
              ports:
              - containerPort: 80
- name: nginx-service
  type: Service
  properties:
  apiVersion: v1
  kind: Service
  metadata:
    name: nginx-service
    spec:
      ports:
      - port: 8000
        targetPort: 80
        protocol: TCP
    selector:
      app: nginx
```

Just as before, we can save these two resources to the file
`replicated-nginx.yaml` and deploy using the command:

```
dm --name replicated-nginx deploy replicated-nginx.yaml
```

## Resource ordering

When your deployment has mulitple resources, DM makes no guarantee on the order
in which they're created or deleted, and it may even parallelize resources when
it can!

## Viewing your creation

You should now have two resources, a `ReplicatedController` and a `Service`.
Use `kubectl` to view them as you normally would:

```
kubectl get rc,services
```

## Deleting things

Again, you'll want to clean up when you're done with your resources:

```
dm delete replicated-nginx
```

## Next steps

Eventually your application is going to have many services talking to each
other, so we better take a look at [connecting services](connecting-services.md)
before we get much further.

