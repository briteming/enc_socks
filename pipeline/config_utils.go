package pipeline

import (
    "errors"
    "strings"
    "strconv"
    "fmt"
)

func Get(mp *map[string]interface{}, path string) (interface{}, error) {
    if mp == nil {
        return nil, errors.New("config nil")
    }
    if len(path) == 0 {
        return *mp, nil
    }
    var v interface{}
    v = *mp
    seps := strings.Split(path, ".")
    for _, sep := range seps {
        switch v.(type) {
        case []interface{}:
            index, err := strconv.ParseInt(sep, 10, 32)
            if err != nil {
                return nil, errors.New(fmt.Sprintf("sub path:%s in array must be int type", sep))
            }
            v = v.([]interface{})[index]
        case map[string]interface{}:
            v = v.(map[string]interface{})[sep]
        default:
            return nil, errors.New(fmt.Sprintf("invalid sub path:%s in json:%v", sep, v))
        }
    }
    return v, nil
}

func GetInt(mp *map[string]interface{}, path string) (int, error) {
    v, e := Get(mp, path)
    if e != nil {
        return 0, e
    }
    switch v.(type) {
    case int:
        return v.(int), nil
    }
    return 0, errors.New("type not match, path:" + path)
}

func GetString(mp *map[string]interface{}, path string) (string, error) {
    v, e := Get(mp, path)
    if e != nil {
        return "", e
    }
    switch v.(type) {
    case string:
        return v.(string), nil
    }
    return "", errors.New("type not match, path:" + path)
}
