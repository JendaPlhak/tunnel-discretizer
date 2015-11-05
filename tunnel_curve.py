from linalg import *


class TunnelCurve:
    def __init__(self, centers, delta):
        self.centers = centers
        self.dirs    = [normalize(centers[i + 1] - centers[i]) \
                        for i in xrange(len(centers) - 1)]
        self.delta   = delta 

    # Returns weighted(smoothen) dir of tunnel in given point.
    # `active_segment_idx` is index of segment between tunnel centers.
    # `d` is distance from beginning of given segment
    def get_weighted_dir(self, active_segment_idx, d):
        i = active_segment_idx

        centers_dist = np.linalg.norm(self.dirs[i])
        w1 = 1 - d / centers_dist 
        w2 = d / centers_dist
        
        return normalize(self.dirs[i]) * w1 + normalize(self.dirs[i + 1]) * w2