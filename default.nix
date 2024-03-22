{ buildGoModule
, lib
}:

buildGoModule rec {
  pname = "prometheus-mgit-exporter";
  # get version from BUILD.bazel
  version = with builtins; elemAt (match ".*version = \"([0-9.]*)\".*" (readFile ./BUILD.bazel)) 0;

  src = ./.;

  vendorHash = "sha256-UQk+mj7ZnE6zKL8VXMzbiGvpcLjKFXB9O/VQkUAJ7TA=";

  meta = with lib; {
    description = "A collection of useful monitoring for Prometheus by mgIT GmbH.";
    homepage = "https://github.com/mgit-at/prometheus-mgit-exporter";
    license = licenses.apsl20;
    maintainers = with maintainers; [ mkg20001 ];
  };
}
