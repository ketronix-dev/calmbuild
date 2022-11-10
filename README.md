solbuild
--------

`solbuild` is a `chroot` based package build system, used to safely and efficiently build Solus packages from source, in a highly controlled and isolated environment. This tool succeeds the `evobuild` tool, originally in Evolve OS, which is now known as Solus. The very core concept of the layered builder has always remained the same, this is the next .. evolution.. of the tool.

Requirements
------------

 - golang (tested with 1.7.4)
 - `libgit2` (Also require `git` at runtime for submodules)
   - Supported versions 1.0.0 to 1.3.0 (tested with 1.3.0)
 - `curl` command
