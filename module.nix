{ modulesPath, ... }: {
  imports = [
    (import "${modulesPath}/services/monitoring/prometheus/mk-downstream-exporter.nix" {
      name = "mgit";
      file = ./mgit-exporter.nix;
    })
  ];
}
