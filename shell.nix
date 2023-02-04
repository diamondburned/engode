{ pkgs ? import <nixpkgs> {} }:

let python-minifier = pkgs.python3Packages.buildPythonPackage (rec {
		pname = "python-minifier";
		version = "2.8.0";
		src = pkgs.python3Packages.fetchPypi {
			inherit version;
			pname = "python_minifier";
			sha256 = "1dvg77asri6smklqdssgba8fqhhzidmy4ginv6wv6pgwcfpfil07";
		};
		propagatedBuildInputs = with pkgs.python3Packages; [
			setuptools-scm
		];
	});

	cminify = 
		let src = pkgs.fetchFromGitHub {
			owner  = "Scylardor";
			repo   = "cminify";
			rev    = "bdde9e6";
			sha256 = "0hn6lralnf064jal7mla9l04rsfd6m64m24a3ivgw9lfqzdq9brg";
		}; in
		pkgs.writeShellScriptBin "cminify" ''
			${pkgs.python27}/bin/python ${src}/minifier.py "$@"
		'';
in

pkgs.mkShell {
	buildInputs = with pkgs; [
		python3
		go
		python-minifier
		cminify
	];
}
