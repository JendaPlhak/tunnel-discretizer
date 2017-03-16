import minball
import numpy as np
import scipy
import time
import random
from scipy import optimize
from geometrical_objects import *
from linalg import *

class Tunnel:

    def load_from_file(self, filename):
        self.t = []
        infile = file(filename)

        for line in infile.readlines():
            words = line.split()
            #print words
            if len(words) > 0 and words[0] == "ATOM":
                center = np.array([float(words[6]), float(words[7]), float(words[8])])
                radius = float(words[9])
                self.t.append(Sphere(center, radius));
            # else:
                # print "Unexpected data: " + line

        infile.close()
        self.check_requirements()
        print "Tunnel readed (" + str(len(self.t)) + " spheres)."

    def get_neighbors(self, sphere_idx):
        first = None
        last  = None

        for i, s in enumerate(self.t):
            if (self.t[sphere_idx].intersect_ball(s)):
                if first == None:
                    first = i
                    last  = i
                else:
                    last = i

        return first, last

    # Return all spheres containing given point
    def get_all_containing_point(self, point):
        spheres = []
        for s in self.t:
            if s.ball_contains(point):
                spheres.append(s)
        return spheres;

    def get_all_intersecting_disk(self, plane, center):
        # print "Containing center %d" % len(self.get_all_containing_point(center))
        cont_spheres = self.get_all_containing_point(center)
        inters       = []
        inter_circs  = set(plane.intersection_sphere(s) for s in cont_spheres)

        circles_count = 0

        while len(inter_circs) != circles_count:
            circles_count = len(inter_circs)
            for s1 in self.t:
                c1 = plane.intersection_sphere(s1)
                if c1 is None:
                    continue
                for ref_circle in set(inter_circs):
                    if ref_circle.has_intersection_circle(c1):
                        inters.append(s1)
                        inter_circs.add(c1)
                        break
        return inters

    def check_requirements(self):
        for i, s1 in enumerate(self.t):
            for s2 in self.t[i+1:]:
                assert(not s1.contains_sphere(s2))
                assert(not s2.contains_sphere(s1))

    def fit_disk(self, normal, center):
        disk_plane  = Plane(center, normal)
        circle_cuts = []

        for sphere in self.get_all_intersecting_disk(disk_plane, center):
            # calculate center of cap that we get by intersection disk_plane
            # and sphere
            cut_circle = disk_plane.intersection_sphere(sphere)
            assert cut_circle is not None

            circle_cuts.append(cut_circle)
        assert circle_cuts

        circles = []
        # print "{"
        for c in circle_cuts:
            # print "%s,"% c.to_geogebra()
            circles.append(minball.Sphere2D(list(c.center), c.radius))
        # print "}"

        min_circle = minball.get_min_sphere2D(circles)
        t, u = min_circle.center
        radius = min_circle.radius

        new_center = disk_plane.get_point_for_param(t, u)
        assert disk_plane.contains(new_center)
        return Disk(new_center, normal, radius)

    def find_minimal_disk(self, point, init_normal, curve):
        def get_axes(normal):
            axis1 = null_space(np.array([normal, null_vec, null_vec]))
            axis2 = null_space(np.array([normal, axis1, null_vec]))
            return axis1, axis2

        def get_rotated_disk(base_normal, theta, phi, axes):
            axis1, axis2 = axes
            v = np.dot(rotation_matrix(axis1, theta), base_normal)
            v = np.dot(rotation_matrix(axis2, phi), v)
            normal = normalize(v)

            disk = self.fit_disk(normal, point)
            if curve.pass_through_disk(disk):
                return disk
            else:
                return None

        best_disk = self.fit_disk(init_normal, point)
        init_radius = best_disk.radius
        for i in xrange(5):
            theta = (math.pi / 3) / 4**i
            # print("Round %d" % i)
            found_better = True
            while found_better:
                found_better = False
                base_normal  = best_disk.normal
                axes         = get_axes(base_normal)

                for phi in np.arange(0, 2*math.pi, 0.1 * (i + 1)):
                    # print theta, phi
                    disk = get_rotated_disk(base_normal, theta, phi, axes)
                    if disk and disk.radius < best_disk.radius:
                        best_disk = disk
                        best_disk.normal *= np.sign(np.dot(best_disk.normal, init_normal))
                        found_better = True
                        # print "Found better!", best_disk.radius

        print "Init radius {}, Optimized: {}".format(init_radius,
            best_disk.radius)
        assert np.dot(best_disk.normal, init_normal) > 0.
        return best_disk
