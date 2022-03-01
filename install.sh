#!/bin/sh

ROBOT_IP="..."

scp ./vacuum_bot root@$ROBOT_IP:/usr/local/bin
scp ./S11vacuum_bot root@$ROBOT_IP:/etc/init
