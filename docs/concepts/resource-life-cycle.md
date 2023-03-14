# Resource Life Cycle

## Resource

A resource represents the current state and desired state of real-world entities. A resource definition look like 

```
type Resource struct {
Resource metadata.
    URN       string
    Name      string
    Kind      string
    Project   string 
    Labels    map[string]string      
    Version   int                   
    CreatedAt time.Time             
    UpdatedAt time.Time             

Resource spec & current-state.
    Spec  Spec 
    State State
}

type Spec struct {
    Configs      map[string]interface{}
    Dependencies map[string]string
}

type State struct {
    ModuleData   json.RawMessage
    Status string
    Output Output
}

type Output map[string]interface{}
```

The `Resource` definition is self explanatory. It has the `Spec` field which holds the `Configs` and `Dependencies` of a resource. The `State` field has three parts, `Status` holds the current status of a resource, `Output` holds the outcome of the latest action performed while `Data` holds the transactory information which might be used to perform actions on the reosurce.

For instance, a [firehose](https://github.com/goto/firehose) resource looks like:

```
{
    "name": "foo",
    "parent": "bar",
    "kind": "firehose",
    "spec": {
            "configs": {
                "log_level": "INFO"
            },
            "dependencies": {
                "deployment_cluster": "orn:entropy:kubernetes:godata"
            }
    },
    "state": {
        "status": "STATUS_PENDING"
    }
}
```

## Resource Lifecycle - The Plan & Sync Approach

We use a Plan and Sync approach for the resource lifecycle in Entropy. For illustration, we will take you through each of the steps in the lifecycle for a firehose resource.

### 1. Create a resource

```
POST /api/v1beta1/resources

{
    "name": "foo",
    "parent": "bar",
    "kind": "firehose",
    "configs": {
        "log_level": "INFO"
    },
    "dependencies": {
        "deployment_cluster": "orn:entropy:kubernetes:godata"
    }
}
```

### 2. Plan phase

Plan validates the action on the current version of the resource and returns the resource with config/status/state changes (if any) applied. Plan DOES NOT have side-effects on anything other thing than the resource.

Here is the resource returned 

```
{
    "urn": "orn:entropy:firehose:bar:foo",
    "created_at": "2022-04-28T11:00:00.000Z",
    "updated_at": "2022-04-28T11:00:00.000Z",
    "name": "foo",
    "parent": "bar",
    "kind": "firehose",
    "configs": {
        "log_level": "INFO"
    },
    "dependencies": {
        "deployment_cluster": "orn:entropy:kubernetes:godata"
    },
    "state": {
        "status": "STATUS_PENDING",
        "output": {},
        "data": {
            "pending": ["helm_release"]
        }
    }
}
```

### 3. The Sync Phase

Sync is called repeatedly by Entropy core until the returned state has `StatusCompleted`.Module implementation is free to execute an action in a single Sync() call or split into multiple steps for better feedback to the end-user about the progress.

A job-queue model is used to handle sync operations. Every mutation (create/update/delete) on resources will lead to enqued jobs which will be processed later by workers.

### 4. Get resource (After Sync completion)

```
{
    "urn": "orn:entropy:firehose:bar:foo",
    "kind": "firehose",
    "name": "foo",
    "project": "bar",
    "created_at": "2022-04-28T11:00:00.000Z",
    "updated_at": "2022-04-28T11:00:00.000Z",
    "configs": {
        "log_level": "INFO"
    },
    "dependencies": {
        "deployment_cluster": "orn:entropy:kubernetes:godata"
    },
    "state": {
        "status": "COMPLETED",
        "output": {
            "app": "entropy-firehose-bar-foo"
        }
    }
}
```

### 5. Execute Action

```
POST /api/v1beta1/resources/orn:entropy:firehose:bar:foo/execute

{
    "action": "increase_log_level"
}
```

This will trigger Plan again, and leave the resource in `STATUS_PENDING` state. These resource will be later picked by the Sync in entropy core.

```
{
    "urn": "orn:entropy:firehose:bar:foo",
    "kind": "firehose",
    "name": "foo",
    "project": "bar",
    "created_at": "2022-04-28T11:00:00.000Z",
    "updated_at": "2022-04-28T11:00:00.000Z",
    "configs": {
        "log_level": "WARN"
    },
    "dependencies": {
        "deployment_cluster": "orn:entropy:kubernetes:godata"
    },
    "state": {
        "status": "PENDING",
        "output": {
            "app": "entropy-firehose-bar-foo"
        },
        "data": {
            "pending": ["helm_release"]
        }
    }
}
```