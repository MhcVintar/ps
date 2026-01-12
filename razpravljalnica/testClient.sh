#!/bin/bash
make buildClient && cd build && ./client -a $1 -p $2
