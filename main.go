package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"net/http"
	"net/url"

	"github.com/garyburd/redigo/redis"
)

const redisHost string = "127.0.0.1"
const redisPost string = "6379"
const redisDB string = "db0"

var rc redis.Conn

func main() {
	//Post accountNO to Create function, create a ShortURL by MD5.
	http.HandleFunc("/create", creatBarcode)

	//Get a accountNO by Get param=MD5 value
	http.HandleFunc("/get", getValue)
	http.ListenAndServe(":8080", nil)
	rc = initRedis()

	fmt.Println("xxxx")

}

type BarcodeType struct {
	accountNO string
	barcode   string
	codeType  int
}

//http method = GET, param= {val:[30]string}, return= {accountNO: [30] string}
//request as  http://127.0.0.1/get?val=xxxxxxxxx
func getValue(rw http.ResponseWriter, req *http.Request) {

	//GET method

	//	---------usual--------
	//	req.ParseForm()

	//	if len(req.Form["val"]) > 0 {
	//		fmt.Println("get val request for %s", req.Form["val"])
	//	}
	//	--------Better--------
	//Better method to GET value by URL in id
	//	<form action="http://localhost:9090/?id=1" method="POST">
	//    <input type="text" name="id" value="2" />
	//    <input type="submit" value="submit" />
	//	</form>
	queryForm, err := url.ParseQuery(req.URL.RawQuery)
	if err == nil && len(queryForm["val"]) > 0 {
		fmt.Fprintln(rw, queryForm["val"][0])
		fmt.Println("get request val as %s", queryForm["val"][0])
	}
}

//http method = POST, param= {accountNO: [30]string}, return= {val: [30]string}
func creatBarcode(rw http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	var accountNO string
	fmt.Println("http method is %s", req.Method)

	if req.Method == "POST" {
		accountNO = req.PostFormValue("accountNO")
		if len(accountNO) != 0 {
			redisSet(accountNO, getMD5(accountNO))

			fmt.Fprintln(rw, accountNO+" value: "+getMD5(accountNO))
		}

	}

	//	barcode := BarcodeType{}

}
func initRedis() (rs redis.Conn) {
	rs, err := redis.Dial("tcp", redisHost)
	defer rs.Close()

	//if connect failed
	if err != nil {
		fmt.Println(err)
		fmt.Println("redis connect Failed!!")
		return
	}
	rs.Do("SELECT", redisDB)
	return
}

//get MD5 value of input
func getMD5(in string) string {
	ret := ""
	if len(in) == 0 {
		return ret
	}
	h := md5.New()
	h.Write([]byte(in))
	fmt.Println("get MD5, value: %s, MD5: %s, cypto: %s", in, h.Sum(nil), hex.EncodeToString(h.Sum(nil)))

	ret = hex.EncodeToString(h.Sum(nil))
	return ret
}

//set a key:value into redis
func redisSet(key string, value string) (ret bool) {
	ret = false
	if rc == nil {
		fmt.Println("redis not connected")
		return
	}
	n, err := rc.Do("SETNX", key, value)
	if err != nil {
		fmt.Println(err)
		return
	}
	//set Expire time as 30min
	if n == int64(1) {
		n, _ := rc.Do("EXPIRE", key, 30*60)
		if n == int64(1) {
			fmt.Println("SUCCESS")
			ret = true
		}
	} else if n == int64(0) {
		fmt.Println("The key is exist")
		ret = false
	}
	return
}
