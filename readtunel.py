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

from visual import *
import visual as vs
import numpy as np
import sys
from decimal import *
import math
import json
from docopt import docopt
from geometrical_objects import *
from linalg import *
from minimal_enclosing import make_circle
import random

        
def fit_disk_tunnel(normal, center, tunnel):
    disk_plane  = Plane(center, normal)
    circle_cuts = []
    n_samples   = 30

    for sphere in tunnel:
        # calculate center of cap that we get by intersection disk_plane 
        # and sphere 
        cap_center = disk_plane.orthogonal_projection(sphere.center)
        # if center is out of the ball, the intersection is empty or one point.
        if not sphere.inner_ball_contains(cap_center):
            continue

        line_dir = disk_plane.get_base_vectors()[0]
        line     = Line(cap_center, line_dir)
        inter_points = sphere.intersection_line(line)
        assert len(inter_points) == 2

        # Result should be symmetrical, therefore half of calculation can
        # be removed.
        v = inter_points[0] - line.point
        r = np.linalg.norm(v)

        cut_circle = Circle(disk_plane.orthogonal_proj_param(cap_center), r)
        circle_cuts.append(cut_circle)

    discrete_approx = []
    for c in circle_cuts:
        approx = c.get_approximation(n_samples)
        discrete_approx.extend([tuple(x) for x in approx])
    t, u, radius = make_circle(discrete_approx)

    new_center = disk_plane.get_point_for_param(t, u)
    # control that calculation went as it was supposed to go.
    control_s = Sphere(new_center, radius)
    for point in discrete_approx:
        control_point = disk_plane.get_point_for_param(point[0], point[1])
        assert control_s.ball_contains(control_point)

    return Disk(new_center, normal, radius)

# return true if point1 and point2 are in the same half-plane based on plane
# normal and line_point 
def in_same_half_plane(normal, line_point, point1, point2):
    v1 = point1 - line_point
    v2 = point2 - line_point
    return np.dot(normal, v1) * np.dot(normal, v2) >= 0

# Calculate vector u, that lies in plane given by v and n, is perpendicular to
# vector v and u, n have the same orientation. Also we require that u is 
# normalized.
def normal_in_plane(n, v):
    # print "n: {} v: {}".format(n, v)
    assert np.dot(v, v) != 0
    k = 1
    l = - np.dot(n, v) / np.dot(v, v)
    u = normalize(k * n + l * v)
    assert np.dot(u, v) < 0.00001
    return u

# For two given points and vector determining plan, calculate disk, that is 
# perpendicular to that plan, and its vertices are point1 and point2
def get_new_disk_points(point1, point2, n):
    center = (point1 + point2) / 2.
    # print "Point 1: {}, Point 2: {}".format(point1, point2)
    # print "New center: {}".format(center)
    normal = normal_in_plane(n, point2 - point1)
    # print "New normal: {}".format(normal)
    return Disk(center, normal, point_dist(point1, point2) / 2.)

def shift_new_disk(new_disk, prev_disk, tunnel):
    new_dir, prev_dir = get_radius_vectors(new_disk, prev_disk)
    # Get proj. disk vertices
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    prev_vert_1 = prev_disk.center + prev_dir
    prev_vert_2 = prev_disk.center - prev_dir

    v1 = new_vert_1 - prev_disk.center 
    v2 = new_vert_2 - prev_disk.center 

    # Determine whether both segment vertices lie in the same half-plane.
    if np.dot(prev_disk.normal, v1) >= 0 and np.dot(prev_disk.normal, v2) >= 0.:
        # if so, return since no shifting is necessary.
        return new_disk

    # print prev_dir, v1, prev_dir, prev_disk.normal
    # print np.dot(prev_dir, v1) * np.dot(prev_dir, prev_disk.normal)

    # print prev_dir, v2, prev_dir, prev_disk.normal
    # print np.dot(prev_dir, v2) * np.dot(prev_dir, prev_disk.normal)


    # At least one of the vertices must be on the right side
    assert np.dot(prev_disk.normal, v1) >= 0. or np.dot(prev_disk.normal, v2) >= 0.
    # Which vertex is in the wrong half plane
    if np.dot(prev_disk.normal, v1) < 0.:
        # print "First one!"
        if in_same_half_plane(prev_dir, prev_disk.center, \
                                new_vert_1, prev_vert_1):
            # print "New_vert 1 and prev_vert 1 are in the same half-plane"
            new_disk = get_new_disk_points(prev_vert_1, new_vert_2, new_disk.normal)
        else:
            # print "New_vert 1 and prev_vert 2 are in the same half-plane"
            new_disk = get_new_disk_points(prev_vert_2, new_vert_2, new_disk.normal)

    elif np.dot(prev_disk.normal, v2) < 0.:
        # print "Second one!"
        # print new_vert_2, prev_vert_1
        if in_same_half_plane(prev_dir, prev_disk.center, \
                                new_vert_2, prev_vert_1):
            # print "New_vert 2 and prev_vert 1 are in the same half-plane"
            new_disk = get_new_disk_points(prev_vert_1, new_vert_1, new_disk.normal)
        else:
            # print "New_vert 2 and prev_vert 2 are in the same half-plane"
            new_disk = get_new_disk_points(prev_vert_2, new_vert_1, new_disk.normal)


    # Check whether our function does what it is supposed to do.
    new_dir, prev_dir = get_radius_vectors(new_disk, prev_disk)

    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    # print new_vert_1
    # print new_vert_2

    v1 = new_vert_1 - prev_disk.center 
    v2 = new_vert_2 - prev_disk.center 

    # print prev_disk.normal, v1
    # print np.dot(prev_disk.normal, v1)

    # perform re-fitting
    new_disk = fit_disk_tunnel(new_disk.normal, new_disk.center, tunnel)

    # print prev_disk.normal, v2
    # print np.dot(prev_disk.normal, v2)
    # if abs(get_radius(new_disk.normal, new_disk.center, tunnel) - new_disk.radius) > 0.1:
    #     new_disk.radius = get_radius(new_disk.normal, new_disk.center, tunnel)
    #     return shift_new_disk(new_disk, prev_disk, tunnel)

        # print abs(get_radius(new_disk.normal, new_disk.center, tunnel) - new_disk.radius)
    # assert np.dot(prev_disk.normal, v1) > -0.0001 and np.dot(prev_disk.normal, v2) > -0.0001
    return new_disk

def is_follower(prev_disk, new_disk):
    prev_dir, new_dir = get_radius_vectors(prev_disk, new_disk)

    # Get proj. disk vertices
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    v1 = new_vert_1 - prev_disk.center 
    v2 = new_vert_2 - prev_disk.center

    return np.dot(prev_disk.normal, v1) > -f_error or np.dot(prev_disk.normal, v2) > -f_error

def load_tunnel_from_file(filename):
    infile = file(filename)
    tunnel = []

    for line in infile.readlines():
        words = line.split()
        #print words
        if len(words) > 0 and words[0] == "ATOM":
            center = np.array([float(words[6]), float(words[7]), float(words[8])])
            radius = float(words[9])
            tunnel.append(Sphere(center, radius));
        else:
            print "Unexpected data: " + line

    infile.close()
    print "Tunnel readed (" + str(len(tunnel)) + " spheres)."
    return tunnel

if __name__ == '__main__':
    arguments = docopt(__doc__)
    tunnel = load_tunnel_from_file(arguments['<filename>'])
    # tunnel = tunnel[0:10]
    draw_ARG = arguments["--draw"]

    # print make_circle([(random.random() * 10, random.random() * 10) for _ in xrange(100000)])
    # sys.exit()
    # draw tunnel

    delta = 0.1
    eps   = delta / 10.
    disks = []
    centers = [s.center for s in tunnel]
    normals = [normalize(centers[i + 1] - centers[i]) for i in xrange(len(centers) - 1)]
    # Calculate disks position
    for i, s in enumerate(tunnel):
        if i == len(tunnel) - 2:
            break

        center      = centers[i]
        next_center = centers[i + 1]
        r1     = s.radius
        r2     = tunnel[i + 1].radius
        normal = normals[i]   
        centers_dist = np.linalg.norm(next_center - center)

        size = 0;
        while size < centers_dist:
            # print "size: {}".format(size)
            # Form initial disk center
            disk_center = normal * size + center

            w1 = 1 - size / centers_dist 
            w2 = size / centers_dist

            new_normal = normal * w1 + normals[i + 1] * w2
            new_disk   = fit_disk_tunnel(new_normal, disk_center, tunnel)

            if (len(disks) > 0):
                if not is_follower(disks[-1], new_disk):
                    size += eps
                    continue
                new_disk = shift_new_disk(new_disk, disks[-1], tunnel)

            if (len(disks) > 1 and disk_dist(new_disk, disks[-2]) < delta):
                disks[-1] = new_disk
            else:
                disks.append(new_disk)
            size += eps

    # draw disks
    if draw_ARG:
        for i, s in enumerate(tunnel):
            sVis = vs.sphere(pos = (s.center[0], s.center[1], s.center[2]), 
                                radius = s.radius, opacity=0.9)
            # central line
            if (i < len(tunnel)-1):
                s2 = tunnel[i+1]
                vVis = vs.arrow(pos=(s.center[0], s.center[1], s.center[2]), 
                                axis=(s2.center[0]-s.center[0], s2.center[1]-s.center[1], s2.center[2]-s.center[2]), 
                                color=(1,0,0), shaftwidth=1)
        for i, disk in enumerate(disks[:]):
            if (i != 0):
                print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
            vs.ring(pos=disk.center, 
                    axis=disk.normal, 
                    radius=disk.radius, 
                    thickness=0.01,
                    color=(1,0,0))


    disks_structured = [disk.to_dict() for disk in disks]
    # print json.dumps(disks_structured, sort_keys=True,
    #                     indent=4, separators=(',', ': '))

