![GitHub License](https://img.shields.io/github/license/antfie/data-tools)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/antfie/data-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/antfie/data-tools)](https://goreportcard.com/report/github.com/antfie/data-tools)
![Docker Image Size](https://img.shields.io/docker/image-size/antfie/data-tools/latest)
![Docker Pulls](https://img.shields.io/docker/pulls/antfie/data-tools)

# Data Tools

Disclaimer: I will not be responsible for any data loss caused by this tool.

If you have a ton of copies of files, backups of backups, and are drowning in duplicate data, then consider using Data Tools to 'ZAP' it all. This program will ultimately place a single instance of every file you have in a single location and all the references are stored in a SQLite database. You can see an example of this below:

![zap.png](docs/images/zap.png)

# Configuration

# Running Order

1. `crawl /some/path`
1. `hash`
1. `zap`
1. `unzap`

# ZAP-ing

When you ZAP your files, every unique file is placed in a folder and all duplicate copies are removed.

Note that empty folders will not be created when un-ZAP-ping, should you desire to re-inflate your disk drive.

It is really, really important that you run crawl AND hash on the same OS. This is due to different filesystems and implementations of the 'file' command which can lead to issues.

## How Can I Support This?

We welcome fixes, features and donations.

# Synology

We need to set CGO_ENABLED=0  and use "github.com/glebarez/sqlite" instead of "gorm.io/driver/sqlite" which is a little less performant.