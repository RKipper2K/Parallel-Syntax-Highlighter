# Parallel-Syntax-Highlighter
Parallel version of my syntax highlighter made during my 4th semester. The purpose of this code is to highlight the difference between running the syntax analysis in serial (file by file) in comparison to running it in parallel, with up to 5 maximum files being processed at the same time.
The program is meant to be run from a terminal, and it receives 2 inputs:
- The path of the folder containing the txt files you wish to have analyzed
- The path where you wish for the program to store the output html files

Once the program is run, it will create two folders in the output path, determined by the second input parameter: one called "secuencial", where it will store the files created sequentially, and one called "paralelo", where it will store the files created in parallel.
The program will also print an output to the terminal, declaring how much time the sequential and parallel highlighting took respectively.
