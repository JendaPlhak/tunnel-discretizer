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
            else:
                print "Unexpected data: " + line

        infile.close()
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
        cont_spheres = self.get_all_containing_point(center)
        inters       = []

        for s1 in self.t:
            for s2 in cont_spheres:
                if s2.intersect_ball(s1) and plane.intersection_sphere(s1) != None:
                    inters.append(s1)
        return inters