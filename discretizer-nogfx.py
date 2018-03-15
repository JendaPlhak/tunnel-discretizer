#!/usr/bin/env python

"""Molecule tunnel discretization tool.

Usage:
  discretizer.py -f | --file <in-filename> [--delta <delta>] [-o <out-filename>]

Options:
  -h --help                         Show this help.
  -f --file                         File containing information about tunnel in molecule in PDB format.
  -o --output-file <out-filename>   Dump disks to file in dsd format.
  --delta <delta>                   Maximal distance between disks.

"""

import random
import json
import sys

from docopt import docopt
from digger import *


if __name__ == '__main__':
    arguments = docopt(__doc__)
    tunnel = Tunnel()
    filename = arguments['<in-filename>']
    tunnel.load_from_file(filename)
    tunnel.t = tunnel.t[:]

    delta = float(arguments["--delta"] or 0.3)
    disks = []
    disks = dig_tunnel(tunnel, DigOpts(delta, filename))

    output_path = arguments.get("--output-file")
    if output_path:
        with open(output_path, "w") as output_file:
            for disk in disks:
                line = "{} {} {} {} {} {} {}\n".format(disk.center[0],
                    disk.center[1], disk.center[2], disk.normal[0], disk.normal[1],
                    disk.normal[2], disk.radius)
                output_file.write(line)

