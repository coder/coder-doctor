# Coder doctor

This repository contains an initial implementation of a command-line
diagnostic tool and library for checking that a given system can run
the Coder control plane and workspaces.

The initial implementation performs a series of preflight checks, as
described in [Preflight cluster health check tool] RFC.

[Preflight cluster health check tool]: https://www.notion.so/coderhq/Preflight-cluster-health-check-tool-07024636e1c741e2a30e482bb796ea0a

This may (or may not) be the final home of this project. Please
provide feedback regarding the plan on the RFC, or comments on the
implementation as comments in the relevant pull requests.

This project is in `alpha` state and Coder offers no compatibility
guarantees, either for the tool itself or any public Go APIs. Most
code is kept in the `internal` package to make this clear, and will
be promoted to an externally-importable package once we enter `beta`
state.
