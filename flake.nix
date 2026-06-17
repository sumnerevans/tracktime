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
            default = tracktime;
            tracktime = pkgs.buildGoModule {
              pname = "tt";
              version = "unstable";
              src = self;
              subPackages = [ "cmd/tt" ];
              vendorHash = "sha256-2ohu2lWPK0ckpdT8dfjnsrgrn9ZuvpzrrieAOoiIBik=";
              propagatedBuildInputs = [ pkgs.typst ];
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
