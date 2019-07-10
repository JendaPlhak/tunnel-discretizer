
for f in `find tunnels_prod -name "*tun_cl_*.pdb"`; do
    # echo $f
    python3.7 -W ignore discretizer.py -f $f --delta 0.3
done
