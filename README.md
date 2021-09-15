# Coder Doctor 🧑‍⚕️

Coder Doctor is a command-line diagnostic tool for checking that a
given platform can run the Coder control plane and workspaces.

## Supported Platforms

Currently, the following platforms are supported, with the following
preflight checks:

### Kubernetes

- Kubernetes Version: checks the cluster version for compatibility
with the requested version of Coder.
- Helm Version: checks the locally-installed Helm version for
compatibility with the requested version of Coder.
- Kubernetes RBAC: checks that the service account has the required
permissions to run Coder.
- Kubernetes Resosurces: checks that the cluster has the required
resource types available to run Coder.

## Usage

To check if your Kubernetes cluster is ready to install Coder, run:

```console
coder-doctor check kubernetes
```

For more information, you can run:

```console
coder-doctor -h
```

## Feedback

We love feedback! If you have any issues or suggestions, please feel
free to submit an [issue](https://github.com/cdr/coder-doctor/issues) or [pull request](https://github.com/cdr/coder-doctor/pulls).

**Note:** This project is in `alpha` state and Coder offers no
compatibility guarantees, either for the tool itself or any public Go
APIs. Most code is kept in the `internal` package to make this clear,
and will be promoted to an externally-importable package once the tool
enters `beta` state.
