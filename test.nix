{ pkgs, lib, ... }:
{
  name = "prometheus-mgit-exporter";

  nodes = {
    server = { lib, pkgs, ... }: {
      services.prometheus.exporters.mgit.enable = true;
    };
  };

  testScript = ''
    start_all()
    server.wait_for_unit("prometheus-mgit-exporter.service")
  '';
}
