package main

import (
	"bytes"
	"os"
	"testing"
)

var htmlDoc = `
<p>Hallo Welt!</p>
<p>Ich bin <b><u>easyPdf</u></b></p>
<p>Projektierung von <i>professionellen</i> PDFs kann manchmal kompliziert sein, insbesondere wenn man besondere Anforderungen hat. Mit <b><u>easyPdf</u></b> wird dieser Prozess erheblich vereinfacht, indem <i>intuitive</i> Funktionen und Schnittstellen bereitgestellt werden. Dabei bleibt die <i>Flexibilität</i> erhalten, sodass man <i>individuelle Lösungen</i> für verschiedene Anforderungen erstellen kann.</p>
<p align="center">Dieses <i>Tool</i> ermöglicht es, PDF-Dokumente mit Leichtigkeit zu erstellen, zu bearbeiten und zu konvertieren. Es bietet auch eine Reihe von Funktionen, um das <i>Layout</i> und <i>Design</i> der Dokumente nach Bedarf anzupassen. Egal, ob Sie Berichte, Rechnungen oder andere Dokumente erstellen möchten, <b><u>easyPdf</u></b> bietet Ihnen die notwendigen <i>Werkzeuge</i> dafür.</p>
<p align="right">Der Benutzer kann <i>Schriftarten</i> und <i>-größen</i>, <i>Farbschemata</i> und andere Elemente mit wenigen Klicks anpassen. Außerdem bietet <b><u>easyPdf</u></b> verschiedene <i>Vorlagen</i>, um den Prozess der Dokumentenerstellung zu beschleunigen und zu vereinfachen. So kann jeder Benutzer, unabhängig von seinen technischen Kenntnissen, qualitativ hochwertige PDF-Dokumente erstellen.</p>
`

func TestCreatePDF(t *testing.T) {

	var buffer bytes.Buffer
	CreatePDF(htmlDoc, &buffer)

	err := os.WriteFile("test_result.pdf", buffer.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write test_result.pdf: %v", err)
	}

	/*referencePDF, err := os.ReadFile("test_reference.pdf")

	if err != nil {
		t.Fatalf("Failed to load reference PDF: %v", err)
	}

	if !bytes.Equal(buffer.Bytes(), referencePDF) {
		t.Fatalf("Generated PDF does not match the reference PDF")

	}*/

}

func BenchmarkCreatePDF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		CreatePDF(htmlDoc, &buffer)
	}
}

func BenchmarkCreatePDF_LargeInput(b *testing.B) {
	// Duplicate htmlDoc to make it larger
	bigHtmlDoc := htmlDoc + htmlDoc
	bigHtmlDoc = bigHtmlDoc + bigHtmlDoc
	bigHtmlDoc = bigHtmlDoc + bigHtmlDoc

	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		CreatePDF(bigHtmlDoc, &buffer)
	}
}
