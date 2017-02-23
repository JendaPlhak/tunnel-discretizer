from linalg import *
from multiprocessing import Process, Queue, cpu_count
import numpy as np
import json



class TunnelCurve:
    def __init__(self, tunnel, delta):
        self.centers = [s.center for s in tunnel.t]
        self.dirs = []
        self.delta = delta
        dump_file = "/home/asd/tmp/curve_dirs.json"
        load_file = "/home/asd/tmp/curve_dirs.json"


        if False:
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
                result_q.put((idx, tunnel.find_minimal_disk(center, normal).normal))

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

