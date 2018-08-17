package main

import (
	"net/http"
	"os"
	"math/rand"
	"time"
	"io/ioutil"
	"fmt"
	"strings"
	"io"
	"bytes"
	"regexp"
	"bufio"
	"log"
)

var userAgent = []string{
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
	"Mozilla/5.0 (Windows NT 6.1; rv,2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0;",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; TencentTraveler 4.0)",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Avant Browser)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1)",
}

type crawlPic struct {
	userAgent     []string
	httpRequester *http.Request
	referer       string
	url           string
	slideShowPath string
	detailPath    string
}

func (c *crawlPic) getRandUserAgent() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return c.userAgent[r.Intn(len(c.userAgent))]
}

func (c *crawlPic) setResHeader(url string, ua string) {
	var err error
	c.httpRequester, err = http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	c.httpRequester.Header.Add("Cache-Control", "no-cache")
	c.httpRequester.Header.Add("User-Agent", ua)
	c.httpRequester.Header.Add("Referer", c.referer)
}

func (c *crawlPic) exec() []byte {
	res, err := http.DefaultClient.Do(c.httpRequester)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic(fmt.Sprintf("返回状态码异常：%s", res.Status))
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return content
}

func (c *crawlPic) disposeContent(content string) {
	if content == "" {
		fmt.Println("网页内容为空")
		return
	}

	//处理图片
	reg := regexp.MustCompilePOSIX(`"original":"(.*)"`)
	regContent := reg.FindAllStringSubmatch(content, -1)
	for _, val := range regContent {
		//保存图片
		if len(val) > 1 {
			c.savePic(c.slideShowPath, val[1])
		}
	}
	//处理详细内容的内容
	detailReg := regexp.MustCompile(`data-tfs-url="(.*?)"`)
	detailContent := detailReg.FindStringSubmatch(content)
	if len(detailContent) > 1 {
		c.disposeDetail(detailContent[1])
	}
}

func (c *crawlPic) disposeDetail(url string) {
	if url == "" {
		fmt.Println("详情url为空")
		return
	}

	ua := c.getRandUserAgent()
	//设置请求头
	c.setResHeader(url, ua)
	//开始请求
	content := c.exec()
	reg := regexp.MustCompile(`src=\\"(.*?)\\"`)
	imgs := reg.FindAllStringSubmatch(string(content), -1)
	for _, val := range imgs {
		//保存图片
		if len(val) > 1 {
			c.savePic(c.detailPath, val[1])
		}
	}
}

func (c *crawlPic) savePic(dir, pic string) {
	if pic == "" {
		fmt.Println("图片地址为空")
		return
	}
	path := strings.Split(pic, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	name = dir + time.Now().Format("20060102150405") + "_" + name
	out, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	ua := c.getRandUserAgent()
	//设置请求头
	c.setResHeader(pic, ua)
	//开始请求
	content := c.exec()
	_, err = io.Copy(out, bytes.NewReader(content))
	if err != nil {
		panic(err)
	}
	fmt.Printf("已抓取：%s\r\n", name)

}

func (c *crawlPic) initPath() {
	var err error
	dirName := time.Now().Format("20060102") + "/" + time.Now().Format("20060102150405")
	c.slideShowPath = dirName + "/轮播图/"
	err = os.MkdirAll(c.slideShowPath, 0755)
	if err != nil {
		panic(err)
	}
	c.detailPath = dirName + "/详情目录/"
	err = os.MkdirAll(c.detailPath, 0755)
	if err != nil {
		panic(err)
	}
}

func (c *crawlPic) run(url string) {
	if url == "" {
		panic("url不能为空")
	}
	c.url = url
	//处理话目录
	c.initPath()
	ua := c.getRandUserAgent()
	//设置请求头
	c.setResHeader(url, ua)
	//开始请求
	content := c.exec()
	//处理内容
	c.disposeContent(string(content))
	fmt.Println("抓取图片完毕...")
}

var logger *log.Logger

func start(url string){
	defer func() {
		err := recover()
		if err != nil{
			fmt.Println(err)
			logger.Println(err)
		}
	}()
	referer := "https://detail.1688.com"
	c := crawlPic{userAgent: userAgent, referer: referer}
	c.run(url)
	fmt.Println("请输入url后按回车")
}

func main() {
	fp, err := os.OpenFile("error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil{
		fmt.Println(err)
		return
	}
	defer fp.Close()
	logger = log.New(fp, "", log.Ldate|log.Ltime|log.Lshortfile)

	fmt.Println("请输入url后按回车")
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		url := strings.Trim(input.Text(), " \r\n")
		start(url)
	}

}
