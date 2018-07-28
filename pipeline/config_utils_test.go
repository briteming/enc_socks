package pipeline

import "testing"

func TestConfigUtils(t *testing.T) {
    mp := make(map[string]interface{})
    obj := make(map[string]interface{})
    arr := make([]interface{}, 2)
    mp["5"] = 6
    mp["test"] = "gello"
    mp["haha"] = "texxxxx"
    arr[0] = "hehe"
    arr[1] = 8888
    mp["obj"] = obj
    mp["arr"] = arr
    obj["xxx"] = "yyy"

    {
        v, err := GetInt(&mp, "5")
        t.Logf("v:%v, err:%v\n", v, err)
    }
    {
        v, err := GetInt(&mp, "5.6")
        t.Logf("v:%v, err:%v\n", v, err)
    }
    {
        v, err := GetString(&mp, "obj.xxx")
        t.Logf("v:%v, err:%v\n", v, err)
    }
    {
        v, err := GetInt(&mp, "arr.1")
        t.Logf("v:%v, err:%v\n", v, err)
    }
}
