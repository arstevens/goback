# Goback
A backup and program written in go that will automatically backup up a folder
whenever the backup drive has been mounted

## Installation
Use the [make](https://www.gnu.org/software/make/) command to install the gobackd
daemon as well as the goback command line utility

```bash
sudo make build
```
If a different build destination is desired simply edit the makefile it is very short
and manageable

## Usage
The gobackd daemon is setup using systemd so to start, stop or look at debug messages
just use systemctl

```bash
systemctl start gobackd
systemctl enable gobackd
systemctl status gobackd
```

For creating and removing new backups use the goback tool

```bash
goback -o="directory/to/backup" -c="location/to/backup"
```

## License
[MIT](https://choosealicense.com/licenses/mit/)
