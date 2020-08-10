// Copyright 2016 - 2020 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to
// and read from XLSX files. Support reads and writes XLSX file generated by
// Microsoft Excel™ 2007 and later. Support save file without losing original
// charts of XLSX. This library needs Go version 1.10 or later.

package excelize

import (
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// STCellFormulaTypeArray defined the formula is an array formula.
	STCellFormulaTypeArray = "array"
	// STCellFormulaTypeDataTable defined the formula is a data table formula.
	STCellFormulaTypeDataTable = "dataTable"
	// STCellFormulaTypeNormal defined the formula is a regular cell formula.
	STCellFormulaTypeNormal = "normal"
	// STCellFormulaTypeShared defined the formula is part of a shared formula.
	STCellFormulaTypeShared = "shared"
)

// GetCellValue provides a function to get formatted value from cell by given
// worksheet name and axis in XLSX file. If it is possible to apply a format
// to the cell value, it will do so, if not then an error will be returned,
// along with the raw value of the cell.
func (f *File) GetCellValue(sheet, axis string) (string, error) {
	return f.getCellStringFunc(sheet, axis, func(x *xlsxWorksheet, c *xlsxC) (string, bool, error) {
		val, err := c.getValueFrom(f, f.sharedStringsReader())
		if err != nil {
			return val, false, err
		}
		return val, true, err
	})
}

// SetCellValue provides a function to set value of a cell. The specified
// coordinates should not be in the first row of the table. The following
// shows the supported data types:
//
//    int
//    int8
//    int16
//    int32
//    int64
//    uint
//    uint8
//    uint16
//    uint32
//    uint64
//    float32
//    float64
//    string
//    []byte
//    time.Duration
//    time.Time
//    bool
//    nil
//
// Note that default date format is m/d/yy h:mm of time.Time type value. You can
// set numbers format by SetCellStyle() method.
func (f *File) SetCellValue(sheet, axis string, value interface{}) error {
	var err error
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		err = f.setCellIntFunc(sheet, axis, v)
	case float32:
		err = f.SetCellFloat(sheet, axis, float64(v), -1, 32)
	case float64:
		err = f.SetCellFloat(sheet, axis, v, -1, 64)
	case string:
		err = f.SetCellStr(sheet, axis, v)
	case []byte:
		err = f.SetCellStr(sheet, axis, string(v))
	case time.Duration:
		_, d := setCellDuration(v)
		err = f.SetCellDefault(sheet, axis, d)
		if err != nil {
			return err
		}
		err = f.setDefaultTimeStyle(sheet, axis, 21)
	case time.Time:
		err = f.setCellTimeFunc(sheet, axis, v)
	case bool:
		err = f.SetCellBool(sheet, axis, v)
	case nil:
		err = f.SetCellStr(sheet, axis, "")
	default:
		err = f.SetCellStr(sheet, axis, fmt.Sprint(value))
	}
	return err
}

// setCellIntFunc is a wrapper of SetCellInt.
func (f *File) setCellIntFunc(sheet, axis string, value interface{}) error {
	var err error
	switch v := value.(type) {
	case int:
		err = f.SetCellInt(sheet, axis, v)
	case int8:
		err = f.SetCellInt(sheet, axis, int(v))
	case int16:
		err = f.SetCellInt(sheet, axis, int(v))
	case int32:
		err = f.SetCellInt(sheet, axis, int(v))
	case int64:
		err = f.SetCellInt(sheet, axis, int(v))
	case uint:
		err = f.SetCellInt(sheet, axis, int(v))
	case uint8:
		err = f.SetCellInt(sheet, axis, int(v))
	case uint16:
		err = f.SetCellInt(sheet, axis, int(v))
	case uint32:
		err = f.SetCellInt(sheet, axis, int(v))
	case uint64:
		err = f.SetCellInt(sheet, axis, int(v))
	}
	return err
}

// setCellTimeFunc provides a method to process time type of value for
// SetCellValue.
func (f *File) setCellTimeFunc(sheet, axis string, value time.Time) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)

	var isNum bool
	cellData.T, cellData.V, isNum, err = setCellTime(value)
	if err != nil {
		return err
	}
	if isNum {
		err = f.setDefaultTimeStyle(sheet, axis, 22)
		if err != nil {
			return err
		}
	}
	return err
}

func setCellTime(value time.Time) (t string, b string, isNum bool, err error) {
	var excelTime float64
	excelTime, err = timeToExcelTime(value)
	if err != nil {
		return
	}
	isNum = excelTime > 0
	if isNum {
		t, b = setCellDefault(strconv.FormatFloat(excelTime, 'f', -1, 64))
	} else {
		t, b = setCellDefault(value.Format(time.RFC3339Nano))
	}
	return
}

func setCellDuration(value time.Duration) (t string, v string) {
	v = strconv.FormatFloat(value.Seconds()/86400.0, 'f', -1, 32)
	return
}

// SetCellInt provides a function to set int type value of a cell by given
// worksheet name, cell coordinates and cell value.
func (f *File) SetCellInt(sheet, axis string, value int) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)
	cellData.T, cellData.V = setCellInt(value)
	return err
}

func setCellInt(value int) (t string, v string) {
	v = strconv.Itoa(value)
	return
}

// SetCellBool provides a function to set bool type value of a cell by given
// worksheet name, cell name and cell value.
func (f *File) SetCellBool(sheet, axis string, value bool) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)
	cellData.T, cellData.V = setCellBool(value)
	return err
}

func setCellBool(value bool) (t string, v string) {
	t = "b"
	if value {
		v = "1"
	} else {
		v = "0"
	}
	return
}

// SetCellFloat sets a floating point value into a cell. The prec parameter
// specifies how many places after the decimal will be shown while -1 is a
// special value that will use as many decimal places as necessary to
// represent the number. bitSize is 32 or 64 depending on if a float32 or
// float64 was originally used for the value. For Example:
//
//    var x float32 = 1.325
//    f.SetCellFloat("Sheet1", "A1", float64(x), 2, 32)
//
func (f *File) SetCellFloat(sheet, axis string, value float64, prec, bitSize int) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)
	cellData.T, cellData.V = setCellFloat(value, prec, bitSize)
	return err
}

func setCellFloat(value float64, prec, bitSize int) (t string, v string) {
	v = strconv.FormatFloat(value, 'f', prec, bitSize)
	return
}

// SetCellStr provides a function to set string type value of a cell. Total
// number of characters that a cell can contain 32767 characters.
func (f *File) SetCellStr(sheet, axis, value string) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)
	cellData.T, cellData.V, cellData.XMLSpace = setCellStr(value)
	return err
}

func setCellStr(value string) (t string, v string, ns xml.Attr) {
	if len(value) > 32767 {
		value = value[0:32767]
	}
	// Leading and ending space(s) character detection.
	if len(value) > 0 && (value[0] == 32 || value[len(value)-1] == 32) {
		ns = xml.Attr{
			Name:  xml.Name{Space: NameSpaceXML, Local: "space"},
			Value: "preserve",
		}
	}
	t = "str"
	v = value
	return
}

// SetCellDefault provides a function to set string type value of a cell as
// default format without escaping the cell.
func (f *File) SetCellDefault(sheet, axis, value string) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(xlsx, col, cellData.S)
	cellData.T, cellData.V = setCellDefault(value)
	return err
}

func setCellDefault(value string) (t string, v string) {
	v = value
	return
}

// GetCellFormula provides a function to get formula from cell by given
// worksheet name and axis in XLSX file.
func (f *File) GetCellFormula(sheet, axis string) (string, error) {
	return f.getCellStringFunc(sheet, axis, func(x *xlsxWorksheet, c *xlsxC) (string, bool, error) {
		if c.F == nil {
			return "", false, nil
		}
		if c.F.T == STCellFormulaTypeShared {
			return getSharedForumula(x, c.F.Si), true, nil
		}
		return c.F.Content, true, nil
	})
}

// FormulaOpts can be passed to SetCellFormula to use other formula types.
type FormulaOpts struct {
	Type *string // Formula type
	Ref  *string // Shared formula ref
}

// SetCellFormula provides a function to set cell formula by given string and
// worksheet name.
func (f *File) SetCellFormula(sheet, axis, formula string, opts ...FormulaOpts) error {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, _, _, err := f.prepareCell(xlsx, sheet, axis)
	if err != nil {
		return err
	}
	if formula == "" {
		cellData.F = nil
		f.deleteCalcChain(f.GetSheetIndex(sheet), axis)
		return err
	}

	if cellData.F != nil {
		cellData.F.Content = formula
	} else {
		cellData.F = &xlsxF{Content: formula}
	}

	for _, o := range opts {
		if o.Type != nil {
			cellData.F.T = *o.Type
		}

		if o.Ref != nil {
			cellData.F.Ref = *o.Ref
		}
	}

	return err
}

// GetCellHyperLink provides a function to get cell hyperlink by given
// worksheet name and axis. Boolean type value link will be ture if the cell
// has a hyperlink and the target is the address of the hyperlink. Otherwise,
// the value of link will be false and the value of the target will be a blank
// string. For example get hyperlink of Sheet1!H6:
//
//    link, target, err := f.GetCellHyperLink("Sheet1", "H6")
//
func (f *File) GetCellHyperLink(sheet, axis string) (bool, string, error) {
	// Check for correct cell name
	if _, _, err := SplitCellName(axis); err != nil {
		return false, "", err
	}

	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return false, "", err
	}
	axis, err = f.mergeCellsParser(xlsx, axis)
	if err != nil {
		return false, "", err
	}
	if xlsx.Hyperlinks != nil {
		for _, link := range xlsx.Hyperlinks.Hyperlink {
			if link.Ref == axis {
				if link.RID != "" {
					return true, f.getSheetRelationshipsTargetByID(sheet, link.RID), err
				}
				return true, link.Location, err
			}
		}
	}
	return false, "", err
}

// SetCellHyperLink provides a function to set cell hyperlink by given
// worksheet name and link URL address. LinkType defines two types of
// hyperlink "External" for web site or "Location" for moving to one of cell
// in this workbook. Maximum limit hyperlinks in a worksheet is 65530. The
// below is example for external link.
//
//    err := f.SetCellHyperLink("Sheet1", "A3", "https://github.com/360EntSecGroup-Skylar/excelize", "External")
//    // Set underline and font color style for the cell.
//    style, err := f.NewStyle(`{"font":{"color":"#1265BE","underline":"single"}}`)
//    err = f.SetCellStyle("Sheet1", "A3", "A3", style)
//
// A this is another example for "Location":
//
//    err := f.SetCellHyperLink("Sheet1", "A3", "Sheet1!A40", "Location")
//
func (f *File) SetCellHyperLink(sheet, axis, link, linkType string) error {
	// Check for correct cell name
	if _, _, err := SplitCellName(axis); err != nil {
		return err
	}

	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	axis, err = f.mergeCellsParser(xlsx, axis)
	if err != nil {
		return err
	}

	var linkData xlsxHyperlink

	if xlsx.Hyperlinks == nil {
		xlsx.Hyperlinks = new(xlsxHyperlinks)
	}

	if len(xlsx.Hyperlinks.Hyperlink) > 65529 {
		return errors.New("over maximum limit hyperlinks in a worksheet")
	}

	switch linkType {
	case "External":
		linkData = xlsxHyperlink{
			Ref: axis,
		}
		sheetPath := f.sheetMap[trimSheetName(sheet)]
		sheetRels := "xl/worksheets/_rels/" + strings.TrimPrefix(sheetPath, "xl/worksheets/") + ".rels"
		rID := f.addRels(sheetRels, SourceRelationshipHyperLink, link, linkType)
		linkData.RID = "rId" + strconv.Itoa(rID)
	case "Location":
		linkData = xlsxHyperlink{
			Ref:      axis,
			Location: link,
		}
	default:
		return fmt.Errorf("invalid link type %q", linkType)
	}

	xlsx.Hyperlinks.Hyperlink = append(xlsx.Hyperlinks.Hyperlink, linkData)
	return nil
}

// SetCellRichText provides a function to set cell with rich text by given
// worksheet. For example, set rich text on the A1 cell of the worksheet named
// Sheet1:
//
//    package main
//
//    import (
//        "fmt"
//
//        "github.com/360EntSecGroup-Skylar/excelize"
//    )
//
//    func main() {
//        f := excelize.NewFile()
//        if err := f.SetRowHeight("Sheet1", 1, 35); err != nil {
//            fmt.Println(err)
//            return
//        }
//        if err := f.SetColWidth("Sheet1", "A", "A", 44); err != nil {
//            fmt.Println(err)
//            return
//        }
//        if err := f.SetCellRichText("Sheet1", "A1", []excelize.RichTextRun{
//            {
//                Text: "blod",
//                Font: &excelize.Font{
//                    Bold:   true,
//                    Color:  "2354e8",
//                    Family: "Times New Roman",
//                },
//            },
//            {
//                Text: " and ",
//                Font: &excelize.Font{
//                    Family: "Times New Roman",
//                },
//            },
//            {
//                Text: " italic",
//                Font: &excelize.Font{
//                    Bold:   true,
//                    Color:  "e83723",
//                    Italic: true,
//                    Family: "Times New Roman",
//                },
//            },
//            {
//                Text: "text with color and font-family,",
//                Font: &excelize.Font{
//                    Bold:   true,
//                    Color:  "2354e8",
//                    Family: "Times New Roman",
//                },
//            },
//            {
//                Text: "\r\nlarge text with ",
//                Font: &excelize.Font{
//                    Size:  14,
//                    Color: "ad23e8",
//                },
//            },
//            {
//                Text: "strike",
//                Font: &excelize.Font{
//                    Color:  "e89923",
//                    Strike: true,
//                },
//            },
//            {
//                Text: " and ",
//                Font: &excelize.Font{
//                    Size:  14,
//                    Color: "ad23e8",
//                },
//            },
//            {
//                Text: "underline.",
//                Font: &excelize.Font{
//                    Color:     "23e833",
//                    Underline: "single",
//                },
//            },
//        }); err != nil {
//            fmt.Println(err)
//            return
//        }
//        style, err := f.NewStyle(&excelize.Style{
//            Alignment: &excelize.Alignment{
//                WrapText: true,
//            },
//        })
//        if err != nil {
//            fmt.Println(err)
//            return
//        }
//        if err := f.SetCellStyle("Sheet1", "A1", "A1", style); err != nil {
//            fmt.Println(err)
//            return
//        }
//        if err := f.SaveAs("Book1.xlsx"); err != nil {
//            fmt.Println(err)
//        }
//    }
//
func (f *File) SetCellRichText(sheet, cell string, runs []RichTextRun) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	cellData, col, _, err := f.prepareCell(ws, sheet, cell)
	if err != nil {
		return err
	}
	cellData.S = f.prepareCellStyle(ws, col, cellData.S)
	si := xlsxSI{}
	sst := f.sharedStringsReader()
	textRuns := []xlsxR{}
	for _, textRun := range runs {
		run := xlsxR{T: &xlsxT{Val: textRun.Text}}
		if strings.ContainsAny(textRun.Text, "\r\n ") {
			run.T.Space = "preserve"
		}
		fnt := textRun.Font
		if fnt != nil {
			rpr := xlsxRPr{}
			if fnt.Bold {
				rpr.B = " "
			}
			if fnt.Italic {
				rpr.I = " "
			}
			if fnt.Strike {
				rpr.Strike = " "
			}
			if fnt.Underline != "" {
				rpr.U = &attrValString{Val: &fnt.Underline}
			}
			if fnt.Family != "" {
				rpr.RFont = &attrValString{Val: &fnt.Family}
			}
			if fnt.Size > 0.0 {
				rpr.Sz = &attrValFloat{Val: &fnt.Size}
			}
			if fnt.Color != "" {
				rpr.Color = &xlsxColor{RGB: getPaletteColor(fnt.Color)}
			}
			run.RPr = &rpr
		}
		textRuns = append(textRuns, run)
	}
	si.R = textRuns
	sst.SI = append(sst.SI, si)
	sst.Count++
	sst.UniqueCount++
	cellData.T, cellData.V = "s", strconv.Itoa(len(sst.SI)-1)
	f.addContentTypePart(0, "sharedStrings")
	rels := f.relsReader("xl/_rels/workbook.xml.rels")
	for _, rel := range rels.Relationships {
		if rel.Target == "sharedStrings.xml" {
			return err
		}
	}
	// Update xl/_rels/workbook.xml.rels
	f.addRels("xl/_rels/workbook.xml.rels", SourceRelationshipSharedStrings, "sharedStrings.xml", "")
	return err
}

// SetSheetRow writes an array to row by given worksheet name, starting
// coordinate and a pointer to array type 'slice'. For example, writes an
// array to row 6 start with the cell B6 on Sheet1:
//
//     err := f.SetSheetRow("Sheet1", "B6", &[]interface{}{"1", nil, 2})
//
func (f *File) SetSheetRow(sheet, axis string, slice interface{}) error {
	col, row, err := CellNameToCoordinates(axis)
	if err != nil {
		return err
	}

	// Make sure 'slice' is a Ptr to Slice
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return errors.New("pointer to slice expected")
	}
	v = v.Elem()

	for i := 0; i < v.Len(); i++ {
		cell, err := CoordinatesToCellName(col+i, row)
		// Error should never happens here. But keep checking to early detect regresions
		// if it will be introduced in future.
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, v.Index(i).Interface()); err != nil {
			return err
		}
	}
	return err
}

// getCellInfo does common preparation for all SetCell* methods.
func (f *File) prepareCell(xlsx *xlsxWorksheet, sheet, cell string) (*xlsxC, int, int, error) {
	var err error
	cell, err = f.mergeCellsParser(xlsx, cell)
	if err != nil {
		return nil, 0, 0, err
	}
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return nil, 0, 0, err
	}

	prepareSheetXML(xlsx, col, row)

	return &xlsx.SheetData.Row[row-1].C[col-1], col, row, err
}

// getCellStringFunc does common value extraction workflow for all GetCell*
// methods. Passed function implements specific part of required logic.
func (f *File) getCellStringFunc(sheet, axis string, fn func(x *xlsxWorksheet, c *xlsxC) (string, bool, error)) (string, error) {
	xlsx, err := f.workSheetReader(sheet)
	if err != nil {
		return "", err
	}
	axis, err = f.mergeCellsParser(xlsx, axis)
	if err != nil {
		return "", err
	}
	_, row, err := CellNameToCoordinates(axis)
	if err != nil {
		return "", err
	}

	lastRowNum := 0
	if l := len(xlsx.SheetData.Row); l > 0 {
		lastRowNum = xlsx.SheetData.Row[l-1].R
	}

	// keep in mind: row starts from 1
	if row > lastRowNum {
		return "", nil
	}

	for rowIdx := range xlsx.SheetData.Row {
		rowData := &xlsx.SheetData.Row[rowIdx]
		if rowData.R != row {
			continue
		}
		for colIdx := range rowData.C {
			colData := &rowData.C[colIdx]
			if axis != colData.R {
				continue
			}
			val, ok, err := fn(xlsx, colData)
			if err != nil {
				return "", err
			}
			if ok {
				return val, nil
			}
		}
	}
	return "", nil
}

// formattedValue provides a function to returns a value after formatted. If
// it is possible to apply a format to the cell value, it will do so, if not
// then an error will be returned, along with the raw value of the cell.
func (f *File) formattedValue(s int, v string) string {
	if s == 0 {
		return v
	}
	styleSheet := f.stylesReader()
	ok := builtInNumFmtFunc[styleSheet.CellXfs.Xf[s].NumFmtID]
	if ok != nil {
		return ok(styleSheet.CellXfs.Xf[s].NumFmtID, v)
	}
	return v
}

// prepareCellStyle provides a function to prepare style index of cell in
// worksheet by given column index and style index.
func (f *File) prepareCellStyle(xlsx *xlsxWorksheet, col, style int) int {
	if xlsx.Cols != nil && style == 0 {
		for _, c := range xlsx.Cols.Col {
			if c.Min <= col && col <= c.Max {
				style = c.Style
			}
		}
	}
	return style
}

// mergeCellsParser provides a function to check merged cells in worksheet by
// given axis.
func (f *File) mergeCellsParser(xlsx *xlsxWorksheet, axis string) (string, error) {
	axis = strings.ToUpper(axis)
	if xlsx.MergeCells != nil {
		for i := 0; i < len(xlsx.MergeCells.Cells); i++ {
			ok, err := f.checkCellInArea(axis, xlsx.MergeCells.Cells[i].Ref)
			if err != nil {
				return axis, err
			}
			if ok {
				axis = strings.Split(xlsx.MergeCells.Cells[i].Ref, ":")[0]
			}
		}
	}
	return axis, nil
}

// checkCellInArea provides a function to determine if a given coordinate is
// within an area.
func (f *File) checkCellInArea(cell, area string) (bool, error) {
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return false, err
	}

	rng := strings.Split(area, ":")
	if len(rng) != 2 {
		return false, err
	}
	coordinates, err := f.areaRefToCoordinates(area)
	if err != nil {
		return false, err
	}

	return cellInRef([]int{col, row}, coordinates), err
}

// cellInRef provides a function to determine if a given range is within an
// range.
func cellInRef(cell, ref []int) bool {
	return cell[0] >= ref[0] && cell[0] <= ref[2] && cell[1] >= ref[1] && cell[1] <= ref[3]
}

// isOverlap find if the given two rectangles overlap or not.
func isOverlap(rect1, rect2 []int) bool {
	return cellInRef([]int{rect1[0], rect1[1]}, rect2) ||
		cellInRef([]int{rect1[2], rect1[1]}, rect2) ||
		cellInRef([]int{rect1[0], rect1[3]}, rect2) ||
		cellInRef([]int{rect1[2], rect1[3]}, rect2) ||
		cellInRef([]int{rect2[0], rect2[1]}, rect1) ||
		cellInRef([]int{rect2[2], rect2[1]}, rect1) ||
		cellInRef([]int{rect2[0], rect2[3]}, rect1) ||
		cellInRef([]int{rect2[2], rect2[3]}, rect1)
}

// getSharedForumula find a cell contains the same formula as another cell,
// the "shared" value can be used for the t attribute and the si attribute can
// be used to refer to the cell containing the formula. Two formulas are
// considered to be the same when their respective representations in
// R1C1-reference notation, are the same.
//
// Note that this function not validate ref tag to check the cell if or not in
// allow area, and always return origin shared formula.
func getSharedForumula(xlsx *xlsxWorksheet, si string) string {
	for _, r := range xlsx.SheetData.Row {
		for _, c := range r.C {
			if c.F != nil && c.F.Ref != "" && c.F.T == STCellFormulaTypeShared && c.F.Si == si {
				return c.F.Content
			}
		}
	}
	return ""
}