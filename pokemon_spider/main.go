package main

import (
	"encoding/json"
	"fmt"
	query "github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"regexp"
)

var (
	wiki_url     = "http://www.pokemon.name/wiki/"
	pokemon_list = wiki_url + "%E5%AE%9D%E5%8F%AF%E6%A2%A6%E5%88%97%E8%A1%A8"
	pwd, _       = os.Getwd()
)

type PokemonLink struct {
	url    string
	avatar string
	key    string
	id     string
}

type PokemonInfo struct {
	Id      int                        `json:"id"`
	Name    []map[string]interface{}   `json:"name"`
	Avatar  string                     `json:"avatar"`
	Message [][]map[string]interface{} `json:"message"`
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

	ctx = d
	return
}

func getDom(url string) *query.Document {
	res, reserr := http.Get(url)
	if reserr != nil {
	}
	defer res.Body.Close()

	doc, err := query.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	return doc
}

func getBase(url string) (pokes []PokemonLink) {
	doc := getDom(url)
	table := doc.Find("table")
	for i := range table.Nodes {
		lis := table.Eq(i).Find("td li")
		for j := range lis.Nodes {
			str := regexp.MustCompile("[0-9]{3}").FindAllString(lis.Eq(j).Text(), -1)
			key := str[0]
			id := lis.Eq(j).Find("a").Text()
			url := wiki_url + id

			pokes = append(pokes, PokemonLink{key: key, id: id, url: url})
		}

	}
	return
}

func (p *PokemonLink) getAvatar() *PokemonLink {
	doc := getDom(p.url)
	image, _ := doc.Find(".infobox-image img").Attr("src")

	p.avatar = image
	return p
}

func getDetail(p PokemonLink) (info PokemonInfo) {
	info.Id, _ = strconv.Atoi(p.key)
	info.Avatar = p.avatar

	doc := getDom(p.url)

	// name
	nameDom := doc.Find("#mw-content-text > table.colortable.colortable-width-full.colortable-colsep-1.colortable-rowsep-2.colorize.colorize-default.colorize-default-default.text-center")
	nameArr := make([]map[string]interface{}, 0)
	for i := range nameDom.Find("tr").Nodes {
		if i == 0 {
			continue
		}
		n := make(map[string]interface{})
		n[nameDom.Find("tr").Eq(i).Find("td").Eq(0).Text()] = nameDom.Find("tr").Eq(i).Find("td").Eq(1).Text()
		nameArr = append(nameArr, n)
		info.Name = nameArr
	}

	// message
	messageDom := doc.Find("#pokemonform-1 > table")
	var (
		baseMsgArr = make([]map[string]interface{}, 0)
		raceArr    = make([]map[string]interface{}, 0)
		cultiArr   = make([]map[string]interface{}, 0)
	)
	for i := range messageDom.Find("tr").Nodes {

		var (
			item  = messageDom.Find("tr").Eq(i)
			m     = make(map[string]interface{})
			title string
		)

		if item.Find("td").Eq(0).Children().Length() > 0 {
			title = item.Find("td").Eq(0).Children().Text()
		} else {
			title = item.Find("td").Eq(0).Text()
		}

		switch {
		case i < 7 && i > 0:
			l := item.Find("td").Eq(1).Children().Length()
			switch {
			case l == 1:
				value := item.Find("td").Eq(1).Children().Text()
				m[title] = value
				break
			case l > 1:
				arr := make([]string, l)
				for j := range item.Find("td").Eq(1).Children().Nodes {
					arr[j] = item.Find("td").Eq(1).Children().Eq(j).Text()
					m[title] = arr
				}
				break
			default:
				td := item.Find("td")
				m[title] = td.Eq(1).Text()
				break
			}

			baseMsgArr = append(baseMsgArr, m)
			break

		case i < 15 && i > 7:
			m[title] = item.Find("td").Eq(1).Text()
			raceArr = append(raceArr, m)
			break

		case i > 19:
			l := item.Find("td").Eq(1).Children().Length()
			switch {
			case l == 1:
				value := item.Find("td").Eq(1).Children().Text()
				m[title] = value
				break
			case l > 1:
				arr := make([]string, l)
				for j := range item.Find("td").Eq(1).Children().Nodes {
					arr[j] = item.Find("td").Eq(1).Children().Eq(j).Text()
					m[title] = arr
				}
				break
			default:
				td := item.Find("td")
				m[title] = td.Eq(1).Text()
				break
			}

			cultiArr = append(cultiArr, m)
			break
		}
	}
	message := make([][]map[string]interface{}, 0)
	message = append(message, baseMsgArr)
	message = append(message, raceArr)
	message = append(message, cultiArr)
	info.Message = message
	return
}

func exist(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

func exportJson(p PokemonInfo) {
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	saveFile(data, strconv.Itoa(p.Id), strconv.Itoa(p.Id)+".json")
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
	base := getBase(pokemon_list)
	for _, p := range base {
		// fmt.Println((&p).id)
		(&p).getAvatar()
		info := getDetail(p)
		exportJson(info)
	}
}
