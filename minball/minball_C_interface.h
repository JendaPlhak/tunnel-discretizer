#ifdef __cplusplus
extern "C"
{
#endif

struct Ball2D {
    double center[2];
    double radius;
};

struct Ball2D compute_minball2D(struct Ball2D* balls, int size);

#ifdef __cplusplus
}
#endif
