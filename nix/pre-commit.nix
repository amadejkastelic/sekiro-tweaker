{
  pkgs,
  preCommitHooks,
}:

let
  deps =
    (pkgs.buildGoModule {
      pname = "sekiro-tweaker-modules";
      version = "0.0.1";
      src = ../.;
      proxyVendor = true;
      vendorHash = pkgs.lib.fakeHash;
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
in
preCommitHooks.run {
  src = ../.;
  hooks = {
    nixfmt-rfc-style.enable = true;
    gofmt.enable = true;
    golangci-lint = {
      enable = true;
      extraPackages = [ goWithProxy ];
    };
    golines = {
      enable = true;
      settings.flags = "-m 100 --dry-run";
    };
    gotest = {
      enable = true;
      extraPackages = [ goWithProxy ];
    };
    govet = {
      enable = true;
      extraPackages = [ goWithProxy ];
    };
  };
}
