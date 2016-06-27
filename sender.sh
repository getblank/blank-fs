#!/bin/bash

curl -X POST $VBOT_ADDR -d $"`verlog`" --header "Authorization: Basic $VBOT_KEY" --header "Content-Type:text/plain;charset=UTF-8"
