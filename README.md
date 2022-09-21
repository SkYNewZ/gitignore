# gitignore

Simple cli to create a `.gitignore` file from https://github.com/github/gitignore.

## Installation

If you have [Go](https://golang.org/) 1.18 (or greater) installed:

```shell
go install github.com/SkYNewZ/gitignore@latest
```

(and make sure your `$GOBIN` is in your `$PATH`.)

## Usage

```
Usage of gitignore [language] [flags]:
  -directory string
        (optional) Output file path. (default ".")
  -filename string
        (optional) Output file name. (default ".gitignore")
  -list
        List available languages on github/gitignore@main.
  -token string
        (optional) GitHub token to use in case of rate-limited.
  -version
        Show version.

```

## Example

````shell
gitignore go
file successfully writed to <current directory>/.gitignore
# Make a .gitignore file with content from https://github.com/github/gitignore/blob/main/Go.gitignore
````