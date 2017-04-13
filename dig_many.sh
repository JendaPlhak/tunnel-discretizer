for f in `find tunells -name "*.pdb"`; do
    echo $f
    time python readtunel.py -f $f --delta 0.3
done
