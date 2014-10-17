# STATUS: WIP.

# distsync: sync files from object stores to your servers.

`distsync` makes synchorizing files from an object store like S3 to your servers very easy.

The primary use case is distributing things like your application tarball or docker export to a group of servers.  It can optionally use BitTorrent to accelerate delivery and reduce S3 costs.

## Usage

1. distsync setup
	Answer propmts
	Creates ~/.distsync
	Creates ~/.distsyncd
1. Copy ~/.distsyncd to your servers.
1. Run distsyncd on servers.
1. distsync upload foo.tar.gz 
1. Voilà! Your files are now on all your servers.


## What does this do?

* `setup` creates two AWS Identities with limited permissions.  The first is for distsync, it allows distsync to upload to a single s3 bucket and send signed notifications to SNS.  The second is for distsyncd which gives it permissions to watch for notifications and list the bucket.
* `upload` encrypts the specified file, uploads it to s3, and notifies servers it is available.
* `distsyncd` watches for notifications, and on a new file being available will download it to the local path using BitTorrent and HTTPS from S3.

