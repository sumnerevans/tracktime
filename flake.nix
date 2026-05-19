{
  description = "tracktime";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
    }:
    (flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" ];
      perSystem =
        {
          pkgs,
          system,
          ...
        }:
        {
          _module.args.pkgs = import inputs.nixpkgs { inherit system; };

          packages = rec {
            default = tt;
            tt = pkgs.buildGoModule {
              pname = "tt";
              version = "1.0.0";
              src = self;
              subPackages = [ "cmd/tt" ];
              vendorHash = "sha256-w+/aj8r7Bi+buexCVSqgD7xcUwK/giQ/bwDXk73NpZY=";
            };
          };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              go-tools
              gotools
              pre-commit
              typst
            ];
          };
        };
    });
}
