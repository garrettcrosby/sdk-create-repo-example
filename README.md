# sdk-create-repo-example

A minimal Go program showing how to call the Chainguard platform API
directly using `chainguard.dev/sdk` — specifically, how to:

1. Take a Chainguard access token from the environment (the kind
   `chainctl auth token` prints).
2. Open platform gRPC clients with bearer credentials.
3. Call `Registry.CreateRepo` to create a new image repository.

The same pattern works for any other gRPC method on the platform —
`IAM`, `Notifications`, `Packages`, etc. Swap the client and request
type.

## Prerequisites

- Go 1.25+ (matches `chainguard.dev/sdk`).
- `chainctl` installed and logged in:
  ```sh
  chainctl auth login
  ```
- A Chainguard organization you can create repos in, and the parent
  group/folder UIDP under which the repo should live (e.g. from
  `chainctl iam folders list` or `chainctl iam organizations list`).

## Build

```sh
go build ./...
```

## Run (dry-run)

```sh
export CHAINGUARD_TOKEN=$(chainctl auth token)

./sdk-create-repo-example \
    --parent <PARENT_GROUP_UIDP> \
    --name   my-new-repo \
    --tier   APPLICATION
```

The dry-run prints the `CreateRepoRequest` it would send and stops
without mutating anything.

## Run (actually create)

Add `--apply`:

```sh
./sdk-create-repo-example \
    --parent <PARENT_GROUP_UIDP> \
    --name   my-new-repo \
    --tier   APPLICATION \
    --apply
```

## Pointing at a non-prod environment

Set both `--api` and the chainctl config that minted the token, e.g.:

```sh
./sdk-create-repo-example \
    --api https://console-api.chainops.dev \
    ...
```

(Make sure `chainctl auth token` was issued against the same
environment.)

## How it maps to chainctl

`chainctl images repos create` does the same gRPC call this example
does. See `chainctl/pkg/commands/images_repos.go:470` in the Chainguard
mono repo:

```go
i.Clients.V1().Registry().Registry().CreateRepo(ctx, &createRepoRequest)
```
