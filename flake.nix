{
  description = "xp-clifford: Crossplane CLI Framework for Resource Data";

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
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem = {
        config,
        pkgs,
        ...
      }: {
        packages.example = pkgs.buildGoModule {
          pname = "tutorial";
          version = "0.0.1";
          src = ./.;
          subPackages = ["tutorial"];
          vendorHash = "sha256-4VAmGMPfMPt+BcDiFHZ68AlMmrOEg6pruxxw7pMo9GQ=";
        };
        devenv.shells = let
          env = {
            GOTOOLCHAIN = pkgs.lib.mkForce "go${config.devenv.shells.default.languages.go.version}";
          };
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
          go-config = {
            enable = true;
            lsp.enable = false;
            version = "1.25.8";
          };
        in {
          ci = {
            inherit env git-hooks;
            packages = with pkgs; [
              pre-commit
            ];
            languages.go = go-config;
          };
          default = {
            inherit env git-hooks;
            packages = with pkgs; [
              pre-commit
              golangci-lint
              markdownlint-cli
              vhs
            ];
            languages.go = go-config;
          };
        };
      };
      flake = {
        githubActions = inputs.nix-github-actions.lib.mkGithubMatrix {
          checks = inputs.nixpkgs.lib.getAttrs ["x86_64-linux"] inputs.self.checks;
        };
      };
    };
}
