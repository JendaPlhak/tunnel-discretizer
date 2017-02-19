import numpy as np
from geometrical_objects import *

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
