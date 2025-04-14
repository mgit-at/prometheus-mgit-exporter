{ buildGoModule
, lib
}:

buildGoModule rec {
  pname = "prometheus-mgit-exporter";
  # get version from BUILD.bazel
  version = with builtins; elemAt (match ".*version = \"([0-9.]*)\".*" (readFile ./BUILD.bazel)) 0;

  src = ./.;

  vendorHash = "sha256-q2e6Q8MbnpOguUNxTumpm3tGbBNes3ada9ma1agcTvs=";

  meta = with lib; {
    description = "A collection of useful monitoring for Prometheus by mgIT GmbH.";
    homepage = "https://github.com/mgit-at/prometheus-mgit-exporter";
    license = licenses.apsl20;
    maintainers = with maintainers; [ mkg20001 ];
    mainProgram = "prometheus-mgit-exporter";
  };
}
