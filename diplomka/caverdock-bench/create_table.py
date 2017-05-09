from os import listdir
from os.path import isfile, join
import re
import csv

def get_ligand_info(path):
    atoms = 0
    torsions = None
    with open(path, "r") as file:
        for line in file:
            if re.match("^ATOM\s+.*", line):
                atoms += 1
            else:
                match = re.match("REMARK  (\d+) active torsions:.*", line)
                if match:
                    torsions = int(match.group(1))
    return atoms, torsions

def get_energies(path):
    energies = []
    with open(path, 'r') as csvfile:
        reader = csv.reader(csvfile, delimiter=' ')
        for row in reader:
            energies.append(float(row[3]))
    return max(energies), sum(energies) / len(energies)

def do_molecule(name):
    result = []
    files = []
    for f in listdir(name):
        if re.match(".*\.dat$", f):
            files.append(f)
    for f in files:
        max_e, avg_e = get_energies(name + "/" + f)
        orig_max_e, orig_avg_e = get_energies(name + "-noconv/" + f)

        match = re.match(".*(m\d\d\d)\.dat", f)
        atoms, torsions = get_ligand_info(name + "/" + match.group(1) + ".pdbqt")
        result.append("{} {} {:.2f} {:.2f} {:.2f} {:.2f} ".format(
            atoms, torsions, orig_max_e, max_e, orig_avg_e, avg_e))
    return sorted(result)

if __name__ == '__main__':
    for line in do_molecule("1BRT"):
        print(line)
    for line in do_molecule("1YGE"):
        print(line)
