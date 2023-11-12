let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
in pkgs.mkShell rec {
  name = "interview";
  buildInputs = with pkgs; [
    gcc
    sqlite
    go 
  ];
  shellHook = ''
	cd ./api
	go run .
  '';
}
