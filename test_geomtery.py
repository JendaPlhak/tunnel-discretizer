import unittest
from geometrical_objects import *

class TestPlane(unittest.TestCase):

  def test_contains(self):
        plane = Plane(np.array([0,0,1]), np.array([0,0,1]))
        self.assertTrue(plane.contains(np.array([0,0,1])))
        self.assertTrue(plane.contains(np.array([1,1,1])))
        self.assertTrue(plane.contains(np.array([5,5,1])))
        self.assertTrue(plane.contains(np.array([0.5,-0.145,1])))

class TestLine(unittest.TestCase):

    def intersection_test(self, plane, line, req_inter):
        intersection = plane.intersection_line(line) 
        self.assertTrue(plane.contains(intersection)) 
        self.assertTrue( np.array_equal(intersection, req_inter) )

    def test_intersection_plane1(self):
        plane = Plane(np.array([0,0,1]), np.array([0,0,1]))
        line  = Line(np.array([0,0,0]), np.array([0,0,1])) 
        self.intersection_test(plane, line, np.array([0,0,1])) 

    def test_intersection_plane2(self):
        plane = Plane(np.array([0,0,1]), np.array([0,0,1]))
        line  = Line(np.array([0,0,0]), np.array([0,0,-1])) 
        self.intersection_test(plane, line, np.array([0,0,1])) 

if __name__ == '__main__':
    unittest.main()