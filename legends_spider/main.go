package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	// detail_url = "http://lol.qq.com/web201310/info-defail.shtml?id="
	detail_url = "http://lol.qq.com/biz/hero/"
	pic_url    = "http://ossweb-img.qq.com/images/lol/img/champion/"
	pwd, _     = os.Getwd()
)

type Hero struct {
	Key    int
	Id     string
	Name   string
	Title  string
	Avatar string
	Type   []interface{}
	Url    string
}

type Detail struct {
	Attack  float64
	Magic   float64
	Defense float64
	Hard    float64
	Story   string
}

type Info struct {
	Hero
	Detail
}

type HeroGroup struct {
	Heros []Hero
}

// getHero ...
func getHero(h Hero) (Info_ Info, err error) {
	dStr := "if(!LOLherojs)var LOLherojs={champion:{}};LOLherojs.champion."
	url := detail_url + h.Id + ".js"
	ctx := request(url)
	newStr := strings.Replace(string(ctx), dStr+h.Id+"=", "", -1)
	newStr = strings.Replace(newStr, ";", "", 1)

	var data interface{}
	json.Unmarshal([]byte(newStr), &data)
	if data == nil {
		err = errors.New("nil")
		return
	}
	json_ := data.(map[string]interface{})["data"].(map[string]interface{})
	info := json_["info"].(map[string]interface{})

	detail := Detail{info["attack"].(float64), info["magic"].(float64), info["defense"].(float64), info["difficulty"].(float64), json_["lore"].(string)}
	return Info{h, detail}, nil
}

func cHero(data interface{}) Hero {

	d := data.(map[string]interface{})
	Key, _ := strconv.Atoi(d["key"].(string))
	Id := d["id"].(string)
	Title := d["title"].(string)
	Name := d["name"].(string)
	Avatar := pic_url + d["image"].(map[string]interface{})["full"].(string)
	Type := d["tags"].([]interface{})
	Url := detail_url + Id

	hero := Hero{Key, Id, Name, Title, Avatar, Type, Url}
	return hero
}

func downImg(url string, filename string) {
	ctx := request(url)

	saveFile(ctx, filename, "avatar.png")
}

func exist(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

func saveFile(data []byte, dirname string, filename string) {

	fmt.Println(dirname, filename)
	dir := pwd + "/test/" + dirname
	isExist := exist(dir)
	if !isExist {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	file, err := os.Create(pwd + "/test/" + dirname + "/" + filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	file.Write(data)
}

func request(url string) (ctx []byte) {
	res, reserr := http.Get(url)
	if reserr != nil {
		return
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	return d
}

func exportJson(obj interface{}) {
	var arr [2]map[string]interface{}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		data := s2map(v.Field(i).Interface())
		if t.Field(i).Name == "Hero" {
			delete(data, "Avatar")
		}
		// fmt.Println(data)
		arr[i] = data
	}
	type ty struct {
		Hero   map[string]interface{}
		Detail map[string]interface{}
	}
	newo := ty{arr[0], arr[1]}
	fData, _ := json.Marshal(newo)
	dirname := arr[0]["Id"].(string)

	saveFile(fData, dirname, "info.json")
}

func s2map(o interface{}) map[string]interface{} {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)

	obj := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		obj[t.Field(i).Name] = v.Field(i).Interface()
	}

	return obj
}

func main() {
	var j interface{}
	file, _ := ioutil.ReadFile(string(pwd + "/champion.json"))
	json.Unmarshal(file, &j)

	m := j.(map[string]interface{})
	hs := m["data"].(map[string]interface{})
	for _, val := range hs {
		hBase := cHero(val)
		info, err := getHero(hBase)
		downImg(hBase.Avatar, hBase.Id)
		if err != nil {
			continue
		}
		exportJson(info)
	}
}
