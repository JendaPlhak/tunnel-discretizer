#!/usr/bin/env python

import visual as vs
import numpy as np
import sys
import math
import json

class Disk:
    def __init__(self, center, normal, radius):
        self.center = center
        self.normal = normal
        self.radius = radius
    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center), 
                "normal" : packer(self.normal), 
                "radius" : self.radius}

class Sphere:
    def __init__(self, center, radius):
        self.center = center
        self.radius = radius
    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center), 
                "normal" : packer(self.normal)}

class Line:
    def __init__(self, dir_, point):
        self.dir   = dir_
        self.point = point
    def get_line_point(self, t):
        return self.point + t * self.dir

infile = file("tunnel.pdb")

tunnel = []

for line in infile.readlines():
    words = line.split()
    #print words
    if words[0] == "ATOM":
        center = np.array([float(words[6]), float(words[7]), float(words[8])])
        radius = float(words[9])
        tunnel.append(Sphere(center, radius));
    else:
        print "Unexpected data: " + line

print "Tunnel readed (" + str(len(tunnel)) + " spheres)."

null_vec = np.array([0,0,0])

def normalize(v):
    norm = np.linalg.norm(v)
    if norm == 0: 
       return v
    return v / norm

def null_space(A, eps=1e-15):
    u, s, vh = np.linalg.svd(A)
    null_space = np.compress(s <= eps, vh, axis=0)
    return null_space[0]

# projection distance of two discs
def disk_dist(d1, d2):
    if (d1.normal == d2.normal).all():
        return np.linalg.norm(d1.center - d2.center)
    else:
        # Get normal vector of plain given by normal vectors of disks.
        normal = null_space(np.array([d1.normal, d2.normal, null_vec]))
        # print "Normal:\n{}".format(normal)
        # Calculate directions of segments created by projection to the plain.
        seg_dir1 = null_space(np.array([d1.normal, normal, null_vec]))
        seg_dir2 = null_space(np.array([d2.normal, normal, null_vec]))
        # Get radius vector
        seg_dir1 = normalize(seg_dir1) * d1.radius
        seg_dir2 = normalize(seg_dir2) * d2.radius
        # print "Radius vector 1:\n{}".format(seg_dir1)
        # print "Radius vector 2:\n{}".format(seg_dir2)

        # for two sets of points calculate distances between pairs
        def distances(v_1, v_2):
            return [np.linalg.norm(x[0] - x[1]) for x in zip(v_1, v_2)]

        seg_1_verts = [d1.center + seg_dir1, d1.center - seg_dir1]
        seg_2_verts = [d2.center + seg_dir2, d2.center - seg_dir2]

        dists_1 = distances(seg_1_verts, seg_2_verts)
        dists_2 = distances(list(reversed(seg_1_verts)), seg_2_verts)

        if sum(dists_1) < sum(dists_2):
            return max(dists_1)
        else:
            return max(dists_2)

        # Calculate transformation matrix from standard basis to 
        # (d1.normal, d2.normal, normal).
        print "Transformation matrix from beta to standard:\n{}".format(
            np.array([d1.normal, d2.normal, normal]))
        A     = np.array([d1.normal, d2.normal, normal])
        A_inv = np.linalg.inv(np.array([d1.normal, d2.normal, normal]))
        # print dot(d1.normal, A_inv) 
        # print dot(d2.normal, A_inv) 
        # print dot(normal, A_inv) 
        # Because 'normal' vector is perpendicular to disk normals, the third
        # coordinate of segment directions will be always zero in this basis.
        seg_dir1_proj = dot(seg_dir1, A_inv) 
        seg_dir2_proj = dot(seg_dir2, A_inv) 

# Calculate points of intersection between line and sphere.
# for reference see https://en.wikipedia.org/wiki/Line%E2%80%93sphere_intersection
def get_intersection_line_sphere(sphere, line):
    line_dir = normalize(line.dir)
    r   = sphere.radius
    vec = line.point - sphere.center

    discrim = np.dot(line_dir, vec)**2 \
        - np.dot(vec, vec) \
        + r**2

    if discrim < 0.0:
        return []

    inter_points = []
    for i in xrange(2):
        # d is distance from starting point
        d = - np.dot(line_dir, vec) + (-1)**i * math.sqrt(discrim)
        inter_points.append(line.point + d * line_dir)

    return inter_points
        

def get_radius(normal, center, tunnel):
    line_dir = normalize(null_space(np.array([normal, null_vec, null_vec])))
    line     = Line(line_dir, center)
    r = sys.float_info.min
    for sphere in tunnel:
        inter_points = get_intersection_line_sphere(sphere, line)
        if len(inter_points) == 2:
            v1 = inter_points[0] - line.point
            v2 = inter_points[1] - line.point
            if np.dot(v1, v2) > 0: # Same direction, disc center lies outside of sphere
                # for now, just ignore this case. It is probably right thing to do.
                pass
            elif np.dot(v1, v2) < 0:
                r1 = np.linalg.norm(v1)
                r2 = np.linalg.norm(v2)
                print r1, r2, sphere.radius
                r = max(r, min(r1, r2))
            else:
                raise ValueError("One of vectors is null")
        else:
            pass # We don't wanna mess with this shit. Ok, seriously, this 
            # cases are not relevant since the intersection is either empty or
            # only one point -> r1 == r2 == 0

    if r == sys.float_info.min:
        raise ValueError("Disk not contained in any of the spheres")
    return r




# draw tunnel
for i, s in enumerate(tunnel):
  sVis = vs.sphere(pos = (s.center[0], s.center[1], s.center[2]), radius = s.radius, opacity=0.3)
  # central line
  if (i < len(tunnel)-1):
    s2 = tunnel[i+1]
    # vVis = vs.arrow(pos=(s.center[0], s.center[1], s.center[2]), 
    #                 axis=(s2.center[0]-s.center[0], s2.center[1]-s.center[1], s2.center[2]-s.center[2]), 
    #                 color=(1,0,0), shaftwidth=1)
delta = 0.1
eps   = delta / 10
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
        disk_center = normal * size + center

        w1 = 1 - size / centers_dist 
        w2 = size / centers_dist

        r = r1 * w1 + r2 * w2
        new_normal = normal * w1 + normals[i + 1] * w2
        r = get_radius(new_normal, disk_center, tunnel)
        # print "w1 = {}, w2 = {}".format(w1, w2)
        # print "r1 = {}, r2 = {}, r = {}".format(r1, r2, r)
        new_disk = Disk(disk_center, new_normal, r)

        if (len(disks) > 1 and disk_dist(new_disk, disks[-2]) < delta):
            disks[-1] = new_disk
        else:
            disks.append(new_disk)
        size += eps

# draw disks
for i, disk in enumerate(disks):
    if (i != 0):
        print "Disk distance: {}".format(disk_dist(disks[i-1], disk))
    vs.ring(pos=disk.center, 
            axis=disk.normal, 
            radius=disk.radius, 
            thickness=0.01)
infile.close()

disks_structured = [disk.to_dict() for disk in disks]
# print json.dumps(disks_structured, sort_keys=True,
#                     indent=4, separators=(',', ': '))

