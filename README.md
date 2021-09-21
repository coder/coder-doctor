# Coder Doctor 🧑‍⚕️

Coder Doctor is a command-line diagnostic tool for checking that a
given platform can run the Coder control plane and workspaces.

## Supported Platforms

Currently, the following platforms are supported, with the following
preflight checks:

### Kubernetes

- Kubernetes Version: checks that the selected Coder version is
  compatible with the Kubernetes control plane.
- Helm Version: checks the locally-installed Helm version for
  compatibility with the requested version of Coder.
- Kubernetes RBAC: checks that the service account has the required
  permissions to run Coder.
- Kubernetes Resources: checks that the cluster has the required
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

## Development

To run from source, clone the repository and run:

```bash
go run . help
```

The `Makefile` also includes various helpful targets to run
linters and tests, but standard Go tools (e.g. `go test`)
should work as well.

## Feedback

We love feedback! Please [open an issue] or [submit a pull request]
with any ideas for improving this.

[open an issue]: https://github.com/cdr/coder-doctor/issues
[submit a pull request]: https://github.com/cdr/coder-doctor/pulls

**Note:** This tool is in `beta` state and Coder offers no compatibility
guarantees, either for the tool itself or any public Go APIs. Most code
is kept in the `internal` package to make this clear, and will be promoted
to an externally-importable package once things stabilize.

## Copyright and License

coder-doctor preflight diagnostic tool

Copyright (C) 2021 Coder Technologies, Inc. &lt;https://coder.com&gt;

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
