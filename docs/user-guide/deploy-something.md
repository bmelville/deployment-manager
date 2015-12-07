# Deploy Something

## Overview

The DM configuration language in a nutshell is a list of resources in YAML.
Each resource has a `name`, `type`, and `properties`.

* `name` will be the name of the resource created,
* `type` is what you would typically provide as the "kind" when interacting with
  kubernetes and `kubectl`,
* `properties` is the properties you would normally use when defining your
  resource with `kubectl`.

## Let's deploy something

Let's jump right in and deploy something. Here is an nginx pod we can deploy:

```
resources:
- name: nginx
  type: Pod
  properties:
    apiVersion: v1
    kind: Pod
    metadata:
      name: nginx
      spec:
        containers:
        - name: nginx
          image: nginx
          ports:
          - containerPort: 80
```

As you can see, the properties of the `nginx` `Pod` is exactly what you would
normally pass to `kubectl` when deploying from the command-line.

Now, save that file into `nginx.yaml` or get it from
[nginx.yaml](https://github.com/kubernetes/deployment-manager/blob/master/examples/user-guide/nginx/nginx.yaml),
and deploy it using the `dm` client:

```
dm --name nginx deploy nginx.yaml
```

## Viewing your creation

Assuming that DM hasn't come back with any errors, you can see the resources
you've created just as you normally do with `kubectl`:

```
kubectl get pods
```

You can always look at the deployment state:

```
dm get nginx
```

**NOTE**: when we deployed we named our deployment `nginx`, otherwise it would
default to the configuration file name.

You can also look at the configuration you deployed by getting the manifest
name from the deployment and looking at the manifest:

```
dm manifest nginx/<manifest-name>
```

## Deleting things

When you're done, you can delete the deployment to clean up all the resources
it deployed:

```
dm delete nginx
```

## Next steps

Next let's take a look at deploying [multiple resources](multiple-resources.md)
to really make use of config to define our whole application.
