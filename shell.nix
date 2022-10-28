{ pkgs ? import <nixpkgs> {} }: with pkgs;
pkgs.mkShell {
  propagatedBuildInputs = with python3Packages; [
    go_1_19
    go-tools
    gopls
    gotools

    chromedriver
    poetry
    python3
    rnix-lsp
    selenium
    wkhtmltopdf
  ];
}
