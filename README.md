# parse-actions-logs

This is a simple program to parse your GitHub actions logs as produced by `go test` and summarise them.

It was originally developed to parse [rclone](https://github.com/rclone/rclone)'s GitHub actions logs but it should work for any Go project which outputs the results of `go test` or `go test -v` into its logs.

It is useful if you have a large project with 1,000s of tests some of which occasionally fail and you would like to get a handle on which ones.

## Install

Parse-Actions-Logs is a Go program and comes as a single binary file.

Download the relevant binary from

- https://github.com/ncw/parse-actions-logs/releases

Or alternatively if you have Go installed use

    go get github.com/ncw/parse-actions-logs

and this will build the binary in `$GOPATH/bin`.

## Usage

Use `parse-actions-logs -h` to see all the options.

```
Usage: parse-actions-logs [options] <log.zip>+
Version: v1.0.0

Parse the logs fetched from GitHub actions. These logs should be the
zip files downloaded for a complete workflow run.

They can be downloaded by https://github.com/ncw/fetch-actions-logs or
by downloading from the web UI.

Example usage:

parse-actions-logs logs.zip  logs2.zip

Full options:
  -o string
    	Output directory (default "output")
```

The recommended way of getting logs is to run
[fetch-actions-logs](https://github.com/ncw/fetch-actions-logs) to get
the zip files of logs.

You can then run this tool against all those logs, eg

```
fetch-actions-logs -conclusion failure -user ncw rclone/rclone logs
parse-actions-logs -o output logs/*.zip
```

This will produce an output directory something like this. Using `tree` (or `rclone tree`) on the output directory shows at a glance what problems there are.

For example this shows that
- There are two occasionally failing tests in `github.com/rclone/rclone/backend/cache`: `TestInternalCachedUpdatedContentMatches` and `TestInternalCachedWrittenContentMatches`
- There is a race test failure in `github.com/rclone/rclone/cmd/serve/sftp` in `TestSftp/AuthProxy`
- There is an occasional test failure in `github.com/rclone/rclone/lib/pool` which only goes wrong on macOS `TestPool/canFail/Flusher`

```
$ tree output/
output/
├── github.com／rclone／rclone／backend／cache
│   ├── TestInternalCachedUpdatedContentMatches
│   │   ├── modules_race／10_Race test.txt-35424342.txt
│   │   └── modules_race／10_Race test.txt-38113485.txt
│   └── TestInternalCachedWrittenContentMatches
│       ├── go1.12／9_Run tests.txt-56903814.txt
│       ├── linux／9_Run tests.txt-38124982.txt
│       ├── modules_race／9_Run tests.txt-36545442.txt
│       ├── modules_race／9_Run tests.txt-65412168.txt
│       └── windows_386／9_Run tests.txt-52206992.txt
├── github.com／rclone／rclone／cmd／serve／sftp
│   └── TestSftp／AuthProxy
│       ├── modules_race／10_Race test.txt-33697335.txt
│       ├── modules_race／10_Race test.txt-36976335.txt
│       ├── modules_race／10_Race test.txt-37080974.txt
│       ├── modules_race／10_Race test.txt-42441438.txt
│       └── modules_race／10_Race test.txt-50183080.txt
└── github.com／rclone／rclone／lib／pool
    └── TestPool／canFail／Flusher
        ├── mac／10_Race test.txt-44851334.txt
        ├── mac／10_Race test.txt-56757602.txt
        ├── mac／10_Race test.txt-59667165.txt
        ├── mac／10_Race test.txt-65412168.txt
        └── mac／10_Race test.txt-86043180.txt

```

Each of the files in the output directory is a snip of the relevant log, so

```
$ cd output/github.com／rclone／rclone／lib／pool/TestPool／canFail／Flusher/
$ ls -l
total 20
-rw-rw-r-- 1 ncw ncw 2338 Apr 26 11:46 'mac／10_Race test.txt-44851334.txt'
-rw-rw-r-- 1 ncw ncw 2338 Apr 26 11:46 'mac／10_Race test.txt-56757602.txt'
-rw-rw-r-- 1 ncw ncw 2338 Apr 26 11:46 'mac／10_Race test.txt-59667165.txt'
-rw-rw-r-- 1 ncw ncw 2338 Apr 26 11:46 'mac／10_Race test.txt-65412168.txt'
-rw-rw-r-- 1 ncw ncw 2338 Apr 26 11:46 'mac／10_Race test.txt-86043180.txt'
```

And

```
$ cat mac／10_Race\ test.txt-44851334.txt 

2020/02/25 11:20:35 Failed to get memory for buffer, waiting for 1ms: failed to allocate memory
2020/02/25 11:20:35 Failed to get memory for buffer, waiting for 1ms: failed to allocate memory
2020/02/25 11:20:35 Failed to free memory: failed to free memory
2020/02/25 11:20:35 Failed to free memory: failed to free memory
--- FAIL: TestPool (0.79s)
    --- FAIL: TestPool/canFail (0.19s)
        --- FAIL: TestPool/canFail/Flusher (0.16s)
            pool_test.go:114: 
                	Error Trace:	pool_test.go:114
                	            				pool_test.go:218
                	Error:      	Not equal: 
                	            	expected: 0
                	            	actual  : 2
                	Test:       	TestPool/canFail/Flusher
FAIL
FAIL	github.com/rclone/rclone/lib/pool	0.898
```

## License

This is free software under the terms of the MIT license (check the
LICENSE file included in this package).

## Contact and support

The project website is at:

- https://github.com/ncw/parse-actions-logs

There you can file bug reports, ask for help or contribute patches.

## Authors

- Nick Craig-Wood <nick@craig-wood.com>
- Your name goes here!
