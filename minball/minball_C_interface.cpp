#include "minball_C_interface.h"

#include <CGAL/Cartesian_d.h>
#include <CGAL/Random.h>
#include <CGAL/Min_sphere_of_spheres_d.h>
#include <array>
#include <vector>
#include <stdio.h>

extern "C" struct Ball2D compute_minball2D(struct Ball2D* balls, int size)
{
    static constexpr int Dim = 2;

    typedef double FT;
    typedef CGAL::Cartesian_d<FT> K;
    typedef CGAL::Min_sphere_of_spheres_d_traits_d<K, FT, Dim> Traits;
    typedef CGAL::Min_sphere_of_spheres_d<Traits> Min_sphere;
    typedef K::Point_d Point;

    std::vector<typename Traits::Sphere> mSpheres;
    for (int i = 0; i < size; ++i) {
        Point center(Dim, balls[i].center, balls[i].center + Dim);
        mSpheres.emplace_back(center, balls[i].radius);
    }

    Min_sphere ms(mSpheres.begin(), mSpheres.end());
    CGAL_assertion(ms.is_valid());

    auto begin = ms.center_cartesian_begin();
    auto end = ms.center_cartesian_end();
    size_t i = 0;

    Ball2D result;
    for (auto it = begin; it != end; ++it, ++i) {
        result.center[i] = *it;
    }
    result.radius = ms.radius();
    return result;
}
