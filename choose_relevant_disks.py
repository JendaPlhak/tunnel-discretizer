#!/usr/bin/env python

"""Tool for choosing representative disks.

Usage:
  choose_relevant_disks.py -f | --file <in-filename> [--ct <ct>] [--nt <nt>] [--rt <rt>] [-o <out-filename>]

Options:
  -h --help                         Show this help.
  -f --file                         File containing information about the disks in dsd format.
  --ct <ct>                         Center threshold, maximal distance between the centers of two disks in the set of representative disks
  --nt <nt>                         Normal threshold, maximal angle in degrees between the normals of two disks in the set of representative disks
  --rt <rt>                         Radius threshold, maximal difference between the radiuses of two disks in the set of representative disks
  -o --output-file <out-filename>   Dump representative disks (not too far, not too close, capturing the abrupt changes) to file in dsd format.

"""

import random
import json
import sys
import visual as vs

from docopt import docopt
from digger import *
from visual import *

def load_disks_from_file(filename):
    disks = []
    if filename:
        with open(filename, "r") as input_file:
            for line in input_file:
                numbers = line.split()
                disk = Disk(np.array([float(numbers[0]), float(numbers[1]), float(numbers[2])]), np.array([float(numbers[3]), float(numbers[4]), float(numbers[5])]), float(numbers[6]))
                disks.append(disk)
        input_file.close()
    return disks


def choose_representative_disks(disks):
    representative_disks = []
    representative_disks.append(disks[0])
    last = disks[0]
    for disk in disks:
        if (point_dist(last.center, disk.center) > center_threshold):
            representative_disks.append(disk)
            last = disk
        if (angle_norm_vectors(last.normal, disk.normal) > normal_threshold):
            representative_disks.append(disk)
            last = disk
        if (last.radius - disk.radius > radius_threshold):
            representative_disks.append(disk)
            last=disk
    representative_disks.append(disks[-1])
    return representative_disks

if __name__ == '__main__':
    arguments = docopt(__doc__)
    filename = arguments['<in-filename>']
    disks = []
    disks = load_disks_from_file(filename)

    center_threshold = float(arguments["--ct"] or 3.0)
    normal_threshold = float(arguments["--nt"] or 20.0)
    radius_threshold = float(arguments["--rt"] or 1.0)

    disks = choose_representative_disks(disks)

    output_path = arguments.get("--output-file")
    if output_path:
        with open(output_path, "w") as output_file:
            for disk in disks:
                line = "{} {} {} {} {} {} {}\n".format(disk.center[0],
                    disk.center[1], disk.center[2], disk.normal[0], disk.normal[1],
                    disk.normal[2], disk.radius)
                output_file.write(line)

