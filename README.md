# ZCanGate
A gatway application reading and writing CAN bus messages for Zehnder ComfoAir Q series devices for monitoring and control.

## State
This is work in progress and under heavy development.

## Usage

### Install dependencies

This project uses [golang dep](https://github.com/golang/dep) for dependency management. 

Install dependencies

    dep ensure
    
Build the project

    go build .
    
Run it

    ./zcangate