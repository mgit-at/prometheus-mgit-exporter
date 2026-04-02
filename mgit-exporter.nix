# nixpkgs/nixos/modules/services/prometheus/exporters/mgit.nix
{ config, lib, pkgs, options, utils, ... }:

with lib;

let
  # for convenience we define cfg here
  cfg = config.services.prometheus.exporters.mgit;
  json = pkgs.formats.json {};
  configurationFile = json.generate "config.json" cfg;
in
{
  port = 9328; # The mgit exporter listens on this port by default

  # `extraOpts` is an attribute set which contains additional options
  # (and optional overrides for default options).
  # Note that this attribute is optional.
  extraOpts = {
    listen = mkOption {
      default = ":${toString cfg.port}";
      type = types.str;
      description = "Listening address";
    };

    certFile = {
      enable = mkEnableOption "certificate file watching";
      globs = mkOption {
        default = [];
        type = types.listOf types.str;
        description = "List of globs";
      };
      exclude_system = mkOption {
        type = types.bool;
        description = "Exclude system certificates";
        default = false;
      };
    };

    mceLog = {
      enable = mkEnableOption "MCE log";
      path = mkOption {
        default = "";
        type = types.str;
        description = "Path to MCE log";
      };
    };

    ptHeartbeat = {
      enable = mkEnableOption "PT heartbeat";
      database = mkOption {
        default = "";
        type = types.str;
        description = "Database for PT heartbeat";
      };
      table = mkOption {
        default = "";
        type = types.str;
        description = "Table for PT heartbeat";
      };
      defaultsFile = mkOption {
        default = "";
        type = types.str;
        description = "Defaults file for PT heartbeat";
      };
      masterId = mkOption {
        default = 0;
        type = types.int;
        description = "Master ID for PT heartbeat";
      };
    };

    fsTab = {
      enable = mkEnableOption "fsTab";
    };

    binLog = {
      enable = mkEnableOption "binLog";
      path = mkOption {
        default = "";
        type = types.str;
        description = "Path to binLog";
      };
    };

    rasDaemon = {
      enable = mkEnableOption "rasDaemon";
      path = mkOption {
        default = "";
        type = types.str;
        description = "Path to rasDaemon";
      };
    };

    elk = {
      enable = mkEnableOption "elk";
      duration = mkOption {
        default = "";
        type = types.str;
        description = "Duration for elk";
      };
      node = mkOption {
        default = "default_node";
        type = types.str;
        description = "Node for elk";
      };
    };

    exec = {
      enable = mkEnableOption "exec";
      scripts = mkOption {
        default = {};
        type = types.attrsOf (types.submodule ({ ... }: {
          options = {
            command = mkOption {
              type = types.listOf types.str;
              description = "Command";
            };

            dir = mkOption {
              type = types.str;
              description = "Working directory";
              default = "";
            };

            timeout = mkOption {
              type = types.str;
              description = "Timeout";
              default = "";
            };
          };
        }));
        description = "Scripts for exec";
      };
    };
  };

  # `serviceOpts` is an attribute set which contains configuration
  # for the exporter's systemd service. One of
  # `serviceOpts.script` and `serviceOpts.serviceConfig.ExecStart`
  # has to be specified here. This will be merged with the default
  # service configuration.
  # Note that by default 'DynamicUser' is 'true'.
  serviceOpts = {
    unitConfig = {
      Description = "mgIT exporter for Prometheus";
      Documentation = "https://prometheus.io/docs/introduction/overview/";
    };
    serviceConfig = {
      MemoryMax = "1G";
      Restart = "on-failure";
      RestartSec = 1;
      DynamicUser = false;
      ExecStart = utils.escapeSystemdExecArgs [
        (getExe pkgs.prometheus-mgit-exporter)
        "-config" (toString configurationFile)
      ];
    };
  };
}
