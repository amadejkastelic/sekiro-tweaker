{
  description = "Sekiro Tweaker - Linux game patcher with GTK4 UI";

  nixConfig = {
    extra-substituters = [ "https://amadejkastelic.cachix.org" ];
    extra-trusted-public-keys = [
      "amadejkastelic.cachix.org-1:EiQfTbiT0UKsynF4q3nbNYjNH6/l7zuhrNkQTuXmyOs="
    ];
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    pre-commit-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      nixpkgs,
      pre-commit-hooks,
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
    };
}
