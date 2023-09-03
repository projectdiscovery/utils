# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<a name="unreleased"></a>
## [Unreleased]

### Keep
- keep a changelog ([#251](https://github.com/projectdiscovery/utils/issues/251))


<a name="v0.0.53"></a>
## [v0.0.53] - 2023-08-29
### Fix
- fix query and fragment issue

### Resolve
- resolve merge conflicts

### Revert
- revert change & fix unit test


<a name="v0.0.52"></a>
## [v0.0.52] - 2023-08-25
### Add
- add HasPrefixAnyI
- add test

### Allow
- allow w/ unsafe

### Decode
- decode unicode chars


<a name="v0.0.51"></a>
## [v0.0.51] - 2023-08-23
### Adding
- adding mutex

### Create
- Create dep-auto-merge.yml

### Excluding
- excluding read-only folders on windows

### Fix
- fix race in synclockMap during Del()
- fix isWritable func
- fix data race warning at connpool

### Improve
- improve home folder detection

### Make
- make it public

### Merge
- Merge branch 'main' into issue-234-data-race
- Merge branch 'main' into issue-234-data-race

### Update
- Update dependabot.yml

### Use
- use os agnostic env variables


<a name="v0.0.50"></a>
## [v0.0.50] - 2023-08-18
### Add
- add default values for TLS_VERIFY and DEBUG
- add UpdateWithEnv func

### Introduce
- introduce generic lockable type

### Move
- move to env

### Support
- support multiple vars

### Using
- using expand prefix


<a name="v0.0.49"></a>
## [v0.0.49] - 2023-08-10
### Fixing
- Fixing conditional build ([#238](https://github.com/projectdiscovery/utils/issues/238))


<a name="v0.0.48"></a>
## [v0.0.48] - 2023-08-09
### Add
- add default permissions

### Adding
- adding value or default context helper

### Fixing
- fixing type

### Merge
- Merge branch 'main' into feat-ctx-or-default


<a name="v0.0.47"></a>
## [v0.0.47] - 2023-08-04

<a name="v0.0.46"></a>
## [v0.0.46] - 2023-07-31
### Add
- add SyncLockMap ctor func
- add FileExistsIn func

### Fix
- fix lint
- fix lint err

### Rejecting
- rejecting allowed relative paths

### Small
- small refactor

### Using
- using variadic options


<a name="v0.0.45"></a>
## [v0.0.45] - 2023-07-28
### Add
- add tests
- add version bump script

### Adding
- adding big number test

### Consider
- consider url ^// as relative


<a name="v0.0.44"></a>
## [v0.0.44] - 2023-07-17
### Adding
- adding arm6
- adding unix 386 fallback

### Using
- using arm


<a name="v0.0.43"></a>
## [v0.0.43] - 2023-07-14
### Add
- add logger to join & split workers
- add channel utils

### Bumping
- bumping go version

### Excluding
- excluding linux with g01.20 with ip extensions

### Fixing
- fixing bugs

### Merge
- Merge branch 'main' into feat-onetimepool
- Merge branch 'main' into feat-onetimepool
- Merge branch 'main' into feat-onetimepool

### Minor
- minor lint stuff

### Moving
- Moving jarm helper to utils

### Release
- release dialer references

### Update
- Update README.md


<a name="v0.0.42"></a>
## [v0.0.42] - 2023-07-12
### Add
- add metrics and model_selection funcs
- add mlutils

### Fixing
- fixing paths


<a name="v0.0.41"></a>
## [v0.0.41] - 2023-07-08
### Add
- add OpenOrCreateFile func
- add guard and remove source dir option
- add async/await

### Fix
- fix MigrateDir func

### Fixing
- fixing func + tests

### Improving
- improving equality check

### Introduce
- introduce MustMigrateDir func

### Removing
- removing dev file


<a name="v0.0.40"></a>
## [v0.0.40] - 2023-06-30
### Add
- add iteration support to orderedparams
- add len method to orderedMap
- add helper for ordered parameters

### Fix
- fix orderMap set method
- fix param order in unit tests

### Fixing
- fixing lint

### More
- more changes

### Orderedparams
- orderedparams deep copy

### Support
- support days unit while parsing time string

### Update
- update and getall methods


<a name="v0.0.39"></a>
## [v0.0.39] - 2023-06-22
### Adding
- adding test with interval

### Changing
- changing endpoint

### Trim
- trim character


<a name="v0.0.38"></a>
## [v0.0.38] - 2023-06-14
### Add
- Add update file permission functio
- Add tests and comment about umask
- Add unix file permissions

### Add
- add common regex patterns

### Adding
- Adding mutex on read ([#176](https://github.com/projectdiscovery/utils/issues/176))

### Adding
- adding note

### Go
- go naming

### Merge
- Merge branch 'main' into pr/163

### Restoring
- restoring cond compilation

### Run
- Run tests only on linux and darwin

### Small
- small refactor

### Update
- Update create file to create temp file to ingore error on windows


<a name="v0.0.37"></a>
## [v0.0.37] - 2023-06-06
### Accept
- accept toolname as param

### Add
- add doc for UserConfigDirOrDefault
- add MigrateDir func
- add UserAppConfigDirOrDefault func
- add UserConfigDirOrDefault func

### Commit
- commit to last commit
- commit to last commit

### Fix
- fix lint err

### Minimize
- minimize return

### Minor
- minor improvements

### Use
- use os


<a name="v0.0.36"></a>
## [v0.0.36] - 2023-05-31
### Add
- add test for syncLockMap
- add Clone to generic map

### Fix
- fix lint error

### Skip
- skip windows race tests

### Use
- use stdlib maps.clone


<a name="v0.0.35"></a>
## [v0.0.35] - 2023-05-30
### Add
- Add file size converter

### Add
- add darwin case to test
- add platform agnostic syscall loadlibrary func

### Use
- use osutils


<a name="v0.0.34"></a>
## [v0.0.34] - 2023-05-28
### Add
- add GetSortedKeys function

### Adding
- adding batcher mechanism

### Changing
- changing condition

### Fixing
- fixing random visit test
- fixing condition


<a name="v0.0.33"></a>
## [v0.0.33] - 2023-05-25
### Add
- add healtcheck api
- add dns resolver
- add connection check hc util
- add PATH env var
- add env healthcheck functionality util
- add exec name to default path list
- add file permission check

### Code
- code refactor

### Fix
- fix ulimit err

### Fixing
- fixing conn error

### Rename
- rename filenames
- rename file to path to make it more generic

### Update
- update test host


<a name="v0.0.32"></a>
## [v0.0.32] - 2023-05-16
### Visit
- Visit helpers + SizeOf ([#152](https://github.com/projectdiscovery/utils/issues/152))


<a name="v0.0.31"></a>
## [v0.0.31] - 2023-05-14
### Add
- add test cases

### Add
- Add ability to set config from env vars

### Fix
- fix ranging over channel

### Rename
- rename func


<a name="v0.0.30"></a>
## [v0.0.30] - 2023-05-09

<a name="v0.0.29"></a>
## [v0.0.29] - 2023-05-08
### Fixing
- Fixing resp body reuse ([#147](https://github.com/projectdiscovery/utils/issues/147))


<a name="v0.0.28"></a>
## [v0.0.28] - 2023-05-06
### Add
- add gitignore

### Adding
- adding tests
- adding safe dereferencing helper


<a name="v0.0.27"></a>
## [v0.0.27] - 2023-05-05
### Check
- Check for Administrator on Windows

### Cleaning
- cleaning up + tests

### Fix
- fix removing semicolon while decoding params ([#145](https://github.com/projectdiscovery/utils/issues/145))

### Fixing
- fixing error


<a name="v0.0.26"></a>
## [v0.0.26] - 2023-04-29

<a name="v0.0.25"></a>
## [v0.0.25] - 2023-04-23
### Add
- Add body in updater error msg

### Fix
- fix panic on err and version empty


<a name="v0.0.24"></a>
## [v0.0.24] - 2023-04-20
### Fix
- fix empty http resp with proxy ([#138](https://github.com/projectdiscovery/utils/issues/138))


<a name="v0.0.23"></a>
## [v0.0.23] - 2023-04-19
### Adding
- Adding process utils


<a name="v0.0.22"></a>
## [v0.0.22] - 2023-04-19
### Add
- add version param & deprecate function
- add path to callback func
- add support for custom org

### Adding
- adding IndexAny

### Fix
- fix lint + tests

### Remove
- remove unused argument from versioncheck

### Render
- render theme update

### Replace
- replace status codes with http.xx variables

### Update
- update if tool installed from dev


<a name="v0.0.21"></a>
## [v0.0.21] - 2023-04-18
### Bug
- bug fix + adds more testcases

### Fix
- fix domain parse error


<a name="v0.0.20"></a>
## [v0.0.20] - 2023-04-11
### Adding
- adding int helper

### Adds
- adds proxy utils
- adds proxy utils

### Fix
- fix send on closed channel

### Fixing
- fixing depbot branch ([#118](https://github.com/projectdiscovery/utils/issues/118))

### Move
- move burp proxy check to proxyutils


<a name="v0.0.19"></a>
## [v0.0.19] - 2023-03-29
### Encoding
- encoding semicolon on windows


<a name="v0.0.18"></a>
## [v0.0.18] - 2023-03-23
### Added
- added also parsing

### Adding
- Adding nil check on url Params
- Adding slice clone helper

### Bug
- bug fix

### Code
- code refactoring

### Fix
- fix on ip return on for loop

### Guard
- guard against concurrent reset ([#109](https://github.com/projectdiscovery/utils/issues/109))

### Ipv4
- ipv4 logic

### Logic
- logic change

### Removing
- removing err details

### Short
- short ip logic

### Unit
- unit test
- unit tests


<a name="v0.0.17"></a>
## [v0.0.17] - 2023-03-18
### Add
- add posix support

### Added
- added unit tests

### Adding
- adding linux + generic interface
- adding console example
- adding keypress reader (win)
- adding isprintable/isctrlc + tests

### Bug
- bug fix upgrade not using timeout

### Chagne
- chagne download asset logic

### ContainsAnyI
- ContainsAnyI added

### Creating
- creating darwin placeholder

### Enabling
- enabling proxy + use defaults

### Finalizing
- finalizing multiplatform keypress

### Fix
- fix resolve conflicts
- fix asset name formation logic

### Fixing
- fixing windows console support

### Increasing
- increasing test coverage

### Merge
- Merge branch 'main' into update-utils-bug-fix
- Merge branch 'main' into feat-console-keypress

### Moving
- moving win syscall
- moving errors to common source file

### Read
- read tagName from method

### Remove
- remove redundant code

### Replacing
- replacing redundant logic with posix constants


<a name="v0.0.16"></a>
## [v0.0.16] - 2023-03-14
### Adds
- adds release notes,version check and more

### Bump
- Bump github.com/ulikunitz/xz from 0.5.7 to 0.5.8 ([#102](https://github.com/projectdiscovery/utils/issues/102))

### Exit
- exit if already updated

### Only
- only skip cert validation in version check

### Selfupdate
- selfupdate callback utils

### Updater
- updater utils


<a name="v0.0.15"></a>
## [v0.0.15] - 2023-03-12
### Removing
- Removing capabilities from linux armv7l


<a name="v0.0.14"></a>
## [v0.0.14] - 2023-03-01
### Adding
- adding key lookup with value helper

### Merge
- Merge permission_darwin.go in permission_other.go

### Remove
- Remove impossible unix buildconstraint


<a name="v0.0.13"></a>
## [v0.0.13] - 2023-02-26
### Added
- added unit tests
- added unit tests
- added generic functions

### Adding
- adding new line

### Check
- check is all have zero items

### Implemented
- implemented fistNonZero

### Moved
- moved all to generic

### Started
- started unit test

### Unit
- unit test


<a name="v0.0.12"></a>
## [v0.0.12] - 2023-02-26
### Bump
- Bump golang.org/x/crypto from 0.0.0-20210921155107-089bfa567519 to 0.1.0 ([#93](https://github.com/projectdiscovery/utils/issues/93))


<a name="v0.0.11"></a>
## [v0.0.11] - 2023-02-24
### Bump
- Bump golang.org/x/net from 0.1.0 to 0.7.0

### Fix
- fix invalid url

### Rename
- rename test variable

### Update
- update url README.md


<a name="v0.0.10"></a>
## [v0.0.10] - 2023-02-12
### Add
- Add truncate test

### Extending
- extending test coverage

### Fix
- Fix issue [#67](https://github.com/projectdiscovery/utils/issues/67)

### Fixing
- fixing typo


<a name="v0.0.9"></a>
## [v0.0.9] - 2023-02-09
### Add
- add url decode test

### Minor
- minor changes

### Use
- use standard url enc format


<a name="v0.0.8"></a>
## [v0.0.8] - 2023-02-09
### Adding
- adding note for goling
- adding synclock tests
- adding sync/lock capabilities to generic map
- adding syncmap prototype
- adding permission to FreeBSD
- adding reverse ptr

### Adding
- Adding proxy utils

### Bug
- bug fixes and improvements

### Bugfix
- bugfix + tests

### Bumping
- bumping go version in GH actions
- bumping go to 1.19

### Fix
- fix missing slash
- fix lint error

### Fixing
- fixing lint errors

### Map
- map with generics and native helpers

### Merge
- Merge branch 'main' into issue-62-rev-ptr

### Minor
- minor improvement & adds documentation

### Small
- small refactor

### Struct
- struct private field get/set via reflect


<a name="v0.0.7"></a>
## [v0.0.7] - 2023-02-03
### Added
- added count with multiple files
- added also count line with separator

### Adding
- adding new err with fmt type

### Bufio
- bufio reader optimizations

### Check
- check on separator

### Code
- code refactor
- code refactoring on CountLinesWithSeparator and CountLines
- code refactor on CountLineLogic()
- code optiomizations

### Comments
- comments + fix

### Finalizing
- finalizing implementation

### Fix
- fix on action error

### Fixed
- fixed problem with windows checks

### Fixes
- fixes and optimizations

### Implemented
- implemented unit tests
- implemented countline feature

### Integration
- integration test files

### Lint
- lint error

### Moving
- moving walk to utils

### Refactoring
- refactoring on Error field

### Removed
- removed old version code

### Removing
- removing third party api test

### Tests
- tests refactoring


<a name="v0.0.6"></a>
## [v0.0.6] - 2023-01-23
### Add
- add context utils

### Addind
- addind os/arch utils

### Allow
- allow localhost as valid hostname

### Fix
- fix lint error
- fix nil pointer dereference in userinfo

### Lint
- lint bypass

### Minor
- minor bug fixes

### New
- new url.URL wrapper

### Remove
- remove omithost (only available in 1.19)

### Update
- update params description

### Use
- use require for unit tests


<a name="v0.0.5"></a>
## [v0.0.5] - 2023-01-20
### Adding
- adding support for nil errors ([#52](https://github.com/projectdiscovery/utils/issues/52))


<a name="v0.0.4"></a>
## [v0.0.4] - 2023-01-19
### Add
- add parameter parsing tests
- add getparams function

### Adding
- adding helpers + tests
- adding extra slice helpers
- adding slice equality helper
- adding port helpers
- adding file helpers
- adding reusable reader

### Adding
- Adding longest sequence helper

### Adds
- adds url encoders
- adds package boolean
- adds errorutils
- adds release checks (closes [#19](https://github.com/projectdiscovery/utils/issues/19))

### Better
- better implementation on isInternal logic
- better login on IsInternal function

### Code
- code formatting

### Created
- created IsInternal functions for ipv4 and ipv6

### Errorutil
- errorutil enriched

### Fix
- fix lint error
- fix + error check

### Fixes
- fixes on package references

### Fixing
- fixing autorelease
- fixing tests
- fixing time.sleep go routine leaks via context with timeout

### Go
- go mod tidy

### Lint
- lint fix and tests

### Moved
- moved naabu routing in utils

### Reader
- reader utils + reusablereadcloser

### Remove
- remove lenreader & add unit tests
- remove extra whitespaces

### Removed
- removed gologger
- removed debug instructions
- removed unused struct

### Removing
- removing redundant check

### Return
- return err

### Unit
- unit tests on IsInternal function

### Updated
- updated sleepTime in TestDeleteFilesOlderThan func
- updated sleep time to 10seconds

### Used
- used log.Fatal, dont create problems with mac action on build
- used panic instead of log.Fatal


<a name="v0.0.3"></a>
## [v0.0.3] - 2022-11-11
### Code
- code refactoring

### Fix
- Fix typo

### Fixing
- fixing linux errors
- fixing match condition

### Merge
- Merge branch 'main' into move-packages-from-naabu

### Misc
- misc updates

### Move
- move packages from naabu

### Renaming
- renaming race to raceutil

### Test
- test cases addition

### Update
- Update mapsutil [#13](https://github.com/projectdiscovery/utils/issues/13) ([#14](https://github.com/projectdiscovery/utils/issues/14))


<a name="v0.0.1"></a>
## [v0.0.1] - 2022-11-03

<a name="v0.0.2"></a>
## v0.0.2 - 2022-11-03
### Added
- added missing label and test action
- added github workflow
- added logutils

### Added
- Added weekly tag + release automation ([#9](https://github.com/projectdiscovery/utils/issues/9))

### CodeQL
- CodeQL Analysis on push

### Fixing
- fixing windows test

### Merge
- Merge branch 'main' into issue-213/logutils

### Misc
- misc changes

### Removed
- removed push from codeql workflow

### Removing
- removing empty line

### Wo
- wo global functions for log util and test cases


[Unreleased]: https://github.com/projectdiscovery/utils/compare/v0.0.53...HEAD
[v0.0.53]: https://github.com/projectdiscovery/utils/compare/v0.0.52...v0.0.53
[v0.0.52]: https://github.com/projectdiscovery/utils/compare/v0.0.51...v0.0.52
[v0.0.51]: https://github.com/projectdiscovery/utils/compare/v0.0.50...v0.0.51
[v0.0.50]: https://github.com/projectdiscovery/utils/compare/v0.0.49...v0.0.50
[v0.0.49]: https://github.com/projectdiscovery/utils/compare/v0.0.48...v0.0.49
[v0.0.48]: https://github.com/projectdiscovery/utils/compare/v0.0.47...v0.0.48
[v0.0.47]: https://github.com/projectdiscovery/utils/compare/v0.0.46...v0.0.47
[v0.0.46]: https://github.com/projectdiscovery/utils/compare/v0.0.45...v0.0.46
[v0.0.45]: https://github.com/projectdiscovery/utils/compare/v0.0.44...v0.0.45
[v0.0.44]: https://github.com/projectdiscovery/utils/compare/v0.0.43...v0.0.44
[v0.0.43]: https://github.com/projectdiscovery/utils/compare/v0.0.42...v0.0.43
[v0.0.42]: https://github.com/projectdiscovery/utils/compare/v0.0.41...v0.0.42
[v0.0.41]: https://github.com/projectdiscovery/utils/compare/v0.0.40...v0.0.41
[v0.0.40]: https://github.com/projectdiscovery/utils/compare/v0.0.39...v0.0.40
[v0.0.39]: https://github.com/projectdiscovery/utils/compare/v0.0.38...v0.0.39
[v0.0.38]: https://github.com/projectdiscovery/utils/compare/v0.0.37...v0.0.38
[v0.0.37]: https://github.com/projectdiscovery/utils/compare/v0.0.36...v0.0.37
[v0.0.36]: https://github.com/projectdiscovery/utils/compare/v0.0.35...v0.0.36
[v0.0.35]: https://github.com/projectdiscovery/utils/compare/v0.0.34...v0.0.35
[v0.0.34]: https://github.com/projectdiscovery/utils/compare/v0.0.33...v0.0.34
[v0.0.33]: https://github.com/projectdiscovery/utils/compare/v0.0.32...v0.0.33
[v0.0.32]: https://github.com/projectdiscovery/utils/compare/v0.0.31...v0.0.32
[v0.0.31]: https://github.com/projectdiscovery/utils/compare/v0.0.30...v0.0.31
[v0.0.30]: https://github.com/projectdiscovery/utils/compare/v0.0.29...v0.0.30
[v0.0.29]: https://github.com/projectdiscovery/utils/compare/v0.0.28...v0.0.29
[v0.0.28]: https://github.com/projectdiscovery/utils/compare/v0.0.27...v0.0.28
[v0.0.27]: https://github.com/projectdiscovery/utils/compare/v0.0.26...v0.0.27
[v0.0.26]: https://github.com/projectdiscovery/utils/compare/v0.0.25...v0.0.26
[v0.0.25]: https://github.com/projectdiscovery/utils/compare/v0.0.24...v0.0.25
[v0.0.24]: https://github.com/projectdiscovery/utils/compare/v0.0.23...v0.0.24
[v0.0.23]: https://github.com/projectdiscovery/utils/compare/v0.0.22...v0.0.23
[v0.0.22]: https://github.com/projectdiscovery/utils/compare/v0.0.21...v0.0.22
[v0.0.21]: https://github.com/projectdiscovery/utils/compare/v0.0.20...v0.0.21
[v0.0.20]: https://github.com/projectdiscovery/utils/compare/v0.0.19...v0.0.20
[v0.0.19]: https://github.com/projectdiscovery/utils/compare/v0.0.18...v0.0.19
[v0.0.18]: https://github.com/projectdiscovery/utils/compare/v0.0.17...v0.0.18
[v0.0.17]: https://github.com/projectdiscovery/utils/compare/v0.0.16...v0.0.17
[v0.0.16]: https://github.com/projectdiscovery/utils/compare/v0.0.15...v0.0.16
[v0.0.15]: https://github.com/projectdiscovery/utils/compare/v0.0.14...v0.0.15
[v0.0.14]: https://github.com/projectdiscovery/utils/compare/v0.0.13...v0.0.14
[v0.0.13]: https://github.com/projectdiscovery/utils/compare/v0.0.12...v0.0.13
[v0.0.12]: https://github.com/projectdiscovery/utils/compare/v0.0.11...v0.0.12
[v0.0.11]: https://github.com/projectdiscovery/utils/compare/v0.0.10...v0.0.11
[v0.0.10]: https://github.com/projectdiscovery/utils/compare/v0.0.9...v0.0.10
[v0.0.9]: https://github.com/projectdiscovery/utils/compare/v0.0.8...v0.0.9
[v0.0.8]: https://github.com/projectdiscovery/utils/compare/v0.0.7...v0.0.8
[v0.0.7]: https://github.com/projectdiscovery/utils/compare/v0.0.6...v0.0.7
[v0.0.6]: https://github.com/projectdiscovery/utils/compare/v0.0.5...v0.0.6
[v0.0.5]: https://github.com/projectdiscovery/utils/compare/v0.0.4...v0.0.5
[v0.0.4]: https://github.com/projectdiscovery/utils/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/projectdiscovery/utils/compare/v0.0.1...v0.0.3
[v0.0.1]: https://github.com/projectdiscovery/utils/compare/v0.0.2...v0.0.1
