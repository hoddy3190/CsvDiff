package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "github.com/jessevdk/go-flags"
    "io"
    "io/ioutil"
    "log"
    "os"
)

type Options struct {
    Pks    []string `long:"pk" description:"primary keys"`
    Output string `short:"o" long:"output" description:"file path where output json is written"`
}

type DiffJson struct {
    Added          []map[string]string `json:"added"`
    Modified       []ModifiedContent `json:"modified"`
    Deleted        []map[string]string `json:"deleted"`
    AddedColumns   []string `json:"addedColumns"`
    DeletedColumns []string `json:"deletedColumns"`
}

type ModifiedContent struct {
    Pks    map[string]string `json:"pks"`
    FromTo map[string]FromTo `json:"fromTo"`
}

type FromTo struct {
    From string `json:"from"`
    To   string `json:"to"`
}

func sliceDiff(slice1 []string, slice2 []string) ([]string, []string) {

    elemsOnlyInSlice1 := []string{}
    elemsOnlyInSlice2 := []string{}

    slice1Map := map[string]bool{}
    for _, slice1Elem := range slice1 {
        slice1Map[slice1Elem] = true
    }
    for _, slice2Elem := range slice2 {
        _, exist := slice1Map[slice2Elem]
        if !exist {
            elemsOnlyInSlice2 = append(elemsOnlyInSlice2, slice2Elem)
        }
        delete(slice1Map, slice2Elem)
    }
    for elemOnlyInSlice1, _ := range slice1Map {
        elemsOnlyInSlice1 = append(elemsOnlyInSlice1, elemOnlyInSlice1)
    }

    return elemsOnlyInSlice1, elemsOnlyInSlice2
}

func getPkIdxList(colNames []string, pks []string) []int {
    pkIdxList := []int{}
    for i, colName := range colNames {
        for _, pk := range pks {
            if colName == pk {
                pkIdxList = append(pkIdxList, i)
            }
        }
    }
    return pkIdxList
}

// return "pk(1):::pk(2):::...:::pk(n)" string
func getPkValuesStr(record []string, pkIdxList []int) string {
    pkValues := make([]byte, 0, 128)
    for i, pkIdx := range pkIdxList {
        pkValues = append(pkValues, record[pkIdx]...)
        if i < len(pkIdxList) - 1 {
            pkValues = append(pkValues, ":::"...)
        }
    }
    return string(pkValues)
}

func parseCsv(baseCsv io.Reader, pks []string) ([]string, map[string][]string) {
    colNames := []string{}
    pkValuesStrToRecord := map[string][]string{}
    pkIdxList := []int{}

    r := csv.NewReader(baseCsv)
    r.LazyQuotes = true

    for {
        record, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

        // column names are written in the first line
        if len(colNames) == 0 {
            colNames = record
            pkIdxList = getPkIdxList(colNames, pks)
            continue
        }

        pkValuesStr := getPkValuesStr(record, pkIdxList)
        pkValuesStrToRecord[pkValuesStr] = record
    }

    return colNames, pkValuesStrToRecord
}

func csvDiff(baseCsv, comparisonCsv io.Reader, pks []string) []byte {
    diffJson := DiffJson{}
    diffJson.Added = []map[string]string{}
    diffJson.Modified = []ModifiedContent{}
    diffJson.Deleted = []map[string]string{}

    baseColNames, basePkValuesStrToRecord := parseCsv(baseCsv, pks)
    pkIdxList := []int{}

    r := csv.NewReader(comparisonCsv)
    r.LazyQuotes = true

    comparisonColNames := []string{}

    for {
        comparisonRecord, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

        // column names are written in the first line
        if len(comparisonColNames) == 0 {
            comparisonColNames = comparisonRecord
            pkIdxList = getPkIdxList(comparisonColNames, pks)
            continue
        }

        pkValuesStr := getPkValuesStr(comparisonRecord, pkIdxList)
        baseRecord, exist := basePkValuesStrToRecord[pkValuesStr]

        comparisonColNameToVal := map[string]string{}
        for i, colName := range comparisonColNames {
            comparisonColNameToVal[colName] = comparisonRecord[i]
        }

        if exist {
            modifiedContent := ModifiedContent{}
            modifiedContent.FromTo = map[string]FromTo{}
            existModifiedVal := false

            for i, colName := range baseColNames {
                comparisonValue, exist := comparisonColNameToVal[colName]
                if !exist {
                    continue
                }
                if comparisonValue != baseRecord[i] {
                    existModifiedVal = true

                    if len(modifiedContent.Pks) == 0 {
                        modifiedContent.Pks = map[string]string{}
                        for _, pkName := range pks {
                            modifiedContent.Pks[pkName] = comparisonColNameToVal[pkName]
                        }
                    }

                    fromTo := FromTo{}
                    fromTo.From = baseRecord[i]
                    fromTo.To = comparisonValue

                    modifiedContent.FromTo[colName] = fromTo
                }
            }

            if existModifiedVal {
                diffJson.Modified = append(diffJson.Modified, modifiedContent)
            }

            delete(basePkValuesStrToRecord, pkValuesStr)

        } else {
            diffJson.Added = append(diffJson.Added, comparisonColNameToVal)
        }
    }

    for _, record := range basePkValuesStrToRecord {
        baseColNameToVal := map[string]string{}
        for i, colName := range baseColNames {
            baseColNameToVal[colName] = record[i]
        }
        diffJson.Deleted = append(diffJson.Deleted, baseColNameToVal)
    }

    deletedColumns, addedColumns := sliceDiff(baseColNames, comparisonColNames)
    diffJson.AddedColumns = addedColumns
    diffJson.DeletedColumns = deletedColumns

    outputJson, err := json.Marshal(&diffJson)
    if err != nil {
        log.Fatal(err)
    }

    return outputJson
}

func main() {
    opts := Options{}
    parser := flags.NewParser(&opts, flags.Default)
    parser.Usage = "base.csv compare.csv [OPTIONS]"
    args, err := parser.Parse()

    // exit 0 early if help option is added
    if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
        os.Exit(0)
    }

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    if len(args) != 2 {
        fmt.Println("argument num is invalid\n")
        parser.WriteHelp(os.Stdout)
        os.Exit(1)
    }

    if len(opts.Pks) == 0 {
        fmt.Println("At least one pk option must be set\n")
        parser.WriteHelp(os.Stdout)
        os.Exit(1)
    }

    baseCsv, err := os.Open(args[0])
    if err != nil {
        log.Fatal(err)
    }
    defer baseCsv.Close()

    comparisonCsv, err := os.Open(args[1])
    if err != nil {
        log.Fatal(err)
    }
    defer comparisonCsv.Close()

    outputJson := csvDiff(baseCsv, comparisonCsv, opts.Pks)
    if len(opts.Output) > 0 {
        ioutil.WriteFile(opts.Output, outputJson, os.ModePerm)
    } else {
        fmt.Println(string(outputJson))
    }
}
