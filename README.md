# A backup tool for Udemy contents

This small tool creates backups of the courses on a given Udemy account.

Udemy is a great place to learn new skills, but the app can be unreliable when the network is slow. I created this tool to easily backup my courses as video files and take them with me while commuting.

In a nutshell, `udemy-backup` downloads videos and related assets, and store them all in a folder.

The tool itself is still work-in-progress, but compared to other solutions I used in the past it is _very fast_: than to go wonderful support of concurrency, it can retreive gigabytes of assets in a matter of seconds, which is very helpful when backuping big courses.

## Getting Started

The tool is a go binary, you can install it using `go get`:

```sh
$ go get -u github.com/ushu/udemy-backup
```

and then use the tool to perform backups:

```sh
$ udemy-backup backup
```

### Setup for development

Clone the repo and use go [`dep`](https://github.com/golang/dep) to fetch dependencies:

```sh
# clone the project into your GOPATH, then:
$ dep ensure
$ go install
```

### Credentials

To use `udemy-backup` you need to know two elements:

- your Udemy "**Client ID**"
- your Udemy "**Access Token**"

This information can be easily retreived from the cookies of your web browser, for ex. on Chrome by visiting the following link:

    chrome://settings/cookies/detail?site=www.udemy.com

then spot the cookies named `client_id` and `access_token` and remember their values.

_NOTE: a good place to store the credentials is the [config file](#global-configuration-file) !_

## Usage

The `udemy-backup` tool is self-documenting, so you can run it without parameters to obtain basic help:

```
$ udemy-backup 
A tool that create backups of Udemy courses, given API crendentials and a course URL.

Usage:
  udemy-backup [command]

Available Commands:
  backup      Backup a course
  help        Help about any command
  list        List all the subscribed courses
  login       Tries to login to the Udemy account

Flags:
      --config string   config file (default is $HOME/.udemy-backup)
  -h, --help            help for udemy-backup
      --id string       Udemy ID for the user
      --quiet           Reduce additional into
      --token string    Udemy Access Token for the user

Use "udemy-backup [command] --help" for more information about a command.
```

You can also request help about a specific command using `udemy-backup help COMMAND` or `udemy-backup COMMAND --help`.

Generally speaking, you will have to enter your credentials whenever, or you can save then in a **config file** as described below.

### `login` command

This command validates the setup by connecting the the Udemy account, and prints basic information to the terminal:

```
 $ udemy-backup login
Using config file: /Volumes/Users/XXX/.udemy-backup.yaml

🍾  SUCCESSFULLY AUTHENTICATED WITH UDEMY
🍾  User name: My Name
🍾  Udemy ID : 12345678
```

### `list` command

This command lists all the subscribed courses for the account, with their respective IDs:

```
$ udemy-backup list
Using config file: /Volumes/Users/XXX/.udemy-backup.yaml
| ID      | Title                                                        |
| 1793828 | Docker and Kubernetes: The Complete Guide                    |
| 1033356 | Modern OpenGL C++ 3D Game Tutorial Series & 3D Rendering     |
#...
```

These IDs can then be used to trigger a backup using the `backup` command.

### `backup` command

This command will trigger a backup of one or more courses.

To backup a specific course, you can either:

- provide the course ID: `udemy-backup backup 12345678`
- omit the ID and be promted with an interactive list: `udemy-backup backup`

once the course selected, the assets will be pulled.

#### Download all the courses

The `--all` flag triggers a backup for all the course associated with the account:

```sh
$ udemy-backup backup --all
```

### Other options

- `--concurrency` specifies the number of simultaneous downloads
- `--dir` specifies a base directory for the downloads
- `--resolution` specifies a preferred resolution for the videos (defaults to "the highest resolution available)
- `--restart` when specifies, `udemy-download` will skip any already-downloaded element

## Global configuration file

You can define a config file in `$HOME/.udemy-backup.yaml` to hold your preferred settings.

A typical setup would be:

```yaml
---
# The Udemy Client ID
id: "xxxxxxxxx"
# The Udemy Access Token
token: "yyyyyy"
# Store all the download in a "udemy-backups" directory (relative to CWD)
dir: "udemy-backups"
# Avoid re-downloading already-downloaded assets
restart: true
# Force concurrency to 8 simultaneous downloads
concurrency: 8
# Don't backup captions
subtitles: false
```

## Contributing

PR are welcome anytime, please consult the **TODO** section below for a basic roadmap, or feel free to add any funcionality you might feel necessary.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## TODO

* [ ] Add backup for slides
* [ ] Improve terminal output (maybe using a progress bar ?)
* [ ] Improve error handling (right now any error fails the whole process)
* [ ] Cleanup the code
