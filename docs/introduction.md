# Introduction

Welcome to intro guide to Entropy! This guide is the best place to start with Entropy. We cover what Entropy is, what problems it can solve, how it compares to existing DevOps practices, and how you can get started using it.

### What is Entropy?

Entropy is a framework to safely and predictably create, change, and improve modern cloud applications and infrastructure using familiar languages, tools, and engineering practices.

### Why is entropy required?

Data engineering has a lot of infrastructure and application deployments happening each day. Currently while some applications are available via self-serve APIs and UI, most of the infrastructure components are still being set up by a gitops based flow.

To reduce the support effort that goes into provisioning and scaling of these infrastructure components we aim to onboard most of them to self-serve APIs.

### What features are we seeking?

 - ***Resource versioning*** Support multiple versions of a resource at a time.

 - ***Config schema versioning*** Schemas shall be versioned and each version of the resource can have a different version of schema.

- ***Payload schema versioning*** Support versioned payload which will allow decoupling the UI from the backend if required.

- ***Rollbacks using config history***Enable maintaining config history in order to support rollback.

- ***Typed config*** We aim to come up with a DSL that would help us define the payload once and allow us to use the same for rendering the UI if required.

### Entropy vs existing solutions

We found three frameworks that are similar to how we visualize entropy.
 - [Pulumi Automation API](https://www.pulumi.com/docs/guides/automation-api/)
 - [Terraform CDK](https://github.com/hashicorp/terraform-cdk/)
 - [Crossplane](https://crossplane.io/)

All have their own share of pros and cons but the major con being they do not allow a way to transform inputs for specific operations and reconcile towards the final state. Whereas entropy gives the ability to apply modifications on top of an existing resource.

Entropy would work as a layer on top of all these and manage configurations and secrets, while the actual deployment can be taken care of by any of the above solutions or more raw k8s APIs, Helm, or k8s + operators for deployments.