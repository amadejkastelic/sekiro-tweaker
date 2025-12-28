{
  description = "Sekiro Tweaker - Linux game patcher with GTK4 UI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    nix-github-actions = {
      url = "github:nix-community/nix-github-actions";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      pre-commit-hooks,
      nix-github-actions,
      ...
    }:
    let
      version = "0.0.1";

      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forAllSystems =
        f:
        builtins.listToAttrs (
          map (system: {
            name = system;
            value = f system;
          }) supportedSystems
        );

      perSystem =
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };

          buildDeps = [
            pkgs.git
            pkgs.go
            pkgs.gtk4
            pkgs.pkg-config
            pkgs.gobject-introspection
            pkgs.wrapGAppsHook4
            pkgs.glib
            pkgs.cairo
            pkgs.pango
            pkgs.gdk-pixbuf
          ];

          devDeps = buildDeps ++ [
            pkgs.golangci-lint
            pkgs.delve
            pkgs.gopls
            pkgs.golines
          ];

          shellHook = ''
            export CGO_ENABLED=1
            export PKG_CONFIG_PATH="${pkgs.gtk4.dev}/lib/pkgconfig:${pkgs.glib.dev}/lib/pkgconfig"
          '';

          preCommitCheck = import ./nix/pre-commit.nix {
            inherit pkgs shellHook;
            preCommitHooks = pre-commit-hooks.lib.${system};
          };

          package = import ./nix/package.nix {
            inherit pkgs version;
          };
        in
        {
          packages.default = package;
          devShells.default = import ./nix/dev-shell.nix {
            inherit pkgs;
            deps = devDeps;
            preCommitCheck = preCommitCheck;
            extraShellHook = shellHook;
          };
          checks.pre-commit-check = preCommitCheck;
        };

    in
    {
      packages = forAllSystems (system: (perSystem system).packages);
      devShells = forAllSystems (system: (perSystem system).devShells);
      checks = forAllSystems (system: (perSystem system).checks);

      githubActions = nix-github-actions.lib.mkGithubMatrix {
        checks = {
          x86_64-linux = self.packages.x86_64-linux;
          aarch64-linux = self.packages.aarch64-linux;
        };
      };
    };
}
