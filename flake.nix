{
  description = "Description for the project";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    devenv.url = "github:cachix/devenv";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    go-overlay = {
      url = "github:purpleclay/go-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nix2container = {
      url = "github:nlewo/nix2container";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    mk-shell-bin.url = "github:rrbutani/nix-mk-shell-bin";
    git-hooks-nix.url = "github:cachix/git-hooks.nix";
    github-actions-nix.url = "github:synapdeck/github-actions-nix";
    nix-github-actions = {
      url = "github:nix-community/nix-github-actions";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.devshell.flakeModule
        inputs.devenv.flakeModule
        inputs.git-hooks-nix.flakeModule
        inputs.github-actions-nix.flakeModule
        # To import an internal flake module: ./other.nix
        # To import an external flake module:
        #   1. Add foo to inputs
        #   2. Add foo as a parameter to the outputs function
        #   3. Add here: foo.flakeModule
      ];
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];
      perSystem = {
        config,
        # self',
        # inputs',
        pkgs,
        # system,
        ...
      }: {
        devenv.shells = let
          go-config = {
            enable = true;
            lsp.enable = false;
            version = "1.24.10";
          };
        in {
          default = {
            # ++ config.pre-commit.settings.enabledPackages;
            # enterShell = config.pre-commit.installationScript;
            packages = [
              pkgs.pre-commit
              pkgs.golangci-lint
              pkgs.markdownlint-cli
            ];
            languages.go = go-config;
            git-hooks = {
              hooks = {
                action-validator.enable = true;
                actionlint.enable = true;
                alejandra.enable = true;
                check-added-large-files.enable = true;
                check-merge-conflicts.enable = true;
                commitizen.enable = true;
                deadnix.enable = true;
                detect-private-keys.enable = true;
                end-of-file-fixer.enable = true;
                flake-checker.enable = true;
                gofmt.enable = true;
                golangci-lint.enable = true;
                gotest.enable = true;
                govet.enable = true;
                markdownlint = {
                  enable = true;
                  settings.configuration = {
                    MD010 = {
                      code_blocks = false;
                    };
                    MD013 = {
                      line_length = 256;
                    };
                    MD033 = {
                      allowed_elements = [
                        "a"
                        "sup"
                      ];
                    };
                  };
                };
                reuse.enable = true;
                ripsecrets.enable = true;
                zizmor.enable = true;
              };
            };
            env = {
              GOTOOLCHAIN = pkgs.lib.mkForce "go${config.devenv.shells.default.languages.go.version}";
            };
            # treefmt = {
            #   enable = true;
            #   config.programs.golangci-lint = {
            #     enable = true;
            #     enableLinters = [
            #       "errcheck"
            #       "ginkgolinter"
            #       "govet"
            #       "ineffassign"
            #       "staticcheck"
            #       "unused"
            #     ];
            #     includes = [""];
            #     excludes = [".devenv/*"];
            #     verbose = true;
            #   };
            # };
            # files = {
            #   "markdownlint.toml" = {
            #     toml = {
            #       MD010 = {
            #         code_blocks = false;
            #       };
            #       MD013 = {
            #         line_length = 256;
            #       };
            #       MD033 = {
            #         allowed_elements = [
            #           "a"
            #           "sup"
            #         ];
            #       };
            #     };
            #   };
            # };
            scripts = {
              go-lint = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  golangci-lint run
                '';
              };
              go-format = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  golangci-lint fmt
                '';
              };
              go-test = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  go test -v ./...
                '';
              };
              gha-lint = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  ghalint run
                '';
                packages = [pkgs.ghalint];
              };
              nix-check = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  nix flake check --impure
                '';
              };
              nix-format = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  alejandra flake.nix
                '';
                packages = [pkgs.alejandra];
              };
              md-lint = {
                exec = ''
                  cd -P -- "$(git rev-parse --show-toplevel)"
                  markdownlint README.md --config markdownlint.toml

                '';
                packages = [pkgs.mado];
              };
              # actions = {
              #   exec = ''
              #     ls ${config.githubActions.workflowsDir}
              #   '';
              # };
              check-all = {
                exec = ''
                  set -euxo pipefail
                  go-format
                  go-lint
                  # go-test
                  # gha-lint
                  # nix-check
                  # nix-format
                  # md-lint
                '';
              };
            };
          };
        };
        # githubActions = {
        #   enable = true;
        #   workflows = {
        #     ci = {
        #       name = "CI";
        #       on = ["push" "pull_request"];
        #       jobs = {
        #         check = {
        #           runsOn = "ubuntu-latest";
        #           steps = [
        #             {
        #               uses = "actions/checkout@6";
        #             }
        #           ];
        #         };
        #       };
        #     };
        #   };
        # };
        # pre-commit = {
        #   check.enable = false;

        #   settings = {
        #     hooks.my = {
        #       enable = true;
        #       entry = "check-all";
        #     };
        #     # hooks.alejandra.enable = true;
        #   };
        # };

        # Per-system attributes can be defined here. The self' and inputs'
        # module parameters provide easy access to attributes of the same
        # system.

        # Equivalent to  inputs'.nixpkgs.legacyPackages.hello;
        # packages.default = pkgs.hello;
      };
      flake = {
        githubActions = inputs.nix-github-actions.lib.mkGithubMatrix {
          checks = inputs.nixpkgs.lib.getAttrs ["x86_64-linux" "x86_64-darwin"] inputs.self.checks;
        };
        # The usual flake attributes can be defined here, including system-
        # agnostic ones like nixosModule and system-enumerating ones, although
        # those are more easily expressed in perSystem.
      };
    };
}
