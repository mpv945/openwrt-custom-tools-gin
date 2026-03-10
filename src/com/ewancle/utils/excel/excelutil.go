package excelutil

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func ReadExcel(filePath string) error {

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	rows, err := f.Rows("Sheet1")
	if err != nil {
		return err
	}
	defer func(rows *excelize.Rows) {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}(rows)

	for rows.Next() {

		row, err := rows.Columns()
		if err != nil {
			return err
		}

		if len(row) < 3 {
			continue
		}

		id := row[0]
		name := row[1]
		email := row[2]

		fmt.Println(id, name, email)
	}

	return nil
}

func WriteExcel(filePath string, total int) error {

	f := excelize.NewFile()

	streamWriter, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	// header
	header := []interface{}{"ID", "Name", "Email"}

	cell, _ := excelize.CoordinatesToCellName(1, 1)
	err1 := streamWriter.SetRow(cell, header)
	if err1 != nil {
		return err1
	}

	for i := 1; i <= total; i++ {

		row := []interface{}{
			i,
			fmt.Sprintf("user_%d", i),
			fmt.Sprintf("user_%d@test.com", i),
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+1)

		err := streamWriter.SetRow(cell, row)
		if err != nil {
			return err
		}
	}

	err2 := streamWriter.Flush()
	if err2 != nil {
		return err2
	}

	return f.SaveAs(filePath)
}
