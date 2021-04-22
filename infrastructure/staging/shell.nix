with import <nixpkgs> {};
let
  my-python-packages = python-packages: [
    python-packages.pip
    python-packages.numpy
  ];
  my-python = python36.withPackages my-python-packages;
in
  pkgs.mkShell {
    buildInputs = [
      bashInteractive
      my-python
      terraform_0_14
      terraform-ls
      ansible
    ];
    shellHook = ''
      export PIP_PREFIX="$(pwd)/.pyenv/pip_packages"
      export PYTHONPATH="$(pwd)/.pyenv/pip_packages/lib/python3.6/site-packages:$PYTHONPATH" 
      unset SOURCE_DATE_EPOCH
    '';
  }
