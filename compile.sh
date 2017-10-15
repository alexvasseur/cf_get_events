#!/bin/bash

GOOS=darwin go build -o bcr-plugin-osx
#GOOS=linux go build -o bcr-plugin-linux
#GOOS=windows GOARCH=amd64 go build -o bcr-plugin.exe
if [ $? != 0 ]; then
   printf "Error when executing compile\n"
   exit 1
fi
cf uninstall-plugin bcr
cf install-plugin -f ./bcr-plugin-osx
