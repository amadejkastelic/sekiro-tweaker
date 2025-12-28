{
  pkgs,
  deps,
  preCommitCheck,
  extraShellHook ? "",
}:
pkgs.mkShell {
  nativeBuildInputs = deps;

  DIRENV_LOG_FORMAT = "";

  hardeningDisable = [ "fortify" ];

  shellHook = ''
    ${preCommitCheck.shellHook}
    ${extraShellHook}
  '';
}
