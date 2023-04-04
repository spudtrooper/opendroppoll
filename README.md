# opendroppoll

Utility to poll using [opendrop](https://github.com/seemoo-lab/opendrop) and write airdrop events to a mongo db.

## Usage

* Install [mongodb](https://www.mongodb.com/)
* Install [opendrop](https://github.com/seemoo-lab/opendrop)
* Run `opendrop find` to run through the certificate installation
* Run this with `go run main.go` or `scripts/run.sh`
* View results in `opendroppoll` table