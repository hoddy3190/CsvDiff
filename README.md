# CsvDiff
[![CircleCI](https://circleci.com/gh/altitude3190/CsvDiff/tree/master.svg?style=svg)](https://circleci.com/gh/altitude3190/CsvDiff/tree/master)

output csv diff smartly

## Usage

```
./CsvDiff base.csv comparison.csv --pk pk1 [--pk pk2 ...] [--output filePath]
```

Csv format has to meet the followings
+ column names are written in the first line
+ values are written in the second and subsequent lines

You can see detail in executing the following command
```
./CsvDiff --help
```

## Output Examples

#### example1

```
$ cat base1.csv
company_id,employee_id,name,created_at
1,1,"Bob","2017/09/24"
1,2,"Tom","2017/08/18"
2,1,"Lisa","2017/04/10"
2,2,"Kate","2016/02/14"

$ cat comparison1.csv
company_id,employee_id,name,created_at
1,1,"Bob","2017/09/24"
1,2,"Ben","2017/06/18"
1,3,"Thomas","2017/03/29"
2,2,"Kate","2016/03/14"

$ ./CsvDiff base1.csv comparison1.csv --pk company_id --pk employee_id | python -m "json.tool"
{
    "added": [
        {
            "company_id": "1",
            "created_at": "2017/03/29",
            "employee_id": "3",
            "name": "Thomas"
        }
    ],
    "addedColumns": [],
    "deleted": [
        {
            "company_id": "2",
            "created_at": "2017/04/10",
            "employee_id": "1",
            "name": "Lisa"
        }
    ],
    "deletedColumns": [],
    "modified": [
        {
            "fromTo": {
                "created_at": {
                    "from": "2017/08/18",
                    "to": "2017/06/18"
                },
                "name": {
                    "from": "Tom",
                    "to": "Ben"
                }
            },
            "pks": {
                "company_id": "1",
                "employee_id": "2"
            }
        },
        {
            "fromTo": {
                "created_at": {
                    "from": "2016/02/14",
                    "to": "2016/03/14"
                }
            },
            "pks": {
                "company_id": "2",
                "employee_id": "2"
            }
        }
    ]
}
```

#### example2

It supports add/delete columns diff, too

```
$ cat base2.csv
id,name,updated_at
1,"Bob","2017/09/24"
2,"Tom","2017/08/18"
3,"Lisa","2016/07/10"

$ cat comparison2.csv
id,name,created_at
1,"Bob","2017/09/24"
2,"Tom","2017/08/18"
3,"Ben","2016/07/10"

$ ./CsvDiff base2.csv comparison2.csv --pk id --output res.json && cat res.json | python -m 'json.tool'
{
    "added": [],
    "addedColumns": [
        "created_at"
    ],
    "deleted": [],
    "deletedColumns": [
        "updated_at"
    ],
    "modified": [
        {
            "fromTo": {
                "name": {
                    "from": "Lisa",
                    "to": "Ben"
                }
            },
            "pks": {
                "id": "3"
            }
        }
    ]
}
```

## Author

[Hodaka Suzuki](https://github.com/altitude3190)


## Lisense

[MIT](https://github.com/altitude3190/CsvDiff/blob/master/LICENSE)

