{ pkgs ? import <nixpkgs> { } }:

pkgs.buildGoModule {
  pname = "ograph-gen";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-lrneUCbx4YmUeYzoHEiNKH7W/0RtpyA+rAoWMmGVDwk=";

  env.CGO_ENABLED = 0;

  nativeBuildInputs = [ pkgs.makeWrapper ];

  postInstall = ''
    wrapProgram $out/bin/ograph-gen \
      --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.librsvg ]}
  '';

  meta = {
    description = "Open Graph image generator for fossunited.org";
    homepage = "https://github.com/fossunited/ograph-gen";
    license = pkgs.lib.licenses.agpl3Only;
    mainProgram = "ograph-gen";
  };
}
