from linalg import *
import numpy as np


class TunnelCurve:
    def __init__(self, centers, delta):
        self.centers = centers
        self.dirs    = [normalize(centers[i + 1] - centers[i]) \
                        for i in xrange(len(centers) - 1)]
        self.delta   = delta

    # Returns weighted(smoothen) dir of tunnel in given point.
    # `active_segment_idx` is index of segment between tunnel centers.
    # `d` is distance from beginning of given segment
    # def get_weighted_dir(self, active_segment_idx, d):
    #     i = active_segment_idx

    #     centers_dist = np.linalg.norm(self.dirs[i])
    #     w1 = 1 - d / centers_dist
    #     w2 = d / centers_dist

    #     return normalize(self.dirs[i]) * w1 + normalize(self.dirs[i + 1]) * w2
    def get_weighted_dir(self, active_segment_idx, d):
        max_dist = self.delta
        i = active_segment_idx
        cent_dist = []
        w_dir = np.array([0.,0.,0.])

        for j in xrange(len(self.dirs)):
            dist = self._center_distance_from_point(j, i, d)
            if dist < max_dist:
                w_dir += self.dirs[j] * (max_dist - dist) ** 3

        return normalize(w_dir)

    def _get_normal_weights(self, active_segment_idx, d, r):
        pass

    def _center_distance_from_point(self, center_idx, active_segment_idx, d):
        c    = center_idx
        step = np.sign(active_segment_idx - center_idx)
        dist = 0.

        while c != active_segment_idx:
            dist += np.linalg.norm(self.centers[c + 1] - self.centers[c])
            c += step

        if step >= 0:
            dist += d
        else: # subtract extra distance.
            dist -= np.linalg.norm(self.centers[c + 1] - self.centers[c]) - d

        return dist

