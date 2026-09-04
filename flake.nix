{
  description = "protoc-gen-go-aip - A protoc plugin that emits Go helpers for Google AIP resource patterns and List-RPC query handling";

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
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "protoc-gen-go-aip";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-go-aip" ];
          vendorHash = "sha256-o4QexSz6AExR5rSKGuSAu3CsY5/v+cXFvxfVCmjQdf8=";
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
