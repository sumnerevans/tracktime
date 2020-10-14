{ pkgs ? import <nixpkgs> {} }: with pkgs;
pkgs.mkShell {
  propagatedBuildInputs = with python38Packages; [
    chromedriver
    poetry
    python38
    rnix-lsp
    selenium
    wkhtmltopdf
  ];
}
