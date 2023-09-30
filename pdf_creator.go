package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/html"
)

type TextLineItem struct {
	fontStyle string
	text      string
}

func (t TextLineItem) Render(ctx *Context) {
	//fmt.Print(t.text)
	ctx.pdf.SetFontStyle(t.fontStyle)
	wordWidth := ctx.pdf.GetStringWidth(ctx.translator(t.text))
	ctx.pdf.CellFormat(wordWidth, 10, ctx.translator(t.text), "0", 0, "", false, 0, "")
}

type LineItem interface {
	Render(*Context)
}

func flushLineBuffer(ctx *Context, attrMap *NestedMap) {

	//fmt.Println("linebuffer flush begin ", ctx.lineBufferLength)

	//fmt.Println("algimnetn ", attrMap.Fetch("align"))

	// TODO HAndling end of Paragraph
	// todo, kann das in den line buffer?
	ctx.pdf.SetX(attrMap.Fetch("CURRENT_AREA_X").(float64))
	gapToFill := attrMap.Fetch("CURRENT_AREA_WIDTH").(float64) - ctx.lineBufferLength

	if attrMap.Fetch("align") == "right" {
		ctx.pdf.SetX(ctx.pdf.GetX() + gapToFill)
	}

	gapBetweenItems := 0.0
	if attrMap.Fetch("align") == "block" {
		itemsCount := len(ctx.lineBuffer)
		gapBetweenItems = gapToFill / float64(itemsCount-1)
	}

	for _, value := range ctx.lineBuffer {
		value.Render(ctx)
		ctx.pdf.SetX(ctx.pdf.GetX() + gapBetweenItems)
	}
	ctx.lineBuffer = ctx.lineBuffer[:0]
	ctx.lineBufferLength = 0.0
	ctx.pdf.Ln(7) // TODO Caluclate based on Line heigth
}

type Context struct {
	pdf        gofpdf.Pdf
	translator func(string) string
	// line buffer needed because items on a single line can be nested in html, like <b><i>...</i></b>
	lineBuffer       []LineItem
	lineBufferLength float64
}

func CreatePDF(htmlDoc string, w io.Writer) error {
	doc, err := html.Parse(strings.NewReader(htmlDoc))
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// better pointer?
	context := Context{
		pdf:              pdf,
		translator:       pdf.UnicodeTranslatorFromDescriptor(""),
		lineBufferLength: 0,
	}
	topAttrMap := NewNestedMap(nil)
	topAttrMap.Set("CURRENT_AREA_X", pdf.GetX())
	topAttrMap.Set("CURRENT_AREA_WIDTH", 180.0)

	processNodeRecur(&context, doc, topAttrMap)

	return pdf.Output(w)
}

func processNodeRecur(ctx *Context, n *html.Node, parentAttrMap *NestedMap) {
	attrMap := NewNestedMap(parentAttrMap)

	if n.Type == html.TextNode {
		processTextNode(attrMap, n, ctx)
		return
	}

	if n.Type == html.ElementNode {
		processElement(n, attrMap, ctx)
		return
	}

	// for non element nodes
	processChdilrenRecu(n, ctx, attrMap)
}

func processChdilrenRecu(n *html.Node, ctx *Context, attrMap *NestedMap) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNodeRecur(ctx, c, attrMap)
	}
}

func processElement(n *html.Node, attrMap *NestedMap, ctx *Context) {
	switch n.Data {
	case "p":
		align := getAligment(n)
		attrMap.Set("align", align)
		processChdilrenRecu(n, ctx, attrMap)
		flushLineBuffer(ctx, attrMap)
		ctx.pdf.Ln(3) // On top of normal line buffer heigth
		return
	case "b":
		attrMap.Set("<B>", "1")
	case "i":
		attrMap.Set("<I>", "1")
	case "u":
		attrMap.Set("<U>", "1")
	case "table":
	case "html":
	case "tbody":
		rows, cols := countTableRowsAndCols(n)
		tableWidth := attrMap.Fetch("CURRENT_AREA_WIDTH").(float64)
		tableX := attrMap.Fetch("CURRENT_AREA_X").(float64)
		tableY := ctx.pdf.GetY()
		columWidth := tableWidth / (float64)(cols)
		fmt.Printf("Table detected: Rows=%d, Cols=%d, X-pos=%f, Y-pos=%f, Width=%f, Column Width=%f\n", rows, cols, tableX, tableY, tableWidth, columWidth)

		// wft, its amazing: https://chat.openai.com/share/b8fdc0c7-9de5-4eca-a3b4-fcbca8017459
		rowNum := 0
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Data == "tr" {
				rowNum++
				rowHeight := 0.0
				rowY := ctx.pdf.GetY()
				colNum := 0
				for td := child.FirstChild; td != nil; td = td.NextSibling {
					if td.Data == "td" || td.Data == "th" {
						colNum++
						x := tableX + columWidth*float64(colNum-1)
						attrMap.Set("CURRENT_AREA_X", x)
						attrMap.Set("CURRENT_AREA_WIDTH", columWidth)
						align := getAligment(td)
						attrMap.Set("align", align)
						ctx.pdf.SetY(rowY)
						fmt.Printf("Rendering Row: %d, Col: %d, X-pos=%f, Y-pos=%f, Width=%f\n", rowNum, colNum, x, rowY, columWidth)
						processChdilrenRecu(td, ctx, attrMap)
						flushLineBuffer(ctx, attrMap)
						colHeight := ctx.pdf.GetY() - rowY
						if colHeight > rowHeight {
							rowHeight = colHeight
						}
					}
				}
				// Drawing the rectangle for each cell in gofpdf to span the full height of the row
				for i := 1; i <= colNum; i++ {
					x := tableX + columWidth*float64(i-1)
					ctx.pdf.Rect(x, rowY, columWidth, rowHeight, "D")
				}
				ctx.pdf.SetY(rowY + rowHeight)
			}
		}

		return
	case "head":
	case "body":
	default:
		log.Fatal("Unbekanntes Element <" + n.Data + "> gefunden")
	}
	processChdilrenRecu(n, ctx, attrMap)

}

func getAligment(n *html.Node) string {
	align := "left"
	for _, attr := range n.Attr {
		if attr.Key == "align" {
			align = attr.Val
			break
		}
	}
	return align
}

func processTextNode(attrMap *NestedMap, n *html.Node, ctx *Context) {
	words := strings.Split(n.Data, " ")
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 0 {

			processSingleLineItem(attrMap, ctx, word+" ")
		}
	}
}

func processSingleLineItem(attrMap *NestedMap, ctx *Context, word string) {
	fontStyle := getCurrentFontstyle(attrMap)

	ctx.pdf.SetFontStyle(fontStyle)
	wordWidth := ctx.pdf.GetStringWidth(ctx.translator(word))

	if (ctx.lineBufferLength + wordWidth) > attrMap.Fetch("CURRENT_AREA_WIDTH").(float64) {
		//fmt.Println("flushing linebuffer because of space", ctx.lineBufferLength, wordWidth)
		flushLineBuffer(ctx, attrMap)
	}

	// TODO performance?
	//	fmt.Println("appending to linebuffer ", word, ctx.lineBufferLength, wordWidth)
	ctx.lineBuffer = append(ctx.lineBuffer, TextLineItem{text: word, fontStyle: fontStyle})
	ctx.lineBufferLength = ctx.lineBufferLength + wordWidth
	//fmt.Println("new linebuffer length  ", ctx.lineBufferLength)

}
func getCurrentFontstyle(attrMap *NestedMap) string {
	fontStyle := ""
	if attrMap.IsSet("<B>") {
		fontStyle += "B"
	}
	if attrMap.IsSet("<I>") {
		fontStyle += "I"
	}
	if attrMap.IsSet("<U>") {
		fontStyle += "U"
	}
	return fontStyle
}

type NestedMap struct {
	Data   map[string]interface{}
	Parent *NestedMap
}

func NewNestedMap(parent *NestedMap) *NestedMap {
	return &NestedMap{
		Data:   make(map[string]interface{}),
		Parent: parent,
	}
}

func (c *NestedMap) Fetch(key string) interface{} {
	value := c.Data[key]

	if value != nil {
		return value
	}

	if c.Parent != nil {
		return c.Parent.Fetch(key)
	}

	return nil
}
func (c *NestedMap) IsSet(key string) bool {
	return c.Fetch(key) != nil
}

func (c *NestedMap) Set(key string, value interface{}) {
	c.Data[key] = value
}

// countTableRowsAndCols nimmt einen HTML-Node, der eine <table> darstellt,
// und gibt die Anzahl der Zeilen und Spalten zurück.
func countTableRowsAndCols(tableNode *html.Node) (int, int) {
	var rows, cols int

	// Durchlaufe alle Kinder des Tabelle-Knotens
	for child := tableNode.FirstChild; child != nil; child = child.NextSibling {
		// Wenn der Kindknoten ein <tr> ist
		if child.Data == "tr" {
			rows++
			var currentCols int
			// Durchlaufe alle Kinder des <tr>-Knotens und zähle <td> oder <th>-Elemente
			for td := child.FirstChild; td != nil; td = td.NextSibling {
				if td.Data == "td" || td.Data == "th" {
					currentCols++
				}
			}
			// Wenn currentCols größer als cols ist, aktualisiere cols
			if currentCols > cols {
				cols = currentCols
			}
		}
	}

	return rows, cols
}
