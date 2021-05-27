{
  description = "A very basic flake";

  outputs = { self, nixpkgs }:
  let
    pkgs = import nixpkgs {
      system = "x86_64-linux";
    };
    version = "1.20.1";
  in
  {

    packages.x86_64-linux.kops = pkgs.buildGoPackage {
      pname = "kops";
      inherit version;

      goPackagePath = "k8s.io/kops";

      src = self;

      nativeBuildInputs = with pkgs; [ go-bindata installShellFiles ];
      subPackages = [ "cmd/kops" ];

      buildFlagsArray = ''
        -ldflags=
            -X k8s.io/kops.Version=${version}
            -X k8s.io/kops.GitVersion=${version}
      '';

      preBuild = ''
        (cd go/src/k8s.io/kops
         go-bindata -o upup/models/bindata.go -pkg models -prefix upup/models/ upup/models/...)
      '';

      postInstall = ''
        for shell in bash zsh; do
          $out/bin/kops completion $shell > kops.$shell
          installShellCompletion kops.$shell
        done
      '';

      meta = with pkgs.lib; {
        description = "Easiest way to get a production Kubernetes up and running";
        homepage = "https://github.com/kubernetes/kops";
        changelog = "https://github.com/kubernetes/kops/tree/master/docs/releases";
        license = licenses.asl20;
        maintainers = with maintainers; [ offline zimbatm diegolelis ];
        platforms = platforms.unix;
      };
    };

    defaultPackage.x86_64-linux = self.packages.x86_64-linux.kops;

  };
}
