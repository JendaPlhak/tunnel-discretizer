#!/usr/bin/env python

import visual as vs
import numpy as np
import sys
import math
import json

f_error = 0.000001

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
    def plot(self):
        vs.ring(pos=self.center, 
                axis=self.normal, 
                radius=self.radius, 
                thickness=0.01)

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

def point_dist(point1, point2):
    return abs(np.linalg.norm(point2 - point1))

def normalize(v):
    norm = np.linalg.norm(v)
    if norm == 0: 
       return v
    return v / norm

def null_space(A, eps=1e-15):
    u, s, vh = np.linalg.svd(A)
    null_space = np.compress(s <= eps, vh, axis=0)
    return null_space[0]

# For two disks calculate vector that realizes radius of segment that emerges by 
# projecting disk into the plain determined by normals of disks d1 and d2.
def get_radius_vectors(d1, d2):
    # Get normal vector of plain given by normal vectors of disks.
    normal = null_space(np.array([d1.normal, d2.normal, null_vec]))
    # print "Normal:\n{}".format(normal)
    # Calculate directions of segments created by projection to the plain.
    seg_dir1 = null_space(np.array([d1.normal, normal, null_vec]))
    seg_dir2 = null_space(np.array([d2.normal, normal, null_vec]))
    # Get radius vector
    seg_dir1 = normalize(seg_dir1) * d1.radius
    seg_dir2 = normalize(seg_dir2) * d2.radius

    return seg_dir1, seg_dir2


# projection distance of two discs
def disk_dist(d1, d2):
    if (d1.normal == d2.normal).all():
        return np.linalg.norm(d1.center - d2.center)
    else:
        # get radius vectors
        seg_dir1, seg_dir2 = get_radius_vectors(d1, d2)
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

# For two given points and vector determining plain, calculate disk, that is 
# perpendicular to that plain, and its vertices are point1 and point2
def get_new_disk_points(point1, point2, n):
    center = (point1 + point2) / 2.
    # print "Point 1: {}, Point 2: {}".format(point1, point2)
    # print "New center: {}".format(center)
    normal = normal_in_plane(n, point2 - point1)
    # print "New normal: {}".format(normal)
    return Disk(center, normal, point_dist(point1, point2) / 2.)

def shift_new_disk(new_disk, prev_disk):
    new_dir, prev_dir = get_radius_vectors(new_disk, prev_disk)
    # Get proj. disk vertices
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    prev_vert_1 = prev_disk.center + prev_dir
    prev_vert_2 = prev_disk.center - prev_dir

    v1 = new_vert_1 - prev_disk.center 
    v2 = new_vert_2 - prev_disk.center 

    # print prev_dir, v1
    # print prev_dir, v2
    # Determine whether both segment vertices lie in the same half-plane.
    # TODO this is probably wrong
    if np.dot(prev_dir, v1) * np.dot(prev_dir, v2) > 0.:
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
            new_disk = get_new_disk_points(prev_vert_1, new_vert_2, new_disk.normal)
        else:
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

    # print prev_disk.normal, v2
    # print np.dot(prev_disk.normal, v2)

    assert np.dot(prev_disk.normal, v1) > -0.001 and np.dot(prev_disk.normal, v2) > -0.0001
    return new_disk


d1 = Disk(np.array([0,0,0]), np.array([1,0,0]), 1)
d2 = Disk(np.array([-1,0,0]), normalize(np.array([1,1,0])), 1)


# new_d = shift_new_disk(d2, d1)
# for disk in [d1, d2]:
#     disk.plot()
# vs.ring(pos=new_d.center, 
#             axis=new_d.normal, 
#             radius=new_d.radius, 
#             color=vs.color.red,
#             thickness=0.01)


# sys.exit()
# tunnel = tunnel[0:4]

# draw tunnel
for i, s in enumerate(tunnel):
  sVis = vs.sphere(pos = (s.center[0], s.center[1], s.center[2]), radius = s.radius, opacity=0.3)
  # central line
  if (i < len(tunnel)-1):
    s2 = tunnel[i+1]
    vVis = vs.arrow(pos=(s.center[0], s.center[1], s.center[2]), 
                    axis=(s2.center[0]-s.center[0], s2.center[1]-s.center[1], s2.center[2]-s.center[2]), 
                    color=(1,0,0), shaftwidth=1)

def is_follower(prev_disk, new_disk):
    prev_dir, new_dir = get_radius_vectors(prev_disk, new_disk)

    # Get proj. disk vertices
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    v1 = new_vert_1 - prev_disk.center 
    v2 = new_vert_2 - prev_disk.center

    return np.dot(prev_disk.normal, v1) > -f_error or np.dot(prev_disk.normal, v2) > -f_error




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

        new_normal = normal * w1 + normals[i + 1] * w2
        r          = get_radius(new_normal, disk_center, tunnel)
        new_disk   = Disk(disk_center, new_normal, r)

        if (len(disks) > 0):
            if not is_follower(disks[-1], new_disk):
                size += eps
                continue
            new_disk = shift_new_disk(new_disk, disks[-1])

        if (len(disks) > 1 and disk_dist(new_disk, disks[-2]) < delta):
            disks[-1] = new_disk
        else:
            disks.append(new_disk)
        size += eps

# draw disks
for i, disk in enumerate(disks[:]):
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

