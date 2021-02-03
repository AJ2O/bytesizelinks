#!/bin/bash
if systemctl list-units --full -all | grep -Fq "bytesizelinks.service"; then
    service bytesizelinks stop
fi
rm -rf /webapps/bytesizelinks