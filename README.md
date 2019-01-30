# A backup tool for Udemy contents

In a nutshell, `udemy-backup` downloads videos and related assets, and store them all in a folder for offline consumption.

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

## Usage

Enter your Udemy credentials and follow the instructions:

```
$ udemy-backup 
Email: xxx@gmail.com
Password: •••••••
```

#### Download all the courses

The `-a` flag triggers a backup for all the course associated with the account:

```sh
$ udemy-backup -a
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
