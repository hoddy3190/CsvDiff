package main

import (
    "os"
    "reflect"
    "strings"
    "testing"
)

func TestCsvDiff(t *testing.T) {

    testCaseList := []map[string]string{
        {
            "baseCsv": "./test_csv/test1.csv",
            "comparisonCsv": "./test_csv/test1.csv",
            "pks": "id",
            "expected": `{"added":[],"modified":[],"deleted":[],"addedColumns":[],"deletedColumns":[]}`,
        },
        {
            "baseCsv": "./test_csv/test1.csv",
            "comparisonCsv": "./test_csv/test2.csv",
            "pks": "id",
            "expected": `{"added":[],"modified":[{"pks":{"id":"2"},"fromTo":{"updated_at":{"from":"2017/08/18","to":"2017/08/20"}}},{"pks":{"id":"3"},"fromTo":{"name":{"from":"Lisa","to":"Ben"},"updated_at":{"from":"2016/07/10","to":"2017/09/12"}}}],"deleted":[],"addedColumns":[],"deletedColumns":[]}`,
        },
        {
            "baseCsv": "./test_csv/test1.csv",
            "comparisonCsv": "./test_csv/test3.csv",
            "pks": "id",
            "expected": `{"added":[{"id":"4","name":"Kate","updated_at":"2017/04/09"}],"modified":[],"deleted":[{"id":"2","name":"Tom","updated_at":"2017/08/18"}],"addedColumns":[],"deletedColumns":[]}`,
        },
        {
            "baseCsv": "./test_csv/test1.csv",
            "comparisonCsv": "./test_csv/test4.csv",
            "pks": "id",
            "expected": `{"added":[],"modified":[{"pks":{"id":"3"},"fromTo":{"name":{"from":"Lisa","to":"Ben"}}}],"deleted":[],"addedColumns":["created_at"],"deletedColumns":["updated_at"]}`,
        },
        {
            "baseCsv": "./test_csv/test5.csv",
            "comparisonCsv": "./test_csv/test6.csv",
            "pks": "company_id,employee_id",
            "expected": `{"added":[{"company_id":"1","created_at":"2017/03/29","employee_id":"3","name":"Thomas"}],"modified":[{"pks":{"company_id":"1","employee_id":"2"},"fromTo":{"created_at":{"from":"2017/08/18","to":"2017/06/18"},"name":{"from":"Tom","to":"Ben"}}},{"pks":{"company_id":"2","employee_id":"2"},"fromTo":{"created_at":{"from":"2016/02/14","to":"2016/03/14"}}}],"deleted":[{"company_id":"2","created_at":"2017/04/10","employee_id":"1","name":"Lisa"}],"addedColumns":[],"deletedColumns":[]}`,
        },
        {
            "baseCsv": "./test_csv/test7.csv",
            "comparisonCsv": "./test_csv/test8.csv",
            "pks": "id,user_id",
            "expected": `{"added":[{"id":"2","user_id":"21"}],"modified":[],"deleted":[{"id":"2","user_id":"20"}],"addedColumns":[],"deletedColumns":[]}`,
        },
    }

    for i, testCase := range testCaseList {

        baseCsv, err := os.Open(testCase["baseCsv"])
        if err != nil {
            t.Fatal(err)
        }

        comparisonCsv, err := os.Open(testCase["comparisonCsv"])
        if err != nil {
            t.Fatal(err)
        }

        expected := ([]byte)(testCase["expected"])
        actual := csvDiff(baseCsv, comparisonCsv, strings.Split(testCase["pks"], ","))

        if !reflect.DeepEqual(expected, actual) {
            t.Error("TestCase No:", i + 1)
            t.Error("expected: ", string(expected))
            t.Error("actual: ", string(actual))
        }

        baseCsv.Close()
        comparisonCsv.Close()

    }

}
