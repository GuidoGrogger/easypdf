# easyPdf

**easyPdf** ist ein Lernprojekt, das zeigt, wie man ein HTML-zu-PDF-Tool in reinem Go erstellt. Es ist ein Projekt zum Spaß.

## Unterstützte HTML-Elemente und Attribute

- `<p>`
  - `align` (Werte: `block`, `right`)
- `<b>`
- `<u>`
- `<i>`
- `<table>`
  - `border`
- `<tr>`
- `<th>`
- `<td>`
  - `align` (Werte: `block`, `right`)

## Ausführung

Um das Tool auszuführen:

```bash
go run .
```

Das Ergebnis wird in der Datei `output.pdf` gespeichert.

## Lizenz

Dieses Projekt steht unter der MIT-Lizenz. Details finden Sie in der LICENSE-Datei.