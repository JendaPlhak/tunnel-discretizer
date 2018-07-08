import json
import minball
import numpy as np
import os
import scipy
import time
import random
from multiprocessing import Process, Queue, cpu_count
from scipy import optimize

from geometrical_objects import *
from linalg import *

class Tunnel:

    def __init__(self, opts):
        self.t = None
        self.centers = None
        self.dirs = None

        self._load_from_file(opts.filename)
        self._compute_optimized_trajectory(opts)

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

    def _load_from_file(self, filename):
        self.t = []
        infile = file(filename)

        for line in infile.readlines():
            words = line.split()
            #print words
            if len(words) > 0 and words[0] == "ATOM":
                center = np.array([float(words[6]), float(words[7]), float(words[8])])
                radius = float(words[9])
                self.t.append(Sphere(center, radius));

        infile.close()
        self.check_requirements()
        print "Tunnel read (" + str(len(self.t)) + " spheres)."

    def _compute_optimized_trajectory(self, opts):
        home_dir = os.path.expanduser("~")
        dump_file = home_dir + "/tmp/" + opts.filename.replace("/", "_") + ".json"
        load_file = dump_file if os.path.exists(dump_file) else None

        if load_file:
            with open(load_file) as infile:
                in_ = json.load(infile)
                self.centers = [np.array(c) for c in in_["centers"]]
                self.dirs = [np.array(d) for d in in_["dirs"]]
        else:
            self.centers, self.dirs = self._compute_tunnel_body()

        if dump_file:
            with open(dump_file, 'w') as outfile:
                out = {
                    "centers": [list(c) for c in self.centers],
                    "dirs": [list(d) for d in self.dirs],
                }
                json.dump(out, outfile)

    def _find_minimal_disk(self, point, init_normal, curve):
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

    def _compute_tunnel_body(self):
        curve = _BallCentersCurve(self.t)

        def get_minimal(self, task_q, result_q):
            while True:
                task = task_q.get()
                if task is None:
                    task_q.put(None)
                    break
                center, normal, idx = task
                result_q.put((idx, self._find_minimal_disk(center, normal, curve)))

        task_q, result_q = Queue(), Queue()
        processes = [Process(target=get_minimal, args=(self, task_q, result_q))
            for __ in xrange(cpu_count())
        ]

        ball_centers = [ball.center for ball in self.t]
        balls_count = len(ball_centers)

        # Add one extra ball center so that each original center has direction
        ball_centers.append(ball_centers[-1] + (ball_centers[-1] - ball_centers[-2]))

        i = 0
        for center, next_center in zip(ball_centers, ball_centers[1:]):
            n = normalize(next_center - center)
            task_q.put((center, n, i))
            i += 1

        task_q.put(None)
        for p in processes:
            p.start()
        for p in processes:
            p.join()

        centers = [None for __ in xrange(balls_count)]
        dirs = [None for __ in xrange(balls_count)]
        while not result_q.empty():
            idx, disk = result_q.get()
            centers[idx] = disk.center
            dirs[idx] = disk.normal
        return centers, dirs

class _BallCentersCurve:

    def __init__(self, balls):
        self.centers = [ball.center for ball in balls]

    # Finds whether given `disk` is passed through by given curve in topological
    # sense.
    def pass_through_disk(self, disk):
        first_pass_sgn = None
        last_pass_sgn = None
        split = None

        for i in xrange(len(self.centers) - 1):
            seg = Segment(self.centers[i], self.centers[i + 1])
            inter = seg.intersection_disk(disk)

            if inter is not None:
                d = self.centers[i + 1] - self.centers[i]
                d_sgn = np.sign(np.dot(disk.normal, d))
                if disk.contains(self.centers[i + 1]):
                    split = split or d
                elif split is not None:
                    sgn = d_sgn * np.sign(np.dot(disk.normal, split))
                    if sgn > 0:
                        first_pass_sgn = first_pass_sgn or d_sgn
                        last_pass_sgn = d_sgn
                    split = None
                else:
                    first_pass_sgn = first_pass_sgn or d_sgn
                    last_pass_sgn = d_sgn
        return first_pass_sgn is not None and first_pass_sgn == last_pass_sgn
