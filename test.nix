{ pkgs, lib, ... }:
{
  name = "prometheus-mgit-exporter";

  nodes = {
    server = { lib, pkgs, ... }: {
      imports = [ ./module.nix ];
      services.prometheus.exporters.mgit.enable = true;
      environment.systemPackages = with pkgs; [ wget ];
    };
  };

  testScript = ''
    start_all()
    server.wait_for_unit("prometheus-mgit-exporter.service")
    server.succeed("wget localhost:9328/metrics")
  '';
}
