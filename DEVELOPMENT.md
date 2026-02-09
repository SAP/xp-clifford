# Development Guide

## Development Environment

The `xp-clifford` repository provides a reproducible development environment powered by [Nix](https://nixos.org/). This environment includes the Go compiler, linters, and all tools needed to contribute to the project.

### Prerequisites

1. Install the [Nix package manager](https://nixos.org/download/)
2. Enable [flake support](https://nixos.wiki/wiki/Flakes)

### Getting Started

After cloning the repository, enter the development environment:

```sh
nix develop --no-pure-eval
```

> **Note:** The initial setup may take several minutes as dependencies are downloaded and configured.

Upon entering the environment, [pre-commit](https://pre-commit.com/) hooks are automatically installed. These hooks run various checks before each Git commit to ensure code quality.

### Direnv Integration (Optional)

For a seamless experience, you can use [direnv](https://direnv.net/) to automatically activate the development environment when entering the project directory.

1. [Install direnv](https://direnv.net/docs/installation.html)
2. Navigate to the cloned repository
3. Allow the direnv configuration:

```sh
direnv allow
```

After this setup, the Nix development environment activates automatically whenever you enter the `xp-clifford` directory.
