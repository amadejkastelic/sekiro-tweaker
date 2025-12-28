{
  pkgs,
  version,
  enableParallelBuilding ? true,
  ...
}:
pkgs.buildGoModule {
  pname = "sekiro-tweaker";
  version = version;

  src = ./..;

  enableParallelBuilding = enableParallelBuilding;

  vendorHash = "sha256-w1Aa7x/spXU50EOrxJ5IOrW9Dj1R1Rh2DR+/8Jsw4+E=";

  nativeBuildInputs = with pkgs; [
    pkg-config
    wrapGAppsHook4
  ];

  buildInputs = with pkgs; [
    gtk4
    glib
    gobject-introspection
  ];

  subPackages = [ "cmd/sekiro-tweaker" ];
}
