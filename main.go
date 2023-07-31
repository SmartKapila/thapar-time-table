package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/utkarsh-1905/thapar-time-table/utils"
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("timetable.xlsx")

	utils.HandleError(err)

	defer func() {
		if err = f.Close(); err != nil {
			panic(err)
		}
	}()

	table, _ := template.ParseFiles("./templates/table.html")
	home, _ := template.ParseFiles("./templates/home.html")
	errorPage, _ := template.ParseFiles("./templates/error.html")

	type HomeData struct {
		Sheets  []string
		Classes map[string]map[int]string
	}

	sheets := f.GetSheetList()

	classes := make(map[string]map[int]string)

	//finding all classes in a sheet
	for _, sheet := range sheets {
		temp := make(map[int]string)
		cols, err := f.GetRows(sheet)
		for i, d := range cols {
			if i == 3 {
				for j, k := range d {
					if k != "" && k != "DAY" && k != "HOURS" && k != "SR NO" && k != "SR.NO" {
						temp[j+1] = k
					}
				}
			}
		}
		classes[sheet] = temp
		utils.HandleError(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			errorPage.Execute(w, "This page is under construction !!(404)")
			return
		}
		h := HomeData{
			Sheets:  sheets,
			Classes: classes,
		}
		home.Execute(w, h)
	})

	http.HandleFunc("/timetable", func(w http.ResponseWriter, r *http.Request) {
		class, _ := strconv.Atoi(r.URL.Query().Get("class"))
		sheet := r.URL.Query().Get("sheet")
		classname := r.URL.Query().Get("classname")

		flag := true
		for i, d := range classes {
			if i == sheet {
				for _, k := range d {
					if classname == k {
						flag = false
					}
				}
				break
			}
		}
		if flag {
			errorPage.Execute(w, "Invalid category/class combination")
			return
		}
		table.Execute(w, utils.GetTableData(sheet, class, f))
	})

	ExcelToJson(classes, f)
	fmt.Println("Server Running at http://localhost:3000")
	err = http.ListenAndServe(":3000", nil)
	utils.HandleError(err)
}

func ExcelToJson(classes map[string]map[int]string, f *excelize.File) {
	file, err := os.OpenFile("./data.json", os.O_TRUNC|os.O_CREATE, 0777)
	utils.HandleError(err)
	defer file.Close()
	data := make(map[string]map[string][][]utils.Data)
	for i, d := range classes {
		temp := make(map[string][][]utils.Data)
		for j, k := range d {
			tc := utils.GetTableData(i, j, f)
			temp[k] = tc
		}
		data[i] = temp
	}
	dj, _ := json.MarshalIndent(data, "", "	")
	_, err = file.Write(dj)
	utils.HandleError(err)
}
