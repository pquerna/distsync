# STATUS: WIP. ETA 3 days. DEAL WITH IT.

# distsync: sync files from object stores to your servers.

`distsync` is the best damn way to distribute tarballs from your CI to many servers. Unlike projects like [syncthing](https://github.com/syncthing/syncthing) is not intended for _personal_ use, all options are optimized for server use.

This means distsync is only optimized to move your application tarball or docker export from your CI, to a group of servers, using public cloud object stores as it's primary storage backend.  It can optionally use BitTorrent to accelerate delivery and reduce S3 costs.

## Features

* __Simple__: Single command to upload from CI, easy daemon mode for servers.
* __Encrypted__: [AEAD Encryption](https://github.com/codahale/etm) of both file contents and file names.
* __Multi-Cloud__: Supports both AWS S3 and Rackspace Cloud Files as storage backends.
* __BitTorrent__: AWS S3 can optionally use BitTorrent to increase speed and reduce transfer costs.
* __Pluggable__: Contributions Welcome: New storage, encryption, and transfer plugins are welcome.

## Usage

1. distsync setup
	Answer propmts
	Creates ~/.distsync
	Creates ~/.distsyncd
1. Copy ~/.distsync to your uploader (eg, Jenkins).
1. Copy ~/.distsyncd to your servers.
1. Run `distsync daemon` on servers.
1. `distsync upload foo.tar.gz` on your uploader.
1. Voilà! Your files are now on all your servers.


## What does this do?

* `distsync setup` creates two identities with limited permissions.  The first is for uploading, it allows distsync to upload to a single bucket.  The second is for downloading which gives it permissions to watch for notifications, list, and download from the bucket.
* `distsync upload` encrypts the specified file, uploads it to s3, and notifies servers it is available.
* `distsync daemon` watches for notifications, and on a new file being available will download it to the local path using BitTorrent and HTTPS from S3.


## Configuration File Reference

The configuration file is in [TOML](https://github.com/toml-lang/toml) syntax.

### Example

```toml

SharedSecret = "<random-secret-here>"
StorageBucket = "distsync-503aa718-89cc-488c-ae82-0d8f6d08ed1c"
Encrypt = "AEAD_AES_128_CBC_HMAC_SHA_256"
Notify = "S3Poll"
Storage = "S3"

[AwsCreds]
  Region = "us-east-1"
  AccessKey = "<access-key here>"
  SecretKey = "<secret-key here>"
```
