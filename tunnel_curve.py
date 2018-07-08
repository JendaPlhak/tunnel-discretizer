import json
import numpy as np
import os
from multiprocessing import Process, Queue, cpu_count
from sortedcontainers import SortedList

from linalg import *
from geometrical_objects import Segment

class PhiCurve(object):

    def __init__(self, tunnel, delta = 3):
        self._delta = delta

        self._centers = tunnel.centers
        self._centers.append(self._centers[-1] + (self._centers[-1] - self._centers[-2]))
        self._c_dists = self._load_distances()

    def _load_distances(self):
        distances = SortedList([0.])
        dst = 0.
        for i in range(1, len(self._centers)):
            center_dst = np.linalg.norm(self._centers[i] - self._centers[i - 1])
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

        alpha = np.linalg.norm(self._centers[i + 1] - self._centers[i])
        v1 = normalize(self._centers[i + 1] - self._centers[i])
        v2 = normalize(self._centers[i + 2] - self._centers[i + 1])

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



