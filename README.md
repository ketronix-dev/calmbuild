calmbuild
--------

`calmbuild` is a `chroot` based package build system, used to safely and efficiently build Calmira packages from source, in a highly controlled and isolated environment.

Requirements
------------

 - golang (tested with 1.7.4)
 - `libgit2` (Also require `git` at runtime for submodules)
   - Supported versions 1.0.0 to 1.3.0 (tested with 1.3.0)
 - `curl` command
