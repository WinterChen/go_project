package base34
import (
	"testing"
	"fmt"
)

func TestInitBaseMap(t *testing.T){
	InitBaseMap()
	if len(baseMap) != 34 {
		t.Errorf("test fail. base: %s, base map: %v", baseStr, baseMap)
	}else{
		t.Logf("test succ. base: %s, base map: %v", baseStr, baseMap)
	}
	
}

func TestBase34(t *testing.T){
	InitBaseMap()
	var num uint64 = 0
	
	str  := Base34(num)
	fmt.Printf("===============base:%d->%s\n", num, string(str))
	n, err := Base34ToNum(str)
	if err != nil {
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码错误:%s",  num, string(str), err.Error())
	} else if (n != num){
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码解成:%d, 不相等！！！",  num, string(str), n)
	} else {
		t.Logf("数字:%d base34编码成%s，解码成功", num, str)
	}
	//200441052
    num = 200441052
	str  = Base34(num)
	fmt.Printf("===============base:%d->%s\n", num, string(str))
	n, err = Base34ToNum(str)
	if err != nil {
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码错误:%s",  num, string(str), err.Error())
	} else if (n != num){
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码解成:%d, 不相等！！！",  num, string(str), n)
	} else {
		t.Logf("数字:%d base34编码成%s，解码成功", num, str)
	}
	//200441052
    num = 1544804416
	str  = Base34(num)
	fmt.Printf("===============base:%d->%s\n", num, string(str))
	n, err = Base34ToNum(str)
	if err != nil {
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码错误:%s",  num, string(str), err.Error())
	} else if (n != num){
		t.Errorf("数字:%d base34编码成%s， 但是反过来解码解成:%d, 不相等！！！",  num, string(str), n)
	} else {
		t.Logf("数字:%d base34编码成%s，解码成功", num, str)
	}

}

func TestBase34ToNum(t *testing.T){

}