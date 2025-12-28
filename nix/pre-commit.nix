{
  pkgs,
  preCommitHooks,
  shellHook,
}:

let
  deps =
    (pkgs.buildGoModule {
      pname = "sekiro-tweaker-modules";
      version = "dev";
      src = ../.;
      proxyVendor = true;
      vendorHash = "sha256-aCB7rxA3bV5tXwE4eOtRitQh50ZAOnplCMVQvZsqoTY=";
      buildInputs = with pkgs; [
        gtk4
        glib
        gobject-introspection
      ];
    }).goModules;

  goWithProxy = pkgs.writeShellScriptBin "go" ''
    export GOPROXY="file://${deps}"
    export GOSUMDB="off"
    exec ${pkgs.go}/bin/go "$@"
  '';

  cgoSetupHook = pkgs.makeSetupHook {
    name = "cgo-setup-hook";
  } (pkgs.writeScript "cgo-setup-hook.sh" shellHook);
in
preCommitHooks.run {
  src = ../.;
  hooks = {
    nixfmt-rfc-style.enable = true;
    golangci-lint = {
      enable = true;
      extraPackages = [
        goWithProxy
        cgoSetupHook
        pkgs.pkg-config
        pkgs.gtk4
        pkgs.glib
        pkgs.gobject-introspection
      ];
    };
    golines = {
      enable = true;
      settings.flags = "-m 100 --dry-run";
    };
    gotest = {
      enable = true;
      extraPackages = [
        goWithProxy
        cgoSetupHook
      ];
    };
  };
}
