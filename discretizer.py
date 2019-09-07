#!/usr/bin/env python

"""Molecule tunnel discretization tool.

Usage:
  discretizer.py -f | --file <in-filename> [-d] [--delta <delta>] [-o <out-filename>]

Options:
  -h --help                         Show this help.
  -f --file                         File containing information about tunnel in molecule in PDB format.
  -d --draw                         Draw scenario into picture using vpython
  -o --output-file <out-filename>   Dump disks to file in dsd format.
  --delta <delta>                   Maximal distance between disks.

"""

import random
import json
import sys
import vpython as vs
import csv

from docopt import docopt
from digger import *
from vpython import *

def load_disks(filename):
    with open(filename, "r") as file:
        reader = csv.reader(file)
        disks = []
        for line in reader:
            disks.append(Disk(
                (float(line[0]), float(line[1]), float(line[2])),
                (float(line[3]), float(line[4]), float(line[5])),
                float(line[6])
            ))
        return disks

if __name__ == '__main__':
    arguments = docopt(__doc__)
    tunnel = Tunnel()
    filename = arguments['<in-filename>']
    tunnel.load_from_file(filename)
    tunnel.t = tunnel.t[:]
    draw_ARG = arguments["--draw"]

    delta = float(arguments["--delta"] or 0.3)
    disks = []
    # disks = dig_tunnel(tunnel, DigOpts(delta, filename))
    disks = load_disks(filename + ".disks")

    # draw disks
    if draw_ARG:
        # set first sphere as a center of scene
        cMin = tunnel.t[0].center.copy()
        cMax = tunnel.t[0].center.copy()
        for i, t in enumerate(tunnel.t):
            #import pdb;pdb.set_trace()
            for j in range(0, 3):
                if cMin[j] > t.center[j]-t.radius:
                    cMin[j] = t.center[j]-t.radius
                if cMax[j] < t.center[j]+t.radius:
                    cMax[j] = t.center[j]+t.radius
        centScene = cMin + (cMax-cMin)/2

        for i, s in enumerate(tunnel.t):
            sVis = vs.sphere(pos = vs.vec(*(s.center - centScene)),
                                radius = s.radius, opacity=0.3)
            # central line
            if (i < len(tunnel.t)-1):
                s2 = tunnel.t[i+1]
                vVis = vs.arrow(pos=vs.vec(*(s.center - centScene)),
                                axis=vs.vec(s2.center[0]-s.center[0],
                                        s2.center[1]-s.center[1],
                                        s2.center[2]-s.center[2]),
                                color=vs.vec(1,0,0), shaftwidth=0.3)
        for i, disk in enumerate(disks[:]):
            # if (i != 0):
            #     print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
            vs.ring(pos=vs.vec(*(disk.center - centScene)),
                    axis=vs.vec(*disk.normal),
                    radius=disk.radius,
                    thickness=0.01,
                    color=vs.vec(1,0,0))
            # if i % 2 == 1:
            #     vVis = vs.arrow(pos=tuple(disk.center - centScene),
            #                     axis=tuple(disk.normal * 0.25) ,
            #                     color=(0,1,0), shaftwidth=0.5)

    output_path = arguments.get("--output-file")
    if output_path:
        with open(output_path, "w") as output_file:
            for disk in disks:
                line = "{} {} {} {} {} {} {}\n".format(disk.center[0],
                    disk.center[1], disk.center[2], disk.normal[0], disk.normal[1],
                    disk.normal[2], disk.radius)
                output_file.write(line)

