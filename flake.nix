{
  description = "A collection of useful monitoring for Prometheus by mgIT GmbH.";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  inputs.patches4nixpkgs.url = "github:mgit-at/patches4nixpkgs/master";

  outputs = { self, patches4nixpkgs, ... }@inputs:
    let
      patchPkgs = patches4nixpkgs.patch inputs.nixpkgs [ self ];
      nixpkgs = patches4nixpkgs.eval patchPkgs;

      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
    in
    {
      overlays.default = final: prev: {
        prometheus-mgit-exporter = prev.callPackage ./. {};
      };

      patches4nixpkgs = nixpkgs: [
        [
          (! builtins.pathExists "${nixpkgs}/nixos/modules/services/monitoring/prometheus/mk-downstream-exporter.nix")
          ./nixos-prom-downstream.patch
        ]
      ];

      packages = forAllSystems (system:
        let
          pkgs = (import nixpkgs {
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
          pkgs = (import nixpkgs {
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
