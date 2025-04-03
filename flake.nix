{
  description = "tracktime";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            pkg-config
            poetry
            python3
            chromedriver
            wkhtmltopdf
            python3Packages.selenium

            pre-commit
          ];

          # Run this command, only after creating the virtual environment
          postVenvCreation = ''
            unset SOURCE_DATE_EPOCH
          '';
        };
      }));
}
