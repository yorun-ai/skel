# skelc Quickstart

This example contains a small public actor, data type, and service contract. It exercises validation and the Go module, TypeScript module, and public Skel generators.

From the repository root, run:

```sh
./examples/quickstart/generate.sh
```

Generated files are written to `examples/quickstart/generated` by default and are ignored by Git. Pass another directory as the first argument to keep output elsewhere:

```sh
./examples/quickstart/generate.sh /tmp/skelc-quickstart
```

The generated Go module targets `skel.DefaultGolangVineVersion`. Compiling it requires that version of `go.yorun.ai/vine` to be published. The generated TypeScript package likewise expects the declared `@yorun-ai/vrpc` dependency to be available from the configured npm registry.
