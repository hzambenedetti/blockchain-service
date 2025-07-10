# Setup instructions 

## Requirements

- Go programming language v1.24.4 can be downloaded in it's [official website](https://go.dev/)
- Internet connection (used to install go packages, not needed otherwise)
- Linux system

## Setup 

Create the database and binaries directories and build the program and build the program:
```bash
    mkdir -p tmp/blocks_0 tmp/blocks_1 tmp/blocks_2 bin
    go build -o bin/server ./cmd/server/main.go
```
you must also create a .env file based on the .env.example one.
If you wish to run the blockchain in localhost, just run the following command:
```bash
cp .env.example .env
```


## Run
To run the program, a node id and port to listen must be provided:
```bash
    ./bin/server -nodeIdx 0 -fiberPort 3100
    ./bin/server -nodeIdx 1 -fiberPort 3200
    ./bin/server -nodeIdx 2 -fiberPort 3300
```
The nodeIdx parameters must be in the range [0,2]. As for the fiberPort the user can choose any port it sees fit.


## Routes 

### POST /upload

Creates a new block in the blockchain 

Body example:
```json
{
    "hash": "1894a19c85ba153acbf743ac4e43fc004c891604b26f8c69e1e83ea2afc7c48f",
    "momId": "bd2702ab7d81edaa3d6ba66c2d1d3dbb4ff4fba8ede520163443c3076fc4a85b",
    "notaryId": "21122ee1-a5bc-4fcc-bead-065acfc38edf",
    "userId": "a101fb26-8b78-4e93-9fab-67d291a28fb7",
    "cnpj": "58.474.125/0001-33"
}
```


### GET /list 

No Body 

### GET /verify 

No body, but requires a query parameter like so:

/verify?hash=1894a19c85ba153acbf743ac4e43fc004c891604b26f8c69e1e83ea2afc7c48f

## Notes 

- If the blockchain is to be run with the fides system at least one of the nodes must use the port 3100.


