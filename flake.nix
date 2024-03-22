{
  description = "A collection of useful monitoring for Prometheus by mgIT GmbH.";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs, ... }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);

      patches = ''
        cp ${./module.nix} nixos/modules/services/monitoring/prometheus/exporters/mgit.nix
        sed 's|"apcupsd"|"apcupsd" "mgit"|g' -i nixos/modules/services/monitoring/prometheus/exporters.nix
      '';

      patchPkgs = nixpkgs: system: let
        origPkgs = import "${nixpkgs}" { inherit system; };
      in origPkgs.stdenv.mkDerivation {
        name = "patched-nixpkgs";
        src = "${nixpkgs}";
        dontBuild = true;
        dontFixup = true;
        installPhase = ''
          ${patches}
          cp -r . $out
        '';
      };
    in
    {
      overlays.default = final: prev: {
        prometheus-mgit-exporter = prev.callPackage ./. {};
      };

      patchNixpkgs = patches;

      packages = forAllSystems (system:
        let
          pkgs = (import (patchPkgs nixpkgs system) {
            inherit system;
            overlays = [ self.overlays.default ];
          });
        in
        {
          inherit (pkgs) prometheus-mgit-exporter;
          default = pkgs.prometheus-mgit-exporter;
        }
      );

      nixosModules = {
        prometheus-mgit-exporter = import ./module.nix;
      };

      checks = forAllSystems (system:
        let
          pkgs = (import (patchPkgs nixpkgs system) {
            inherit system;
            overlays = [ self.overlays.default ];
          });
        in
        {
          prometheus-mgit-exporter = pkgs.testers.runNixOSTest (import ./test.nix);
        }
      );
  };
}
