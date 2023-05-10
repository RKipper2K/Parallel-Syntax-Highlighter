// Rene Gerardo Kipper Peña - A01283516
// Actividad Integradora 2: Resaltador de Sintaxis Paralelo
// reflexion al final
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var palabras_reservadas = []string{
	"False", "await", "else", "import", "pass", "None", "break", "except", "in", "raise",
	"True", "class", "finally", "is", "return", "and", "continue", "for", "lambda", "try",
	"as", "def", "from", "nonlocal", "while", "assert", "del", "global", "not", "with",
	"async", "elif", "if", "or", "yield"}

// //Tokens
var ASIG = 100 //Asignación (=)
var SUM = 101  // Suma (+)
var MULT = 102 // Multiplicación (*)
var POW = 103  // Potencia (^)
var LRP = 104  // Abre Paréntesis (
var RRP = 105  // Cierra Paréntesis )
var VAR = 106  // Nombre de Variable Ej. atributo_1
var INT = 107  // Número Entero
var FLT = 108  // Número Flotante
var DIV = 109  // División
var RES = 110  // Resta
var COM = 111  // Comentario
var PER = 112  // Coma
var FUNC = 113 // Funcion
var DOU = 114  // Dos Puntos
var STR = 115  // String
var OBR = 116  // [
var CBR = 117  // ]
var PUN = 118  // .
var ESP = 119  // " "
var ERR = 200  // Error, token inválido

var MT = [][]int{
	{1, ERR, 2, 3, 1, ASIG, SUM, MULT, 6, POW, LRP, RRP, ESP, 8, PER, DOU, 9, OBR, CBR, 7},               // State 0 - Initial
	{1, 1, 1, VAR, 1, VAR, VAR, VAR, VAR, VAR, FUNC, VAR, VAR, VAR, VAR, VAR, VAR, VAR, VAR, VAR},        // State 1 - Variable
	{ERR, ERR, 2, 3, ERR, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT, INT},     // State 2 - Entero
	{ERR, ERR, 3, ERR, 4, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT},     // State 3 - Float Antes del Exponencial E
	{ERR, ERR, 4, ERR, ERR, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, 5, FLT, FLT, FLT, FLT, FLT, FLT},     // State 4 - Float Después del Exponencial E
	{ERR, ERR, 5, ERR, ERR, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT, FLT},   // State 5 - Float Negativo
	{DIV, ERR, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV, DIV}, // State 6 - Comentario/División
	{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7},                                         // State 7 - Comentario
	{8, RES, 2, 2, 3, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES, RES},         // State 8 - Resta o Real
	{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, STR, 9, 9, 9}}                                       //State 9 - String

func checkStringAlphabet(str string) bool {
	for _, charVariable := range str {
		if (charVariable < 'a' || charVariable > 'z') && (charVariable < 'A' || charVariable > 'Z') {
			return false
		}
	}
	return true
}

func filter(c string) int {
	alphaNum := []byte(c)
	if c == "0" || c == "1" || c == "2" || c == "3" || c == "4" || c == "5" || c == "6" || c == "7" || c == "8" || c == "9" {
		return 2
	} else if c == " " || alphaNum[0] == 9 || alphaNum[0] == 10 || alphaNum[0] == 13 {
		return 12
	} else if c == "e" || c == "E" {
		return 4
	} else if c == "-" {
		return 13
	} else if checkStringAlphabet(c) {
		return 0
	} else if c == "_" {
		return 1
	} else if c == "." {
		return 3
	} else if c == "=" || c == "<" || c == ">" || c == "!" {
		return 5
	} else if c == "+" {
		return 6
	} else if c == "*" || c == "%" {
		return 7
	} else if c == "/" {
		return 8
	} else if c == "^" {
		return 9
	} else if c == "(" {
		return 10
	} else if c == ")" {
		return 11
	} else if c == "," {
		return 14
	} else if c == ":" {
		return 15
	} else if c == `"` || c == `\'` || c == `'` { // borrar el || c == `'`
		return 16
	} else if c == "[" || c == "{" {
		return 17
	} else if c == "]" || c == "}" {
		return 18
	} else if c == "#" {
		return 19
	} else {
		return 12
	}
}
func isInTheArray(lexeme string, pal []string) bool {
	for _, b := range pal {
		if b == lexeme {
			return true
		}
	}
	return false
}

func scanning_the_text(linea string, archivo string) {
	state := 0
	lexeme := ""
	tokens := []int{}
	read := true
	charNum := 0
	c := ""
	linea = linea + "\n"
	contador := len(linea)
	file_Output, err := os.OpenFile(archivo, os.O_APPEND|os.O_WRONLY, 0644) // Abre el archivo de salida

	if err != nil {
		fmt.Println(err)
	}
	for contador > 0 {
		for state < 100 {
			if read {
				if charNum < len(linea) {
					c = string(linea[charNum])
					charNum = charNum + 1
					if c == "\t" {
						file_Output.WriteString(`&emsp;`)
					}
				}
			} else {
				read = true
			}
			state = MT[state][filter(string(c))]
			if state < 100 && state != 0 {
				lexeme += string(c)
			}
			contador = contador - 1
			if charNum == len(linea) && state == 7 {
				lexeme = strings.TrimSpace(lexeme)
				state = COM
			}
			if contador == 0 {
				break
			}
		}
		if state == INT {
			read = false
			file_Output.WriteString(`<span style="color:#50A905">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
			contador = contador + 1
		} else if state == FLT {
			read = false
			file_Output.WriteString(`<span style="color:#50A905">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
			contador = contador + 1
		} else if state == ASIG {

			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#8E44AD">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == SUM {

			lexeme += string(c)
			file_Output.WriteString(`<span style="color:red">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == MULT {

			lexeme += string(c)
			file_Output.WriteString(`<span style="color:red">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == POW {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:red">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == LRP {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#B7950B">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == RRP {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#B7950B">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == VAR {
			read = false
			if isInTheArray(lexeme, palabras_reservadas) {
				file_Output.WriteString(`<span style="color:#6401BF">`)
			} else {
				file_Output.WriteString(`<span style="color:#61A1DD">`)
			}
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
			contador = contador + 1
		} else if state == DIV {
			lexeme = string(lexeme[0])
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:red">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == RES {
			lexeme += string(c)
			lexeme = string(lexeme[0])
			file_Output.WriteString(`<span style="color:red">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == COM {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#00FF44">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
			contador = contador + 1
		} else if state == PER {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:black">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == FUNC {
			read = false
			file_Output.WriteString(`<span style="color:#117A65">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
			contador = contador + 1
		} else if state == DOU {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:black">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == STR {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:orange">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == OBR {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#D2B4DE">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == CBR {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:#D2B4DE">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == PUN {
			lexeme += string(c)
			file_Output.WriteString(`<span style="color:black">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")

		} else if state == ESP {
			file_Output.WriteString(`<span style="color:#1E8449">`)
			file_Output.WriteString(`&nbsp;`)
			file_Output.WriteString("</span>\n")

		} else if state == ERR {
			read = false
			file_Output.WriteString(`<span style="color:black">`)
			file_Output.WriteString(lexeme)
			file_Output.WriteString("</span>\n")
		}
		tokens = append(tokens, state)
		lexeme = ""
		state = 0
	}
	defer file_Output.Close()
}

func reading_Files(f string, o string, sel int) {

	file := f // Lee el archivo de entrada
	fileName, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}

	output := o + `\` + "A01283516_"
	// toma el nombre del archivo (sin el path), y reemplaza .txt por .html, y establece su path como el folder determinado por o
	if sel == 1 {
		output += "secuencial_"
	} else {
		output += "paralelo_"
	}
	output += filepath.Base(strings.Replace(f, ".txt", ".html", -1))

	file_Output, err := os.Create(output) // Lee el archivo de salida
	if err != nil {
		fmt.Println(err)
	}
	// Escribe la fun cabeza
	const head_del_texto = (`
    <!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Lexer</title>
</head>
<body>
    <h3>
    <div align = "left">
	<style>
	h1{
        text-align:center;
    }
    </style>
	`)
	defer file_Output.Close()

	// Escribe lo principal
	file_Output.WriteString(head_del_texto)
	scanner := bufio.NewScanner(fileName)

	// Aqui va leyendo las lineas
	var datos []string
	for scanner.Scan() { // toma cada linea del archivo y la mete a datos
		datos = append(datos, scanner.Text())
	}
	defer fileName.Close()

	file_Output, err = os.OpenFile(output, os.O_APPEND|os.O_WRONLY, 0644) // Abre el archivo de salida
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range datos {

		scanning_the_text(line, output) // corre la funcion de escaneo en cada linea
		file_Output.WriteString("<br>")
	}
	file_Output.WriteString(`</div></h3></body></html>`) // al terminar de leer todas las lineas, cierra las divisiones para poder cerrar el archivo
}

func main() {
	// estructuras para el codigo
	var files []string
	var maxGo int = 10
	var wg sync.WaitGroup

	// lectura de los paths desde la consola
	root := os.Args[1]
	outpath := os.Args[2]

	pathSeq := outpath + `\` + "secuencial"
	pathPar := outpath + `\` + "paralelo"

	if _, err := os.Stat(pathSeq); errors.Is(err, os.ErrNotExist) { // crea la carpeta para secuencial si no existe
		err := os.Mkdir(pathSeq, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(pathPar); errors.Is(err, os.ErrNotExist) { // crea la carpeta para paralelo si no existe
		err := os.Mkdir(pathPar, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	//lee los archivos dentro de la carpeta root y pone sus directorios en files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Ejecución Secuencial
	startSeq := time.Now()

	for _, f := range files { // por cada archivo, version secuencial
		fi, err := os.Stat(f)
		if err != nil {
			fmt.Println(err)
		}
		if !os.FileInfo.IsDir(fi) {
			reading_Files(f, pathSeq, 1) // funcion a ejecutar

		}
	}

	elapsedSeq := time.Since(startSeq)
	parsedSeq := elapsedSeq.Seconds()
	parsedSeq_str := strconv.FormatFloat(parsedSeq, 'f', 5, 64)
	fmt.Println("Tiempo de ejecucion secuencial: " + parsedSeq_str + " seg")

	// Ejecución Paralela
	// abre una cantidad maxGo (maximo de gorutinas) de canales
	c1 := make(chan string, maxGo)

	start := time.Now()

	for _, f := range files { // por cada archivo, version paralela
		fi, err := os.Stat(f)
		if err != nil {
			fmt.Println(err)
		}
		if !os.FileInfo.IsDir(fi) {
			wg.Add(1)                       // agrega uno al contador del waitGroup
			c1 <- "work"                    // "ocupa" un canal, frena temporalmente el codigo si los 5 canales estan llenos
			go func(fl string, ou string) { // abre una gorutina
				defer wg.Done()          // cuando se termine el trabajo, resta uno del waitGroup
				reading_Files(fl, ou, 2) // funcion a ejecutar
				<-c1                     // vacía un canal
			}(f, pathPar) // ejecuta esa gorutina con f como valor de fl y outpath como valor de ou

		}

	}

	elapsed := time.Since(start)

	if err != nil {
		log.Fatal(err)
	}

	parsed := elapsed.Seconds()
	parsed_str := strconv.FormatFloat(parsed, 'f', 5, 64)
	fmt.Println("Tiempo de ejecucion paralela: " + parsed_str + " seg")

	wg.Wait() // el codigo no concluye hasta que el waitGroup esté vacío
}

/*
Para calcular la complejidad, podemos analizar la cantidad de ciclos que se tienen que recorrer en el codigo.
Similarmente a la actividad 3.4, los ciclos son "por cada archivo => por cada linea => por cada caracter"; es decir,
n => n => n. La diferencia es que en la 3.4 había un solo archivo por lo que la complejidad quedaba 1>n>n, que resultaba
en un O(n^2); al tener varios archivos este se convierte a n>n>n por lo que nos quedamos con O(n^3) en la ejecucion
secuencial. Sin embargo, al emplear paralelismo en el código la complejidad se divide por un numero de canales 'c'
(expresado en el código por la variable maxGo), por lo que quedaría como un O(n^3)/c. Como en este código la variable
 maxGo tiene un valor de 5, la complejidad terminaría siendo O(n^3)/5 en el caso de ejecución paralela.

 Tras ejecutarlo en múltiples ocasiones pude ver una diferencia de hasta 10 milisegundos entre la ejecución secuencial
 y la paralela, con 7 archivos cortos de texto y 5 ejecuciones paralelas. En casos en los que se trabajen con muchos
 gigabytes de informacion, asumiendo que se tuviera un procesador lo suficientemente capaz y una cantidad bastante
 superior de rutinas concurrentes, podría implicar un ahorro significativo en complejidad temporal del manejo de dichos
 archivos, lo que implicaría un ahorro de recursos tanto computacionales como energéticos.
*/
