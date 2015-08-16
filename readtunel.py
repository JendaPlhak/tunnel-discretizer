#!/usr/bin/env python

import visual as vs
import numpy as np

infile = file("tunnel.pdb")

tunnel=[]

for line in infile.readlines():
  words = line.split()
  #print words
  if words[0] == "ATOM":
    s = (float(words[6]), float(words[7]), float(words[8]), float(words[9]))
    tunnel.append(s);
  else:
    print "Unexpected data: " + line

print "Tunnel readed (" + str(len(tunnel)) + " spheres)."

null_vec = np.array([0,0,0])

class Disk:
    def __init__(self, center, normal, radius):
        self.center = center
        self.normal = normal
        self.radius = radius

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
def proj_dist(d1, d2):
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





d1 = Disk(np.array([0,0,0]), np.array([1,0,0]), 1)
d2 = Disk(np.array([1,0,0]), np.array([1,1,0]), 1)

print proj_dist(d1, d2)


vectors = []

# draw tunnel
for i, s in enumerate(tunnel):
  sVis = vs.sphere(pos = (s[0], s[1], s[2]), radius = s[3], opacity=0.3)
  # central line
  if (i < len(tunnel)-1):
    s2 = tunnel[i+1]
    vVis = vs.arrow(pos=(s[0], s[1], s[2]), 
                    axis=(s2[0]-s[0], s2[1]-s[1], s2[2]-s[2]), 
                    color=(1,0,0), shaftwidth=1)
delta = 0.1
disks = []
centers = [np.array([s[0], s[1], s[2]]) for s in tunnel]
normals = [normalize(centers[i + 1] - centers[i]) for i in xrange(len(centers) - 1)]
# Calculate disks position
for i, s in enumerate(tunnel):
    if i == len(tunnel) - 2:
        break

    center      = centers[i]
    next_center = centers[i + 1]
    r1     = s[3]
    r2     = tunnel[i + 1][3]
    normal = normals[i]   
    centers_dist = np.linalg.norm(next_center - center)

    step = 0;
    while step * delta < centers_dist:
        size = step * delta
        disk_center = normal * size + center

        w1 = 1 - size / centers_dist 
        w2 = size / centers_dist

        r = r1 * w1 + r2 * w2
        # print "w1 = {}, w2 = {}".format(w1, w2)
        # print "r1 = {}, r2 = {}, r = {}".format(r1, r2, r)
        disks.append(Disk(disk_center, normal * w1 + normals[i + 1] * w2, r))

        step += 1

# draw disks
for i, disk in enumerate(disks):
    if (i != 0):
        print "Disk distance: {}".format(proj_dist(disks[i-1], disk))
    vs.ring(pos=disk.center, 
            axis=disk.normal, 
            radius=disk.radius, 
            thickness=0.1)




infile.close()


