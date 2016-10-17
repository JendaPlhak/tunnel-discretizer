import numpy as np
import math

f_error = 0.0000001
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

def is_3D_basis(v1, v2, v3):
    return abs(np.dot(v1, v2)) < f_error  \
        and abs(np.dot(v1, v3)) < f_error \
        and abs(np.dot(v2, v3)) < f_error

# return true if point1 and point2 are in the same half-plane based on plane
# normal and line_point
def in_same_half_plane(normal, line_point, point1, point2):
    v1 = point1 - line_point
    v2 = point2 - line_point
    return np.dot(normal, v1) * np.dot(normal, v2) >= 0

# Calculate vector u, that lies in plane given by v and n, is perpendicular to
# vector v and u, n have the same orientation. Also we require that u is
# normalized.
def get_normal_in_plane(n, v):
    # print "n: {} v: {}".format(n, v)
    assert np.dot(v, v) != 0
    k = 1
    l = - np.dot(n, v) / np.dot(v, v)
    u = normalize(k * n + l * v)
    assert np.dot(u, v) < 0.00001
    return u

# For two disks calculate vector that realizes radius of segment that emerges by
# projecting disk into the plan determined by normals of disks d1 and d2.
def get_radius_vectors(d1, d2):
    # Get normal vector of plan given by normal vectors of disks.
    normal = null_space(np.array([d1.normal, d2.normal, null_vec]))
    # print "Normal:\n{}".format(normal)
    # Calculate directions of segments created by projection to the plan.
    seg_dir1 = null_space(np.array([d1.normal, normal, null_vec]))
    seg_dir2 = null_space(np.array([d2.normal, normal, null_vec]))
    # Get radius vector
    seg_dir1 = normalize(seg_dir1) * d1.radius
    seg_dir2 = normalize(seg_dir2) * d2.radius

    return seg_dir1, seg_dir2


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
