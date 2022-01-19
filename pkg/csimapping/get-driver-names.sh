#!/bin/bash
curl "https://raw.githubusercontent.com/kubernetes-csi/docs/master/book/src/drivers.md"   | grep "^\["|  awk -F "|" '{printf "%s,%s\n",$1,$2}' | grep "\`"| sed 's/`//g' | sed 's/ //g' | sed 's/\[//g' | sed 's/\]//g'| sed 's/[(|)]//g'
