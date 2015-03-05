package envparse

import (
    "fmt"
    "os"
    "strconv"
)

func GetSettingStr(key string, def string) string {
    val := os.Getenv(key)
    if val == "" {
        val = def
    }
    return val
}

func GetSettingInt(key string, def int) int {
    var valInt int
    val := os.Getenv(key)
    if val == "" {
        valInt = def
    } else {
        valParsed, err := strconv.Atoi(val)
        if err != nil {
            fmt.Printf("Expecting an integer for key %s", key)
            os.Exit(1)
        }
        valInt = valParsed
    }
    return valInt
}
