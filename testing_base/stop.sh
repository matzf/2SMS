#!/bin/bash

SIG="-TERM"
killall "$SIG" endpoint &>/dev/null
killall "$SIG" node_exporter &>/dev/null
killall "$SIG" scraper &>/dev/null
killall "$SIG" prometheus &>/dev/null
killall "$SIG" manager &>/dev/null

SIG="-KILL"
killall "$SIG" endpoint &>/dev/null
killall "$SIG" node_exporter &>/dev/null
killall "$SIG" scraper &>/dev/null
killall "$SIG" prometheus &>/dev/null
killall "$SIG" manager &>/dev/null

exit 0
