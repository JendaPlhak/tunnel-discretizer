#!/usr/bin/env python

"""Molecule tunnel discretization tool.

Usage:
  show_disks.py -f | --file <in-filename> [-d <all-disks>] [-r <representative-disks>]

Options:
  -h --help                         Show this help.
  -f --file                         File containing information about tunnel in molecule in PDB format.
  -d                                File containing information about the discrete disks in the tunnel in dsd format, output of discretizer script.
  -r                                File containing information about the representative disks in the tunnel in dsd format, output of choose_relevant_disks script.

"""

import random
import json
import sys
import visual as vs

from docopt import docopt
from digger import *
from choose_relevant_disks import *
from visual import *


if __name__ == '__main__':
    arguments = docopt(__doc__)
    tunnel = Tunnel()
    filename = arguments['<in-filename>']
    tunnel.load_from_file(filename)
    tunnel.t = tunnel.t[:]

    all_disks = load_disks_from_file(arguments['<all-disks>'])
    relevant_disks = load_disks_from_file(arguments['<representative-disks>'])

    draw_ARG = True
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
            sVis = vs.sphere(pos = tuple(s.center - centScene),
                                radius = s.radius, opacity=0.3)
            # central line
            if (i < len(tunnel.t)-1):
                s2 = tunnel.t[i+1]
                vVis = vs.arrow(pos=s.center - centScene,
                                axis=(s2.center[0]-s.center[0],
                                        s2.center[1]-s.center[1],
                                        s2.center[2]-s.center[2]),
                                color=(1,0,0), shaftwidth=0.3)
        for i, disk in enumerate(all_disks[:]):
            # if (i != 0):
            #     print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
            vs.ring(pos=disk.center - centScene,
                    axis=disk.normal,
                    radius=disk.radius,
                    thickness=0.01,
                    color=(1,0,0))
        for i, disk in enumerate(relevant_disks[:]):
            # if (i != 0):
            #     print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
            vs.ring(pos=disk.center - centScene,
                    axis=disk.normal,
                    radius=disk.radius,
                    thickness=0.1,
                    color=(0,1,0))
            # if i % 2 == 1:
            #     vVis = vs.arrow(pos=tuple(disk.center - centScene),
            #                     axis=tuple(disk.normal * 0.25) ,
            #                     color=(0,1,0), shaftwidth=0.5)


