{
  description = "protoc-gen-go-aip - A protoc plugin that emits Go helpers for Google AIP resource patterns and List-RPC query handling";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nix-release-bin = {
      url = "github:nixos-contrib/nix-release-bin";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      nix-release-bin,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        release = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";
        # release-please advances the manifest on the release commit itself, so
        # every commit after it still reads as the previous release. Append the
        # revision the plugin was actually built from, so a build off main is
        # not mistaken for the release it trails.
        version = "${release}+${self.shortRev or self.dirtyShortRev or "dirty"}";

        source = pkgs.buildGoModule {
          pname = "protoc-gen-go-aip";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-go-aip" ];
          vendorHash = "sha256-vgdp+XOdEvVyy+zdJLJFfJI5mKzv9CZ1KHjyL6irbGo=";
          ldflags = [
            "-s"
            "-w"
          ];
          meta = with pkgs.lib; {
            description = "A protoc plugin that emits Go helpers for Google AIP resource patterns and List-RPC query handling";
            license = licenses.mit;
            mainProgram = "protoc-gen-go-aip";
          };
        };
      in
      {
        packages = {
          # The latest release binary, where it has one for the system: CI pins
          # them in nix/release.json once the release has published them.
          default = nix-release-bin.lib.mkReleaseBinary {
            inherit pkgs;
            lock = ./nix/release.json;
            pname = "protoc-gen-go-aip";
            fallback = source;
          };
          inherit source;
        };

        devShells.default = pkgs.mkShell {
          name = "protoc-gen-go-aip";
          packages = [
            pkgs.go
            pkgs.protobuf
            pkgs.buf
          ];
        };
      }
    );
}
