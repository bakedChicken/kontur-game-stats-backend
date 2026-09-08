{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";

    devenv = {
      url = "github:cachix/devenv";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ nixpkgs, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
      ];

      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem = { pkgs, ... }: {
        formatter = pkgs.nixfmt-tree;

        devenv.shells.default = {
          starship.config.enable = true;
          languages.nix.enable = true;

          packages = with pkgs; [
            go-swagger
          ];
        };

        devenv.shells.original = {
          languages.dotnet.enable = true;

          enterShell = ''
            echo "Welcome to the .NET shell!"
          '';
        };

        devenv.shells.revisited = {
          languages.go.enable = true;
          languages.go.package = pkgs.go_1_27;

          enterShell = ''
            echo "Welcome to the Go shell!"
          '';
        };
      };
    };
}
