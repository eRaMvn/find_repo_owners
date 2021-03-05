# `find_repo_owner`

find_repo_owner is a CLI tool to gather all repo owners based on CODEOWNERS files in github.com.

The tool is written in Golang and can make 20 requests concurrently.

## Pre-requisites

You will need:

1. Linux Environment (Windows Environment is under development)
2. Github Access key (which can be found under [setting](https://github.com/settings/tokens))

## Installing

`git clone https://github.com/eRaMvn/find_repo_owners.git`

Build executable

```bash
#!/bin/bash
go build
```

Or you can grab one of the executables under `Releases`

## Example Commands

Set up environment variable

```bash
#!/bin/bash
export GITHUB_TOKEN="qwdsfdsfsdf"
```

1. Gather all owners in the CODEOWNERS file

`
find_repo_owner -o eRaMvn
`

The command will generates a csv file with the name `results_from_repos.csv`

To change the name of the output file:

`
find_repo_owner -o eRaMvn --of some_other_name
`

2. If you know exactly what you want to look for and just want to know what repo has that owners

`
find_repo_owner -o eRaMvn -f owners_to_watch.txt
`

The format of `owners_to_watch.txt` will be like this

```txt
@user1
@group1
@group2
```
