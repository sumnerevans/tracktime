{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  propagatedBuildInputs = with pkgs; [
    poetry
    python38
    wkhtmltopdf
  ];
}
