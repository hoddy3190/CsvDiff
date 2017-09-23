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

func parseBaseCsv(baseCsv io.Reader, pks []string) (baseColNames []string, pkValuesStrToRecord map[string][]string) {

    baseColNames = []string{}
    pkValuesStrToRecord = map[string][]string{}
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
        if len(baseColNames) == 0 {
            baseColNames = record
            for i, colName := range baseColNames {
                for _, pk := range pks {
                    if colName == pk {
                        pkIdxList = append(pkIdxList, i)
                    }
                }
            }
            continue
        }

        pkValues := make([]byte, 0, 128)
        for i, pkIdx := range pkIdxList {
            pkValues = append(pkValues, record[pkIdx]...)
            if i < len(pkIdxList) - 1 {
                pkValues = append(pkValues, ":::"...)
            }
        }
        pkValuesStr := string(pkValues)

        pkValuesStrToRecord[pkValuesStr] = record
    }

    return baseColNames, pkValuesStrToRecord
}

func csvDiff(baseCsv, comparisonCsv io.Reader, pks []string) (outputJson []byte) {
    diffJson := DiffJson{}

    baseColNames, basePkValuesStrToRecord := parseBaseCsv(baseCsv, pks)
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
            for i, colName := range comparisonColNames {
                for _, pk := range pks {
                    if colName == pk {
                        pkIdxList = append(pkIdxList, i)
                    }
                }
            }
            continue
        }

        pkValues := make([]byte, 0, 128)
        for i, pkIdx := range pkIdxList {
            pkValues = append(pkValues, comparisonRecord[pkIdx]...)
            if i < len(pkIdxList) - 1 {
                pkValues = append(pkValues, ":::"...)
            }
        }
        pkValuesStr := string(pkValues)

        baseRecord, exist := basePkValuesStrToRecord[pkValuesStr]

        comparisonColNameToVal := map[string]string{}
        for i, colName := range comparisonColNames {
            comparisonColNameToVal[colName] = comparisonRecord[i]
        }

        if exist {
            modifiedContent := ModifiedContent{}
            existModifiedVal := false

            for i, colName := range baseColNames {
                comparisonValue, exist := comparisonColNameToVal[colName]
                if !exist {
                    continue
                }
                if comparisonValue != baseRecord[i] {
                    existModifiedVal = true

                    for _, pkName := range pks {
                        modifiedContent.Pks = map[string]string{}
                        modifiedContent.Pks[pkName] = comparisonColNameToVal[pkName]
                    }

                    fromTo := FromTo{}
                    fromTo.From = baseRecord[i]
                    fromTo.To = comparisonValue

                    modifiedContent.FromTo = map[string]FromTo{}
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


    baseColNameMap := map[string]bool{}
    for _, baseColName := range baseColNames {
        baseColNameMap[baseColName] = true
    }
    for _, comparisonColName := range comparisonColNames {
        _, exist := baseColNameMap[comparisonColName]
        if !exist {
            diffJson.AddedColumns = append(diffJson.AddedColumns, comparisonColName)
        }
        delete(baseColNameMap, comparisonColName)
    }
    for baseColName, _ := range baseColNameMap {
        diffJson.DeletedColumns = append(diffJson.DeletedColumns, baseColName)
    }

    outputJson, err = json.Marshal(&diffJson)
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