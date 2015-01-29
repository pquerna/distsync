# STATUS: WIP. ETA 2.3 days remaining. DEAL WITH IT.

# distsync

[![Build Status](https://travis-ci.org/pquerna/distsync.svg?branch=master)](https://travis-ci.org/pquerna/distsync)

`distsync` is the best damn way to distribute tarballs from your build infrastructure to production servers. Unlike projects like [syncthing](https://github.com/syncthing/syncthing), distsync is not intended for _personal_ use, instead all options and design choices are optimized for __servers__.

This means distsync is only optimized to move your application tarball or docker export from your CI, to a group of servers, using public cloud object stores as it's primary storage backend.

## Features

* __Simple__: Single command to upload from CI, and a daemon mode for servers.
* __Encrypted__: [AEAD Encryption](https://github.com/codahale/etm) of both file contents and file names.
* __Multi-Cloud__: Supports both AWS S3 and Rackspace Cloud Files as storage backends.
* __Pluggable__: Contributions Welcome: New storage, encryption, and transfer plugins are welcome.

## Usage

1. `distsync setup` Answer the prompts, it will create a `~/.distsync` and `~/.distsyncd`.
1. Copy `~/.distsync` to your uploader (eg, Jenkins).
1. Copy `~/.distsyncd` to your servers.
1. Run `distsync daemon` on servers.
1. `distsync upload foo.tar.gz` on your uploader.
1. Voilà! Your files are now on all your servers.


## What does this do?

* `distsync setup` creates two identities with limited permissions.  The first is for uploading, it allows distsync to upload to a single bucket.  The second is for downloading which gives it permissions to watch for notifications, list, and download from the bucket.
* `distsync upload` encrypts the specified file, uploads it to s3, and notifies servers it is available.
* `distsync daemon` watches for notifications, and on a new file being available will download it to the local path using  HTTPS from S3.


## Configuration File Reference

The configuration file is in [TOML](https://github.com/toml-lang/toml) syntax.  When invoked as `distsync daeomn`, `~/.distsyncd` is read by default. For all other invocations, `~/.distsync` is read by default. All commands also take a `-c path/to/conf` argument to specify the path to the configuration file.

### Example

```toml

SharedSecret = "<random-secret-here>"
StorageBucket = "distsync-503aa718-89cc-488c-ae82-0d8f6d08ed1c"
Encrypt = "AEAD_AES_128_CBC_HMAC_SHA_256"
Notify = "S3Poll"
Storage = "S3"

[Aws]
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
* CloudFilesb


#### Section: Aws

Credentials to use against AWS.  The user associated with these credentials should be setup with [AWS IAM](http://aws.amazon.com/iam/) to have limited privileges.

TODO: Document IAM policy that is created with `distsync setup`

#### Aws.Region

__Default Value__: us-east-1

__Type__: Enum String

__Details__: Region to use.  Must be one of:

* ap-northeast-1
* ap-southeast-1
* ap-southeast-2
* cn-north-1
* eu-central-1
* eu-west-1
* sa-east-1
* us-east-1
* us-gov-west-1
* us-west-1
* us-west-2


#### Aws.AccessKey

__Default Value__: None

__Type__: String

__Details__: Access Key to use with AWS.


#### Aws.SecretKey

__Default Value__: None

__Type__: String

__Details__: Secret Key to use with AWS.


#### Section: Rackspace

Credentials to use against Rackspace.  The user associated with these credentials should be setup with [RBAC](http://www.rackspace.com/knowledge_center/article/overview-role-based-access-control-rbac) to limit permissions.

By default `distsync setup` creates two users:

* `distsyncUpload${UUID}`: API Key only user with the `object-store:admin` role. For use with `distsync upload`.
* `distsyncDownload${UUID}`: API Key only user with the `object-store:observer` role. For use with `distsync daemon`.


#### Rackspace.Region

__Default Value__: None

__Type__: Enum String

__Details__: Region to use.  Must be one of:

* DFW
* HKG
* IAD
* ORD
* SYD


#### Rackspace.Username

__Default Value__: None

__Type__: String

__Details__: Username to use with Rackspace.


#### Rackspace.ApiKey

__Default Value__: None

__Type__: String

__Details__: API Key associated with the user, to use with Rackspace.


# License

`distsync` was created by [Paul Querna](http://paul.querna.org/) is licensed under the [Apache Software License 2.0](./LICENSE)

