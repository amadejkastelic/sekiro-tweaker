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

  vendorHash = "sha256-zrWxoPGbFdPaWsjZMFhSkaea0ONAI8636GN+IRI11ZI=";

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
