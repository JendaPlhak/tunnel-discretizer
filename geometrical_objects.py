from linalg import *

class Disk:
    def __init__(self, center, normal, radius):
        self.center = center
        self.normal = normal
        self.radius = radius
    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center), 
                "normal" : packer(self.normal), 
                "radius" : self.radius}
    def plot(self):
        vs.ring(pos=self.center, 
                axis=self.normal, 
                radius=self.radius, 
                thickness=0.01)

class Plane:
    def __init__(self, point, normal):
        self.point  = point
        self.normal = normal
    def contains(self, point):
        diff = np.dot(self.normal, self.point) - np.dot(self.normal, point)
        return abs(diff) < f_error
    def intersection_line(self, line):
        return line.intersection_plane(self)

class Sphere:
    def __init__(self, center, radius):
        self.center = center
        self.radius = radius
    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center), 
                "normal" : packer(self.normal)}
    def intersection_line(self, line):
        return line.intersection_sphere(self)

class Line:
    def __init__(self, point, dir_):
        self.dir   = dir_
        self.point = point
    def get_line_point(self, t):
        return self.point + t * self.dir
    def intersection_sphere(self, sphere):
        return get_intersection_line_sphere(sphere, self)
    def intersection_plane(self, plane):
        # First get vectors for plane parametric form
        # print "plane", plane.point, plane.normal
        # print "line", self.point, self.dir
        v1 = normalize(null_space(np.array([plane.normal, null_vec, null_vec])))
        v2 = normalize(null_space(np.array([plane.normal, v1, null_vec])))
        assert is_3D_basis(plane.normal, v1, v2)

        r_side = np.transpose(np.array([plane.point - self.point]))
        l_side = np.transpose(np.array([self.dir, v1, v2]))
        # print "left side:\n", l_side
        # print "right side:\n", r_side

        l_side_inv = np.linalg.inv(l_side)
        parameters = np.dot(l_side_inv, r_side)

        # print "Result:"
        # print "parameters:\n", parameters
        # print "dir:\n", self.dir
        # print self.get_line_point(parameters[0])
        intersection = self.get_line_point(parameters[0])
        assert plane.contains(intersection)
        return intersection




