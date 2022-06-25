# Layout

Here is how the basic layout of Entropy looks like:

```
+ entropy
|--+ cli/
|  |--+ serve.go                 {Load configs, invoke server.Serve()}
|  |--+ migrate.go               {Load configs, execute store migration}
|--+ core/
|  |--+ resource/
|  |  |--+ resource.go                 {Resource type, Store iface, pure functions on Resource}
|  |  |--+ service.go                  {Service struct that operates using the above}
|  |--+ module/
|  |  |--+ module.go                   {Module iface, module related Error types}
|  |  |--+ service.go                  {Service that operates using Module}
|--+ modules/
|  |--+ firehose/
|  |  |--+ firehose.go                 {Module implementation}
|  |--+ kubernetes/
|  |  |--+ kubernetes.go               {Module implementation}
|--+ pkg/
|  |--+ logger/
|  |--+ telemetry/                  {NR, StatsD, opencensus integrations}
|--+ internal/
|  |--+ server/
|  |  |--+ v1/
|  |  |  |--+ resource.go           {Resource CRUD handlers}
|  |  |--+ server.go                {Server setup, Serve() function, etc.}
|  |--+ store/
|  |  |--+ mongodb                  {Implement resource.Store, module.Store using MongoDB}
|  |  |--+ inmemory                 {Implement resource.Store, module.Store using InMem}
|--+ docs/
|--+ main.go                        {Setup cobra+viper & add commands}

```

***Some highlights:***

Domain oriented packages inside `core/` (resource, module, provider)

`internal/` for keeping packages that should not be imported by any other projects (e.g., server, store, etc.)

`pkg/` for truly reusable (independent of entropy specific things) packages.

Interfaces are defined on the client side (e.g., ResourceStore interface is in resource package along with the resource-service which is the actual user of the store.).

Mocks for interfaces are defined close to the interface definition in an isolated `mocks/` package.