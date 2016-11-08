class SmoothOpts:
    def __init__(self, max_diff):
        self.max_diff = max_diff


def smoothen_tunnel(disks, opts):
    print(opts.max_diff)
    print(opts)
    for i in xrange(len(disks) - 1):
        if disks[i].radius - disks[i+1].radius > opts.max_diff:
            disks[i+1].radius = disks[i].radius - opts.max_diff

        j = i
        while j >= 0 and disks[j+1].radius - disks[j].radius > opts.max_diff:
            disks[j].radius = disks[j+1].radius - opts.max_diff
            j -= 1
    return disks

