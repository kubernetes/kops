# kOps Project Overview

kOps is a command-line tool for creating, destroying, upgrading, and maintaining production-grade, highly available Kubernetes clusters. It is written in Go and supports multiple cloud providers, including AWS, GCP, DigitalOcean, Hetzner, OpenStack, and Azure.

The project is well-structured, with a clear separation of concerns between the different packages. The `cmd` directory contains the main entry points for the `kops` CLI and other related commands. The `pkg` directory contains the core logic for managing clusters, and the `upup` directory contains the code for provisioning cloud infrastructure.

The project has a comprehensive test suite, including unit tests, integration tests, and end-to-end tests. It also has a robust CI/CD pipeline that runs these tests on every pull request.

## Building and Running

### Prerequisites

* `make`

### Building

To build the `kops` binary, run the following command:

```bash
make kops
```

This will create the `kops` binary in the `.build/dist/<os>/<arch>` directory.

To build all the binaries, including `kops`, `protokube`, `nodeup`, and `channels`, run the following command:

```bash
make all
```

### Running

To run the `kops` binary, you can either run it directly from the `dist` directory or install it to your `$GOPATH/bin` directory by running the following command:

```bash
make install
```

### Testing

To run the unit tests, run the following command:

```bash
make test
```

To run the verification scripts, run the following command:

```bash
make verify
```

To run the full suite of CI checks, run the following command:

```bash
make ci
```

## Development Conventions

### Guidelines for Programming Assistance

When assisting with programming tasks, you will adhere to the following principles:

* **Follow Requirements**: Carefully follow the user's requirements to the letter.
* **Plan First**: For any non-trivial change, first describe a detailed, step-by-step plan, including the files you intend to modify and the tests you will add or update.
* **Test Thoroughly**: Implement comprehensive tests to ensure correctness and prevent regressions.
* **Comment Intelligently**: Add comments to explain the "why" behind complex or non-obvious code, keeping in mind that the reader may not be a Kubernetes expert.
* **No TODOs**: Leave no `TODO` comments, placeholders, or incomplete implementations.
* **Prioritize Correctness**: Always prioritize security, scalability, and maintainability in your implementations.

### Avoiding Loops

When performing complex tasks, especially those involving code modifications and verification, it's important to avoid getting into loops. A loop can occur when the agent repeatedly tries the same action without success, or when it gets stuck in a cycle of analysis, action, and failure.

To avoid loops:
*   **Analyze the problem carefully**: Before taking any action, make sure you understand the problem and have a clear plan to solve it.
*   **Break down the problem**: Break down complex problems into smaller, more manageable steps.
*   **Verify each step**: After each step, verify that it was successful before moving on to the next one.
*   **Don't repeat failed actions**: If an action fails, don't just repeat it. Analyze the cause of the failure and try a different approach.
*   **Ask for help**: If you're stuck, don't hesitate to ask for help from the user.

### Code Style

The project follows the standard Go code style and the official [Kubernetes coding conventions](https://www.k8s.dev/docs/guide/coding-convention/). All code should be formatted with `gofmt` and `goimports`. You can format the code by running the following commands:

```bash
make gofmt
make goimports
```

### Linting

The project uses `golangci-lint` to lint the code. You can run the linter by running the following command:

```bash
make verify-golangci-lint
```

### Dependencies

The project uses Go modules to manage dependencies. To add a new dependency, add it to the `go.mod` file and then run the following command:

```bash
make gomod
```

### Commits

The project follows the conventional commit message format.

### Contributions

Contributions are welcome! Before submitting a pull request, please open an issue to discuss your proposed changes. All pull requests must be reviewed and approved by a maintainer before they can be merged.

