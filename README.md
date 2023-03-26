# pantest

A tool to run multiple programs on a set of tests.

## Dependencies

You should have [Go](https://go.dev/dl/) and the [timeout](http://www.gnu.org/software/coreutils/manual/html_node/timeout-invocation.html#timeout-invocation) command installed.

## Installation

```bash
# Clone the repo.
git clone https://github.com/pandrei7/pantest

# Download the Go dependencies.
cd pantest
go mod tidy
```

## Usage

First create a configuration file using `go run . init` and
[customize](#configuration) it to your needs.

Then execute `go run . run` to run the tests.

To check that two programs have the same output on random tests run
`go run . same --help`.

## Configuration

The input and reference files should have `.in` and `.ref` extensions,
respectively. Ideally, they should also contain a number in the name.

For similarity checking, you should provide a command to generate a random test
case (written to standard out) and a directory where these tests will be saved.

Make sure all sources **use standard IO to read and write**, not file IO.

For each executable, you can configure:

- `cmd`: a command to run the executable
- `name` (optional): a name to be displayed
- `time` (optional, defaults to `1`): the time limit for one run, in seconds
