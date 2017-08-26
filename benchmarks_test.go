package xlsx_test

import (
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize"
	"math/rand"
	"testing"
	"time"
	"strconv"

	ooxml "github.com/plandem/xlsx"
	"github.com/plandem/xlsx/format"
)

const simpleFile = "./test_files/simple.xlsx"
const bigFile = "./test_files/example_big.xlsx"
const hugeFile = "./test_files/example_huge.xlsx"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type openFileFn func(fileName string) (interface{}, interface{})

func tealegOpen(fileName string) (interface{}, interface{}) {
	xl, err := xlsx.OpenFile(fileName)
	if err != nil {
		panic(err)
	}

	return xl, xl.Sheets[0]
}

func excelizeOpen(fileName string) (interface{}, interface{}) {
	xl, err := excelize.OpenFile(fileName)
	if err != nil {
		panic(err)
	}

	return xl, "Sheet1"
}

func xlsxOpen(fileName string) (interface{}, interface{}) {
	xl, err := ooxml.Open(fileName)
	if err != nil {
		panic(err)
	}

	return xl, xl.Sheet(0)
}

func BenchmarkRandomGet(b *testing.B) {
	benchmarks := []struct {
		name     string
		open     openFileFn
		callback func(f interface{}, s interface{}, value *string, x int, y int)
	}{
		{"excelize", excelizeOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			xl := f.(*excelize.File)
			col := excelize.ToAlphaString(rand.Intn(maxCols))
			row := strconv.Itoa(1 + rand.Intn(maxRows))
			axis := col + row
			*value = xl.GetCellValue("Sheet1", axis)
		}},
		{"tealeg", tealegOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			sheet := s.(*xlsx.Sheet)
			*value = sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).Value
		}},
		{"xlsx", xlsxOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			sheet := s.(*ooxml.Sheet)
			*value = sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).Value()
		}},
	}

	const maxCols = 100
	const maxRows = 100
	var value string

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			f, sheet := bm.open(simpleFile)
			for i := 0; i < b.N; i++ {
				bm.callback(f, sheet, &value, maxCols, maxRows)
			}
		})
	}
}

func BenchmarkRandomSet(b *testing.B) {
	benchmarks := []struct {
		name     string
		open     openFileFn
		callback func(f interface{}, s interface{}, value *string, x int, y int)
	}{
		{"excelize", excelizeOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			xl := f.(*excelize.File)
			col := excelize.ToAlphaString(rand.Intn(maxCols))
			row := strconv.Itoa(1 + rand.Intn(maxRows))
			axis := col + row
			xl.SetCellValue("Sheet1", axis, rand.Intn(100))
		}},
		{"tealeg", tealegOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			sheet := s.(*xlsx.Sheet)
			sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).SetValue(rand.Intn(100))
		}},
		{"xlsx", xlsxOpen, func(f interface{}, s interface{}, value *string, maxCols, maxRows int) {
			sheet := s.(*ooxml.Sheet)
			sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).SetValue(rand.Intn(100))
		}},
	}

	const maxCols = 100
	const maxRows = 100
	var value string

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			f, sheet := bm.open(simpleFile)
			for i := 0; i < b.N; i++ {
				bm.callback(f, sheet, &value, maxCols, maxRows)
			}
		})
	}
}

func BenchmarkRandomSetStyle(b *testing.B) {
	benchmarks := []struct {
		name     string
		open     openFileFn
		create   func(f interface{}) interface{}
		callback func(f interface{}, s interface{}, ss interface{}, x int, y int)
	}{
		{"excelize", excelizeOpen, func(f interface{}) interface{} {
			xl := f.(*excelize.File)
			style, err := xl.NewStyle(`{"custom_number_format": "[$-380A]dddd\\,\\ dd\" de \"mmmm\" de \"yyyy;@"}`)
			if err != nil {
				panic(err)
			}

			return style
		}, func(f interface{}, s interface{}, ss interface{}, maxCols, maxRows int) {
			xl := f.(*excelize.File)
			styleId := ss.(int)

			col := excelize.ToAlphaString(rand.Intn(maxCols))
			row := strconv.Itoa(1 + rand.Intn(maxRows))
			axis := col + row

			xl.SetCellStyle("Sheet1", axis, axis, styleId)
		}},
		{"tealeg", tealegOpen, func(f interface{}) interface{} {
			style := xlsx.NewStyle()
			font := *xlsx.NewFont(12, "Verdana")
			font.Bold = true
			font.Italic = true
			font.Underline = true
			style.Font = font
			fill := *xlsx.NewFill("solid", "00FF0000", "FF000000")
			style.Fill = fill
			border := *xlsx.NewBorder("thin", "thin", "thin", "thin")
			style.Border = border
			style.ApplyBorder = true

			return style
		}, func(f interface{}, s interface{}, ss interface{}, maxCols, maxRows int) {
			sheet := s.(*xlsx.Sheet)
			style := ss.(*xlsx.Style)
			sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).SetStyle(style)
		}},
		{"xlsx", xlsxOpen, func(f interface{}) interface{} {
			xl := f.(*ooxml.Spreadsheet)

			style := format.New(
				format.Font.Name("Calibri"),
				format.Font.Size(12),
				format.Font.Color("#FF0000"),
				format.Font.Scheme(format.FontSchemeMinor),
				format.Font.Family(format.FontFamilySwiss),

				format.Fill.Type(format.PatternTypeNone),

				format.Alignment.VAlign(format.VAlignBottom),
				format.Alignment.HAlign(format.HAlignFill),
				format.Border.Color("#ff00ff"),
				format.Border.Type(format.BorderStyleDashDot),
				format.Protection.Hidden,
				format.Protection.Locked,
				//format.NumberFormat("#.### usd"),
				format.Fill.Type(format.PatternTypeDarkDown),
				format.Fill.Color("#FFFFFF"),
				format.Fill.Background("#FF0000"),
			)

			return xl.AddFormatting(style)
		}, func(f interface{}, s interface{}, ss interface{}, maxCols, maxRows int) {
			sheet := s.(*ooxml.Sheet)
			styleId := ss.(format.StyleRefID)
			sheet.Cell(rand.Intn(maxCols), rand.Intn(maxRows)).SetFormatting(styleId)
		}},
	}

	const maxCols = 100
	const maxRows = 100

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			f, sheet := bm.open(simpleFile)
			style := bm.create(f)
			for i := 0; i < b.N; i++ {
				bm.callback(f, sheet, style, maxCols, maxRows)
			}
		})
	}
}

func BenchmarkReadBigFile(b *testing.B) {
	benchmarks := []struct {
		name     string
		open     openFileFn
		callback func(f interface{}, s interface{}, value *string)
	}{
		{"excelize", excelizeOpen, func(f interface{}, s interface{}, value *string) {
			xl := f.(*excelize.File)
			rows := xl.GetRows("Sheet1")

			for _, row := range rows {
				*value = row[0]
			}
		}},
		{"tealeg", tealegOpen, func(f interface{}, s interface{}, value *string) {
			sheet := s.(*xlsx.Sheet)
			for row_i, row_max := 0, len(sheet.Rows); row_i < row_max; row_i++ {
				*value = sheet.Cell(0, row_i).Value
			}
		}},
		{"xlsx", xlsxOpen, func(f interface{}, s interface{}, value *string) {
			sheet := s.(*ooxml.Sheet)
			for row_i, row_max := 0, sheet.TotalRows(); row_i < row_max; row_i++ {
				*value = sheet.Cell(0, row_i).Value()
			}
		}},
	}

	var value string
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f, s := bm.open(bigFile)
				bm.callback(f, s, &value)
			}
		})
	}
}

func BenchmarkReadHugeFile(b *testing.B) {
	benchmarks := []struct {
		name     string
		open     openFileFn
		callback func(f interface{}, s interface{}, value *string)
	}{
		//{"excelize", excelizeOpen, func(f interface{}, s interface{}, value *string) {
		//	xl := f.(*excelize.File)
		//	rows := xl.GetRows("Sheet1")
		//
		//	for _, row := range rows {
		//		*value = row[0]
		//	}
		//}},
		//{"tealeg", tealegOpen, func(f interface{}, s interface{}, value *string) {
		//	sheet := s.(*xlsx.Sheet)
		//	for row_i, row_max := 0, len(sheet.Rows); row_i < row_max; row_i++ {
		//		*value = sheet.Cell(0, row_i).Value
		//	}
		//}},
		{"xlsx", xlsxOpen, func(f interface{}, s interface{}, value *string) {
			sheet := s.(*ooxml.Sheet)
			for row_i, row_max := 0, sheet.TotalRows(); row_i < row_max; row_i++ {
				*value = sheet.Cell(0, row_i).Value()
			}
		}},
	}

	var value string
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f, s := bm.open(hugeFile)
				bm.callback(f, s, &value)
			}
		})
	}
}

