import math
import numpy as np
from decimal import *

from geometrical_objects import *
from linalg import *
from tunnel import Tunnel
from minimal_enclosing import make_circle
from digger import *
from tunnel_curve import PhiCurve


class DigOpts:
    def __init__(self, delta, filename):
        self.delta = delta
        self.eps   = delta * 0.1
        self.look_ahead = 2 * delta
        self.filename = filename

def dig_tunnel(tunnel, opts):
    centers = tunnel.centers
    normals = [normalize(d) for d in tunnel.dirs]

    phi_curve = PhiCurve(tunnel, 6.)
    print(phi_curve.get_direction(0))
    disks = [
        tunnel.fit_disk(phi_curve.get_direction(0), centers[0])
    ]

    # Calculate disks position
    for i, __ in enumerate(tunnel.t):
        if i == len(tunnel.t) - 1:
            break

        print("Processing ball NO. %d" % i)
        center       = centers[i]
        next_center  = centers[i + 1]
        centers_dist = np.linalg.norm(next_center - center)
        # disks.append(tunnel.fit_disk(curve.dirs[i], center))
        # continue

        # line between centers
        line = Line(center, next_center - center)

        size = 0;
        while True:
            # Form initial disk center
            disk_plane = disks[-1].get_plane()
            inter = line.intersection_plane(disk_plane)
            size  = np.linalg.norm(inter - center) + opts.eps
            if size > centers_dist:
                break

            if is_sharp_turn(tunnel, disks[-1], opts):
                # print("Curving!")
                disk_center = disks[-1].center + disks[-1].normal * opts.look_ahead
                new_normal  = disks[-1].normal
                shift_fun   = shift_sharp_turn
                # print "new disk distance: ", disk_dist(disks[-1], new_disk)
            else:
                # print("Moving!")
                disk_center = disks[-1].center + disks[-1].normal * opts.eps

                t = phi_curve._c_dists[i] + size
                new_normal  = phi_curve.get_direction(t)
                # new_normal  = curve.get_weighted_dir(i, size)
                shift_fun   = shift_new_disk

            try:
                new_disk = tunnel.fit_disk(new_normal, disk_center)
                new_disk = shift_fun(disks[-1], new_disk, tunnel, opts)
            except:
                raise
                return disks

            if (len(disks) > 1 and disk_dist(new_disk, disks[-2]) < opts.delta):
                disks[-1] = new_disk
            else:
                disks.append(new_disk)

    return disks

# For two given points and vector determining plane, calculate disk that is
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
def get_vertices(disk1, disk2, normal = None):
    dir1, dir2 = get_radius_vectors(disk1, disk2, normal = normal)
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

    assert disk1.contains(disk1_vert_1) and disk1.circle_contains(disk1_vert_1)
    assert disk1.contains(disk1_vert_2) and disk1.circle_contains(disk1_vert_2)
    assert disk2.contains(disk2_vert_1) and disk2.circle_contains(disk2_vert_1)
    assert disk2.contains(disk2_vert_2) and disk2.circle_contains(disk2_vert_2)
    return disk1_vert_1, disk1_vert_2, disk2_vert_1, disk2_vert_2

# For two disks calculate normal of plane which gives maximum distance
# for disks d1, d2 in orthogonal projection.
def find_max_distance(d1, d2):
    assert np.linalg.norm(d1.normal - d2.normal) < f_error

    v = normalize(d2.center - d1.center)
    if (d1.normal == v).all():
        v = null_space(np.array([d1.normal, null_vec, null_vec]))
    return null_space(np.array([d1.normal, v, null_vec]))

def is_sharp_turn(tunnel, prev_disk, opts):
    disk_center = prev_disk.center + prev_disk.normal * opts.look_ahead
    new_disk = tunnel.fit_disk(prev_disk.normal, disk_center)

    plane_normal = find_max_distance(prev_disk, new_disk)
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 \
        = get_vertices(new_disk, prev_disk, normal = plane_normal)

    v1 = new_vert_1 - prev_vert_1
    v2 = new_vert_2 - prev_vert_2
    d1 = np.linalg.norm(v1)
    d2 = np.linalg.norm(v2)

    # print abs(d1 - d2) / d1
    return abs(d1 - d2) / ((d1 + d2) / 2.) > 0.35

def shift_new_disk(prev_disk, new_disk, tunnel, opts):
    # print "\n\n"
    # print "Previous Disk:"
    # print prev_disk.to_geogebra()
    # print "New Disk:"
    # print new_disk.to_geogebra()
    delta = opts.delta
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 \
        = get_vertices(new_disk, prev_disk)

    v1 = new_vert_1 - prev_disk.center
    v2 = new_vert_2 - prev_disk.center

    # At least one of the vertices must be on the right side
    assert np.dot(prev_disk.normal, v1) >= 0. or np.dot(prev_disk.normal, v2) >= 0.

    # Determine whether both segment vertices lie in the same half-plane. If not
    # fix it.
    if not (np.dot(prev_disk.normal, v1) >= 0 and np.dot(prev_disk.normal, v2) >= 0.):
        # print prev_disk.normal, v1, prev_disk.normal, prev_disk.normal
        # print np.dot(prev_disk.normal, v1) * np.dot(prev_disk.normal, prev_disk.normal)

        # print prev_disk.normal, v2, prev_disk.normal, prev_disk.normal
        # print np.dot(prev_disk.normal, v2) * np.dot(prev_disk.normal, prev_disk.normal)

        # Find which vertex is in the wrong half plane.
        # Then shift appropriate vertex so that prev_disk is followed by
        # new disk.
        if np.dot(prev_disk.normal, v1) < 0.:
            # print "First one!"
            shift_vert = prev_vert_1 + prev_disk.normal * delta * 0.01
            new_disk   = get_new_disk_points(shift_vert, new_vert_2, prev_disk.normal)
        elif np.dot(prev_disk.normal, v2) < 0.:
            # print "Second one!"
            shift_vert = prev_vert_2 + prev_disk.normal * delta * 0.01
            new_disk   = get_new_disk_points(shift_vert, new_vert_1, prev_disk.normal)

    assert is_follower(prev_disk, new_disk)
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 = get_vertices(new_disk, prev_disk)
    # print("({},{}),".format(new_vert_1[0],new_vert_1[1]))
    # print("({},{}),".format(new_vert_2[0],new_vert_2[1]))
    # print("({},{}),".format(prev_vert_1[0],prev_vert_1[1]))
    # print("({},{}),".format(prev_vert_2[0],prev_vert_2[1]))


    v1 = new_vert_1 - prev_vert_1
    v2 = new_vert_2 - prev_vert_2
    d1 = np.linalg.norm(v1)
    d2 = np.linalg.norm(v2)
    # print "Before: ", np.linalg.norm(v1), np.linalg.norm(v2)

    # Ensure that disks are not too far from each other.
    if d1 > delta:
        new_vert_1 = prev_vert_1 + normalize(v1) * delta * 0.95
    if d2 > delta:
        new_vert_2 = prev_vert_2 + normalize(v2) * delta * 0.95

    # print "Disk distance: %f" % disk_dist(new_disk, prev_disk)
    new_disk = get_new_disk_points(new_vert_1, new_vert_2, prev_disk.normal)
    # print "Disk distance: %f" % disk_dist(new_disk, prev_disk)
    assert is_follower(prev_disk, new_disk)
    # perform re-fitting
    new_disk = tunnel.fit_disk(new_disk.normal, new_disk.center)
    # print "Revised disk : {}".format(new_disk.to_geogebra())
    # print "Disk distance: %f > %f\n" % (disk_dist(new_disk, prev_disk), delta)

    if disk_dist(new_disk, prev_disk) > delta or not is_follower(prev_disk, new_disk):
        new_disk = shift_new_disk(prev_disk, new_disk, tunnel, opts)
    assert is_follower(prev_disk, new_disk)

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

    assert np.dot(prev_disk.normal, v1) > -f_error
    assert np.dot(prev_disk.normal, v2) > -f_error
    # assert disk_dist(prev_disk, new_disk) < delta
    return new_disk

def shift_sharp_turn(prev_disk, new_disk, tunnel, opts):
    plane_normal = find_max_distance(prev_disk, new_disk)
    new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 = \
        get_vertices(new_disk, prev_disk, normal = plane_normal)

    # print("({},{}),".format(new_vert_1[0],new_vert_1[1]))
    # print("({},{}),".format(new_vert_2[0],new_vert_2[1]))
    # print("({},{}),".format(prev_vert_1[0],prev_vert_1[1]))
    # print("({},{}),".format(prev_vert_2[0],prev_vert_2[1]))

    v1 = new_vert_1 - prev_vert_1
    v2 = new_vert_2 - prev_vert_2
    d1 = np.linalg.norm(v1)
    d2 = np.linalg.norm(v2)

    # Swap distance variables if we detect decrease in radius
    if prev_disk.radius > new_disk.radius:
        d1, d2 = d2, d1

    if d1 > d2:
        new_vert_1 = prev_vert_1
    else:
        new_vert_1 = prev_vert_1 + normalize(v1) * opts.delta * 0.95
    if d1 < d2:
        new_vert_2 = prev_vert_2
    else:
        new_vert_2 = prev_vert_2 + normalize(v1) * opts.delta * 0.95

    new_disk = get_new_disk_points(new_vert_1, new_vert_2, prev_disk.normal)
    new_disk = tunnel.fit_disk(new_disk.normal, new_disk.center)

    if disk_dist(prev_disk, new_disk) >= opts.delta:
        new_disk = shift_new_disk(prev_disk, new_disk, tunnel, opts)
    # new_vert_1, new_vert_2, prev_vert_1, prev_vert_2 = \
    #     get_vertices(new_disk, prev_disk)
    # print "After"
    # print("({},{}),".format(new_vert_1[0],new_vert_1[1]))
    # print("({},{}),".format(new_vert_2[0],new_vert_2[1]))
    # print("({},{}),".format(prev_vert_1[0],prev_vert_1[1]))
    # print("({},{}),".format(prev_vert_2[0],prev_vert_2[1]))

    # print("Distance:", disk_dist(prev_disk, new_disk))
    assert disk_dist(prev_disk, new_disk) < opts.delta

    return new_disk
