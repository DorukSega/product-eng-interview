let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
in pkgs.mkShell rec {
  name = "interview";
  buildInputs = with pkgs; [
    go 
    sqlite
  ];
  shellHook = ''
	cd ./api
	go run .
  '';
}
