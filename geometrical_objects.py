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

class Sphere:
    def __init__(self, center, radius):
        self.center = center
        self.radius = radius
    def to_dict(self):
        packer  = lambda c : tuple([c[0], c[1], c[2]])
        return {"center" : packer(self.center), 
                "normal" : packer(self.normal)}

class Line:
    def __init__(self, dir_, point):
        self.dir   = dir_
        self.point = point
    def get_line_point(self, t):
        return self.point + t * self.dir