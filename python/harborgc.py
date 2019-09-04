#! /usr/bin/env python3
# -*- coding:utf-8 -*-
# 本脚本只是harbor页面tags清理
# 物理清理：docker run -it --name gc --rm --volumes-from registry goharbor/registry-photon:v2.6.2-v1.6.3 garbage-collect /etc/registry/config.yml
import requests

class RequestClient(object):

    def __init__(self, login_url, username, password):
        self.username = username
        self.password = password
        self.login_url = login_url
        self.session = requests.Session()
        self.login()

    def login(self):
        requests.packages.urllib3.disable_warnings()
        self.session.post(self.login_url, params={"principal": self.username, "password": self.password}, verify=False)


class ClearHarbor(object):

    def __init__(self, harbor_domain, username, password, num, schema="https", ):
        self.num = num
        self.schema = schema
        self.harbor_domain = harbor_domain
        self.harbor_url = self.schema + "://" + self.harbor_domain
        self.login_url = self.harbor_url + "/login"
        self.api_url = self.harbor_url + "/api"
        self.pro_url = self.api_url + "/projects"
        self.repos_url = self.api_url + "/repositories"
        self.username = username
        self.password = password
        self.client = RequestClient(self.login_url, self.username, self.password)

    def __fetch_pros_obj(self):
        # 获取所有项目名称
        self.pros_obj = self.client.session.get(self.pro_url).json()
        # for n in self.pros_obj:
        #     print("分组：",n.get("name"))
        return self.pros_obj

    def fetch_pros_id(self):
        # 获取所有项目ID
        self.pros_id = []
        pro_res = self.__fetch_pros_obj()
        for i in pro_res:
            self.pros_id.append(i['project_id'])
        # print("所有项目ID：",self.pros_id)
        return self.pros_id

    def fetch_del_repos_name(self, pro_id):
        # 镜像tag数量大于self.num的镜像仓库名称
        self.del_repos_name = []
        repos_res = self.client.session.get(self.repos_url, params={"project_id": pro_id})
        # print("项目信息：",repos_res.json())
        for repo in repos_res.json():
            if repo["tags_count"] > self.num:
                # print("镜像仓库名称：",repo['name'])
                self.del_repos_name.append(repo['name'])
        return self.del_repos_name

    def fetch_del_repos(self, repo_name):
        # 删除镜像仓库tag
        self.del_res = []
        tag_url = self.repos_url + "/" + repo_name + "/tags"
        # 项目镜像仓库的所有tags,按创建时间排序
        tags = self.client.session.get(tag_url).json()
        tags_sort = sorted(tags, key=lambda a: a["created"])
        # print(len(tags_sort),tags_sort)
        # 除了最新的self.num个，其他的tag都添加到待删除列表del_tags
        del_tags = tags_sort[0:len(tags_sort) - self.num]
        # print(del_tags)
        for tag in del_tags:
            del_repo_tag_url = tag_url + "/" + tag['name']
            # print(del_repo_tag_url)
            del_res = self.client.session.delete(del_repo_tag_url)
            self.del_res.append("镜像: %s 删除状态: %s" % (del_repo_tag_url,del_res))
        return self.del_res

    def work(self):
        # 遍历project id
        for i in self.fetch_pros_id():
            # 获取所有tag超过self.num的repos
            repos = self.fetch_del_repos_name(i)
            if repos:
                for repo in repos:
                    del_repos = self.fetch_del_repos(repo)
                    print(del_repos)


if __name__ == "__main__":
    clean = ClearHarbor(harbor_domain="harbor.test.com",
                        username="username",
                        password="password",
                        num=10)
    clean.work()