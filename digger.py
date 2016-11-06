import math
import numpy as np
from decimal import *

from geometrical_objects import *
from linalg import *
from tunnel import Tunnel
from minimal_enclosing import make_circle
from digger import *
from tunnel_curve import TunnelCurve


class DigOpts:
    def __init__(self, delta):
        self.delta = delta
        self.eps   = delta / 10.

def dig_tunnel(tunnel, opts):
    centers = [s.center for s in tunnel.t]
    normals = [normalize(centers[i + 1] - centers[i]) \
               for i in xrange(len(centers) - 1)]
    disks   = [fit_disk_tunnel(normals[0], centers[0], 0, tunnel, opts.delta)]

    curve = TunnelCurve(centers, 8.)

    # Calculate disks position
    for i, s in enumerate(tunnel.t):
        if i == len(tunnel.t) - 2:
            break

        print("Processing ball NO. %d" % i)
        normal       = normals[i]
        center       = centers[i]
        next_center  = centers[i + 1]
        centers_dist = np.linalg.norm(next_center - center)

        # line between centers
        line = Line(center, next_center - center)

        size = 0;
        while size < centers_dist:
            # Form initial disk center
            if (size != 0):
                disk_plane = disks[-1].get_plane()
                inter = line.intersection_plane(disk_plane)
                size  = np.linalg.norm(inter - center) + opts.eps

            disk_center = center + normal * size
            new_normal  = curve.get_weighted_dir(i, size)

            new_disk = fit_disk_tunnel(new_normal, disk_center, i, tunnel, opts.delta)
            new_disk = shift_new_disk(new_disk, disks[-1], i, tunnel, opts.delta)
            disk_dist(new_disk, disks[-1]) < opts.delta + f_error

            if (len(disks) > 1 and disk_dist(new_disk, disks[-2]) < opts.delta):
                disks[-1] = new_disk
            else:
                disks.append(new_disk)
            size += opts.eps

    return disks


def fit_disk_tunnel(normal, center, ball_idx, tunnel, delta):
    disk_plane  = Plane(center, normal)
    circle_cuts = []
    n_samples   = 15

    for sphere in tunnel.get_all_intersecting_disk(disk_plane, center):
        # calculate center of cap that we get by intersection disk_plane
        # and sphere
        cut_circle = disk_plane.intersection_sphere(sphere)
        if cut_circle == None:
            continue

        circle_cuts.append(cut_circle)

    assert(circle_cuts)
    discrete_approx = []
    for c in circle_cuts:
        approx = c.get_approximation_delta(delta)
        discrete_approx.extend([tuple(x) for x in approx])
    t, u, radius = make_circle(discrete_approx)

    new_center = disk_plane.get_point_for_param(t, u)
    # control that calculation went as it was supposed to go.
    # control_s = Sphere(new_center, radius)
    # for point in discrete_approx:
    #     control_point = disk_plane.get_point_for_param(point[0], point[1])
    #     assert control_s.ball_contains(control_point)

    return Disk(new_center, normal, radius)

# For two given points and vector determining plane, calculate disk, that is
# perpendicular to that plane, and its vertices are point1 and point2
def get_new_disk_points(point1, point2, n):
    center = (point1 + point2) / 2.
    # print "Point 1: {}, Point 2: {}".format(point1, point2)
    # print "New center: {}".format(center)
    normal = get_normal_in_plane(n, point2 - point1)
    # print "New normal: {}".format(normal)
    new_disk = Disk(center, normal, point_dist(point1, point2) / 2.)
    assert new_disk.contains(point1) and new_disk.circle_contains(point1)
    assert new_disk.contains(point2) and new_disk.circle_contains(point2)
    return new_disk

def is_follower(prev_disk, new_disk):
    prev_dir, new_dir = get_radius_vectors(prev_disk, new_disk)

    # Get proj. disk vertices
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir

    v1 = new_vert_1 - prev_disk.center
    v2 = new_vert_2 - prev_disk.center

    return np.dot(prev_disk.normal, v1) > -f_error \
        and np.dot(prev_disk.normal, v2) > -f_error

# get vertices of segments determined by disks in orthogonal projection.
def get_vertices(disk1, disk2):
    dir1, dir2 = get_radius_vectors(disk1, disk2)
    # Get proj. disk vertices
    disk1_vert_1 = disk1.center + dir1
    disk1_vert_2 = disk1.center - dir1
    disk2_vert_1 = disk2.center + dir2
    disk2_vert_2 = disk2.center - dir2

    # Ensure that vertices corresponds to each other in given order.
    dist11 = point_dist(disk1_vert_1, disk2_vert_1)
    dist12 = point_dist(disk1_vert_1, disk2_vert_2)
    dist21 = point_dist(disk1_vert_2, disk2_vert_1)
    dist22 = point_dist(disk1_vert_2, disk2_vert_2)
    if dist11 + dist22 > dist12 + dist21:
        disk2_vert_1, disk2_vert_2 = disk2_vert_2, disk2_vert_1

    return disk1_vert_1, disk1_vert_2, disk2_vert_1, disk2_vert_2

def shift_new_disk(new_disk, prev_disk, ball_idx, tunnel, delta):
    # print "\n\n"
    # print "Previous Disk:"
    # print prev_disk.to_geogebra()
    # print "New Disk:"
    # print new_disk.to_geogebra()
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 \
        = get_vertices(new_disk, prev_disk)

    v1 = new_vert_1 - prev_disk.center
    v2 = new_vert_2 - prev_disk.center

    # Determine whether both segment vertices lie in the same half-plane. If not
    # fix it.
    if not (np.dot(prev_disk.normal, v1) >= 0 and np.dot(prev_disk.normal, v2) >= 0.):
        # print prev_disk.normal, v1, prev_disk.normal, prev_disk.normal
        # print np.dot(prev_disk.normal, v1) * np.dot(prev_disk.normal, prev_disk.normal)

        # print prev_disk.normal, v2, prev_disk.normal, prev_disk.normal
        # print np.dot(prev_disk.normal, v2) * np.dot(prev_disk.normal, prev_disk.normal)


        # At least one of the vertices must be on the right side
        assert np.dot(prev_disk.normal, v1) >= 0. or np.dot(prev_disk.normal, v2) >= 0.

        # Find which vertex is in the wrong half plane.
        # Then shift appropriate vertex so that prev_disk is followed by
        # new disk.
        if np.dot(prev_disk.normal, v1) < 0.:
            # print "First one!"
            shift_vert = prev_vert_1 + prev_disk.normal * (delta / 5)
            new_disk   = get_new_disk_points(shift_vert, new_vert_2, prev_disk.normal)
        elif np.dot(prev_disk.normal, v2) < 0.:
            # print "Second one!"
            shift_vert = prev_vert_2 + prev_disk.normal * (delta / 5)
            new_disk   = get_new_disk_points(shift_vert, new_vert_1, prev_disk.normal)

            new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 \
                = get_vertices(new_disk, prev_disk)
            v1 = new_vert_1 - prev_disk.center
            v2 = new_vert_2 - prev_disk.center

    assert is_follower(prev_disk, new_disk)

    # Ensure that disks are not too far from each other.
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 = get_vertices(new_disk, prev_disk)
    v = new_vert_1 - prev_vert_1
    if (not np.linalg.norm(v) < delta):
        shifted_vert = prev_vert_1 + normalize(v) * delta
        new_disk = get_new_disk_points(shifted_vert, new_vert_2, new_disk.normal)

    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 = get_vertices(new_disk, prev_disk)
    v = new_vert_2 - prev_vert_2
    if (not np.linalg.norm(v) < delta):
        shifted_vert = prev_vert_2 + normalize(v) * delta
        new_disk = get_new_disk_points(new_vert_1, shifted_vert, new_disk.normal)

    assert is_follower(prev_disk, new_disk)
    # perform re-fitting
    new_disk = fit_disk_tunnel(new_disk.normal, new_disk.center, ball_idx, tunnel, delta)
    # print "Revised disk : {}".format(new_disk.to_geogebra())

    assert is_follower(prev_disk, new_disk)
    if not is_follower(prev_disk, new_disk):
        # print "Recursive shift!"
        new_disk = shift_new_disk(new_disk, prev_disk, ball_idx, tunnel, delta)
        # print "Complete!"

    # Check whether our function does what it is supposed to do.
    new_dir, prev_dir = get_radius_vectors(new_disk, prev_disk)

    #     print abs(get_radius(new_disk.normal, new_disk.center, tunnel) - new_disk.radius)
    new_vert_1 = new_disk.center + new_dir
    new_vert_2 = new_disk.center - new_dir
    # print "New vertex 1: '{}', New vertex 2: '{}'".format(new_vert_1, new_vert_2)

    v1 = new_vert_1 - prev_disk.center
    v2 = new_vert_2 - prev_disk.center

    # print prev_disk.normal, v1
    # print np.dot(prev_disk.normal, v1)

    # print prev_disk.normal, v2
    # print np.dot(prev_disk.normal, v2)

    assert np.dot(prev_disk.normal, v1) > -0.0001
    assert np.dot(prev_disk.normal, v2) > -0.0001
    return new_disk
