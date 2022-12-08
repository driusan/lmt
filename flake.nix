{
  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs;
    flake-utils.url = github:numtide/flake-utils;
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      packages = self.packages.${system};
    in
    {
      defaultPackage = packages.lmt;
      packages.lmt = pkgs.buildGoPackage {
        name = "lmt";
        src = ./.;
        goPackagePath = "main";
        postInstall = /*bash*/ ''
          mv $out/bin/main $out/bin/lmt
        '';
        meta = with pkgs.lib; {
          description = " literate markdown tangle";
          homepage = "https://github.com/driusan/lmt";
          license = licenses.mit;
        };
      };
    });
}
