import json
import numpy as np
import os
from multiprocessing import Process, Queue, cpu_count
from sortedcontainers import SortedList

from linalg import *
from geometrical_objects import Segment

class TunnelCurve(object):
    def __init__(self, tunnel, delta, opts):
        self.centers = [s.center for s in tunnel.t]
        self.dirs = []
        self.delta = delta

        home_dir = os.path.expanduser("~")
        dump_file = home_dir + "/tmp/" + opts.filename.replace("/", "_") + ".json"
        load_file = dump_file if os.path.exists(dump_file) else None

        if load_file:
            with open(load_file) as infile:
                self.dirs = [np.array(d) for d in json.load(infile)]
        else:
            self.dirs = self._compute_dirs(tunnel)

        if dump_file:
            with open(dump_file, 'w') as outfile:
                json.dump([list(d) for d in self.dirs], outfile)

    def _compute_dirs(self, tunnel):
        def get_minimal(task_q, result_q, tunnel):
            while True:
                task = task_q.get()
                if task is None:
                    task_q.put(None)
                    break
                center, normal, idx = task
                result_q.put((idx, tunnel.find_minimal_disk(center, normal, self).normal))

        N_CORES = cpu_count()
        task_q = Queue()
        result_q = Queue()
        processes = [Process(target=get_minimal, args=(task_q, result_q, tunnel))
            for __ in xrange(N_CORES)
        ]

        dirs_count = len(self.centers) - 1
        for i in xrange(dirs_count):
            n = normalize(self.centers[i + 1] - self.centers[i])
            task_q.put((self.centers[i], n, i))

        task_q.put(None)
        for p in processes:
            p.start()
        for p in processes:
            p.join()
        dirs = [None for __ in xrange(dirs_count)]
        while not result_q.empty():
            idx, normal = result_q.get()
            dirs[idx] = normal
        return dirs

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

class PhiCurve(TunnelCurve):

    def __init__(self, *args):
        super(PhiCurve, self).__init__(*args)
        self._delta = 3

        self.centers.append(self.centers[-1] + (self.centers[-1] - self.centers[-2]))
        self._c_dists = self._load_distances()
        self.dirs.append(self.dirs[-1])

    def _load_distances(self):
        distances = SortedList([0.])
        dst = 0.
        for i in range(1, len(self.centers)):
            center_dst = np.linalg.norm(self.centers[i] - self.centers[i - 1])
            dst += center_dst
            distances.add(dst)

        return distances

    def get_direction(self, t0):
        t1 = max(t0 - self._delta, 0)
        t2 = min(t0 + self._delta, self._max_dst())

        g1 = self._get_prev_center(t1)
        g0 = self._get_prev_center(t0)
        g2 = self._get_prev_center(t2)

        v = self._integrate_interval(t0, t1, self._c_dists[g1], is_before_center = True)
        v += self._integrate_between_centers(t0, g1 + 1, g0, is_before_center = True)

        v += self._integrate_interval(t0, self._c_dists[g0], t0, is_before_center = True)
        v += self._integrate_interval(t0, t0, self._c_dists[g0 + 1], is_before_center = False)

        v += self._integrate_between_centers(t0, g0 + 1, g2, is_before_center = False)
        v += self._integrate_interval(t0, self._c_dists[g2], t2, is_before_center = False)

        return normalize(v)

    def _max_dst(self):
        return self._c_dists[-1]

    def _get_prev_center(self, t):
        i = max(0, self._c_dists.bisect_left(t) - 1)
        assert self._c_dists[i] <= t
        return i

    def _integrate_between_centers(self, t0, first, last, is_before_center):
        u = np.array([0.,0.,0.])
        sign = 1 if is_before_center else -1
        for i in range(first, last):
            u += self._integrate_interval(t0, self._c_dists[i], self._c_dists[i+1],
                is_before_center)
        return u

    def _integrate_interval(self, t0, x, y, is_before_center):
        sign = 1 if is_before_center else -1

        # assert self._get_prev_center(p) == self._get_prev_center(q)
        i = self._get_prev_center(x)

        d = self._delta + sign * (-t0 + x)
        assert self._delta >= d >= 0

        alpha = np.linalg.norm(self.centers[i + 1] - self.centers[i])
        # v1 = normalize(self.centers[i + 1] - self.centers[i])
        # v2 = normalize(self.centers[i + 2] - self.centers[i + 1])
        v1 = normalize(self.dirs[i])
        v2 = normalize(self.dirs[i + 1])

        p = (x - self._c_dists[i]) / alpha
        q = (y - self._c_dists[i]) / alpha

        def eval_indef(t):
            return t * (
                12*d**2 * v1
                + 3*t**3 * alpha**2 * (v2 - v1)
                - 6*t*d * (d*(v1 - v2) -sign * 2*alpha*v1)
                + 4 * alpha * t**2 * (-sign * 2*d * (v1 - v2) + alpha * v1)
            ) / (12 * alpha)

        return eval_indef(q) - eval_indef(p)



