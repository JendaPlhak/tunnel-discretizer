from linalg import *

class Disk:

    def __init__(self, center, normal, radius):
        self.center = center
        self.normal = normalize(normal)
        self.radius = radius

    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center),
                "normal" : packer(self.normal),
                "radius" : self.radius}

    def to_geogebra(self):
        tup_pack = lambda t : "({}, {}, {})".format(t[0], t[1], t[2])
        return "Circle[{0}, {1}, Vector[(0, 0, 0), {2}]]"\
            .format(tup_pack(self.center), self.radius, tup_pack(self.normal))
        return {"center" : packer(self.center),
                "normal" : packer(self.normal),
                "radius" : self.radius}
    # Calculate plane that is determined by disk
    def get_plane(self):
        return Plane(self.center, self.normal)

    def plot(self):
        vs.ring(pos=self.center,
                axis=self.normal,
                radius=self.radius,
                thickness=0.01)

    def contains(self, point):
        vec = point - self.center
        if not abs(np.dot(vec, self.normal)) < f_error:
            return False
        return np.linalg.norm(vec) <= self.radius + f_error

    def circle_contains(self, point):
        vec = point - self.center
        if not abs(np.dot(vec, self.normal)) < f_error:
            return False
        return abs(np.linalg.norm(vec) - self.radius) < f_error


class Plane:

    def __init__(self, point, normal):
        self.point      = point
        self.normal     = normalize(normal)
        self._basis     = None
        self._projector = None

    def contains(self, point):
        diff = np.dot(self.normal, self.point) - np.dot(self.normal, point)
        return abs(diff) < f_error

    def intersection_line(self, line):
        return line.intersection_plane(self)

    def intersection_sphere(self, sphere):
        cap_center = self.orthogonal_projection(sphere.center)

        if not sphere.inner_ball_contains(cap_center):
            return None

        line_dir = self.get_base_vectors()[0]
        line     = Line(cap_center, line_dir)
        inter_points = sphere.intersection_line(line)
        assert len(inter_points) == 2

        # Result should be symmetrical, therefore half of calculation can
        # be removed.
        v = inter_points[0] - line.point
        r = np.linalg.norm(v)

        return Circle(self.orthogonal_proj_param(cap_center), r)

    def get_base_vectors(self):
        if self._basis == None:
            v1 = normalize(null_space(np.array([self.normal, null_vec, null_vec])))
            v2 = normalize(null_space(np.array([self.normal, v1, null_vec])))
            assert abs(np.dot(v1, v2)) < f_error
            assert abs(np.dot(v1, self.normal)) < f_error
            assert abs(np.dot(v2, self.normal)) < f_error
            self._basis = (v1, v2)
        return self._basis

    def get_orthogonal_projector(self):
        if self._projector == None:
            v1, v2      = self.get_base_vectors()
            base_matrix = np.transpose(np.array([v1, v2]))
            self._projector = np.linalg.pinv(base_matrix)
        return self._projector

    def orthogonal_projection(self, point):
        point_proj = point - np.dot(point - self.point, self.normal) * self.normal
        assert self.contains(point_proj)
        return point_proj

    def get_point_for_param(self, t, u):
        v1, v2 = self.get_base_vectors()
        assert self.contains(self.point + t * v1 + u * v2)
        return self.point + t * v1 + u * v2

    def orthogonal_proj_param(self, point):
        v1, v2 = self.get_base_vectors()
        projector  = self.get_orthogonal_projector()
        parameters = np.dot(projector, np.transpose(point - self.point))
        return parameters

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

    def intersect_ball(self, ball):
        return np.linalg.norm(self.center - ball.center) \
            <= self.radius + ball.radius

    # Does ball defined by sphere contains given point (< operator used)
    def inner_ball_contains(self, point):
        return np.linalg.norm(self.center - point) < self.radius

    def ball_contains(self, point):
        return np.linalg.norm(self.center - point) - self.radius < f_error

class Line:

    def __init__(self, point, dir_):
        self.dir   = dir_
        self.point = point
        self._norm_dir = normalize(dir_)

    def contains(self, point):
        if (point == self.point).all():
            return True
        v = normalize(point - self.point)
        return np.linalg.norm(v - self._norm_dir) < f_error

    def get_line_point(self, t):
        return self.point + t * self.dir

    def orthogonal_proj(self, point):
        p = self.point \
            + np.dot(point - self.point, self._norm_dir) * self._norm_dir
        assert self.contains(p)
        return p


    def intersection_sphere(self, sphere):
        return get_intersection_line_sphere(sphere, self)

    # Compute intersection with given plane
    def intersection_plane(self, plane):
        # First get vectors for plane parametric form
        v1, v2 = plane.get_base_vectors()
        assert is_3D_basis(plane.normal, v1, v2)
        assert is_perpendicular(v1, plane.normal)
        assert is_perpendicular(v2, plane.normal)

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
        assert self.contains(intersection)
        print("Ok")
        return intersection

class Circle:

    def __init__(self, center, radius):
        self.center = center
        self.radius = radius

    # For parameter in float in [0, 1] return proportional point on circle.
    def get_point(self, c_percetage):
        assert 0 <= c_percetage and c_percetage <= 1
        x = self.radius * math.cos(c_percetage * 2 * math.pi)
        y = self.radius * math.sin(c_percetage * 2 * math.pi)
        return self.center + np.array([x, y])

    # Calculates equidistant approximation of circle
    def get_approximation(self, n_samples):
        return [self.get_point(float(x) / n_samples) for x in xrange(n_samples)]



