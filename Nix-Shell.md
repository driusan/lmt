# Nix Shell
Nix development shells can be used to get development environments which can be accessed using the `nix-shell` command.

Firstly, we must define a function. Nix expressions are written in a functional language (also called Nix) and use the function as their basic unit. This one takes a parameter called `pkgs` with a default value of the primary Nix package collection.
```nix "Function definition"
{ pkgs ? import <nixpkgs> {} }:
```

Now, we can create a shell using the `pkgs.mkShell` function. This function creates the shell that will be accessed and defines properties of it. In our case, we use the `buildInputs` parameter to include our build dependencies (ie. Golang) in this shell.
```nix "Create shell"
pkgs.mkShell {
  buildInputs = with pkgs; [ go ];
}
```

<!--
```nix shell.nix
<<<Function definition>>>
<<<Create shell>>>
-->
