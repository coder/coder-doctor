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

## Installation

You can manually download the latest [release](https://github.com/cdr/coder-doctor/releases):

1. Click a release and download the tar file for your operating system
   (ex: coder-doctor-linux-amd64.tar.gz)
1. (Optional, recommended) Download the `checksums.txt` for the desired
   release and verify the `sha256` checksum of the file you downloaded.
1. Extract the `coder-doctor` binary.
1. Move the `coder-doctor` binary to somewhere in your `$PATH`.

Alternatively, use the below command, replacing `$VERSION`, `$OSTYPE`
(one of `darwin`, `linux`, `windows`) and `$ARCH` (one of `amd64`, `arm64`)
as required:

```shell-session
$ curl -fSsL "https://github.com/cdr/coder-doctor/releases/latest/download/coder-doctor_${VERSION}_${OSTYPE}_${ARCH}.tar.gz" | tar -xzvf -
README.md
coder-doctor
```

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
