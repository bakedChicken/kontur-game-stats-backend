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

        packages.gamestats = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
          pname = "gamestats";
          version = "0.1.0";
          src = ./gamestats;
          env.CGO_ENABLED = 1;

          vendorHash = "sha256-IUyZTlMqht5sx010M3kc96okTBsRrL1e+6m1VVWvK0k=";
        };

        packages.modernized-server = pkgs.buildDotnetModule {
          pname = "kontur-gamestats-server";
          version = "0.1.0";
          src = ./Kontur.GameStats.Modern;
          projectFile = "Server/Server.csproj";
          nugetDeps = ./Kontur.GameStats.Modern/deps.json;
          dotnet-sdk = pkgs.dotnet-sdk_10;
          dotnet-runtime = pkgs.dotnet-aspnetcore_10;
          meta.mainProgram = "Server";
        };

        packages.modernized-data-generator = pkgs.buildDotnetModule {
          pname = "kontur-gamestats-data-generator";
          version = "0.1.0";
          src = ./Kontur.GameStats.Modern;
          projectFile = "DataGenerator/DataGenerator.csproj";
          nugetDeps = ./Kontur.GameStats.Modern/deps.json;
          dotnet-sdk = pkgs.dotnet-sdk_10;
          dotnet-runtime = pkgs.dotnet-aspnetcore_10;
          executables = [ "DataGenerator" ];
          meta.mainProgram = "DataGenerator";
        };

        devenv.shells.default = {
          enterShell = ''
            echo "Welcome to the default dev shell!"
          '';

          languages.nix.enable = true;

          packages = with pkgs; [
            go-swagger
          ];

          processes = {
            openapi.exec = "swagger serve openapi.yaml";
          };
        };

        devenv.shells.modernized = {
          enterShell = ''
            echo "Welcome to the .NET shell!"
            dotnet --version
          '';

          languages.dotnet = {
            enable = true;
            package = pkgs.dotnet-sdk_10;
          };
        };

        devenv.shells.gamestats = {
          enterShell = ''
            echo "Welcome to the Go shell!"
            go version
            sqlite3 --version
          '';

          languages.go.enable = true;
          languages.go.package = pkgs.go_1_27;

          packages = with pkgs; [
            sqlite
          ];
        };
      };
    };
}
