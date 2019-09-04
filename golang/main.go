package main

import (
"crypto/tls"
"encoding/json"
"fmt"
"github.com/bitly/go-simplejson"
"github.com/spf13/viper"
"io/ioutil"
"net/http"
"net/http/cookiejar"
"net/url"
"os"
"sort"
"strconv"
)

var (
	num int64
	username string
	password string
	harbor_url string
	api_url string
	login_url string
	pro_url string
	repos_url string
)

type Tag struct {
	Created string `json:"created"`
	Name string `json:"name"`
}

type Tags []Tag

func (I Tags) Len() int {
	return len(I)
}

func (I Tags) Less(i,j int) bool {
	return I[i].Created > I[j].Created
}

func (I Tags) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func init(){
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
		os.Exit(1)
	}
	num = viper.GetInt64("num")
	username = viper.GetString("harbor.username")
	password = viper.GetString("harbor.password")
	harbor_url = viper.GetString("harbor.url")
	login_url = harbor_url+"/login"
	api_url = harbor_url + "/api"
	pro_url = api_url + "/projects"
	repos_url = api_url + "/repositories"
}

func harborClient()(cli *http.Client) {
	tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, }
	jar,err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	cli = &http.Client{Transport: tr, Jar: jar}
	return
}

func harborLogin(cli * http.Client) {
	form := make(url.Values)
	form.Set("principal",username)
	form.Add("password",password)
	resp,_ := cli.PostForm(login_url, form)
	defer resp.Body.Close()
}

func getProjects(cli *http.Client) (project_ids []string) {
	resp,_ := cli.Get(pro_url)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		panic(err.Error())
	}
	rows, err := js.Array()
	for _, row := range rows {
		if each_map, ok := row.(map[string]interface{}); ok {
			project_ids = append(project_ids,each_map["project_id"].(json.Number).String())
		}
	}
	return
}

func fetch_del_repos_name(cli *http.Client, id string) (del_repos_name []string){
	resp,_ := cli.Get(repos_url+"?project_id="+id)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		panic(err.Error())
	}
	rows, err := js.Array()
	for _, row := range rows {
		if each_map, ok := row.(map[string]interface{}); ok {
			if tags_count, ok := each_map["tags_count"].(json.Number); ok {
				count, _ := strconv.ParseInt(string(tags_count), 10, 0)
				if count > num {
					del_repos_name = append(del_repos_name, each_map["name"].(string))
				}
			}

		}
	}
	return
}

func del_tags(cli *http.Client, repo_name string) {
	var tags Tags
	tag_url := repos_url + "/" + repo_name + "/tags"
	resp,_ := cli.Get(tag_url)
	data, _ := ioutil.ReadAll(resp.Body)
	js, err := simplejson.NewJson([]byte(data))
	if err != nil {
		panic(err.Error())
	}
	rows, err := js.Array()
	for _, row := range rows {
		if each_map, ok := row.(map[string]interface{}); ok {
			var tag Tag
			tag.Name = each_map["name"].(string)
			tag.Created = each_map["created"].(string)
			tags = append(tags,tag)
		}
	}
	sort.Sort(&tags)
	del_tags := tags[num:]
	for _,t := range del_tags {
		del_repo_tag_url := tag_url + "/" + t.Name
		req,_ := http.NewRequest("DELETE", del_repo_tag_url,nil)
		resp, _ := cli.Do(req)
		defer resp.Body.Close()
		fmt.Printf("删除状态: %v 删除镜像tag: %v\n",resp.StatusCode,del_repo_tag_url)
	}
}

func work(){
	cli := harborClient()
	harborLogin(cli)
	project_ids := getProjects(cli)
	for _,id := range project_ids {
		del_repos_names := fetch_del_repos_name(cli,id)
		if len(del_repos_names) > 0 {
			for _,image := range del_repos_names {
				del_tags(cli, image)
			}
		}
	}
}

func main(){
	work()
}

