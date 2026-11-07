{
  description = "protoc-gen-go-aip - A protoc plugin that emits Go helpers for Google AIP resource patterns and List-RPC query handling";

  nixConfig = {
    extra-substituters = [ "https://ralch.cachix.org" ];
    extra-trusted-public-keys = [ "ralch.cachix.org-1:ti5RSqCJwd+VIv7dlxJEAVD98KIJDLxj1yJY+S80Q3A=" ];
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
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

        # Generated code imports the standard library's `uuid` package (Go 1.27),
        # and so do this module's own fixtures.
        go = pkgs.go_1_27;

        package = (pkgs.buildGoModule.override { inherit go; }) {
          pname = "protoc-gen-go-aip";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-go-aip" ];
          vendorHash = "sha256-Q1uixcJFD05AspyWWcjWiu2HnlYTUg04tmJe80a20p4=";
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
        packages.default = package;

        devShells.default = pkgs.mkShell {
          name = "protoc-gen-go-aip";
          packages = [
            go
            pkgs.protobuf
            pkgs.buf
          ];
        };
      }
    );
}
