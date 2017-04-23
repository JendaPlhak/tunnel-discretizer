# Tunnel discretizer
Tunnel discretization tool used to transform tunnel in protein molecule
leading to active site of it to a sequence of disks that enclose the tunnel and
are not consequently distant by more than `delta`.

## Usage
Generally used as `python discretizer.py -f tunnel.pdb --delta 0.3`. For
more details see `python discretizer.py --help`.

## Install
This project uses some external libraries. These will be initialized and
build using command `make setup`.
