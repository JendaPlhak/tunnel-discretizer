#!/usr/bin/env python

"""Tunnel generator.

Usage:
  readtunel.py (-f | --file) <filename> (-d | --draw)
  readtunel.py (-f | --file) <filename>

Options:
  -h --help     Show this screen.
  -f --file     File containing information about tunnel in molecule.
  -d --draw     Draw scenario into picture using vpython

"""

import random
import json
import sys
import visual as vs

from visual import *
from docopt import docopt
from digger import *
from smoother import smoothen_tunnel, SmoothOpts


if __name__ == '__main__':
    arguments = docopt(__doc__)
    tunnel = Tunnel()
    tunnel.load_from_file(arguments['<filename>'])
    tunnel.t = tunnel.t[:]
    draw_ARG = arguments["--draw"]

    # print make_circle([(random.random() * 10, random.random() * 10) for _ in xrange(100000)])
    # sys.exit()
    # draw tunnel

    disks = dig_tunnel(tunnel, DigOpts(0.1))
    disks = smoothen_tunnel(disks, SmoothOpts(0.025))

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
                                color=(1,0,0), shaftwidth=1)
        # for i, disk in enumerate(disks[:]):
        #     # if (i != 0):
        #     #     print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
        #     vs.ring(pos=disk.center - centScene,
        #             axis=disk.normal,
        #             radius=disk.radius,
        #             thickness=0.01,
        #             color=(1,0,0))
        #     if i % 2 == 1:
        #         vVis = vs.arrow(pos=tuple(disk.center - centScene),
        #                         axis=tuple(disk.normal * 0.25) ,
        #                         color=(0,1,0), shaftwidth=0.5)


    disks_structured = [disk.to_dict() for disk in disks]
    # print json.dumps(disks_structured, sort_keys=True,
    #                     indent=4, separators=(',', ': '))

