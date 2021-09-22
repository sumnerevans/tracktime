{ pkgs ? import <nixpkgs> {} }: with pkgs;
pkgs.mkShell {
  propagatedBuildInputs = with python3Packages; [
    chromedriver
    poetry
    python3
    rnix-lsp
    selenium
    wkhtmltopdf
  ];
}
