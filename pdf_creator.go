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
	Pdf        gofpdf.Pdf
	Translator func(string) string
	// line buffer needed because items on a single line can be nested in html, like <b><i>...</i></b>
	LineBuffer         []LineItem
	LineBufferWidth    float64
	ElementNodeContext ElementNodeContext
}

type ElementNodeContext struct {
	FontStyle string
	Aligment  string
	X         float64
	Width     float64
}

type LineItem interface {
	Render(*Context)
}

type TextLineItem struct {
	fontStyle string
	text      string
}

func CreatePDF(htmlDoc string, w io.Writer) error {
	doc, err := html.Parse(strings.NewReader(htmlDoc))
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	nodeCtx := ElementNodeContext{
		FontStyle: "",
		Aligment:  "left",
		X:         pdf.GetX(),
		Width:     180.0,
	}

	context := Context{
		Pdf:                pdf,
		Translator:         pdf.UnicodeTranslatorFromDescriptor(""),
		LineBufferWidth:    0.0,
		LineBuffer:         []LineItem{},
		ElementNodeContext: nodeCtx,
	}

	processNodeRecur(&context, doc)

	return pdf.Output(w)
}

func processNodeRecur(ctx *Context, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		processTextNode(n, ctx)

	case html.ElementNode:
		nodeCtxBefore := ctx.ElementNodeContext
		processElement(n, ctx)
		ctx.ElementNodeContext = nodeCtxBefore

	default:
		processChildrenRecu(n, ctx)
	}
}

func processElement(n *html.Node, ctx *Context) {
	switch n.Data {
	case "p":
		align := getAligment(n)
		ctx.ElementNodeContext.Aligment = align
		processChildrenRecu(n, ctx)
		flushLineBuffer(ctx)
		ctx.Pdf.Ln(3) // On top of normal line buffer heigth
		return
	case "b":
		ctx.ElementNodeContext.FontStyle += "B"
	case "i":
		ctx.ElementNodeContext.FontStyle += "I"
	case "u":
		ctx.ElementNodeContext.FontStyle += "U"
	case "table":
	case "html":
	case "tbody":
		rows, cols := countTableRowsAndCols(n)
		tableWidth := ctx.ElementNodeContext.Width
		tableX := ctx.ElementNodeContext.X
		tableY := ctx.Pdf.GetY()
		columWidth := tableWidth / (float64)(cols)
		fmt.Printf("Table detected: Rows=%d, Cols=%d, X-pos=%f, Y-pos=%f, Width=%f, Column Width=%f\n", rows, cols, tableX, tableY, tableWidth, columWidth)

		// its amazing: https://chat.openai.com/share/b8fdc0c7-9de5-4eca-a3b4-fcbca8017459
		rowNum := 0
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			// Drawing the rectangle for each cell in gofpdf to span the full height of the row
			processTableRow(child, rowNum, ctx, tableX, columWidth)
		}
		return
	case "head":
	case "body":
	default:
		log.Fatal("Unbekanntes Element <" + n.Data + "> gefunden")
	}
	processChildrenRecu(n, ctx)

}

func processTableRow(child *html.Node, rowNum int, ctx *Context, tableX float64, columWidth float64) {
	if child.Data == "tr" {
		rowNum++
		maxRowHeight := 0.0
		rowY := ctx.Pdf.GetY()
		colNum := 0
		for td := child.FirstChild; td != nil; td = td.NextSibling {
			if td.Data == "td" || td.Data == "th" {
				colNum++
				x := tableX + columWidth*float64(colNum-1)
				ctx.ElementNodeContext.X = x
				ctx.ElementNodeContext.Width = columWidth
				ctx.ElementNodeContext.Aligment = getAligment(td)
				ctx.Pdf.SetY(rowY)
				fmt.Printf("Rendering Row: %d, Col: %d, X-pos=%f, Y-pos=%f, Width=%f\n", rowNum, colNum, x, rowY, columWidth)
				processChildrenRecu(td, ctx)
				flushLineBuffer(ctx)
				colHeight := ctx.Pdf.GetY() - rowY
				if colHeight > maxRowHeight {
					maxRowHeight = colHeight
				}
			}
		}

		// draw cell borders
		for i := 1; i <= colNum; i++ {
			x := tableX + columWidth*float64(i-1)
			ctx.Pdf.Rect(x, rowY, columWidth, maxRowHeight, "D")
		}
		ctx.Pdf.SetY(rowY + maxRowHeight)
	}
}

func processTextNode(n *html.Node, ctx *Context) {
	words := strings.Split(n.Data, " ")
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 0 {
			processSingleWord(ctx, word+" ")
		}
	}
}

func processChildrenRecu(n *html.Node, ctx *Context) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNodeRecur(ctx, c)
	}
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

func processSingleWord(ctx *Context, word string) {
	fontStyle := ctx.ElementNodeContext.FontStyle

	ctx.Pdf.SetFontStyle(fontStyle)
	wordWidth := ctx.Pdf.GetStringWidth(ctx.Translator(word))

	if (ctx.LineBufferWidth + wordWidth) > ctx.ElementNodeContext.Width {
		flushLineBuffer(ctx)
	}

	ctx.LineBuffer = append(ctx.LineBuffer, TextLineItem{text: word, fontStyle: fontStyle})
	ctx.LineBufferWidth = ctx.LineBufferWidth + wordWidth
}

func (t TextLineItem) Render(ctx *Context) {
	ctx.Pdf.SetFontStyle(t.fontStyle)
	wordWidth := ctx.Pdf.GetStringWidth(ctx.Translator(t.text))
	ctx.Pdf.CellFormat(wordWidth, 10, ctx.Translator(t.text), "0", 0, "", false, 0, "")
}

func flushLineBuffer(ctx *Context) {
	// to the line buffer?
	ctx.Pdf.SetX(ctx.ElementNodeContext.X)
	gapToFill := ctx.ElementNodeContext.Width - ctx.LineBufferWidth

	if ctx.ElementNodeContext.Aligment == "right" {
		ctx.Pdf.SetX(ctx.Pdf.GetX() + gapToFill)
	}

	gapBetweenItems := 0.0
	if ctx.ElementNodeContext.Aligment == "block" {
		itemsCount := len(ctx.LineBuffer)
		gapBetweenItems = gapToFill / float64(itemsCount-1)
	}

	for _, value := range ctx.LineBuffer {
		value.Render(ctx)
		ctx.Pdf.SetX(ctx.Pdf.GetX() + gapBetweenItems)
	}
	ctx.LineBuffer = (ctx.LineBuffer)[:0]
	ctx.LineBufferWidth = 0.0
	ctx.Pdf.Ln(7) // TODO Caluclate based on Line heigth
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
