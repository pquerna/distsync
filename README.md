# STATUS: WIP. ETA 3 days. DEAL WITH IT.

# distsync

`distsync` is the best damn way to distribute tarballs from your build infrastructure to production servers. Unlike projects like [syncthing](https://github.com/syncthing/syncthing) is not intended for _personal_ use, instead all options and design choices are optimized for server use.

This means distsync is only optimized to move your application tarball or docker export from your CI, to a group of servers, using public cloud object stores as it's primary storage backend.  It can optionally use BitTorrent to accelerate delivery and reduce S3 costs.

## Features

* __Simple__: Single command to upload from CI, easy daemon mode for servers.
* __Encrypted__: [AEAD Encryption](https://github.com/codahale/etm) of both file contents and file names.
* __Multi-Cloud__: Supports both AWS S3 and ~~Rackspace Cloud Files~~ as storage backends.
* __BitTorrent__: AWS S3 can optionally use BitTorrent to increase speed and reduce transfer costs.
* __Pluggable__: Contributions Welcome: New storage, encryption, and transfer plugins are welcome.


_Rackspace Cloud Support_: Waiting on v0.2 rewrite of the [Gophercloud SDK](https://github.com/rackspace/gophercloud) to add support for Rackspace Identity and Rackspace Cloud Files as backends.

## Usage

1. `distsync setup` Answer the propmts, it will create a `~/.distsync` and `~/.distsyncd`.
1. Copy `~/.distsync` to your uploader (eg, Jenkins).
1. Copy `~/.distsyncd` to your servers.
1. Run `distsync daemon` on servers.
1. `distsync upload foo.tar.gz` on your uploader.
1. Voilà! Your files are now on all your servers.


## What does this do?

* `distsync setup` creates two identities with limited permissions.  The first is for uploading, it allows distsync to upload to a single bucket.  The second is for downloading which gives it permissions to watch for notifications, list, and download from the bucket.
* `distsync upload` encrypts the specified file, uploads it to s3, and notifies servers it is available.
* `distsync daemon` watches for notifications, and on a new file being available will download it to the local path using BitTorrent and HTTPS from S3.


## Configuration File Reference

The configuration file is in [TOML](https://github.com/toml-lang/toml) syntax.  When invoked as `distsync daeomn`, `~/.distsyncd` is read by default. For all other invocations, `~/.distsync` is read by default. All commands also take a `-c path/to/conf` argument to specify the path to the configuration file.

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

### Reference


#### SharedSecret

__Default Value__: None

__Type__: String

__Details__: A base64 encoded shared secret used to encrypt and HMAC all objects.  Generally created by `distsync setup`.



#### StorageBucket

__Default Value__: None

__Type__: String

__Details__: Name of the bucket to use in the storage backend.


#### Encrypt

__Default Value__: AEAD_AES_128_CBC_HMAC_SHA_256

__Type__: Enum String

__Details__: Type of encryption and HMAC to use on objects. Must be one of:

* AEAD_AES_128_CBC_HMAC_SHA_256


#### Notify

__Default Value__: S3Poll

__Type__: Enum String

__Details__: Method to detect new files are available. Must be one of:

* S3Poll


#### Storage

__Default Value__: S3

__Type__: Enum String

__Details__: Storage backend used to upload and download files. Must be one of:

* S3
* S3+BitTorrent


#### Section: AwsCreds

Credentials to use against AWS.  The user associated with these credentials should be setup with [AWS IAM](http://aws.amazon.com/iam/) to have limited privileges.

TODO: Document IAM policy that is created with `distsync setup`

#### AwsCreds.Region

__Default Value__: us-east-1

__Type__: Enum String

__Details__: Region to use.  Must be one of:

* ap-northeast-1
* ap-southeast-1
* ap-southeast-2
* cn-north-1
* eu-west-1
* sa-east-1
* us-east-1
* us-gov-west-1
* us-west-1
* us-west-2


#### AwsCreds.AccessKey

__Default Value__: None

__Type__: String

__Details__: Access Key to use with AWS.


#### AwsCreds.SecretKey

__Default Value__: None

__Type__: String

__Details__: Secret Key to use with AWS.


# License

`distsync` was created by [Paul Querna](http://paul.querna.org/) is licensed under the [Apache Software License 2.0](./LICENSE)

