#!/bin/sh

id
echo "hello action from container " > /action/container.txt
echo "hello host from container" > /host/container.txt
ls /action
ls /host
