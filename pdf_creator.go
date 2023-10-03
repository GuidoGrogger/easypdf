package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/html"
)

type Context struct {
	pdf        gofpdf.Pdf
	translator func(string) string
	// line buffer needed because items on a single line can be nested in html, like <b><i>...</i></b>
	lineBuffer       *[]LineItem
	lineBufferLength *float64
	fontStyle        string
	CurrentAreaX     float64
	CurrentAreaWidth float64
	Aligment         string
}

func CreatePDF(htmlDoc string, w io.Writer) error {
	doc, err := html.Parse(strings.NewReader(htmlDoc))
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	context := Context{
		pdf:              pdf,
		translator:       pdf.UnicodeTranslatorFromDescriptor(""),
		lineBufferLength: new(float64),
		lineBuffer:       new([]LineItem),
		fontStyle:        "",
		CurrentAreaX:     pdf.GetX(),
		CurrentAreaWidth: 180.0,
		Aligment:         "left",
	}
	context.CurrentAreaX = pdf.GetX()

	processNodeRecur(context, doc)

	return pdf.Output(w)
}

func processNodeRecur(ctx Context, n *html.Node) {

	if n.Type == html.TextNode {
		processTextNode(n, &ctx)
		return
	}

	if n.Type == html.ElementNode {
		processElement(n, &ctx)
		return
	}

	// for non element nodes
	processChildrenRecu(n, &ctx)
}

func processChildrenRecu(n *html.Node, ctx *Context) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNodeRecur(*ctx, c)
	}
}

func processElement(n *html.Node, ctx *Context) {
	switch n.Data {
	case "p":
		align := getAligment(n)
		ctx.Aligment = align
		processChildrenRecu(n, ctx)
		flushLineBuffer(ctx)
		ctx.pdf.Ln(3) // On top of normal line buffer heigth
		return
	case "b":
		ctx.fontStyle += "B"
	case "i":
		ctx.fontStyle += "I"
	case "u":
		ctx.fontStyle += "U"
	case "table":
	case "html":
	case "tbody":
		rows, cols := countTableRowsAndCols(n)
		tableWidth := ctx.CurrentAreaWidth
		tableX := ctx.CurrentAreaX
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
						ctx.CurrentAreaX = x
						ctx.CurrentAreaWidth = columWidth
						align := getAligment(td)
						ctx.Aligment = align
						ctx.pdf.SetY(rowY)
						fmt.Printf("Rendering Row: %d, Col: %d, X-pos=%f, Y-pos=%f, Width=%f\n", rowNum, colNum, x, rowY, columWidth)
						processChildrenRecu(td, ctx)
						flushLineBuffer(ctx)
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
	processChildrenRecu(n, ctx)

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

func processTextNode(n *html.Node, ctx *Context) {
	words := strings.Split(n.Data, " ")
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 0 {
			processSingleLineItem(ctx, word+" ")
		}
	}
}

func processSingleLineItem(ctx *Context, word string) {
	fontStyle := ctx.fontStyle

	ctx.pdf.SetFontStyle(fontStyle)
	wordWidth := ctx.pdf.GetStringWidth(ctx.translator(word))

	if (*ctx.lineBufferLength + wordWidth) > ctx.CurrentAreaWidth {
		flushLineBuffer(ctx)
	}

	*ctx.lineBuffer = append(*ctx.lineBuffer, TextLineItem{text: word, fontStyle: fontStyle})
	*ctx.lineBufferLength = *ctx.lineBufferLength + wordWidth
}

type TextLineItem struct {
	fontStyle string
	text      string
}

func (t TextLineItem) Render(ctx *Context) {
	ctx.pdf.SetFontStyle(t.fontStyle)
	wordWidth := ctx.pdf.GetStringWidth(ctx.translator(t.text))
	ctx.pdf.CellFormat(wordWidth, 10, ctx.translator(t.text), "0", 0, "", false, 0, "")
}

type LineItem interface {
	Render(*Context)
}

func flushLineBuffer(ctx *Context) {
	// to the line buffer?
	ctx.pdf.SetX(ctx.CurrentAreaX)
	gapToFill := ctx.CurrentAreaWidth - *ctx.lineBufferLength

	if ctx.Aligment == "right" {
		ctx.pdf.SetX(ctx.pdf.GetX() + gapToFill)
	}

	gapBetweenItems := 0.0
	if ctx.Aligment == "block" {
		itemsCount := len(*ctx.lineBuffer)
		gapBetweenItems = gapToFill / float64(itemsCount-1)
	}

	for _, value := range *ctx.lineBuffer {
		value.Render(ctx)
		ctx.pdf.SetX(ctx.pdf.GetX() + gapBetweenItems)
	}
	*ctx.lineBuffer = (*ctx.lineBuffer)[:0]
	*ctx.lineBufferLength = 0.0
	ctx.pdf.Ln(7) // TODO Caluclate based on Line heigth
}

func countTableRowsAndCols(tableNode *html.Node) (int, int) {
	var rows, cols int

	for child := tableNode.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "tr" {
			rows++
			var currentCols int
			for td := child.FirstChild; td != nil; td = td.NextSibling {
				if td.Data == "td" || td.Data == "th" {
					currentCols++
				}
			}
			if currentCols > cols {
				cols = currentCols
			}
		}
	}

	return rows, cols
}
