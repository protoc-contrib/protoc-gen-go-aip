{
  description = "protoc-gen-go-aip - A protoc plugin that emits Go helpers for Google AIP resource patterns and List-RPC query handling";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "protoc-gen-go-aip";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-go-aip" ];
          vendorHash = "sha256-+7F7rkODNogajQpiDR1oaeLXD3XY81/01X1EfD4wpj8=";
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
