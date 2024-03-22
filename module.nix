{ config, pkgs, lib, options, utils, modulesPath, ... }@args:

with (import "${modulesPath}/services/monitoring/prometheus/mk-exporter.nix" lib);

{
  options.services.prometheus.exporters = mkDownstreamOptions "mgit" (import ./mgit-exporter.nix args);
  config = mkDownstreamConfig "mgit" (import ./mgit-exporter.nix args) config.services.prometheus.exporters.mgit;
}
