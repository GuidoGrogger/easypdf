package main

import (
	"log"
	"os"

	_ "net/http/pprof"
)

func main() {
	/*	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/

	htmlDoc := `
	<p>Hallo Welt!</p>
	<p>Ich bin <b><u>easyPdf</u></b></p>
	<p>Projektierung von <i>professionellen</i> PDFs kann manchmal kompliziert sein, insbesondere wenn man besondere Anforderungen hat. Mit <b><u>easyPdf</u></b> wird dieser Prozess erheblich vereinfacht, indem <i>intuitive</i> Funktionen und Schnittstellen bereitgestellt werden. Dabei bleibt die <i>Flexibilität</i> erhalten, sodass man <i>individuelle Lösungen</i> für verschiedene Anforderungen erstellen kann.</p>
	<p align="block">Dieses <i>Tool</i> ermöglicht es, PDF-Dokumente mit Leichtigkeit zu erstellen, zu bearbeiten und zu konvertieren. Es bietet auch eine Reihe von Funktionen, um das <i>Layout</i> und <i>Design</i> der Dokumente nach Bedarf anzupassen. Egal, ob Sie Berichte, Rechnungen oder andere Dokumente erstellen möchten, <b><u>easyPdf</u></b> bietet Ihnen die notwendigen <i>Werkzeuge</i> dafür.</p>
	<p align="right">Der Benutzer kann <i>Schriftarten</i> und <i>-größen</i>, <i>Farbschemata</i> und andere Elemente mit wenigen Klicks anpassen. Außerdem bietet <b><u>easyPdf</u></b> verschiedene <i>Vorlagen</i>, um den Prozess der Dokumentenerstellung zu beschleunigen und zu vereinfachen. So kann jeder Benutzer, unabhängig von seinen technischen Kenntnissen, qualitativ hochwertige PDF-Dokumente erstellen.</p>
	<p>Table1</p>
	<table border="1">
				<tr>
					<th>Header 1</th>
					<th>Header 2</th>
				</tr>
				<tr>
					<td>Row 1, Cell 1</td>
					<td>Row 1, Cell 2 And a veryyyyyy longggg text .....</td>
					<td>Row 1, Cell 3</td>
					<td align="block">Row 1, Cell 4 And a veryyyyyy longggg text .....</td>
				</tr>
				<tr>
					<td>Row 2, Cell 1</td>
					<td align="right">Row 2, Cell 2 And a veryyyyyy longggg text .....</td>
				</tr>
			</table>
	<p> Table2 (nested)</p>
	<table border="1">
		<tr>
			<td>
				<table>
					<tr>
						<td>Inner 1</td>						
						<td>Inner 2</td>						
					</tr>
					<tr>
						<td>Inner 3</td>								
						<td>Inner 4</td>						
					</tr>
				</table>
			</td>
			<td>Row 1, Cell 2 And a veryyyyyy longggg text .....</td>
		</tr>
		<tr>
			<td>Row 2, Cell 1</td>
			<td align="right">Row 2, Cell 2 And a veryyyyyy longggg text .....</td>
		</tr>
	</table>
`

	/*	for {
		time.Sleep(1 * time.Second)
		createAndSavePDF(htmlDoc)
	}*/

	createAndSavePDF(htmlDoc)

}

func createAndSavePDF(htmlDoc string) {

	file, err := os.Create("output.pdf")
	if err != nil {
		log.Fatal(err)
	}

	err = CreatePDF(htmlDoc, file)
	if err != nil {
		log.Fatal(err)
	}
}
