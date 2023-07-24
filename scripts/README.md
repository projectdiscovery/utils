# scripts
The package contains various scripts


## versionbump
This Go script can automatically bump the semantic version number defined in a Go source file. It parses the specified Go source file with `go/ast`, finds the given variable (which is assumed to contain a semantic version string), increments the specified part of the version number (major, minor, or patch) with `github.com/Masterminds/semver/v3`, and rewrites the file with the updated version.

```
go run versionbump.go -file /path/to/your/file.go -var YourVersionVariable
```

By default, the patch version is incremented. To increment the major or minor versions instead, specify -part major or -part minor respectively:

```
go run versionbump.go -file /path/to/your/file.go -var YourVersionVariable -part minor
```
