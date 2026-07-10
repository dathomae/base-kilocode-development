# Shell Scripts

This directory contains utility shell scripts for the Chronocone Planning Language (CPL) toolchain.

## `compile_and_run.sh`

Compiles CPL source files into a Go program and immediately runs the generated executable. This script orchestrates the full compilation pipeline: parse → analyze → synthesize → execute.

### Prerequisites

- Go 1.21+ installed and available on `PATH`
- The current working directory must be within the `chronocone_planning_language` Go module (the script resolves its own location and changes to the module root automatically)

### Example Directory Structure

The script operates on subdirectories under `examples/with_data/`. Each example directory must contain:

| File          | Description                                              |
| ------------- | -------------------------------------------------------- |
| `iospec.json` | IO specification describing connectors and data sources  |
| `*.cpl`       | One or more CPL source files to compile                  |
| `*.parquet`   | One or more Parquet data files referenced by the IO spec |

Example layout:

```
examples/with_data/
└── random_data_no_time/
    ├── iospec.json
    ├── test.cpl
    └── testdata.parquet
```

### Usage

```bash
./shell_scripts/compile_and_run.sh <directory_name>
```

Where `<directory_name>` is the name of a subdirectory under `examples/with_data/`.

### Example

```bash
./shell_scripts/compile_and_run.sh random_data_no_time
```

This will:

1. Validate that `examples/with_data/random_data_no_time/` exists
2. Verify `iospec.json`, at least one `.cpl` file, and at least one `.parquet` file are present
3. Run the CPL compiler (`src/main/go/cli/compiler/cpl_compiler.go`) with the IO spec and all `.cpl` files
4. Write the generated Go program to `generated/cmd/random_data_no_time/main.go`
5. Execute the generated program with `go run`

### Output

Generated Go code is written to:

```
generated/cmd/<directory_name>/main.go
```

This directory is created automatically if it does not exist.

### Error Handling

The script uses `set -e` and performs explicit validation at each step:

- Missing or incorrect argument count
- Missing example directory
- Missing `iospec.json`
- Missing `.cpl` files
- Missing `.parquet` files
- CPL compilation failures

If any step fails, the script prints a descriptive error message and exits with a non-zero status code.

### Notes

- The script discovers all `.cpl` files in the target directory automatically; you do not need to list them individually.
- The script also discovers all `.parquet` files but only verifies that at least one exists; it does not enumerate them.
- Filenames containing spaces are handled safely via `find -print0`.
