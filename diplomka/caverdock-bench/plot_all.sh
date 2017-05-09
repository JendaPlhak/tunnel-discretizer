plot_in_folder () {
    for file in $1/*.dat; do
        base=$(basename "$file" .dat)
        gnuplot -e "set terminal png size 1920, 1350; set output 'img/$1_${base}.png';set xlabel 'Ligand position';set ylabel 'Energy';plot '$1/${base}.dat' u 1:4 w l t 'covergence', '$1-noconv/${base}.dat' u 1:4 w l t 'orig';"
        # gnuplot -e "set output 'img/$1_${base}.png'"
    done
}
# gnuplot -e

plot_in_folder "1BRT"
plot_in_folder "1YGE"
