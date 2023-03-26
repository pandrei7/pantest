# PAntest

A tool to run multiple programs on a set of tests. It can help when creating
programming assignments.

You can run multiple programs on a set of tests.

![Example output of the `run` command](docs/img/example-run.png)

You can also check if two programs agree on random tests.

![Example output of the `same` command](docs/img/example-same.png)

## Dependencies

You should have [Go](https://go.dev/dl/) and the
[timeout](https://www.gnu.org/software/coreutils/manual/html_node/timeout-invocation.html)
command installed.

## Getting the binary

To install the `pantest` binary simply run this command:

```bash
go install github.com/pandrei7/pantest@latest
```

## Getting the code

If you want to work on the code, the repository is hosted
[on Github](https://github.com/pandrei7/pantest.git):

```bash
# Clone the repo.
git clone https://github.com/pandrei7/pantest

# Download the Go dependencies.
cd pantest
go mod tidy
```

## Usage

First create a configuration file using `pantest init` and
[customize](#configuration) it to your needs.

If you have a set of tests you want to evaluate, run `pantest run`.

If you want to check whether two programs are "equivalent", run `pantest same`.

Use the `--help` flag after any command whenever you need more details.

## Configuration

### Running fixed tests (`run` command)

The input and reference files should have `.in` and `.ref` extensions,
respectively. Ideally, they should also contain a number in the name.

### Running random tests (`same` command)

For similarity checking, you should provide a command to generate a random test
case (written to standard out) and a directory where these tests will be saved.
Tests which do not generate a mismatch will be removed automatically.

### Executables

Make sure all sources **use standard IO to read and write**, not file IO.

For each executable, you can configure:

- `cmd`: a command to run the executable
- `name` (optional): a name to be displayed
- `time` (optional, defaults to `1`): the time limit for one run, in seconds
